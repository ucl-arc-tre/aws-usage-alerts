package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLoopDelay(t *testing.T) {
	os.Setenv("UPDATE_DELAY_SECONDS", "")
	assert.Equal(t, 1*time.Minute, ManagerLoopDelayDuration())
}

func TestLoopDelayFromEnv(t *testing.T) {
	os.Setenv("UPDATE_DELAY_SECONDS", "10")
	assert.Equal(t, 10*time.Second, ManagerLoopDelayDuration())
}
