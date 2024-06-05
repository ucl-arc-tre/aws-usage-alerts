package efs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueOrZero(t *testing.T) {
	assert.Equal(t, float64(0), valueOrZero(nil))
	value := int64(1)
	assert.Equal(t, float64(1), valueOrZero(&value))
}
