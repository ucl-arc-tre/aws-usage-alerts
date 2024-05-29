package pricing

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	awsPricing "github.com/aws/aws-sdk-go-v2/service/pricing"
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

func (c *Client) PriceListJSON(serviceCode string, region string) ([]byte, error) {
	arn, err := c.priceListARN(serviceCode, c.aws.Options().Region)
	if err != nil {
		return []byte{}, err
	}
	url, err := c.priceListFileURL(arn)
	if err != nil {
		return []byte{}, err
	}
	response, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	return body, err
}

func (c *Client) priceListARN(serviceCode string, region string) (string, error) {
	if c.aws == nil {
		return "", errors.New("no aws client")
	}
	now := time.Now()
	output, err := c.aws.ListPriceLists(
		context.Background(),
		&awsPricing.ListPriceListsInput{
			CurrencyCode:  &currencyCode,
			ServiceCode:   &serviceCode,
			EffectiveDate: &now,
			RegionCode:    &region,
		},
	)
	if err != nil {
		return "", err
	} else if len(output.PriceLists) != 1 {
		return "", errors.New("found more than one matching price list")
	}
	priceListArn := output.PriceLists[0].PriceListArn
	if priceListArn == nil {
		return "", errors.New("price list ARN for was nil")
	} else {
		return *priceListArn, nil
	}
}

func (c *Client) priceListFileURL(priceListARN string) (string, error) {
	if c.aws == nil {
		return "", errors.New("no aws client")
	}
	output, err := c.aws.GetPriceListFileUrl(
		context.Background(),
		&awsPricing.GetPriceListFileUrlInput{
			FileFormat:   &fileFormat,
			PriceListArn: &priceListARN,
		},
	)
	if err != nil {
		return "", err
	} else if output.Url == nil {
		return "", errors.New("output url undefined")
	} else {
		return *output.Url, nil
	}
}
