package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
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
