package cgcaws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/pkg/errors"
)

func FetchStackOutputValue(ctx context.Context, client *cloudformation.Client, stackName string, outputKey string) (string, error) {
	stacks, err := client.DescribeStacksRequest(&cloudformation.DescribeStacksInput{
		StackName: &stackName,
	}).Send(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "could not describe the stack %s", stackName)
	}
	outputs := stacks.Stacks[0].Outputs
	for _, output := range outputs {
		if aws.StringValue(output.OutputKey) == outputKey {
			return aws.StringValue(output.OutputValue), nil
		}
	}
	return "", errors.Errorf("could not find the output key %s in the stack %s", outputKey, stackName)
}