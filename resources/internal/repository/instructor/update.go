package instructor

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
)

func (r *repository) Update(ctx context.Context, schoolID, instructorID string, patch model.InstructorPatch) (model.Instructor, error) {
	repoPatch := converter.InstructorPatchFromModel(patch)
	builder := def.Builder().
		Update("instructors").
		Set("updated_at", sq.Expr("now()")).
		Where("id = ?", instructorID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")
	changed := false

	if repoPatch.BranchID != nil {
		builder = builder.Set("branch_id", *repoPatch.BranchID)
		changed = true
	}
	if repoPatch.FullName != nil {
		builder = builder.Set("full_name", *repoPatch.FullName)
		changed = true
	}
	if repoPatch.Phone.Set {
		builder = builder.Set("phone", repoPatch.Phone.Value)
		changed = true
	}
	if repoPatch.Email.Set {
		builder = builder.Set("email", repoPatch.Email.Value)
		changed = true
	}
	if repoPatch.Status != nil {
		builder = builder.Set("status", *repoPatch.Status)
		changed = true
	}

	if !changed {
		return r.GetByID(ctx, schoolID, instructorID)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Instructor{}, err
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return model.Instructor{}, def.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.Instructor{}, def.ErrNotFound
	}

	return r.GetByID(ctx, schoolID, instructorID)
}
