package config

import "os"

type Config struct {
	Port        string
	Environment string
}

func New() *Config {
	return &Config{
		Port:        getEnvOrDefault("PORT", "5000"),
		Environment: getEnvOrDefault("ENV", "development"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
