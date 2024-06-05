package types

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

type YearAndMonth string

type GroupsUsage map[Group]AWSAccumulatedCost

type StateVersion string

type StateWithVersionVersion struct {
	Version StateVersion `json:"version"`
}

type StateV1alpha1 struct {
	Version            StateVersion                 `json:"version"`
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
	groupsUsage := s.GroupsUsageInMonth[YearAndMonthNow()]
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
	s.GroupsUsageInMonth[YearAndMonthNow()] = groupsUsage
	log.Debug().Any("state", s).Msg("Added usage")
	// todo: ec2
}

// Usage for every group in the current month
func (s *StateV1alpha1) GroupsUsage() GroupsUsage {
	return s.GroupsUsageInMonth[YearAndMonthNow()]
}

func (s *StateV1alpha1) addCurrentMonthIfRequired() {
	yearAndMonth := YearAndMonthNow()
	_, exists := s.GroupsUsageInMonth[yearAndMonth]
	if !exists {
		log.Debug().Any("yearAndMonth", yearAndMonth).Msg("Adding new year/month")
		s.GroupsUsageInMonth[yearAndMonth] = GroupsUsage{}
	}
}

func (s *StateV1alpha1) Marshal() string {
	if result, err := json.Marshal(s); err != nil {
		log.Err(err).Msg("Failed to marshal. Using an empty string")
		return ""
	} else {
		return string(result)
	}
}

func YearAndMonthNow() YearAndMonth {
	return YearAndMonth(time.Now().Format("2006-01"))
}
