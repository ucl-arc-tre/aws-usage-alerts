package email

import "github.com/ucl-arc-tre/aws-cost-alerts/internal/types"

type Controller struct {
}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) Send(state *types.StateV1alpha1) {

}
