package db

import "github.com/ucl-arc-tre/aws-cost-alerts/internal/types"

type Database interface {
	Load() *types.StateV1alpha1
	Store(state *types.StateV1alpha1)
}
