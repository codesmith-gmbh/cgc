package cgcaws

import (
	"context"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/pkg/errors"
)

const (
	errorPhysicalId = "🌪️"
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

type EventProcessor interface {
	ProcessEvent(context.Context, cfn.Event) (string, map[string]interface{}, error)
}

type ConstantErrorEventProcessor struct {
	Error error
}

func (p *ConstantErrorEventProcessor) ProcessEvent(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
	return event.PhysicalResourceID, nil, p.Error
}

func WrapForErrorPhysicalId(proc func(context.Context, cfn.Event) (string, map[string]interface{}, error)) func(context.Context, cfn.Event) (string, map[string]interface{}, error) {
	return func(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
		if event.PhysicalResourceID == errorPhysicalId && event.RequestType == cfn.RequestDelete {
			return errorPhysicalId, nil, nil
		}
		physicalId, data, err := proc(ctx, event)
		if err != nil || physicalId == "" {
			if physicalId == "" {
				physicalId = errorPhysicalId
			}
			if err == nil {
				err = errors.New("Physical ID not valid")
			}
		}
		return physicalId, data, err
	}
}
