package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
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
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("region"),
	)
	if err != nil {
		t.Fatal()
	}
	output := &sts.GetSessionTokenOutput{
		Credentials: &types.Credentials{
			AccessKeyId:     aws.String("id"),
			SecretAccessKey: aws.String("key"),
			SessionToken:    aws.String("token"),
		},
	}
	setEnv(cfg, output)
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
				AssumedRoleUser: &types.AssumedRoleUser{
					Arn:           aws.String("arn:aws:sts::000000000000:assumed-role/AdminX/test"),
					AssumedRoleId: aws.String("XXXXXXXXXXXXXXXXXXXXX:test"),
				},
				Credentials: &types.Credentials{
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
				Credentials: &types.Credentials{
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
			t.Errorf("load(%q); = %v want %v", vt.pc, res, vt.res)
		}
		if *res.Credentials.SecretAccessKey != *vt.res.Credentials.SecretAccessKey {
			t.Errorf("load(%q); = %v, want %v", vt.pc, res, vt.res)
		}
		if *res.Credentials.SessionToken != *vt.res.Credentials.SessionToken {
			t.Errorf("load(%q); = %v, want %v", vt.pc, res, vt.res)
		}
	}
	os.RemoveAll(filepath.Join(testHomeA, cacheDir))
	os.RemoveAll(filepath.Join(testHomeB, cacheDir))
}
