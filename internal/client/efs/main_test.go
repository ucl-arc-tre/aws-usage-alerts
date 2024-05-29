package efs

import (
	"context"
	"os"
	"testing"

	awsEFS "github.com/aws/aws-sdk-go-v2/service/efs"
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

func NewMockPricingClient() *MockPricingClient {
	data, err := os.ReadFile("testdata/price_list.json")
	if err != nil {
		panic("failed to read price list file")
	}
	return &MockPricingClient{priceListContent: data}
}

func (c *MockPricingClient) PriceListJSON(serviceCode string, region string) ([]byte, error) {
	return c.priceListContent, nil
}

func NewMockClient() *Client {
	return &Client{
		aws:     &MockAWSClient{},
		pricing: NewMockPricingClient(),
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
	client := NewMockClient()
	cost, err := client.CurrentCostPerUnit()
	assert.NoError(t, err)
	assert.Equal(t, cost.Standard.Dollars, types.USD(0.33))
	assert.Zero(t, cost.Archive.Dollars)
	assert.Zero(t, cost.IA)
	assert.Equal(t, cost.Standard.PerTime.Hours(), 730.5)
	assert.Equal(t, cost.Standard.PerUnit, types.GB)
}
