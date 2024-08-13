package ec2

import (
	"context"
	"strconv"
	"strings"
	"time"

	awsEC2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/rs/zerolog/log"
	pricingClient "github.com/ucl-arc-tre/aws-cost-alerts/internal/client/pricing"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

const (
	serviceCode = "AmazonEC2"
)

var (
	instanceStateNameKey      = "instance-state-name"
	runningInstanceStateValue = "running"
	maxResults                = int32(1000)
)

type Client struct {
	aws     awsClientInterface
	pricing pricingClient.Interface
}

func New() *Client {
	client := Client{
		aws:     awsEC2.NewFromConfig(config.AWS()),
		pricing: pricingClient.New(),
	}
	return &client
}

func (c *Client) RunningInstances() ([]Instance, error) {
	instances, err := c.accumulateRunningInstancesWithGroup([]Instance{}, nil)
	if err != nil {
		if len(instances) == 0 {
			return instances, err
		} else {
			log.Err(err).Msg("Found an error but some instances. Ignoring")
		}
	}
	return instances, nil
}

func (c *Client) accumulateRunningInstancesWithGroup(instances []Instance, nextToken *string) ([]Instance, error) {
	output, err := c.aws.DescribeInstances(
		context.Background(),
		&awsEC2.DescribeInstancesInput{
			Filters: []awsEC2Types.Filter{{
				Name:   &instanceStateNameKey,
				Values: []string{runningInstanceStateValue},
			}},
			MaxResults: &maxResults,
			NextToken:  nextToken,
		},
	)
	if err != nil {
		return instances, err
	}
	log.Debug().Int("number", len(output.Reservations)).Msg("Found filtered ec2 reservations")
	for _, reservation := range output.Reservations {
		for _, awsInstance := range reservation.Instances {
			group, exists := awsInstanceGroup(awsInstance)
			if exists {
				instance := Instance{
					Group: group,
					Type:  typeFromAWSInstance(awsInstance),
				}
				instances = append(instances, instance)
			} else {
				log.Debug().Any("tags", awsInstance.Tags).Msg("Skipping instance - is ungrouped")
			}
		}
	}
	log.Debug().Int("number", len(instances)).Msg("Added AWS instances")
	if output.NextToken != nil {
		return c.accumulateRunningInstancesWithGroup(instances, nextToken)
	}
	return instances, nil
}

func (c *Client) InstanceCosts(instances []Instance) (InstanceCosts, error) {
	costs := InstanceCosts{}
	for _, instance := range instances {
		_, exists := costs[instance.Type]
		if !exists {
			cost, err := c.averageInstanceCost(instance)
			if err == nil {
				costs[instance.Type] = cost
			} else {
				log.Err(err).Any("type", instance.Type).Msg("Failed to get instance cost")
			}
		}
	}
	return costs, nil
}

func (c *Client) averageInstanceCost(instance Instance) (InstanceCost, error) {
	filters := pricingClient.ProductFilters{{
		Field: "instanceType",
		Value: string(instance.Type),
	}}
	priceLists, err := c.pricing.PriceLists(serviceCode, filters)
	if err != nil {
		return InstanceCost{}, err
	}
	log.Debug().
		Int("number", len(priceLists)).
		Any("type", instance.Type).
		Msg("Found price lists for instance type")
	priceLists = filterPriceListsForBoxUsage(priceLists)
	values := usdPerHourForOnDemandInPriceLists(priceLists)
	cost := InstanceCost{
		Cost: types.Cost{
			Dollars: types.USD(mean(values)),
			Per:     time.Hour,
		},
	}
	return cost, nil
}

// filter price lists for "box usage" aka. the price associated with the EC2 compute
func filterPriceListsForBoxUsage(priceLists []pricingClient.ProductPriceList) []pricingClient.ProductPriceList {
	res := []pricingClient.ProductPriceList{}
	for _, pl := range priceLists {
		usageType, exists := pl.Product.Attributes["usagetype"]
		if exists {
			if strings.Contains(usageType, "BoxUsage") {
				res = append(res, pl)
			}
		} else {
			log.Error().Msg("usageType not present in response")
		}
	}
	log.Debug().Int("number", len(res)).Msg("Filtered price lists for 'BoxUsage'")
	return res
}

func usdPerHourForOnDemandInPriceLists(priceLists []pricingClient.ProductPriceList) []float64 {
	values := []float64{}
	for _, pl := range priceLists {
		onDemand, exists := pl.Terms["OnDemand"]
		if !exists {
			log.Error().Msg("missing onDemand in price list terms")
			continue
		}
		onDemandMap, ok := onDemand.(map[string]any)
		if !ok {
			log.Error().Msg("Failed to parse onDemand object as map")
			continue
		}
		for _, sku := range onDemandMap {
			skuMap, ok := sku.(map[string]any)
			if !ok {
				log.Error().Msg("Failed to parse sku object as map")
				continue
			}
			priceDimMap, ok := skuMap["priceDimensions"].(map[string]any)
			if !ok {
				log.Error().Msg("Failed to parse inner sku object as map")
				continue
			}
			for _, priceDimInner := range priceDimMap {
				priceDimInnerMap, ok := priceDimInner.(map[string]any)
				if !ok {
					log.Error().Msg("Failed to parse price dim inner sku object as map")
					continue
				}
				if priceDimInnerMap["unit"] != "Hrs" {
					log.Warn().Msg("unit was not Hours")
					continue
				}
				pricePerUnitMap, ok := priceDimInnerMap["pricePerUnit"].(map[string]any)
				if !ok {
					log.Error().Msg("Failed to parse price dim inner per unit sku object as map")
					continue
				}
				value, err := strconv.ParseFloat(pricePerUnitMap["USD"].(string), 64)
				if err != nil {
					log.Err(err).Msg("Failed to get USD value")
					continue
				}
				values = append(values, value)
			}
		}
	}
	return values
}

func mean(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func typeFromAWSInstance(instance awsEC2Types.Instance) InstanceType {
	return InstanceType(instance.InstanceType)
}

func awsInstanceGroup(instance awsEC2Types.Instance) (types.Group, bool) {
	for _, tag := range instance.Tags {
		if tag.Key == nil || tag.Value == nil {
			continue
		}
		if *tag.Key == config.GroupTagKey() {
			return types.Group(*tag.Value), true
		}
	}
	return "", false
}
