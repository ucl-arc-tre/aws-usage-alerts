package config

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Load configuration from environment variables & files and configure zerolog
func Init() {
	if os.Getenv("DEBUG") == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	initViper()
}

func envOrDefault(key string, defaultValue string) string {
	if value := os.Getenv(key); value == "" {
		return defaultValue
	} else {
		return value
	}
}

// Initialise viper. See: github.com/spf13/viper
func initViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(envOrDefault("CONFIG_DIR", "/etc/aws-usage-alerts"))
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed to open config file: %w", err))
	}
	viper.WatchConfig()
}
