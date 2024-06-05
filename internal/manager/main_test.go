package manager

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/client/efs"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/controller/aws"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/controller/email"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/db"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

var (
	oneUSDPerGBHour = types.CostPerUnit{Dollars: types.USD(1.0), PerUnit: types.GB, PerTime: 1 * time.Hour}
	testGroup       = types.Group("test")
)

type MockSNSClient struct{}

func (c *MockSNSClient) Send(content string) error {
	return nil
}

type MockEFSClient struct{}

func (c *MockEFSClient) FileSystems() []efs.EFSFileSystem {
	fs := efs.EFSFileSystem{
		Group: testGroup,
		Size: struct {
			StandardBytes float64
			ArchiveBytes  float64
			IABytes       float64
		}{
			StandardBytes: 5000.0,
		},
	}
	return []efs.EFSFileSystem{fs}
}
func (c *MockEFSClient) CurrentCostPerUnit() (efs.EFSCostPerUnit, error) {
	return efs.EFSCostPerUnit{Standard: oneUSDPerGBHour}, nil
}

type MockEC2Client struct{}

func newMockManager() *Manager {
	manager := Manager{
		aws:   aws.NewWithClients(&MockEC2Client{}, &MockEFSClient{}),
		email: email.NewWithClient(&MockSNSClient{}),
		db:    db.NewInMemory(),
	}
	viper.AddConfigPath("testdata")
	return &manager
}

// todo: make this much better
func TestManageLoopInitThenSingleLoadStore(t *testing.T) {
	m := newMockManager()
	m.manage() // should not panic
	state1, err := m.db.Load()
	m.db.Store(state1)
	assert.Nil(t, err)
	assert.Len(t, state1.GroupsUsageInMonth, 1)
	monthNow := types.YearAndMonthNow()
	usageTotal1 := state1.GroupsUsageInMonth[monthNow][testGroup].Total()
	m.manage() // should add some amount of usage to the state
	usageTotal2 := state1.GroupsUsageInMonth[monthNow][testGroup].Total()
	assert.NotEqual(t, usageTotal1, usageTotal2)
}
