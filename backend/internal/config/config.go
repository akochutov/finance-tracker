package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	RedisURL    string
}

func LoadConfig() (*Config, error) {
	cfg := defaultConfig()

	cacheURL, ok := os.LookupEnv("REDIS_URL")
	if !ok || cacheURL == "" {
		return nil, fmt.Errorf("env variable REDIS_URL must be set")
	}
	cfg.RedisURL = cacheURL

	dbURL, ok := os.LookupEnv("DB_URL")
	if !ok || dbURL == "" {
		return nil, fmt.Errorf("env variable DB_URL must be set")
	}
	cfg.DatabaseURL = dbURL

	return cfg, nil
}

func defaultConfig() *Config {
	cfg := &Config{}
	cfg.HTTPAddr = ":8080"

	return cfg
}
