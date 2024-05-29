package types

import (
	"time"

	"github.com/rs/zerolog/log"
)

type StateV1alpha1 struct {
	Version      string                       `json:"version"`
	GroupsUsage  map[Group]AWSAccumulatedCost `json:"groups_usage"`
	EmailsSentAt map[EmailAddress]time.Time   `json:"emails_sent_at"`
}

func MakeState() StateV1alpha1 {
	return StateV1alpha1{
		Version:      "v1alpha1",
		GroupsUsage:  map[Group]AWSAccumulatedCost{},
		EmailsSentAt: map[EmailAddress]time.Time{},
	}
}

func (s *StateV1alpha1) AddUsage(usage AWSUsage) {
	if s.EmailsSentAt == nil || s.GroupsUsage == nil {
		log.Error().Msg("Cannot add usage with undefined maps")
		return
	}
	log.Debug().Msg("Adding resource usage")
	for group, cost := range usage.EFS {
		accCost, exists := s.GroupsUsage[group]
		if exists {
			duration := time.Since(accCost.EFS.At)
			accCost.EFS.Dollars += USD(float64(cost.Dollars) * (duration.Seconds() / cost.Per.Seconds()))
			accCost.EFS.At = time.Now()
			log.Trace().Any("cost", accCost).Any("group", group).Msg("Group accumulated cost")
			s.GroupsUsage[group] = accCost
		} else {
			s.GroupsUsage[group] = AWSAccumulatedCost{
				EFS: AccumulatedCost{At: time.Now()},
			}
		}
	}
	log.Debug().Any("state", s).Msg("Added usage")
	// todo: ec2
}
