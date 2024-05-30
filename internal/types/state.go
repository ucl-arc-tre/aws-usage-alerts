package types

import (
	"time"

	"github.com/rs/zerolog/log"
)

type YearAndMonth string

type GroupsUsage map[Group]AWSAccumulatedCost

type StateV1alpha1 struct {
	Version            string                       `json:"version"`
	GroupsUsageInMonth map[YearAndMonth]GroupsUsage `json:"groups_usage"`
	EmailsSentAt       map[EmailAddress]time.Time   `json:"emails_sent_at"`
}

func MakeState() StateV1alpha1 {
	return StateV1alpha1{
		Version:            "v1alpha1",
		GroupsUsageInMonth: map[YearAndMonth]GroupsUsage{},
		EmailsSentAt:       map[EmailAddress]time.Time{},
	}
}

func (s *StateV1alpha1) AddUsage(usage AWSUsage) {
	if s.EmailsSentAt == nil || s.GroupsUsageInMonth == nil {
		log.Error().Msg("Cannot add usage with undefined maps")
		return
	}
	s.addCurrentMonthIfRequired()
	groupsUsage := s.GroupsUsageInMonth[yearAndMonthNow()]
	log.Debug().Msg("Adding resource usage")
	for group, cost := range usage.EFS {
		accCost, exists := groupsUsage[group]
		if exists {
			duration := time.Since(accCost.EFS.At)
			accCost.EFS.Dollars += USD(float64(cost.Dollars) * (duration.Seconds() / cost.Per.Seconds()))
			accCost.EFS.At = time.Now()
			log.Trace().Any("cost", accCost).Any("group", group).Msg("Group accumulated cost")
			groupsUsage[group] = accCost
		} else {
			groupsUsage[group] = AWSAccumulatedCost{
				EFS: AccumulatedCost{At: time.Now()},
			}
		}
	}
	s.GroupsUsageInMonth[yearAndMonthNow()] = groupsUsage
	log.Debug().Any("state", s).Msg("Added usage")
	// todo: ec2
}

// Usage for every group in the current month
func (s *StateV1alpha1) GroupsUsage() GroupsUsage {
	return s.GroupsUsageInMonth[yearAndMonthNow()]
}

func (s *StateV1alpha1) addCurrentMonthIfRequired() {
	yearAndMonth := yearAndMonthNow()
	_, exists := s.GroupsUsageInMonth[yearAndMonth]
	if !exists {
		log.Debug().Any("yearAndMonth", yearAndMonth).Msg("Adding new year/month")
		s.GroupsUsageInMonth[yearAndMonth] = GroupsUsage{}
	}
}

func yearAndMonthNow() YearAndMonth {
	return YearAndMonth(time.Now().Format("2006-01"))
}
