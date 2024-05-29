package db

import (
	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

type ConfigMap struct {
}

func NewConfigMap() *ConfigMap {
	log.Debug().Msg("Creating new config map storage backend")
	return &ConfigMap{}
}

func (cm *ConfigMap) Load() *types.StateV1alpha1 {
	return nil // todo
}

func (cm *ConfigMap) Store(state *types.StateV1alpha1) {
	panic("not implemented")
}

// todo set lease. wait for lease to not exist on create
