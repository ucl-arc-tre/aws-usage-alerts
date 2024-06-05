package aws

import (
	"github.com/rs/zerolog/log"
	ec2Client "github.com/ucl-arc-tre/aws-cost-alerts/internal/client/ec2"
	efsClient "github.com/ucl-arc-tre/aws-cost-alerts/internal/client/efs"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

type Controller struct {
	ec2 ec2Client.Interface
	efs efsClient.Interface
}

func New() *Controller {
	return NewWithClients(ec2Client.New(), efsClient.New())
}

func NewWithClients(ec2 ec2Client.Interface, efs efsClient.Interface) *Controller {
	controller := Controller{
		ec2: ec2,
		efs: efs,
	}
	return &controller
}

func (c *Controller) Usage() types.AWSUsage {
	log.Debug().Msg("Getting AWS usage information")
	usage := types.AWSUsage{
		EFS: c.efsUsage(),
	}
	return usage
}

func (c *Controller) efsUsage() types.ResourceUsage {
	perUnitCost, err := c.efs.CurrentCostPerUnit()
	if err != nil {
		log.Err(err).Msg("Failed to get the current cost. Skipping EFS usage")
		return types.ResourceUsage{}
	}
	usage := types.ResourceUsage{}
	for _, fs := range c.efs.FileSystems() {
		fsCost := fs.Cost(perUnitCost)
		if groupUsage, ok := usage[fs.Group]; ok {
			groupUsage.Add(fsCost)
		} else {
			usage[fs.Group] = fsCost
		}
	}
	log.Debug().Any("usage", usage).Msg("")
	return usage
}
