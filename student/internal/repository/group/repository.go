package group

import (
	"github.com/jackc/pgx/v5/pgxpool"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

var _ def.GroupRepository = (*repository)(nil)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}
