package efs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

func TestCostOfFileSystemMath(t *testing.T) {
	efs := EFSFileSystem{}
	efs.Size.StandardBytes = 0.1 * 1e9 // 0.1 GB
	perUnitCost := EFSCostPerUnit{
		Standard: types.CostPerUnit{
			Dollars: 10,
			PerUnit: types.GB,
		},
	}
	cost := efs.Cost(perUnitCost)
	assert.Equal(t, time.Duration(0), cost.Per)
	assert.Equal(t, 1, int(cost.Dollars))
}
