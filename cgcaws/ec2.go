package cgcaws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	AwsUrl                    = "http://checkip.amazonaws.com"
	IpifyUrl                  = "https://api.ipify.org"
	DuplicateIngressErrorCode = "InvalidPermission.Duplicate"
)

type SGS struct {
	Ec2        *ec2.Client
	IpCheckUrl string
}

func NewSecurityGroupService(ec2 *ec2.Client, ipCheckUrl string) *SGS {
	return &SGS{
		Ec2:        ec2,
		IpCheckUrl: ipCheckUrl,
	}
}

func NewAwsSecurityGroupService(ec2 *ec2.Client) *SGS {
	return NewSecurityGroupService(ec2, AwsUrl)
}

func NewIpifySecurityGroupService(ec2 *ec2.Client) *SGS {
	return NewSecurityGroupService(ec2, IpifyUrl)
}

func (sgs *SGS) OpenSecurityGroup(ctx context.Context, groupId, description string) error {
	cidr, err := sgs.PublicIpAsCidr()
	if err != nil {
		return err
	}
	if err = sgs.EnsureDescribedIngressRevoked(ctx, groupId, description); err != nil {
		return err
	}
	return sgs.AuthorizeDescribedIngress(ctx, groupId, cidr, description)
}

func (sgs *SGS) PublicIpAsCidr() (string, error) {
	resp, err := http.Get(sgs.IpCheckUrl)
	if err != nil {
		return "", errors.Wrap(err, "could not get the ip of the server")
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()
	ipBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "could not read the body of the request to %s", sgs.IpCheckUrl)
	}
	ipAddress := strings.TrimSpace(string(ipBytes))
	cidr := ipAddress + "/32"
	return cidr, nil
}

func (sgs *SGS) EnsureDescribedIngressRevoked(ctx context.Context, groupId, description string) error {
	groups, err := sgs.Ec2.DescribeSecurityGroupsRequest(&ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{groupId},
	}).Send(ctx)
	if err != nil {
		return errors.Wrapf(err, "describe the security group %s", groupId)
	}
	group := groups.SecurityGroups[0]
	for _, perm := range group.IpPermissions {
		for _, ipRange := range perm.IpRanges {
			if ipRange.Description != nil && *ipRange.Description == description {
				// We found an existing Ingress Rule with the correct description -> we can delete the rule
				_, err = sgs.Ec2.RevokeSecurityGroupIngressRequest(&ec2.RevokeSecurityGroupIngressInput{
					GroupId: &groupId,
					IpPermissions: []ec2.IpPermission{
						{
							FromPort:   perm.FromPort,
							ToPort:     perm.ToPort,
							IpProtocol: perm.IpProtocol,
							IpRanges:   []ec2.IpRange{ipRange},
						},
					},
				}).Send(ctx)
				if err != nil {
					return errors.Wrap(err, "could not revoke the ingress rule")
				}
			}
		}
	}
	return nil
}

func (sgs *SGS) AuthorizeDescribedIngress(ctx context.Context, groupId, cidr, description string) error {
	_, err := sgs.Ec2.AuthorizeSecurityGroupIngressRequest(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: &groupId,
		IpPermissions: []ec2.IpPermission{
			{
				FromPort:   aws.Int64(5432),
				ToPort:     aws.Int64(5432),
				IpProtocol: aws.String("tcp"),
				IpRanges: []ec2.IpRange{
					{
						CidrIp:      &cidr,
						Description: &description,
					},
				},
			},
		},
	}).Send(ctx)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == DuplicateIngressErrorCode {
			return nil
		}
		return errors.Wrapf(err, "could not authorize for the group %s, cidr %s and description %s", groupId, cidr, description)
	}
	return nil
}
