package instructor

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
)

func (r *repository) Create(ctx context.Context, schoolID string, input model.InstructorCreate) (model.Instructor, error) {
	repoInput := converter.InstructorCreateFromModel(input)
	columns := []string{"school_id", "branch_id", "full_name", "phone", "email"}
	values := []any{schoolID, repoInput.BranchID, repoInput.FullName, repoInput.Phone, repoInput.Email}

	if repoInput.Status != nil {
		columns = append(columns, "status")
		values = append(values, *repoInput.Status)
	}

	builder := def.Builder().
		Insert("instructors").
		Columns(columns...).
		Values(values...)

	query, args, err := builder.Suffix("RETURNING id::text").ToSql()
	if err != nil {
		return model.Instructor{}, err
	}

	var instructorID string
	if err = r.pool.QueryRow(ctx, query, args...).Scan(&instructorID); err != nil {
		return model.Instructor{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, instructorID)
}
