package types

import (
	"time"

	"github.com/rs/zerolog/log"
)

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
	Errors  []error
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

func (a *AccumulatedCost) AddCostToNow(cost Cost) {
	duration := time.Since(a.At)
	a.Dollars += USD(float64(cost.Dollars) * (duration.Seconds() / cost.Per.Seconds()))
	a.At = time.Now()
	log.Trace().Any("cost", a).Msg("Added accumulated cost")
}

type AWSAccumulatedCost struct {
	EFS AccumulatedCost
	EC2 AccumulatedCost
}

func makeAWSAccumulatedCostNow() AWSAccumulatedCost {
	return AWSAccumulatedCost{
		EFS: AccumulatedCost{At: time.Now()},
		EC2: AccumulatedCost{At: time.Now()},
	}
}

func (a AWSAccumulatedCost) Total() AccumulatedCost {
	total := AccumulatedCost{At: time.Now()}
	total.Dollars += a.EC2.Dollars
	total.Dollars += a.EFS.Dollars
	return total
}
