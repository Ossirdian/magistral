package branch

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
	def "github.com/ossirdian/magistral/saas/internal/repository"
	"github.com/ossirdian/magistral/saas/internal/repository/converter"
)

func (r *repository) Create(ctx context.Context, input model.BranchCreate) (model.Branch, error) {
	repoInput := converter.BranchCreateFromModel(input)
	columns := []string{"school_id", "name", "address"}
	values := []any{repoInput.SchoolID, repoInput.Name, repoInput.Address}

	if repoInput.Status != nil {
		columns = append(columns, "status")
		values = append(values, *repoInput.Status)
	}

	builder := def.Builder().
		Insert("branches").
		Columns(columns...).
		Values(values...)

	query, args, err := builder.Suffix("RETURNING id::text").ToSql()
	if err != nil {
		r.logError(ctx, "create.build", err, "school_id", input.SchoolID, "name", input.Name)
		return model.Branch{}, err
	}

	var branchID string
	if err = r.pool.QueryRow(ctx, query, args...).Scan(&branchID); err != nil {
		err = def.NormalizeError(err)
		r.logError(ctx, "create.query", err, "school_id", input.SchoolID, "name", input.Name)
		return model.Branch{}, err
	}

	return r.GetByID(ctx, branchID)
}
