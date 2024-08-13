package ec2

import (
	"fmt"
	"testing"

	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/client/pricing"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
)

func TestGroupFromTaggedInstanceIsOK(t *testing.T) {
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

func TestGroupFromUnTaggedInstanceIsNotOK(t *testing.T) {
	viper.Set("groupTagKey", "group")
	_, exists := awsInstanceGroup(awsEC2Types.Instance{})
	assert.False(t, exists)
}

func TestMean(t *testing.T) {
	assert.Equal(t, 2.0, mean([]float64{1.0, 2.0, 3.0}))
}

func TestUSDFromPriceListsSimple(t *testing.T) {
	sku := "ABCD1234"
	expectedUSDValue := 4.2
	priceList := pricing.ProductPriceList{
		Product: pricing.ProductWithAttributes{
			Attributes: map[string]string{},
			SKU:        sku,
		},
		Terms: map[string]any{
			"OnDemand": map[string]any{
				"sku": map[string]any{
					"priceDimensions": map[string]any{
						sku: map[string]any{
							"unit": "Hrs",
							"pricePerUnit": map[string]any{
								"USD": fmt.Sprintf("%v", expectedUSDValue),
							},
						},
					},
				},
			},
		},
	}
	usd := usdPerHourForOnDemandInPriceLists([]pricing.ProductPriceList{priceList})
	assert.Equal(t, []float64{expectedUSDValue}, usd)
}
