package settings

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// MongoDB contains configuration settings for a MongoDB database.
type MongoDB struct {
	MongodbURI      string `split_words:"true" required:"true"`
	MongodbDatabase string `split_words:"true" required:"true"`
}

// Logger contains configuration settings for a logger.
type Logger struct {
	LogLevel string `split_words:"true" default:"INFO"`
}

// Monitoring contains configuration settings for the monitoring endpoints.
type Monitoring struct {
	// MonitoringPort defines the TCP port for the monitoring endpoints.
	MonitoringPort string `split_words:"true" default:"8000"`
	PprofEnabled   bool   `split_words:"true" default:"false"`
}

// LoadFromEnv loads the configuration settings from environment variables.
//
// If there is a .env file in the current directory, it will be used to
// populate the environment variables.
func LoadFromEnv[T any]() (*T, error) {

	// Load .env file (if it exists)
	_ = godotenv.Load()

	// Load environment variables into a struct
	var settings T
	err := envconfig.Process("", &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from environment: %w", err)
	}

	return &settings, nil
}
