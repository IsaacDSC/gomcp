package config

import (
	"os"
	"strings"
)

const (
	defaultLogLevel = "INFO"
)

type Config struct {
	LogLevel string
}

func Load() Config {
	return Config{
		LogLevel: readEnv("LOG_LEVEL", defaultLogLevel),
	}
}

func readEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
