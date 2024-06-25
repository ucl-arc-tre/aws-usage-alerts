package types

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
	// should be no usage for the current year/month
	_, exists := s.GroupsUsageNow()[group]
	assert.False(t, exists)
	s.GroupsUsageInMonth[YearAndMonthNow()] = s.GroupsUsageInMonth[yearAndMonth]
	// but there should when it's set'
	value, exists := s.GroupsUsageNow()[group]
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
	assert.True(t, strings.Contains(s.Marshal(), `"version"`))
	assert.True(t, strings.HasPrefix(s.Marshal(), "{"))
	assert.True(t, strings.HasSuffix(s.Marshal(), "}"))
}
