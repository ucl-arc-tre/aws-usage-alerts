package pricing

type Interface interface {
	PriceLists(serviceCode string, filters ProductFilters) ([]ProductPriceList, error)
}
