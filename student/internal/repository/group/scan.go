package group

import (
	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/student/internal/model"
)

func scanGroup(row pgx.Row) (model.Group, error) {
	var item model.Group
	err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.BranchID,
		&item.Name,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}
