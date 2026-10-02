package efs

import (
	"context"

	awsEFS "github.com/aws/aws-sdk-go-v2/service/efs"
)

type Interface interface {
	FileSystems() ([]EFSFileSystem, error)
	CostPerUnit() (EFSCostPerUnit, error)
}

type awsClientInterface interface {
	DescribeFileSystems(ctx context.Context, params *awsEFS.DescribeFileSystemsInput, optFns ...func(*awsEFS.Options)) (*awsEFS.DescribeFileSystemsOutput, error)
}
