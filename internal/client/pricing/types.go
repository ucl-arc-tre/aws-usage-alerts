package pricing

import (
	awsTypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
)

type ProductFilter struct {
	Field string
	Value string
}

type ProductWithAttributes struct {
	Attributes map[string]string
	SKU        string
}

type ProductPriceList struct {
	Product         ProductWithAttributes `json:"product"`
	Terms           map[string]any        `json:"terms"`
	Version         string                `json:"version"`
	PublicationDate string                `json:"publicationDate"`
}

type ProductFilters []ProductFilter

func (p *ProductFilters) appendRegionFilter() {
	*p = append(*p, ProductFilter{
		Field: "regionCode",
		Value: config.AWS().Region,
	})
}

func (p ProductFilters) ToAWS() []awsTypes.Filter {
	awsFilters := []awsTypes.Filter{}
	for _, filter := range p {
		awsFilters = append(awsFilters, awsTypes.Filter{
			Field: &filter.Field,
			Type:  awsTypes.FilterTypeTermMatch,
			Value: &filter.Value,
		})
	}
	log.Trace().Int("number", len(awsFilters)).Msg("AWS filters")
	return awsFilters
}
