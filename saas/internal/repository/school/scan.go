package school

import (
	"time"

	repomodel "github.com/ossirdian/magistral/saas/internal/repository/model"
)

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanSchool(row scanner) (repomodel.School, error) {
	var item repomodel.School
	var deletedAt *time.Time

	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
	); err != nil {
		return repomodel.School{}, err
	}

	item.DeletedAt = deletedAt
	return item, nil
}
