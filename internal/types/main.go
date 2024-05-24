package types

import "time"

type Cost struct {
	MicroCents int
	Per        time.Duration
}

type Group string

type AWSUsage struct {
	EFS map[Group]Cost
	EC2 map[Group]Cost
}
