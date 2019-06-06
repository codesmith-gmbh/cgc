package cgcaws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/codesmith-gmbh/cgc/testUtil"
	"testing"
)

const (
	TestStackName          = "CGCTestStack"
	SecurityGroupOutputKey = "SecurityGroup"
	IngressDescription     = "IngressTest"
)

func TestSecurityGroupService(t *testing.T) {
	ctx := context.TODO()
	cfg := testUtil.MustTestConfig()
	ec2Client := ec2.New(cfg)
	groupId, err := FetchStackOutputValue(ctx, cloudformation.New(cfg), TestStackName, SecurityGroupOutputKey)
	if err != nil {
		panic(err)
	}
	t.Logf("Group ID %s", groupId)
	sgs := NewIpifySecurityGroupService(ec2Client)
	err = sgs.OpenSecurityGroup(ctx, groupId, IngressDescription)
	if err != nil {
		t.Fatal(err)
	}
	//noinspection GoUnhandledErrorResult
	defer sgs.EnsureDescribedIngressRevoked(ctx, groupId, IngressDescription)
}
