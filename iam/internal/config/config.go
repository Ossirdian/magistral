package config

import "time"

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type GRPCConfig interface {
	Addr() string
}

type HTTPConfig interface {
	Addr() string
	AllowedOrigins() []string
}

type PostgresConfig interface {
	DSN() string
	MigrationsDir() string
	AutoMigrate() bool
}

type AuthConfig interface {
	JWTSecret() string
	Issuer() string
	AccessTokenTTL() time.Duration
	RefreshTokenTTL() time.Duration
	ResetTokenTTL() time.Duration
}

type BootstrapConfig interface {
	SuperAdminEmail() string
	SuperAdminPassword() string
}

type TracingConfig interface {
	CollectorEndpoint() string
	ServiceName() string
	Environment() string
	ServiceVersion() string
}

type Config interface {
	Logger() LoggerConfig
	GRPC() GRPCConfig
	HTTP() HTTPConfig
	Postgres() PostgresConfig
	Auth() AuthConfig
	Bootstrap() BootstrapConfig
	Tracing() TracingConfig
}
