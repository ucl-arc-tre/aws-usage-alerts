package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
	awsController "github.com/ucl-arc-tre/aws-cost-alerts/internal/controller/aws"
	emailController "github.com/ucl-arc-tre/aws-cost-alerts/internal/controller/email"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/db"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

type Manager struct {
	aws   *awsController.Controller
	email *emailController.Controller
	db    db.Database
}

func New() *Manager {
	manager := Manager{
		aws:   awsController.New(),
		email: emailController.New(),
	}
	switch backend := config.StorageBackend(); backend {
	case "inMemory":
		manager.db = db.NewInMemory()
	case "configMap":
		manager.db = db.NewConfigMap()
	default:
		panic(fmt.Sprintf("Unrecognized storage backend [%v]", backend))
	}
	return &manager
}

func (m *Manager) Loop(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	var lastTotalsLog time.Time
	for {
		state, err := m.manage()
		if err != nil {
			log.Err(err).Msg("Failed to manage")
		} else if time.Since(lastTotalsLog) >= time.Hour {
			logCumulativeUsage(state)
			lastTotalsLog = time.Now()
		}
		select {
		case <-ctx.Done():
			log.Info().Msg("Exiting manager loop")
			return
		case <-time.After(config.ManagerLoopDelayDuration()):
			continue
		}
	}
}

func (m *Manager) manage() (*types.StateV1alpha1, error) {
	usage := m.aws.Usage()
	state, err := m.db.Load()
	if err != nil {
		return state, err
	}
	state.AddUsage(usage)
	m.email.Send(state, usage.Errors())
	m.db.Store(state)
	return state, nil
}

func logCumulativeUsage(state *types.StateV1alpha1) {
	totals := map[types.Group]types.USD{}
	for _, projectsUsage := range state.GroupsUsageInMonth {
		for group, usage := range projectsUsage {
			value, exists := totals[group]
			if !exists {
				value = 0
			}
			totals[group] = value + usage.Total().Dollars
		}
	}
	for group, total := range totals {
		log.Info().Any("group", group).Int("$", int(total)).Msg("Cumulative total")
	}
}
