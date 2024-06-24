package pricing

import (
	"context"
	"encoding/json"

	awsPricing "github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
)

var (
	currencyCode = "USD"
	region       = "us-east-1"
	fileFormat   = "json"
)

type Client struct {
	aws *awsPricing.Client
}

func New() *Client {
	awsConfig := config.AWS()
	awsConfig.Region = region // not all regions have a pricing API..
	return &Client{awsPricing.NewFromConfig(awsConfig)}
}

func (c *Client) PriceLists(serviceCode string, filters ProductFilters) ([]ProductPriceList, error) {
	filters.appendRegionFilter()
	output, err := c.aws.GetProducts(
		context.Background(),
		&awsPricing.GetProductsInput{
			ServiceCode: &serviceCode,
			Filters:     filters.ToAWS(),
		},
	)
	if err != nil {
		return []ProductPriceList{}, err
	}
	if output.NextToken != nil {
		log.Warn().Str("serviceCode", serviceCode).Msg("Had paginated response. Ignoring!")
	}
	priceLists := []ProductPriceList{}
	for _, jsonStr := range output.PriceList {
		var productPriceList ProductPriceList
		if err := json.Unmarshal([]byte(jsonStr), &productPriceList); err != nil {
			log.Err(err).Msg("Failed to unmarshal price list JSON for product")
			continue
		} else {
			priceLists = append(priceLists, productPriceList)
		}
	}
	return priceLists, nil
}
