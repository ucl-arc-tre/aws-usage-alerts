package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddCost(t *testing.T) {
	a := Cost{Dollars: USD(0.5)}
	a.Add(Cost{Dollars: USD(0.5)})
	assert.EqualValues(t, Cost{Dollars: USD(1)}, a)
}

func TestTotalAccumulatedCost(t *testing.T) {
	assert.EqualValues(t,
		AccumulatedCost{Dollars: USD(1)}.Dollars,
		AWSAccumulatedCost{
			EFS: AccumulatedCost{Dollars: USD(0.5)},
			EC2: AccumulatedCost{Dollars: USD(0.5)},
		}.Total().Dollars,
	)
}
