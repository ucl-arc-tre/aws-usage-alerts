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
	for {
		m.manage()
		select {
		case <-ctx.Done():
			log.Info().Msg("Exiting manager loop")
			return
		case <-time.After(config.ManagerLoopDelayDuration()):
			continue
		}
	}
}

func (m *Manager) manage() {
	usage := m.aws.Usage()
	state, err := m.db.Load()
	if err != nil {
		log.Err(err).Msg("Failed to load the state - cannot continue")
		return
	}
	state.AddUsage(usage)
	m.email.Send(state, usage.Errors())
	m.db.Store(state)
}
