package efs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	cost.Dollars += types.USD(float64(perUnitCost.IA.Dollars) * e.Size.IABytes * bytesPerGB)
	cost.Dollars += types.USD(float64(perUnitCost.Archive.Dollars) * e.Size.ArchiveBytes * bytesPerGB)
	return cost
}

type EFSCostPerUnit struct {
	Standard types.CostPerUnit
	IA       types.CostPerUnit
	Archive  types.CostPerUnit
	Errors   []error
}

type SkuId string

type awsEFSPriceListProductAttributes struct {
	StorageClass string `json:"storageClass"`
	UsageType    string `json:"usageType"`
}

type awsEFSPriceListProduct struct {
	Sku        SkuId                            `json:"sku"`
	Attributes awsEFSPriceListProductAttributes `json:"attributes"`
}

type awsEFSPriceListTermSkuPriceDimension struct {
	RateCode     string            `json:"rateCode"`
	Unit         string            `json:"unit"`
	PricePerUnit map[string]string `json:"pricePerUnit"`
}

type awsEFSPriceListTermSku struct {
	Sku             SkuId `json:"sku"`
	PriceDimensions map[string]awsEFSPriceListTermSkuPriceDimension
}

type EFSPriceList struct {
	Products map[SkuId]awsEFSPriceListProduct                       `json:"products"`
	Terms    map[string]map[SkuId]map[string]awsEFSPriceListTermSku `json:"terms"`
}

func (e *EFSPriceList) skuFromStorageClass(storageClass string) (SkuId, error) {
	for _, object := range e.Products {
		if object.Attributes.StorageClass == storageClass &&
			strings.Contains(object.Attributes.UsageType, "ByteHrs") {
			return object.Sku, nil
		}
	}
	return SkuId(""), fmt.Errorf("sku not found for [%v]", storageClass)
}

func (e *EFSPriceList) costOfSku(sku SkuId) (types.CostPerUnit, error) {
	onDemand, ok := e.Terms["OnDemand"]
	if !ok {
		return types.CostPerUnit{}, errors.New("failed to find OnDemand in price list terms")
	}
	skuMap, ok := onDemand[sku]
	if !ok {
		return types.CostPerUnit{}, fmt.Errorf("failed to find [%v] in OnDemand price list", sku)
	}
	log.Trace().Any("skuMap", skuMap).Msg("")
	for _, skuMapInner := range skuMap {
		for _, priceDim := range skuMapInner.PriceDimensions {
			cost := types.CostPerUnit{}
			switch unit := priceDim.Unit; unit {
			case "GB-Mo":
				cost.PerUnit = types.GB
				cost.PerTime = time.Duration(time.Minute * minutesPerMonth)
			default:
				panic(fmt.Errorf("unsupported unit [%v]", unit))
			}
			dollarsString, ok := priceDim.PricePerUnit["USD"]
			if !ok {
				return cost, errors.New("failed to find cost code")
			}
			if v, err := strconv.ParseFloat(dollarsString, 64); err != nil {
				return cost, err
			} else {
				cost.Dollars = types.USD(v)
			}
			return cost, nil
		}
	}
	return types.CostPerUnit{}, errors.New("failed to find any sku data")
}

func (e *EFSPriceList) CostOfStorageClass(storageClass string) (types.CostPerUnit, error) {
	if sku, err := e.skuFromStorageClass(storageClass); err != nil {
		return types.CostPerUnit{}, err
	} else {
		return e.costOfSku(sku)
	}
}
