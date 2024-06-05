package pricing

type Interface interface {
	PriceListJSON(serviceCode string, region string) ([]byte, error)
}
