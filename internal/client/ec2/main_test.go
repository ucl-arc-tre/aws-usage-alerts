package ec2

import (
	"testing"

	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
)

func TestInstanceGroupOK(t *testing.T) {
	key := "group"
	viper.Set("groupTagKey", key)
	value := "example"
	assert.Equal(t, key, config.GroupTagKey())
	instance := awsEC2Types.Instance{
		Tags: []awsEC2Types.Tag{{
			Key:   &key,
			Value: &value,
		}},
	}
	group, exists := awsInstanceGroup(instance)
	assert.True(t, exists)
	assert.Equal(t, value, string(group))
}

func TestInstanceGroupDoesNotExistWithNoTags(t *testing.T) {
	viper.Set("groupTagKey", "group")
	_, exists := awsInstanceGroup(awsEC2Types.Instance{})
	assert.False(t, exists)
}

func TestMean(t *testing.T) {
	assert.Equal(t, 2.0, mean([]float64{1.0, 2.0, 3.0}))
}
