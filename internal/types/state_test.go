package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type PartialState struct {
	Version string `json:"version"`
}

func TestMakeState(t *testing.T) {
	state := MakeState()
	assert.NotNil(t, state.EmailsSentAt)
	assert.NotNil(t, state.Version)
	assert.NotNil(t, state.GroupsUsageInMonth)
}

func TestGroupsUsage(t *testing.T) {
	s := MakeState()
	group := Group("a")
	yearAndMonth := YearAndMonth("2006-01")
	s.GroupsUsageInMonth[yearAndMonth] = GroupsUsage{
		Group("a"): AWSAccumulatedCost{
			EFS: AccumulatedCost{
				Dollars: USD(0.1),
				At:      time.Now(),
			},
		},
	}
	// should be no usage for a random current year/month
	yearAndMonthAtTimeZero := YearAndMonthAt(time.Time{})
	_, exists := s.GroupsUsageAt(yearAndMonthAtTimeZero)[group]
	assert.False(t, exists)
	// but there should when it's set
	value, exists := s.GroupsUsageAt(yearAndMonth)[group]
	assert.True(t, exists)
	assert.NotZero(t, value.EFS.Dollars)
}

func TestAddUsage(t *testing.T) {
	s := MakeState()
	group := Group("a")
	awsUsage := AWSUsage{
		EFS: ResourceUsage{
			group: Cost{Dollars: USD(0.3), Per: time.Hour},
		},
		EC2: ResourceUsage{
			group: Cost{Dollars: USD(0.6), Per: time.Hour},
		},
	}
	s.AddUsage(awsUsage)
	ec2AccCost := s.GroupsUsageNow()[group].EC2.Dollars
	assert.Less(t, ec2AccCost, 1e-5)
	efsAccCost := s.GroupsUsageNow()[group].EFS.Dollars
	assert.Less(t, efsAccCost, 1e-5)
	time.Sleep(10 * time.Millisecond)
	s.AddUsage(awsUsage)
	assert.Greater(t, s.GroupsUsageNow()[group].EC2.Dollars, ec2AccCost)
	assert.Greater(t, s.GroupsUsageNow()[group].EFS.Dollars, efsAccCost)
}

func TestStateMarshaling(t *testing.T) {
	s := MakeState()
	var partialState PartialState
	err := json.Unmarshal([]byte(s.Marshal()), &partialState)
	assert.NoError(t, err)
}

func TestAddUsageDoesNotChargeIdleResources(t *testing.T) {
	for _, resource := range []string{"EFS", "EC2", "both"} {
		t.Run(resource, func(t *testing.T) {
			s := MakeState()
			group := Group("a")
			old := time.Now().Add(-72 * time.Hour)
			s.GroupsUsageInMonth[YearAndMonthNow()] = GroupsUsage{
				group: AWSAccumulatedCost{
					EFS: AccumulatedCost{Dollars: 10, At: old},
					EC2: AccumulatedCost{Dollars: 10, At: old},
				},
			}
			cost := Cost{Dollars: 1, Per: time.Hour}
			usage := AWSUsage{
				EFS: ResourceUsage{group: cost},
				EC2: ResourceUsage{group: cost},
			}
			if resource == "EFS" || resource == "both" {
				usage.EFS = nil
			}
			if resource == "EC2" || resource == "both" {
				usage.EC2 = nil
			}

			before := time.Now()
			s.AddUsage(usage)
			after := time.Now()
			idle := s.GroupsUsageNow()[group]
			for name, accumulated := range map[string]AccumulatedCost{"EFS": idle.EFS, "EC2": idle.EC2} {
				if resource == name || resource == "both" {
					assert.Equal(t, USD(10), accumulated.Dollars)
					assert.False(t, accumulated.At.Before(before))
					assert.False(t, accumulated.At.After(after))
				} else {
					assert.GreaterOrEqual(t, accumulated.Dollars, USD(82))
				}
			}

			s.AddUsage(AWSUsage{
				EFS: ResourceUsage{group: cost},
				EC2: ResourceUsage{group: cost},
			})
			resumed := s.GroupsUsageNow()[group]
			maxIncrease := USD(time.Since(before).Hours())
			assert.InDelta(t, float64(idle.EFS.Dollars), float64(resumed.EFS.Dollars), float64(maxIncrease)+1e-9)
			assert.InDelta(t, float64(idle.EC2.Dollars), float64(resumed.EC2.Dollars), float64(maxIncrease)+1e-9)
		})
	}
}
