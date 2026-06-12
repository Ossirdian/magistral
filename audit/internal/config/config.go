package config

import (
	"net"
	"os"
	"strconv"
)

type Config struct {
	HTTPHost        string
	HTTPPort        string
	PostgresDSN     string
	AutoMigrate     bool
	LoggerLevel     string
	LoggerAsJSON    bool
	TracingEndpoint string
	TracingEnv      string
	TracingVersion  string
}

func FromEnv() Config {
	return Config{
		HTTPHost:        env("HTTP_HOST", "0.0.0.0"),
		HTTPPort:        env("HTTP_PORT", "8088"),
		PostgresDSN:     env("POSTGRES_DSN", ""),
		AutoMigrate:     envBool("AUTO_MIGRATE", true),
		LoggerLevel:     env("LOGGER_LEVEL", "info"),
		LoggerAsJSON:    envBool("LOGGER_AS_JSON", false),
		TracingEndpoint: env("TRACING_COLLECTOR_ENDPOINT", ""),
		TracingEnv:      env("TRACING_ENVIRONMENT", "local"),
		TracingVersion:  env("TRACING_SERVICE_VERSION", "dev"),
	}
}

func (c Config) Addr() string { return net.JoinHostPort(c.HTTPHost, c.HTTPPort) }

func (c Config) CollectorEndpoint() string { return c.TracingEndpoint }
func (c Config) ServiceName() string       { return "audit" }
func (c Config) Environment() string       { return c.TracingEnv }
func (c Config) ServiceVersion() string    { return c.TracingVersion }

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
