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
		EC2: c.ec2Usage(),
	}
	return usage
}

func (c *Controller) efsUsage() types.ResourceUsage {
	perUnitCost, err := c.efs.CostPerUnit()
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
	log.Trace().Any("usage", usage).Msg("efs")
	return usage
}

func (c *Controller) ec2Usage() types.ResourceUsage {
	instances, err := c.ec2.Instances()
	if err != nil {
		log.Err(err).Msg("Failed to get EC2 instances. Skipping EC2 usage")
		return types.ResourceUsage{}
	}
	log.Debug().Int("number", len(instances)).Msg("Found running ec2 instances to group")
	instancePricing, err := c.ec2.InstanceCosts(instances)
	if err != nil {
		log.Err(err).Msg("Failed to get EC2 instance pricing. Skipping EC2 usage")
		return types.ResourceUsage{}
	}
	usage := types.ResourceUsage{}
	for _, instance := range instances {
		ec2Cost, err := instance.Cost(instancePricing)
		if err != nil {
			log.Err(err).Msg("Failed to get costs for instance")
			continue
		}
		if groupUsage, ok := usage[instance.Group]; ok {
			groupUsage.Add(ec2Cost)
		} else {
			usage[instance.Group] = ec2Cost
		}
	}
	log.Debug().Any("usage", usage).Msg("ec2")
	return usage
}
