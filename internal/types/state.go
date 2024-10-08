package types

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

type YearAndMonth string

type GroupsUsage map[Group]AWSAccumulatedCost

type StateVersion string

type StateWithVersion struct {
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
			accCost.EFS.AddCostToNow(cost)
			groupsUsage[group] = accCost
		} else {
			groupsUsage[group] = makeAWSAccumulatedCostNow()
		}
	}
	for group, cost := range usage.EC2 {
		accCost, exists := groupsUsage[group]
		if exists {
			accCost.EC2.AddCostToNow(cost)
			groupsUsage[group] = accCost
		} else {
			groupsUsage[group] = makeAWSAccumulatedCostNow()
		}
	}
	s.GroupsUsageInMonth[YearAndMonthNow()] = groupsUsage
	log.Debug().Any("state", s).Msg("Added usage")
}

// Usage for every group in the current month
func (s *StateV1alpha1) GroupsUsageNow() GroupsUsage {
	return s.GroupsUsageAt(YearAndMonthNow())
}

func (s *StateV1alpha1) GroupsUsageAt(yearAndMonth YearAndMonth) GroupsUsage {
	usage, exists := s.GroupsUsageInMonth[yearAndMonth]
	if exists {
		return usage
	} else {
		log.Error().Msg("Groups usage currently did not exist")
		return GroupsUsage{}
	}
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
	return YearAndMonthAt(time.Now())
}

func YearAndMonthAt(instant time.Time) YearAndMonth {
	return YearAndMonth(instant.Format("2006-01"))
}
