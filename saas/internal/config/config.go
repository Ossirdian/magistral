package config

type LoggerConfig interface {
	Level() string
	AsJSON() bool
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

type IAMConfig interface {
	BaseURL() string
}

type TracingConfig interface {
	CollectorEndpoint() string
	ServiceName() string
	Environment() string
	ServiceVersion() string
}

type Config interface {
	Logger() LoggerConfig
	HTTP() HTTPConfig
	Postgres() PostgresConfig
	IAM() IAMConfig
	Tracing() TracingConfig
}
