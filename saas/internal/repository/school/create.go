package school

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
	def "github.com/ossirdian/magistral/saas/internal/repository"
	"github.com/ossirdian/magistral/saas/internal/repository/converter"
)

func (r *repository) Create(ctx context.Context, input model.SchoolCreate) (model.School, error) {
	repoInput := converter.SchoolCreateFromModel(input)
	columns := []string{"name"}
	values := []any{repoInput.Name}

	if repoInput.Status != nil {
		columns = append(columns, "status")
		values = append(values, *repoInput.Status)
	}

	builder := def.Builder().
		Insert("schools").
		Columns(columns...).
		Values(values...)

	query, args, err := builder.Suffix("RETURNING id::text").ToSql()
	if err != nil {
		r.logError(ctx, "create.build", err)
		return model.School{}, err
	}

	var schoolID string
	if err = r.pool.QueryRow(ctx, query, args...).Scan(&schoolID); err != nil {
		err = def.NormalizeError(err)
		r.logError(ctx, "create.query", err, "name", input.Name)
		return model.School{}, err
	}

	return r.GetByID(ctx, schoolID)
}
