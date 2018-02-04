package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

const (
	testHomeA = "./test/a"
	testHomeB = "./test/b"
	awsCred   = ".aws/credentials"
	awsConf   = ".aws/config"
)

func TestGetProfileEnv(t *testing.T) {
	var vtests = []struct {
		defValue  string
		profValue string
		expected  string
	}{
		{"def", "prof", "def"},
		{"", "prof", "prof"},
	}
	for _, vt := range vtests {
		env.AWSDefaultProfile = vt.defValue
		env.AWSProfile = vt.profValue
		r := getProfileEnv()
		if r != vt.expected {
			t.Errorf("AWSDefaultProfile=%q,AWSProfile=%q,getProfileEnv() = %q, want %q", vt.defValue, vt.profValue, r, vt.expected)
		}
	}
}
func TestSetEnv(t *testing.T) {
	prof := profileConfig{
		Region: "region",
	}
	role := &sts.AssumeRoleOutput{
		Credentials: &sts.Credentials{
			AccessKeyId:     aws.String("id"),
			SecretAccessKey: aws.String("key"),
			SessionToken:    aws.String("token"),
		},
	}
	setEnv(prof, role)
	if os.Getenv("AWS_PROFILE") != "" {
		t.Error("AWS_PROFILE")
	}
	if os.Getenv("AWS_DEFAULT_PROFILE") != "" {
		t.Error("AWS_DEFAULT_PROFILE")
	}
	if os.Getenv("AWS_ACCESS_KEY_ID") != "id" {
		t.Error("id")
	}
	if os.Getenv("AWS_SECRET_ACCESS_KEY") != "key" {
		t.Error("key")
	}
	if os.Getenv("AWS_SESSION_TOKEN") != "token" {
		t.Error("token")
	}
	if os.Getenv("AWS_REGION") != "region" {
		t.Error("region")
	}
}

func TestEnvExportPrints(t *testing.T) {
	var vtests = []struct {
		setEnvs   []string
		unsetEnvs []string
		expected  []string
	}{
		{
			[]string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_REGION"},
			[]string{},
			[]string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_REGION"},
		}, {
			[]string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN"},
			[]string{"AWS_REGION"},
			[]string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN"},
		},
	}
	for _, vt := range vtests {
		for _, env := range vt.setEnvs {
			os.Setenv(env, env)
		}
		for _, env := range vt.unsetEnvs {
			os.Unsetenv(env)
		}
		var b bytes.Buffer
		envExportPrints(&b)
		expected := ""
		for _, env := range vt.expected {
			expected += fmt.Sprintf("export %s=\"%s\"\n", env, env)
		}
		if b.String() != expected {
			t.Errorf("setEnvs:%q, unsetEnvs:%q, envExportPrint() = %q, want %q", vt.setEnvs, vt.unsetEnvs, b.String(), expected)
		}
	}
}
func TestAwsFilePath(t *testing.T) {
	var vtests = []struct {
		envValue         string
		defaultPathParam string
		expected         string
	}{
		{
			envValue:         filepath.Join("~", awsCred),
			defaultPathParam: awsCred,
			expected:         filepath.Join(testHomeA, awsCred),
		}, {
			envValue:         filepath.Join("~", awsConf),
			defaultPathParam: awsConf,
			expected:         filepath.Join(testHomeA, awsConf),
		}, {
			envValue:         "",
			defaultPathParam: ".aws/credentials",
			expected:         filepath.Join(testHomeA, awsCred),
		}, {
			envValue:         "",
			defaultPathParam: awsConf,
			expected:         filepath.Join(testHomeA, awsConf),
		},
	}

	env.Home = testHomeA
	for _, vt := range vtests {
		r := awsFilePath(vt.envValue, vt.defaultPathParam, testHomeA)
		if r != vt.expected {
			t.Errorf("awsFilePath(%q, %q) = %q, want %q", vt.envValue, vt.defaultPathParam, r, vt.expected)
		}
	}
}

type mockedSts struct {
	stsiface.STSAPI
	Resp sts.AssumeRoleOutput
}

func (m mockedSts) AssumeRole(in *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	// Only need to return mocked response output
	return &m.Resp, nil
}

func TestAssumeRole(t *testing.T) {
	cases := []struct {
		Resp        sts.AssumeRoleOutput
		expectedKey string
	}{
		{
			Resp: sts.AssumeRoleOutput{
				AssumedRoleUser: &sts.AssumedRoleUser{
					Arn:           aws.String("arn:..."),
					AssumedRoleId: aws.String("xxxx"),
				},
				Credentials: &sts.Credentials{
					AccessKeyId:     aws.String("id"),
					Expiration:      &time.Time{},
					SecretAccessKey: aws.String("key"),
					SessionToken:    aws.String("token"),
				},
			},
			expectedKey: "id",
		},
	}

	for i, c := range cases {
		mock := mockedSts{Resp: c.Resp}
		res, err := assumeRole(&mock, "")
		if err != nil {
			t.Fatalf("%d, unexpected error:%s", i, err)
		}
		if c.expectedKey != *res.Credentials.AccessKeyId {
			t.Fatalf("%d, expected %q messages, got %q", i, c.expectedKey, res.Credentials.AccessKeyId)
		}
	}
}

func TestGetProfileConfig(t *testing.T) {
	var vtests = []struct {
		home     string
		profile  string
		err      *string
		expected profileConfig
	}{
		{
			testHomeA,
			"testprof",
			nil,
			profileConfig{
				RoleARN:    "arn:aws:iam::123456789012:role/Admin",
				Region:     "ap-northeast-1",
				SrcProfile: "srcprof",
				//SrcRegion:    "us-east-1",
				//SrcAccountID: "000000000000",
			},
		},
		{
			testHomeB,
			"not_profile_prefix",
			nil,
			profileConfig{
				RoleARN:    "arn:aws:iam::123456789011:role/a",
				Region:     "ap-northeast-1",
				SrcProfile: "srcprof",
				//SrcRegion:    "us-east-1",
				//SrcAccountID: "000000000000",
			},
		},
		{
			testHomeB,
			"src_default",
			nil,
			profileConfig{
				RoleARN:    "arn:aws:iam::123456789011:role/b",
				Region:     "ap-northeast-1",
				SrcProfile: "default",
				//SrcRegion:    "us-east-1",
				//SrcAccountID: "000000000000",
			},
		},
		{
			testHomeB,
			"none",
			aws.String("not found ini section err:section 'profile none' does not exist"),
			profileConfig{
				RoleARN:    "",
				Region:     "",
				SrcProfile: "",
				//SrcRegion:    "us-east-1",
				//SrcAccountID: "000000000000",
			},
		},
	}
	for _, vt := range vtests {
		env.Home = vt.home
		res, err := getProfileConfig(vt.profile)
		if err != nil && vt.err == nil {
			t.Errorf("err getProfileConfig(%q) = err:%s", vt.profile, err)
		}
		if err != nil {
			if err.Error() != *vt.err {
				t.Errorf("err getProfileConfig(%q) = err:%s", vt.profile, err)
			}
		}
		if res != vt.expected {
			t.Errorf("getProfileConfig(%q); = %q, want %q", vt.profile, res, vt.expected)
		}
	}
}
func TestGetExitCode(t *testing.T) {
	var vtests = []struct {
		cmd      []string
		expected int
	}{
		{[]string{"ls", "-abcefghijk"}, 2},
		{[]string{"ls", "-la"}, 0},
	}
	for _, vt := range vtests {
		cmd := exec.Command(vt.cmd[0], vt.cmd[1:]...) // nolint: gas
		res := getExitCode(cmd.Run())
		if res != vt.expected {
			t.Errorf("getExitCode(cmd:%q); = %q, want %q", vt.cmd, res, vt.expected)
		}
	}
}
func TestCreateCacheKey(t *testing.T) {
	var vtests = []struct {
		profile  string
		roleARN  string
		expected string
	}{
		{"a", "b", "da23614e02469a0d7c7bd1bdab5c9c474b1904dc"},
		{"testprof", "arn:aws:iam::123456789012:role/Admin", "44b61a3deae05f4e4bb01555bc3966dcae87f121"},
	}
	for _, vt := range vtests {
		res := createCacheKey(vt.profile, vt.roleARN)
		if res != vt.expected {
			t.Errorf("createCacheKey(%q, %q); = %q, want %q", vt.profile, vt.roleARN, res, vt.expected)
		}
	}
}
func TestLoadStoreCache(t *testing.T) {
	var vtests = []struct {
		res    sts.AssumeRoleOutput
		pc     profileConfig
		expire bool
	}{
		{
			sts.AssumeRoleOutput{
				AssumedRoleUser: &sts.AssumedRoleUser{
					Arn:           aws.String("arn:aws:sts::000000000000:assumed-role/AdminX/test"),
					AssumedRoleId: aws.String("XXXXXXXXXXXXXXXXXXXXX:test"),
				},
				Credentials: &sts.Credentials{
					AccessKeyId:     aws.String("XXXXXXXXXXXXXXXXXXXX"),
					Expiration:      aws.Time(time.Now().Add(3600 * time.Second)),
					SecretAccessKey: aws.String("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"),
					SessionToken:    aws.String("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"),
				},
			},
			profileConfig{
				RoleARN:    "arn:aws:iam::123456789012:role/Admin",
				SrcProfile: "srcprof",
			},
			false,
		},
		{
			sts.AssumeRoleOutput{
				Credentials: &sts.Credentials{
					Expiration: aws.Time(time.Now().Add(-3600 * time.Second)),
				},
			},
			profileConfig{
				RoleARN:    "arn:aws:iam::123456789012:role/Admin",
				SrcProfile: "srcprof",
			},
			true,
		},
	}
	env.Home = testHomeA
	for _, vt := range vtests {
		err := storeCache(vt.pc, &vt.res)
		if err != nil {
			t.Error(err)
		}
		res, err := loadCache(vt.pc)
		if err != nil {
			if vt.expire {
				continue
			}
			t.Fatal(err)
		}
		if *res.Credentials.AccessKeyId != *vt.res.Credentials.AccessKeyId {
			t.Errorf("load(%q); = %q, want %q", vt.pc, res, vt.res)
		}
		if *res.Credentials.SecretAccessKey != *vt.res.Credentials.SecretAccessKey {
			t.Errorf("load(%q); = %q, want %q", vt.pc, res, vt.res)
		}
		if *res.Credentials.SessionToken != *vt.res.Credentials.SessionToken {
			t.Errorf("load(%q); = %q, want %q", vt.pc, res, vt.res)
		}
	}
	os.RemoveAll(filepath.Join(testHomeA, cacheDir))
	os.RemoveAll(filepath.Join(testHomeB, cacheDir))
}
