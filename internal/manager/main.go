package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	default:
		panic(fmt.Sprintf("Unrecognized storage backend [%v]", backend))
	}
	return &manager
}

func (m *Manager) Loop(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			usage := m.aws.Usage()
			state := m.db.Load()
			state.AddUsage(usage)
			m.email.Send(state)
			m.db.Store(state)
			time.Sleep(config.ManagerLoopDelayDuration())
		}
	}
}
