package aws

import "github.com/ucl-arc-tre/aws-cost-alerts/internal/types"

type Controller struct {
}

func New() *Controller {
	controller := Controller{}
	return &controller
}

func (c *Controller) Usage() types.AWSUsage {
	usage := types.AWSUsage{}

	return usage
}
