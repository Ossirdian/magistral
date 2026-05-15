package env

import (
	"fmt"
	"net"
	"strings"

	cenv "github.com/caarlos0/env/v11"
	def "github.com/ossirdian/magistral/payment/internal/config"
)

type rawConfig struct {
	LoggerLevel     string `env:"LOGGER_LEVEL" envDefault:"info"`
	LoggerAsJSON    bool   `env:"LOGGER_AS_JSON" envDefault:"false"`
	HTTPHost        string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	HTTPPort        string `env:"HTTP_PORT" envDefault:"8084"`
	AllowedOrigins  string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`
	PostgresDSN     string `env:"POSTGRES_DSN,required"`
	MigrationsDir   string `env:"MIGRATIONS_DIR" envDefault:"./payment/migrations"`
	AutoMigrate     bool   `env:"AUTO_MIGRATE" envDefault:"false"`
	IAMGRPCAddr     string `env:"IAM_GRPC_ADDR" envDefault:""`
	StudentGRPCAddr string `env:"STUDENT_GRPC_ADDR" envDefault:""`
	TracingEndpoint string `env:"TRACING_COLLECTOR_ENDPOINT" envDefault:""`
	TracingEnv      string `env:"TRACING_ENVIRONMENT" envDefault:"local"`
	TracingVersion  string `env:"TRACING_SERVICE_VERSION" envDefault:"dev"`
}

type Config struct {
	raw rawConfig
}

type (
	loggerConfig   struct{ raw rawConfig }
	httpConfig     struct{ raw rawConfig }
	postgresConfig struct{ raw rawConfig }
	iamConfig      struct{ raw rawConfig }
	studentConfig  struct{ raw rawConfig }
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

func (c *Config) Logger() def.LoggerConfig     { return &loggerConfig{raw: c.raw} }
func (c *Config) HTTP() def.HTTPConfig         { return &httpConfig{raw: c.raw} }
func (c *Config) Postgres() def.PostgresConfig { return &postgresConfig{raw: c.raw} }
func (c *Config) IAM() def.IAMConfig           { return &iamConfig{raw: c.raw} }
func (c *Config) Student() def.StudentConfig   { return &studentConfig{raw: c.raw} }
func (c *Config) Tracing() def.TracingConfig   { return &tracingConfig{raw: c.raw} }

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

func (c *postgresConfig) DSN() string           { return c.raw.PostgresDSN }
func (c *postgresConfig) MigrationsDir() string { return c.raw.MigrationsDir }
func (c *postgresConfig) AutoMigrate() bool     { return c.raw.AutoMigrate }

func (c *iamConfig) Addr() string { return c.raw.IAMGRPCAddr }

func (c *studentConfig) Addr() string {
	return strings.TrimSpace(c.raw.StudentGRPCAddr)
}

func (c *tracingConfig) CollectorEndpoint() string { return c.raw.TracingEndpoint }
func (c *tracingConfig) ServiceName() string       { return "payment" }
func (c *tracingConfig) Environment() string       { return c.raw.TracingEnv }
func (c *tracingConfig) ServiceVersion() string    { return c.raw.TracingVersion }
