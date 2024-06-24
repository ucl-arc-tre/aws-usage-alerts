package efs

import (
	"context"
	"fmt"
	"strings"

	awsEFS "github.com/aws/aws-sdk-go-v2/service/efs"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/rs/zerolog/log"
	pricingClient "github.com/ucl-arc-tre/aws-cost-alerts/internal/client/pricing"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

const (
	serviceCode = "AmazonEFS"
)

type Client struct {
	aws     awsClientInterface
	pricing pricingClient.Interface
}

func New() *Client {
	client := Client{
		aws:     awsEFS.NewFromConfig(config.AWS()),
		pricing: pricingClient.New(),
	}
	return &client
}

func (c *Client) FileSystems() []EFSFileSystem {
	return c.accumulateFileSystems([]EFSFileSystem{}, nil)
}

func (c *Client) accumulateFileSystems(fileSystems []EFSFileSystem, nextMarker *string) []EFSFileSystem {
	output, err := c.aws.DescribeFileSystems(
		context.Background(),
		&awsEFS.DescribeFileSystemsInput{
			Marker: nextMarker,
		},
	)
	if err != nil {
		log.Err(err).Msg("Failed to describe EFS file systems")
		return fileSystems
	}
	for _, fs := range output.FileSystems {
		if fs.FileSystemId != nil {
			if fs.SizeInBytes == nil {
				log.Warn().Msg("File system had no available size")
				continue
			}
			fileSystem := EFSFileSystem{Id: *fs.FileSystemId}
			if group, ok := tagValue(fs, config.GroupTagKey()); ok {
				fileSystem.Group = types.Group(group)
			} else { // tag doesn't exist
				continue
			}
			fileSystem.Size.StandardBytes = valueOrZero(fs.SizeInBytes.ValueInStandard)
			fileSystem.Size.IABytes = valueOrZero(fs.SizeInBytes.ValueInIA)
			fileSystem.Size.ArchiveBytes = valueOrZero(fs.SizeInBytes.ValueInArchive)
			fileSystems = append(fileSystems, fileSystem)
		}
	}
	if output.NextMarker == nil {
		return fileSystems
	} else {
		return c.accumulateFileSystems(fileSystems, output.NextMarker)
	}
}

func (c *Client) CostPerUnit() (EFSCostPerUnit, error) {
	cost := EFSCostPerUnit{}
	priceList, err := c.priceList()
	if err != nil {
		cost.Errors = append(cost.Errors, err)
		return cost, err
	}
	if v, err := priceList.Standard.CostPerUnit(); err != nil {
		cost.Errors = append(cost.Errors, err)
	} else {
		cost.Standard = v
	}
	if v, err := priceList.IA.CostPerUnit(); err != nil {
		cost.Errors = append(cost.Errors, err)
	} else {
		cost.IA = v
	}
	if v, err := priceList.Archive.CostPerUnit(); err != nil {
		cost.Errors = append(cost.Errors, err)
	} else {
		cost.Archive = v
	}
	if len(cost.Errors) > 0 {
		return cost, fmt.Errorf("%v", cost.Errors)
	} else {
		return cost, nil
	}
}

func (c *Client) priceList() (EFSPriceList, error) {
	priceLists, err := c.pricing.PriceLists(serviceCode, pricingClient.ProductFilters{})
	if err != nil {
		return EFSPriceList{}, err
	}
	efsPriceList := EFSPriceList{}
	for _, priceList := range priceLists {
		usageType, exists := priceList.Product.Attributes["usagetype"]
		if exists && strings.Contains(usageType, "TimedStorage-ByteHrs") {
			storageClass, storageClassExists := priceList.Product.Attributes["storageClass"]
			if storageClassExists {
				switch storageClass {
				case "Infrequent Access":
					efsPriceList.IA = priceList.Terms
				case "General Purpose":
					efsPriceList.Standard = priceList.Terms
				case "Archive":
					efsPriceList.Archive = priceList.Terms
				}
			} else {
				log.Error().Any("priceList", priceList).Msg("storageClass attribute not found")
			}
		}
	}
	return efsPriceList, err
}

func tagValue(fs awsTypes.FileSystemDescription, tagKey string) (string, bool) {
	for _, tag := range fs.Tags {
		if tag.Key != nil && *tag.Key == tagKey {
			if tag.Value == nil {
				return "", true
			} else {
				return *tag.Value, true
			}
		}
	}
	return "", false
}
