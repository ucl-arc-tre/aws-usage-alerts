package aws

import (
	ec2Client "github.com/ucl-arc-tre/aws-cost-alerts/internal/client/ec2"
	efsClient "github.com/ucl-arc-tre/aws-cost-alerts/internal/client/efs"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

type Controller struct {
	ec2 ec2Client.EC2Client
	efs efsClient.EFSClient
}

func New() *Controller {
	controller := Controller{
		ec2: ec2Client.New(),
		efs: efsClient.New(),
	}
	return &controller
}

func (c *Controller) Usage() types.AWSUsage {
	usage := types.AWSUsage{
		EFS: map[types.Group]types.Cost{},
		EC2: map[types.Group]types.Cost{},
	}

	return usage
}
