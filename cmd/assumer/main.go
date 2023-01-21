package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/kelseyhightower/envconfig"
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

// STSGetSessionToken mock
type STSGetSessionToken interface {
	GetSessionToken(ctx context.Context, params *sts.GetSessionTokenInput, optFns ...func(*sts.Options)) (*sts.GetSessionTokenOutput, error)
}

func main() {
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
	if len(env.Home) == 0 {
		env.Home, err = os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
	}
	profile := getProfileEnv()
	if len(profile) > 0 {
		cfg, err := config.LoadDefaultConfig(
			context.TODO(),
			config.WithSharedConfigProfile(profile),
		)
		if err != nil {
			log.Fatal(err)
		}
		cred, err := cfg.Credentials.Retrieve(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
		setEnv(cfg, cred)
	}
	args := flag.Args()
	if len(args) <= 0 {
		envExportPrints(os.Stdout)
		return
	}
	cmd := exec.Command(args[0], args[1:]...) // nolint: gas
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	os.Exit(cmd.ProcessState.ExitCode())
}

// see: https://github.com/boto/botocore/blob/2f0fa46380a59d606a70d76636d6d001772d8444/botocore/session.py#L82
func getProfileEnv() (profile string) {
	if env.AWSDefaultProfile != "" {
		return env.AWSDefaultProfile
	}
	return env.AWSProfile
}

func envExportPrints(out io.Writer) {
	envExportPrint(out, "AWS_ACCESS_KEY_ID")
	envExportPrint(out, "AWS_SECRET_ACCESS_KEY")
	envExportPrint(out, "AWS_SESSION_TOKEN")
	envExportPrint(out, "AWS_REGION")
}

func envExportPrint(out io.Writer, env string) {
	value := os.Getenv(env)
	if len(value) > 0 {
		fmt.Fprintf(out, "export %s=\"%s\"\n", env, value)
	}
}

func setEnv(cfg aws.Config, cred aws.Credentials) {
	os.Unsetenv("AWS_PROFILE")                               // nolint errcheck
	os.Unsetenv("AWS_DEFAULT_PROFILE")                       // nolint errcheck
	os.Setenv("AWS_ACCESS_KEY_ID", cred.AccessKeyID)         // nolint errcheck
	os.Setenv("AWS_SECRET_ACCESS_KEY", cred.SecretAccessKey) // nolint errcheck
	os.Setenv("AWS_SESSION_TOKEN", cred.SessionToken)        // nolint errcheck
	if len(cfg.Region) > 0 {
		os.Setenv("AWS_REGION", cfg.Region) // nolint errcheck
	}
}
