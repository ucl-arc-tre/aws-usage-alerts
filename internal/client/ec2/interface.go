package ec2

import (
	"context"

	awsEC2 "github.com/aws/aws-sdk-go-v2/service/ec2"
)

type Interface interface {
	RunningInstances() ([]Instance, error)
	InstanceCosts([]Instance) (InstanceCosts, error)
}

type awsClientInterface interface {
	DescribeInstances(ctx context.Context, params *awsEC2.DescribeInstancesInput, optFns ...func(*awsEC2.Options)) (*awsEC2.DescribeInstancesOutput, error)
}
