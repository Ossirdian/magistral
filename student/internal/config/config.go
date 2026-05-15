package config

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type HTTPConfig interface {
	Addr() string
	AllowedOrigins() []string
}

type GRPCConfig interface {
	Addr() string
}

type PostgresConfig interface {
	DSN() string
	MigrationsDir() string
	AutoMigrate() bool
}

type IAMConfig interface {
	Addr() string
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
	GRPC() GRPCConfig
	Postgres() PostgresConfig
	LessonPostgres() PostgresConfig
	IAM() IAMConfig
	Tracing() TracingConfig
}
