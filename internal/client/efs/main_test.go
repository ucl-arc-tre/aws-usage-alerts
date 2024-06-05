package efs

import (
	"context"
	"os"
	"testing"

	awsEFS "github.com/aws/aws-sdk-go-v2/service/efs"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

type MockAWSClient struct {
}

func (c *MockAWSClient) DescribeFileSystems(ctx context.Context, params *awsEFS.DescribeFileSystemsInput, optFns ...func(*awsEFS.Options)) (*awsEFS.DescribeFileSystemsOutput, error) {
	return &awsEFS.DescribeFileSystemsOutput{}, nil
}

type MockPricingClient struct {
	priceListContent []byte
}

func NewMockPricingClient(filepath string) *MockPricingClient {
	data, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	return &MockPricingClient{priceListContent: data}
}

func (c *MockPricingClient) PriceListJSON(serviceCode string, region string) ([]byte, error) {
	return c.priceListContent, nil
}

func NewMockClient() *Client {
	return &Client{
		aws:     &MockAWSClient{},
		pricing: NewMockPricingClient("testdata/price_list.json"),
	}
}

func TestPriceListParse(t *testing.T) {
	client := NewMockClient()
	priceList, err := client.priceList()
	assert.NoError(t, err)
	assert.Len(t, priceList.Products, 1)
	assert.Len(t, priceList.Terms, 1)

	price := priceList.Terms["OnDemand"]["9D8Y226KZNKRMDQK"]["9D8Y226KZNKRMDQK.JRTCKXETXF"].PriceDimensions["9D8Y226KZNKRMDQK.JRTCKXETXF.6YS6EN2CT7"].PricePerUnit["USD"]
	assert.Equal(t, "0.3300000000", price)
}

func TestCurrentCost(t *testing.T) {
	t.Setenv("AWS_REGION", "eu-west-2")
	client := NewMockClient()
	cost, err := client.CurrentCostPerUnit()
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
	cost, err := client.CurrentCostPerUnit()
	assert.Error(t, err)
	assert.True(t, len(cost.Errors) > 1)
}
