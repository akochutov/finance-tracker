package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func LoadConfig() (*Config, error) {
	cfg := defaultConfig()

	val, ok := os.LookupEnv("DB_URL")
	if !ok || val == "" {
		return nil, fmt.Errorf("env variable DB_URL must be set")
	}
	cfg.DatabaseURL = val

	return cfg, nil
}

func defaultConfig() *Config {
	cfg := &Config{}
	cfg.HTTPAddr = ":8080"

	return cfg
}
