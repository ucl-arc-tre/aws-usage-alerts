package ec2

import (
	"errors"

	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

type InstanceType string

type Instance struct {
	Type  InstanceType // e.g. t3.small
	Group types.Group
}

func (i *Instance) Cost(instanceCosts InstanceCosts) (types.Cost, error) {
	instanceTypeCost, exists := instanceCosts[i.Type]
	if exists {
		return instanceTypeCost.Cost, nil
	} else {
		return types.Cost{}, errors.New("instanceCosts did not include this type")
	}
}

type InstanceCost struct {
	types.Cost
}

type InstanceCosts map[InstanceType]InstanceCost
