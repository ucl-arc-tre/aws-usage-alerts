package aws

import (
	"errors"

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

func (c *Controller) Usage() (types.AWSUsage, error) {
	log.Debug().Msg("Getting AWS usage information")
	usage := types.AWSUsage{}
	if efs, err := c.efsUsage(); err != nil {
		return usage, err
	} else {
		usage.EFS = efs
	}
	if ec2, err := c.ec2Usage(); err != nil {
		return usage, err
	} else {
		usage.EC2 = ec2
	}
	return usage, nil
}

func (c *Controller) efsUsage() (types.ResourceUsage, error) {
	perUnitCost, err := c.efs.CostPerUnit()
	if err != nil {
		log.Err(err).Msg("Failed to get the current cost. Skipping EFS usage")
		return types.ResourceUsage{}, err
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
	return usage, nil
}

func (c *Controller) ec2Usage() (types.ResourceUsage, error) {
	instances, err := c.ec2.RunningInstances()
	if err != nil {
		return types.ResourceUsage{}, err
	}
	log.Debug().Int("number", len(instances)).Msg("Found running ec2 instances to group")
	instancePricing, err := c.ec2.InstanceCosts(instances)
	if err != nil {
		return types.ResourceUsage{}, err
	}
	usage := types.ResourceUsage{}
	errs := []error{}
	for _, instance := range instances {
		ec2Cost, err := instance.Cost(instancePricing)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if groupUsage, ok := usage[instance.Group]; ok {
			groupUsage.Add(ec2Cost)
		} else {
			usage[instance.Group] = ec2Cost
		}
	}
	log.Debug().Any("usage", usage).Msg("ec2")
	return usage, errors.Join(errs...)
}
