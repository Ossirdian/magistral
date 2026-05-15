package env

import (
	"fmt"
	"net"
	"strings"

	cenv "github.com/caarlos0/env/v11"
	def "github.com/ossirdian/magistral/reporting/internal/config"
)

type rawConfig struct {
	LoggerLevel          string `env:"LOGGER_LEVEL" envDefault:"info"`
	LoggerAsJSON         bool   `env:"LOGGER_AS_JSON" envDefault:"false"`
	HTTPHost             string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	HTTPPort             string `env:"HTTP_PORT" envDefault:"8086"`
	AllowedOrigins       string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`
	IAMGRPCAddr          string `env:"IAM_GRPC_ADDR" envDefault:""`
	StudentPostgresDSN   string `env:"STUDENT_POSTGRES_DSN,required"`
	LessonPostgresDSN    string `env:"LESSON_POSTGRES_DSN,required"`
	ResourcesPostgresDSN string `env:"RESOURCES_POSTGRES_DSN,required"`
	PaymentPostgresDSN   string `env:"PAYMENT_POSTGRES_DSN,required"`
	TracingEndpoint      string `env:"TRACING_COLLECTOR_ENDPOINT" envDefault:""`
	TracingEnv           string `env:"TRACING_ENVIRONMENT" envDefault:"local"`
	TracingVersion       string `env:"TRACING_SERVICE_VERSION" envDefault:"dev"`
}

type Config struct {
	raw rawConfig
}

type (
	loggerConfig   struct{ raw rawConfig }
	httpConfig     struct{ raw rawConfig }
	iamConfig      struct{ raw rawConfig }
	postgresConfig struct{ dsn string }
	tracingConfig  struct{ raw rawConfig }
)

var _ def.Config = (*Config)(nil)

func New() (*Config, error) {
	var raw rawConfig
	if err := cenv.Parse(&raw); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}
	return &Config{raw: raw}, nil
}

func (c *Config) Logger() def.LoggerConfig            { return &loggerConfig{raw: c.raw} }
func (c *Config) HTTP() def.HTTPConfig                { return &httpConfig{raw: c.raw} }
func (c *Config) IAM() def.IAMConfig                  { return &iamConfig{raw: c.raw} }
func (c *Config) StudentPostgres() def.PostgresConfig { return &postgresConfig{dsn: c.raw.StudentPostgresDSN} }
func (c *Config) LessonPostgres() def.PostgresConfig  { return &postgresConfig{dsn: c.raw.LessonPostgresDSN} }
func (c *Config) ResourcesPostgres() def.PostgresConfig {
	return &postgresConfig{dsn: c.raw.ResourcesPostgresDSN}
}
func (c *Config) PaymentPostgres() def.PostgresConfig { return &postgresConfig{dsn: c.raw.PaymentPostgresDSN} }
func (c *Config) Tracing() def.TracingConfig          { return &tracingConfig{raw: c.raw} }

func (c *loggerConfig) Level() string { return c.raw.LoggerLevel }
func (c *loggerConfig) AsJSON() bool  { return c.raw.LoggerAsJSON }

func (c *httpConfig) Addr() string {
	return net.JoinHostPort(c.raw.HTTPHost, c.raw.HTTPPort)
}

func (c *httpConfig) AllowedOrigins() []string {
	parts := strings.Split(c.raw.AllowedOrigins, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return result
}

func (c *iamConfig) Addr() string  { return strings.TrimSpace(c.raw.IAMGRPCAddr) }
func (c *postgresConfig) DSN() string { return c.dsn }

func (c *tracingConfig) CollectorEndpoint() string { return c.raw.TracingEndpoint }
func (c *tracingConfig) ServiceName() string       { return "reporting" }
func (c *tracingConfig) Environment() string       { return c.raw.TracingEnv }
func (c *tracingConfig) ServiceVersion() string    { return c.raw.TracingVersion }
