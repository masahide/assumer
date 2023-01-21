package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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
	output := aws.Credentials{
		AccessKeyID:     "id",
		SecretAccessKey: "key",
		SessionToken:    "token",
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
