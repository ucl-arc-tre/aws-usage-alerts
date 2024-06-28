package pricing

type Interface interface {
	// Price lists for a product within an AWS service (e.g. AmazonEC2) filtered by a list of filters
	PriceLists(serviceCode string, filters ProductFilters) ([]ProductPriceList, error)
}
