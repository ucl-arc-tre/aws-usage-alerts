package efs

import (
	"context"
	"encoding/json"

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

func (c *Client) CurrentCostPerUnit() (EFSCostPerUnit, error) {
	cost := EFSCostPerUnit{}
	priceList, err := c.priceList()
	if err != nil {
		return cost, err
	}
	if v, err := priceList.CostOfStorageClass("General Purpose"); err != nil {
		log.Err(err).Msg("Failed to get general sku cost")
	} else {
		cost.Standard = v
	}
	if v, err := priceList.CostOfStorageClass("Infrequent Access"); err != nil {
		log.Err(err).Msg("Failed to get IA sku cost")
	} else {
		cost.IA = v
	}
	if v, err := priceList.CostOfStorageClass("Archive"); err != nil {
		log.Err(err).Msg("Failed to get archive sku cost")
	} else {
		cost.Archive = v
	}
	return cost, nil
}

func (c *Client) priceList() (EFSPriceList, error) {
	content, err := c.pricing.PriceListJSON(serviceCode, config.AWS().Region)
	if err != nil {
		return EFSPriceList{}, err
	}
	var priceList EFSPriceList
	err = json.Unmarshal(content, &priceList)
	log.Trace().Any("priceList", priceList).Msg("unmarshalled")
	return priceList, err
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
