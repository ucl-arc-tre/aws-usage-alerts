package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

const (
	testEmailAddr = types.EmailAddress("alice@example.com")
)

func TestInMemLoadSaveConcurrent(t *testing.T) {
	db := NewInMemory()
	state1, err := db.Load()
	assert.NoError(t, err)
	go func() {
		state2, err := db.Load()
		assert.NoError(t, err)
		state2.EmailsSentAt[testEmailAddr] = time.Time{}
		db.Store(state2)
	}()
	expectedTime := time.Now()
	state1.EmailsSentAt[testEmailAddr] = expectedTime
	db.Store(state1)

	state3, err := db.Load()
	assert.NoError(t, err)
	assert.Len(t, state3.EmailsSentAt, 1)
	assert.Equal(t, expectedTime, state3.EmailsSentAt[testEmailAddr])
}

func TestCanSaveNilStateAndDoesNothing(t *testing.T) {
	db := NewInMemory()
	_, err := db.Load()
	assert.NoError(t, err)
	db.Store(nil)
	state, err := db.Load()
	assert.NoError(t, err)
	assert.NotNil(t, state)
	db.Store(state)
}
