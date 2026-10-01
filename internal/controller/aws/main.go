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
	filesystems, err := c.efs.FileSystems()
	if err != nil {
		return types.ResourceUsage{}, err
	}
	for _, fs := range filesystems {
		fsCost := fs.Cost(perUnitCost)
		groupUsage, exists := usage[fs.Group]
		if !exists {
			groupUsage = types.Cost{Per: fsCost.Per}
		}
		groupUsage.Add(fsCost)
		usage[fs.Group] = groupUsage
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
	instanceCost, err := c.ec2.InstanceCosts(instances)
	if err != nil {
		return types.ResourceUsage{}, err
	}
	usage := types.ResourceUsage{}
	errs := []error{}
	for _, instance := range instances {
		ec2Cost, err := instance.Cost(instanceCost)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		groupUsage, exists := usage[instance.Group]
		if !exists {
			groupUsage = types.Cost{Per: ec2Cost.Per}
		}
		groupUsage.Add(ec2Cost)
		usage[instance.Group] = groupUsage
	}
	log.Debug().Any("usage", usage).Msg("ec2")
	return usage, errors.Join(errs...)
}
