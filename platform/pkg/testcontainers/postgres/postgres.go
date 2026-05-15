package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type Container struct {
	container *tcpostgres.PostgresContainer
	dbName    string
	user      string
	password  string
	host      string
	port      string
}

func NewContainer(ctx context.Context, dbName, user, password string) (*Container, error) {
	c, err := tcpostgres.Run(ctx,
		"postgres:17.0-alpine3.20",
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(user),
		tcpostgres.WithPassword(password),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve postgres host: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, fmt.Errorf("resolve postgres port: %w", err)
	}

	return &Container{
		container: c,
		dbName:    dbName,
		user:      user,
		password:  password,
		host:      host,
		port:      mappedPort.Port(),
	}, nil
}

func (c *Container) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		c.user,
		c.password,
		net.JoinHostPort(c.host, c.port),
		c.dbName,
	)
}

func (c *Container) OpenDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("pgx", c.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres db: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres db: %w", err)
	}

	return db, nil
}

func (c *Container) Terminate(ctx context.Context) error {
	return c.container.Terminate(ctx)
}
