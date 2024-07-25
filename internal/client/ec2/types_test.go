package ec2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

func TestInstanceCost(t *testing.T) {
	intanceType := InstanceType("a")
	instance := Instance{Type: intanceType}
	cost := types.Cost{Dollars: types.USD(0.1)}
	costs := InstanceCosts{
		intanceType: InstanceCost{Cost: cost},
	}
	instanceCost, err := instance.Cost(costs)
	assert.NoError(t, err)
	assert.Equal(t, cost.Dollars, instanceCost.Dollars)
}

func TestInstanceCostWithUnknownTypeErrors(t *testing.T) {
	intanceType := InstanceType("a")
	instance := Instance{Type: intanceType}
	_, err := instance.Cost(InstanceCosts{})
	assert.Error(t, err)
}
