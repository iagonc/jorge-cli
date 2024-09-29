package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
    APIBaseURL string
    Timeout    time.Duration
    Version    string
}

func LoadConfig() (*Config, error) {
    viper.SetDefault("API_BASE_URL", "http://localhost:8080/api/v1")
    viper.SetDefault("TIMEOUT", 10)
    viper.SetDefault("VERSION", "v1.0.0")

    viper.AutomaticEnv()

    cfg := &Config{
        APIBaseURL: viper.GetString("API_BASE_URL"),
        Timeout:    viper.GetDuration("TIMEOUT") * time.Second,
        Version:    viper.GetString("VERSION"),
    }

    // Validate configurations
    if cfg.APIBaseURL == "" {
        return nil, errors.New("API_BASE_URL is required")
    }

    return cfg, nil
}
