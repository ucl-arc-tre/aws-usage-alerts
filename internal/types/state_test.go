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
