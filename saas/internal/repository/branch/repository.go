package branch

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	def "github.com/ossirdian/magistral/saas/internal/repository"
)

var _ def.BranchRepository = (*repository)(nil)

type repository struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

func NewRepository(pool *pgxpool.Pool, log logger.Logger) *repository {
	if log == nil {
		log = &logger.NoopLogger{}
	}

	return &repository{
		pool: pool,
		log:  log,
	}
}

func (r *repository) logError(ctx context.Context, op string, err error, args ...interface{}) {
	fields := append([]interface{}{"repository", "branch", "op", op, "error", err}, args...)
	r.log.WithContext(ctx).Error("saas repository error", fields...)
}
