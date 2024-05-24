package db

import (
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

type InMemory struct {
	mutex sync.RWMutex
	state types.StateV1alpha1
}

func NewInMemory() *InMemory {
	return &InMemory{}
}

func (d *InMemory) Load() *types.StateV1alpha1 {
	d.mutex.Lock()
	return &d.state
}

func (d *InMemory) Store(state *types.StateV1alpha1) {
	defer d.mutex.Unlock()
	if state != nil {
		d.state = *state
	} else {
		log.Error().Msg("Attempted to save a nil state")
	}
}
