package efs

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	awsEFS "github.com/aws/aws-sdk-go-v2/service/efs"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/client/pricing"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

type MockAWSClient struct {
}

func (c *MockAWSClient) DescribeFileSystems(ctx context.Context, params *awsEFS.DescribeFileSystemsInput, optFns ...func(*awsEFS.Options)) (*awsEFS.DescribeFileSystemsOutput, error) {
	return &awsEFS.DescribeFileSystemsOutput{}, nil
}

type MockPricingClient struct {
	data []byte
}

func NewMockPricingClient(filepath string) *MockPricingClient {
	data, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	return &MockPricingClient{data: data}
}

func (c *MockPricingClient) PriceLists(serviceCode string, filters pricing.ProductFilters) ([]pricing.ProductPriceList, error) {
	var priceLists []pricing.ProductPriceList
	if err := json.Unmarshal(c.data, &priceLists); err != nil {
		panic(err)
	}
	return priceLists, nil
}

func NewMockClient() *Client {
	return &Client{
		aws:     &MockAWSClient{},
		pricing: NewMockPricingClient("testdata/price_list.json"),
	}
}

func TestPriceListParse(t *testing.T) {
	client := NewMockClient()
	pl, err := client.priceList()
	assert.NoError(t, err)
	cost, err := pl.Standard.CostPerUnit()
	assert.NoError(t, err)
	assert.NotZero(t, cost.Dollars)
	_, err = pl.IA.CostPerUnit()
	assert.Error(t, err) // not present
}

func TestCurrentCost(t *testing.T) {
	t.Setenv("AWS_REGION", "eu-west-2")
	client := NewMockClient()
	cost, err := client.CostPerUnit()
	assert.Error(t, err) // ther is no IA + Archive sku defined
	assert.Equal(t, cost.Standard.Dollars, types.USD(0.33))
	assert.Zero(t, cost.Archive.Dollars)
	assert.Zero(t, cost.IA)
	assert.Equal(t, cost.Standard.PerTime.Hours(), 730.5)
	assert.Equal(t, cost.Standard.PerUnit, types.GB)
}

func TestTagValue(t *testing.T) {
	tagKey := "a"
	value := "b"
	fsDesc := awsTypes.FileSystemDescription{
		Tags: []awsTypes.Tag{{
			Key:   &tagKey,
			Value: &value,
		}},
	}
	actualVal, exists := tagValue(fsDesc, tagKey)
	assert.True(t, exists)
	assert.Equal(t, value, actualVal)
	_, existsOther := tagValue(fsDesc, "non-existant-key")
	assert.False(t, existsOther)
}

func TestCurrentCostAccumulatesErrors(t *testing.T) {
	client := Client{
		aws:     &MockAWSClient{},
		pricing: NewMockPricingClient("testdata/price_list_empty.json"),
	}
	cost, err := client.CostPerUnit()
	assert.Error(t, err)
	assert.True(t, len(cost.Errors) > 1)
}
