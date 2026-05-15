package env

import (
	"fmt"
	"net"
	"strings"
	"time"

	cenv "github.com/caarlos0/env/v11"
	def "github.com/ossirdian/magistral/iam/internal/config"
)

type rawConfig struct {
	LoggerLevel     string        `env:"LOGGER_LEVEL" envDefault:"info"`
	LoggerAsJSON    bool          `env:"LOGGER_AS_JSON" envDefault:"false"`
	GRPCHost        string        `env:"GRPC_HOST" envDefault:"0.0.0.0"`
	GRPCPort        string        `env:"GRPC_PORT" envDefault:"50053"`
	HTTPHost        string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	HTTPPort        string        `env:"HTTP_PORT" envDefault:"8082"`
	AllowedOrigins  string        `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`
	PostgresDSN     string        `env:"POSTGRES_DSN,required"`
	MigrationsDir   string        `env:"MIGRATIONS_DIR" envDefault:"./iam/migrations"`
	AutoMigrate     bool          `env:"AUTO_MIGRATE" envDefault:"false"`
	JWTSecret       string        `env:"JWT_SECRET,required"`
	JWTIssuer       string        `env:"JWT_ISSUER" envDefault:"magistral-iam"`
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"720h"`
	ResetTokenTTL   time.Duration `env:"RESET_TOKEN_TTL" envDefault:"1h"`
	SuperAdminEmail string        `env:"BOOTSTRAP_SUPER_ADMIN_EMAIL" envDefault:"superadmin@example.com"`
	SuperAdminPass  string        `env:"BOOTSTRAP_SUPER_ADMIN_PASSWORD" envDefault:"Passw0rd!"`
	TracingEndpoint string        `env:"TRACING_COLLECTOR_ENDPOINT" envDefault:""`
	TracingEnv      string        `env:"TRACING_ENVIRONMENT" envDefault:"local"`
	TracingVersion  string        `env:"TRACING_SERVICE_VERSION" envDefault:"dev"`
}

type (
	Config          struct{ raw rawConfig }
	loggerConfig    struct{ raw rawConfig }
	grpcConfig      struct{ raw rawConfig }
	httpConfig      struct{ raw rawConfig }
	postgresConfig  struct{ raw rawConfig }
	authConfig      struct{ raw rawConfig }
	bootstrapConfig struct{ raw rawConfig }
	tracingConfig   struct{ raw rawConfig }
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
func (c *Config) GRPC() def.GRPCConfig         { return &grpcConfig{raw: c.raw} }
func (c *Config) HTTP() def.HTTPConfig         { return &httpConfig{raw: c.raw} }
func (c *Config) Postgres() def.PostgresConfig { return &postgresConfig{raw: c.raw} }
func (c *Config) Auth() def.AuthConfig         { return &authConfig{raw: c.raw} }
func (c *Config) Bootstrap() def.BootstrapConfig {
	return &bootstrapConfig{raw: c.raw}
}
func (c *Config) Tracing() def.TracingConfig { return &tracingConfig{raw: c.raw} }

func (c *loggerConfig) Level() string { return c.raw.LoggerLevel }
func (c *loggerConfig) AsJSON() bool  { return c.raw.LoggerAsJSON }

func (c *grpcConfig) Addr() string {
	return net.JoinHostPort(c.raw.GRPCHost, c.raw.GRPCPort)
}

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

func (c *authConfig) JWTSecret() string              { return c.raw.JWTSecret }
func (c *authConfig) Issuer() string                 { return c.raw.JWTIssuer }
func (c *authConfig) AccessTokenTTL() time.Duration  { return c.raw.AccessTokenTTL }
func (c *authConfig) RefreshTokenTTL() time.Duration { return c.raw.RefreshTokenTTL }
func (c *authConfig) ResetTokenTTL() time.Duration   { return c.raw.ResetTokenTTL }

func (c *bootstrapConfig) SuperAdminEmail() string    { return c.raw.SuperAdminEmail }
func (c *bootstrapConfig) SuperAdminPassword() string { return c.raw.SuperAdminPass }

func (c *tracingConfig) CollectorEndpoint() string { return c.raw.TracingEndpoint }
func (c *tracingConfig) ServiceName() string       { return "iam" }
func (c *tracingConfig) Environment() string       { return c.raw.TracingEnv }
func (c *tracingConfig) ServiceVersion() string    { return c.raw.TracingVersion }
