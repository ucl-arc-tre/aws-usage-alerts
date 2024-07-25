package db

import "github.com/ucl-arc-tre/aws-cost-alerts/internal/types"

type Database interface {
	// Load the current state and lock for edits
	Load() (*types.StateV1alpha1, error)

	// Store the state and release the state lock
	Store(state *types.StateV1alpha1)
}
