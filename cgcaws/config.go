package cgcaws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
)

func MustConfig(configs ...external.Config) aws.Config {
	cfg, err := external.LoadDefaultAWSConfig(configs...)
	if err != nil {
		panic(err)
	}
	return cfg
}
