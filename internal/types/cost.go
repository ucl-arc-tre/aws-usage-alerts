package types

import "time"

type Unit string

type USD float64

const (
	GB Unit = "GB"
)

type CostPerUnit struct {
	Dollars USD
	PerTime time.Duration
	PerUnit Unit
}

type Cost struct {
	Dollars USD
	Per     time.Duration
}

func (c *Cost) Add(other Cost) {
	if c.Per != other.Per {
		panic("attempted to add costs with different time intervals")
	}
	c.Dollars += other.Dollars
}

type AccumulatedCost struct {
	Dollars USD
	At      time.Time
}

type AWSAccumulatedCost struct {
	EFS AccumulatedCost
	EC2 AccumulatedCost
}

func (a *AWSAccumulatedCost) Total() AccumulatedCost {
	total := AccumulatedCost{At: time.Now()}
	total.Dollars += a.EC2.Dollars
	total.Dollars += a.EFS.Dollars
	return total
}
