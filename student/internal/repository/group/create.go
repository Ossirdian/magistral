package group

import (
	"context"

	"github.com/ossirdian/magistral/student/internal/model"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

func (r *repository) Create(ctx context.Context, schoolID string, input model.GroupCreate) (model.Group, error) {
	query, args, err := def.Builder().
		Insert("groups").
		Columns("school_id", "branch_id", "name").
		Values(schoolID, input.BranchID, input.Name).
		Suffix("RETURNING id::text").
		ToSql()
	if err != nil {
		return model.Group{}, err
	}

	var groupID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&groupID); err != nil {
		return model.Group{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, groupID)
}
