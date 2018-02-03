package main

import (
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/go-ini/ini"
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/go-homedir"
)

const (
	credPath = ".aws/credentials"
	confPath = ".aws/config"

	iniRoleARN    = "role_arn"
	iniSrcProfile = "source_profile"
	iniRegion     = "region"
	cacheDir      = ".assumer/cache/"
)

type environments struct {
	AWSSharedCredentialsFile string `envconfig:"AWS_SHARED_CREDENTIALS_FILE"`
	AWSConfigFile            string `envconfig:"AWS_CONFIG_FILE"`
	AWSDefaultProfile        string `envconfig:"AWS_DEFAULT_PROFILE"`
	AWSProfile               string `envconfig:"AWS_PROFILE"`
	MaxExpiration            int64  `envconfig:"MAX_DURATION" default:"3600"`
	MinExpiration            int64  `envconfig:"MIN_EXPIRATION" default:"600"`
	Home                     string `envconfig:"HOME"`
}

type profileConfig struct {
	RoleARN    string
	SrcProfile string
	Region     string
	//SrcRegion    string
	//SrcAccountID string
}

type cache struct {
	Version string
	Role    sts.AssumeRoleOutput
}

var (
	env     environments
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	showVersion := false
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()
	if showVersion {
		fmt.Printf("%s version %v, commit %v, built at %v\n", filepath.Base(os.Args[0]), version, commit, date)
		os.Exit(0)
	}

	err := envconfig.Process("", &env)
	if err != nil {
		log.Fatal(err)
	}
	if env.MaxExpiration > 3600 {
		log.Fatal("Member must have MAX_DURATION less than or equal to 3600")
	}
	if len(env.Home) == 0 {
		env.Home, err = homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	conf, err := getProfileConfig(getProfileEnv())
	if err != nil {
		log.Fatal(err)
	}
	res, err := getCred(conf)
	if err != nil {
		log.Fatal(err)
	}
	setEnv(conf, res)
	args := flag.Args()
	if len(args) <= 1 {
		envExportPrints()
		return
	}
	cmd := exec.Command(args[1], args[2:]...) // nolint: gas
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	code := getExitCode(cmd.Run())
	os.Exit(code)
}

// see: https://github.com/boto/botocore/blob/2f0fa46380a59d606a70d76636d6d001772d8444/botocore/session.py#L82
func getProfileEnv() (profile string) {
	if env.AWSDefaultProfile != "" {
		return env.AWSDefaultProfile
	}
	return env.AWSProfile
}

func envExportPrints() {
	envExportPrint("AWS_ACCESS_KEY_ID")
	envExportPrint("AWS_SECRET_ACCESS_KEY")
	envExportPrint("AWS_SESSION_TOKEN")
	envExportPrint("AWS_REGION")
}

func envExportPrint(env string) {
	value := os.Getenv(env)
	if len(value) > 0 {
		fmt.Printf("export %s=\"%s\"\n", env, value)
	}
}

func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		log.Fatal(err)
	}
	s, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		log.Fatal(err)
	}
	return s.ExitStatus()
}

func setEnv(conf profileConfig, assumeRole *sts.AssumeRoleOutput) {
	os.Unsetenv("AWS_PROFILE")                                                  // nolint errcheck
	os.Unsetenv("AWS_DEFAULT_PROFILE")                                          // nolint errcheck
	os.Setenv("AWS_ACCESS_KEY_ID", *assumeRole.Credentials.AccessKeyId)         // nolint errcheck
	os.Setenv("AWS_SECRET_ACCESS_KEY", *assumeRole.Credentials.SecretAccessKey) // nolint errcheck
	os.Setenv("AWS_SESSION_TOKEN", *assumeRole.Credentials.SessionToken)        // nolint errcheck
	if len(conf.Region) > 0 {
		os.Setenv("AWS_REGION", conf.Region) // nolint errcheck
	}
}

func loadCache(conf profileConfig) (*sts.AssumeRoleOutput, error) {
	file := filepath.Join(env.Home, cacheDir, createCacheKey(conf.SrcProfile, conf.RoleARN))
	cf, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer cf.Close() // nolint errcheck
	var c cache
	if err = json.NewDecoder(cf).Decode(&c); err != nil {
		return nil, err
	}
	if c.Version != version {
		defer os.Remove(file) // nolint errcheck
		return nil, fmt.Errorf("Cache version:%s is different from current:%s", c.Version, version)
	}
	if c.Role.Credentials == nil || c.Role.Credentials.Expiration == nil {
		defer os.Remove(file) // nolint errcheck
		return nil, fmt.Errorf("Illegal cache file:%s", file)
	}
	if time.Since(*c.Role.Credentials.Expiration) > 0 {
		defer os.Remove(file) // nolint errcheck
		return nil, fmt.Errorf("Credential expired. Expiration:%s", time.Since(*c.Role.Credentials.Expiration))
	}
	return &c.Role, err

}
func storeCache(conf profileConfig, res *sts.AssumeRoleOutput) error {
	c := cache{
		Version: version,
		Role:    *res,
	}
	os.MkdirAll(filepath.Join(env.Home, cacheDir), 0700) // nolint errcheck
	cf, err := os.Create(filepath.Join(env.Home, cacheDir, createCacheKey(conf.SrcProfile, conf.RoleARN)))
	if err != nil {
		return err
	}
	defer cf.Close() // nolint errcheck
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	_, err = cf.Write(b)
	return err
}

func getCred(conf profileConfig) (*sts.AssumeRoleOutput, error) {
	res, err := loadCache(conf)
	if err != nil {
		stsSvc := sts.New(session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewSharedCredentials(awsFilePath(env.AWSSharedCredentialsFile, credPath, env.Home), conf.SrcProfile),
		})))
		res, err = assumeRole(stsSvc, conf.RoleARN)
		storeCache(conf, res) // nolint errcheck
		return res, err
	}

	return res, err
}
func assumeRole(stsSvc stsiface.STSAPI, roleARN string) (*sts.AssumeRoleOutput, error) { // nolint interfacer
	input := &sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(env.MaxExpiration),
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String("test"),
	}
	return stsSvc.AssumeRole(input)
}

func awsFilePath(filePath, defaultPath, home string) string {
	if filePath != "" {
		if filePath[0] == '~' {
			return filepath.Join(home, filePath[1:])
		}
		return filePath
	}
	if home == "" {
		return ""
	}

	return filepath.Join(home, defaultPath)
}
func getProfileConfig(profile string) (res profileConfig, err error) {
	res, err = getProfile(profile, confPath)
	if err != nil {
		return res, err
	}
	if len(res.SrcProfile) > 0 && len(res.RoleARN) > 0 {
		return res, err
	}
	return getProfile(profile, credPath)
}

func getProfile(profile, confFileNmae string) (res profileConfig, err error) {
	cnfPath := awsFilePath(env.AWSConfigFile, confFileNmae, env.Home)
	config, err := ini.Load(cnfPath)
	if err != nil {
		return res, fmt.Errorf("failed to load shared credentials file. err:%s", err)
	}
	sec, err := config.GetSection(profile)
	if err != nil {
		// reference code -> https://github.com/aws/aws-sdk-go/blob/fae5afd566eae4a51e0ca0c38304af15618b8f57/aws/session/shared_config.go#L173-L181
		sec, err = config.GetSection(fmt.Sprintf("profile %s", profile))
		if err != nil {
			return res, fmt.Errorf("not found ini section err:%s", err)
		}
	}
	res.RoleARN = sec.Key(iniRoleARN).String()
	res.SrcProfile = sec.Key(iniSrcProfile).String()
	res.Region = sec.Key(iniRegion).String()
	return res, nil
}

func createCacheKey(roleARN, sessionName string) string {
	h := sha1.New()
	io.WriteString(h, roleARN)     // nolint errcheck
	io.WriteString(h, sessionName) // nolint errcheck
	return fmt.Sprintf("%x", h.Sum(nil))
}
