package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	SlackBotToken string
	SlackAppToken string
	DatabasePath  string
	PeopleChannel string
	Debug         bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (optional)
	_ = godotenv.Load()

	config := &Config{
		SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
		SlackAppToken: os.Getenv("SLACK_APP_TOKEN"),
		DatabasePath:  getEnvOrDefault("DATABASE_PATH", "fambot.db"),
		PeopleChannel: getEnvOrDefault("PEOPLE_CHANNEL", "people"),
		Debug:         os.Getenv("DEBUG") == "true",
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validate ensures all required configuration is present
func (c *Config) validate() error {
	if c.SlackBotToken == "" {
		return fmt.Errorf("SLACK_BOT_TOKEN is required")
	}
	if c.SlackAppToken == "" {
		return fmt.Errorf("SLACK_APP_TOKEN is required")
	}
	return nil
}

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
