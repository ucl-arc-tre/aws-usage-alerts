package efs

import (
	"errors"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

const (
	bytesPerGB = 1e-9

	daysPerYear     = 365.25
	hoursPerDay     = 24.0
	monthsPerYear   = 12.0
	hoursPerMonth   = (hoursPerDay * daysPerYear) / monthsPerYear
	minutesPerMonth = hoursPerMonth * 60
)

type EFSFileSystem struct {
	Id    string
	Group types.Group
	Size  struct {
		StandardBytes float64
		ArchiveBytes  float64
		IABytes       float64 // infrequent access
	}
}

func (e *EFSFileSystem) Cost(perUnitCost EFSCostPerUnit) types.Cost {
	cost := types.Cost{
		Per:    perUnitCost.Standard.PerTime, // assumes standard == IA == archive
		Errors: perUnitCost.Errors,
	}
	cost.Dollars += types.USD(float64(perUnitCost.Standard.Dollars) * e.Size.StandardBytes * bytesPerGB)
	cost.Dollars += types.USD(float64(perUnitCost.InfrequentAccess.Dollars) * e.Size.IABytes * bytesPerGB)
	cost.Dollars += types.USD(float64(perUnitCost.Archive.Dollars) * e.Size.ArchiveBytes * bytesPerGB)
	return cost
}

type EFSCostPerUnit struct {
	Standard         types.CostPerUnit
	InfrequentAccess types.CostPerUnit
	Archive          types.CostPerUnit
	Errors           []error
}

type AWSPriceListTerms map[string]any

type EFSPriceList struct {
	Standard AWSPriceListTerms
	Archive  AWSPriceListTerms
	IA       AWSPriceListTerms
}

func (t AWSPriceListTerms) CostPerUnit() (types.CostPerUnit, error) {
	onDemand, exists := t["OnDemand"]
	if !exists {
		return types.CostPerUnit{}, errors.New("missing OnDemand in price list terms")
	}
	onDemandMap := onDemand.(map[string]any)
	for _, sku := range onDemandMap {
		skuMap, ok := sku.(map[string]any)
		if !ok {
			return types.CostPerUnit{}, errors.New("failed to parse sku object as map")
		}
		priceDimMap, ok := skuMap["priceDimensions"].(map[string]any)
		if !ok {
			return types.CostPerUnit{}, errors.New("failed to parse inner sku object as map")
		}
		for _, priceDimInner := range priceDimMap {
			priceDimInnerMap, ok := priceDimInner.(map[string]any)
			if !ok {
				return types.CostPerUnit{}, errors.New("failed to parse inner sku object as map")
			}
			if priceDimInnerMap["unit"] != "GB-Mo" {
				log.Warn().Msg("unit was not GB-Mo")
				continue
			}
			pricePerUnitMap, ok := priceDimInnerMap["pricePerUnit"].(map[string]any)
			if !ok {
				log.Error().Msg("Failed to parse price dim inner per unit sku object as map")
				continue
			}
			value, err := strconv.ParseFloat(pricePerUnitMap["USD"].(string), 64)
			cost := types.CostPerUnit{
				Dollars: types.USD(value),
				PerTime: time.Duration(time.Minute * minutesPerMonth),
				PerUnit: types.GB,
			}
			if err == nil {
				return cost, nil
			} else {
				return types.CostPerUnit{}, err
			}
		}
	}
	return types.CostPerUnit{}, errors.New("failed to parse OnDemand object pricing terms")

}
