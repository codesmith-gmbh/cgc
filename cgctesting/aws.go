package cgctesting

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"log"
	"os"
	"strconv"
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

func MustEnvString(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("env var %s not defined\n", key)
	}
	return val
}

func MustEnvInt(key string) int {
	val, err := strconv.Atoi(MustEnvString(key))
	if err != nil {
		log.Fatal(err, fmt.Sprintf("env var %s undefined or not int", key))
	}
	return val
}
