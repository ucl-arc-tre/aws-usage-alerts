package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeResourceUsageWithError() ResourceUsage {
	usage := ResourceUsage{
		Group("a"): Cost{Errors: []error{
			errors.New("err"),
		}},
	}
	return usage
}

func TestResourceUsageErrors(t *testing.T) {
	assert.Len(t, makeResourceUsageWithError().Errors(), 1)
}

func TestAWSUsageErrors(t *testing.T) {
	usage := AWSUsage{
		EC2: makeResourceUsageWithError(),
		EFS: ResourceUsage{},
	}
	assert.Len(t, usage.Errors(), 1)
}
