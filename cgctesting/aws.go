package cgctesting

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/codesmith-gmbh/cgc/cgcaws"
	"os"
)

// AWS Configuration

func MustTestConfig() aws.Config {
	codeBuildId := os.Getenv("CODEBUILD_BUILD_ID")
	if codeBuildId == "" {
		testProfile := os.Getenv("CGC_TEST_AWS_PROFILE")
		if testProfile == "" {
			panic("the env var CGC_TEST_AWS_PROFILE is not defined")
		}
		return cgcaws.MustConfig(external.WithSharedConfigProfile(testProfile))
	}
	return cgcaws.MustConfig()
}
