package config

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type HTTPConfig interface {
	Addr() string
	AllowedOrigins() []string
}

type IAMConfig interface {
	Addr() string
}

type PostgresConfig interface {
	DSN() string
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
	IAM() IAMConfig
	StudentPostgres() PostgresConfig
	LessonPostgres() PostgresConfig
	ResourcesPostgres() PostgresConfig
	PaymentPostgres() PostgresConfig
	Tracing() TracingConfig
}
