package testUtil

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"os"
)

// AWS Configuration

func MustConfig(configs ...external.Config) aws.Config {
	cfg, err := external.LoadDefaultAWSConfig(configs...)
	if err != nil {
		panic(err)
	}
	return cfg
}

func MustTestConfig() aws.Config {
	codeBuildId := os.Getenv("CODEBUILD_BUILD_ID")
	if codeBuildId == "" {
		testProfile := os.Getenv("CGC_TEST_AWS_PROFILE")
		if testProfile == "" {
			panic("the env var CGC_TEST_AWS_PROFILE is not defined")
		}
		return MustConfig(external.WithSharedConfigProfile(testProfile))
	}
	return MustConfig()
}
