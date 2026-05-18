package config

import (
	"os"
	"strconv"
)

type Config struct {
	App AppConfig
}

type AppConfig struct {
	Name string
	Port string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3400"
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "transfer-service"
	}

	return &Config{
		App: AppConfig{
			Name: appName,
			Port: port,
		},
	}, nil
}

func GetEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}
