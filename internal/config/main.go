package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

const (
	defaultAWSRegion = "eu-west-2"
)

// Load configuration from environment variables & files and configure zerolog
func Init() {
	if os.Getenv("TRACE") == "true" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if os.Getenv("DEBUG") == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	initViper()
}

// Initialise viper. See: github.com/spf13/viper
func initViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(envOrDefault("CONFIG_DIR", "."))
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed to open config file: %w", err))
	}
	viper.WatchConfig()
}

// Get the value of an environment variable and fallback to a default if it's unset
func envOrDefault(key string, defaultValue string) string {
	if value := os.Getenv(key); value == "" {
		return defaultValue
	} else {
		return value
	}
}

// Get the value of an environment variable and log an error if it's unset
func env(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Warn().Str("key", key).Msg("Environment variable unset")
	}
	return value
}

func StorageBackend() string {
	return viper.GetString("storageBackend")
}

func GroupTagKey() string {
	return viper.GetString("groupTagKey")
}

func ManagerLoopDelayDuration() time.Duration {
	seconds, err := strconv.Atoi(envOrDefault("UPDATE_DELAY_SECONDS", "60"))
	if err != nil {
		panic(err)
	}
	return time.Duration(seconds) * time.Second
}

func AWS() aws.Config {
	config, err := awsConfig.LoadDefaultConfig(context.Background()) // Loads from ENV vars
	if err != nil {
		panic(fmt.Errorf("unable to load AWS SDK config, %v", err))
	}
	if config.Region == "" {
		log.Warn().
			Str("defaultAWSRegion", defaultAWSRegion).
			Msg("AWS region not explicity set. Please set 'AWS_REGION' as an env var")
		config.Region = defaultAWSRegion
	}
	return config
}

func HealthPort() string {
	return envOrDefault("HEALTH_PORT", "8080")
}

func SNSTopicARN() string {
	return env("SNS_TOPIC_ARN")
}

func GroupThreshold(group types.Group) types.USD {
	return types.USD(viper.GetFloat64("groups." + string(group) + ".threshold"))
}

func AdminEmails() []types.EmailAddress {
	emails := []types.EmailAddress{}
	for _, email := range viper.GetStringSlice("adminEmails") {
		emails = append(emails, types.EmailAddress(email))
	}
	return emails
}
