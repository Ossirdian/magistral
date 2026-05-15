package branch

import (
	"time"

	repomodel "github.com/ossirdian/magistral/saas/internal/repository/model"
)

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanBranch(row scanner) (repomodel.Branch, error) {
	var item repomodel.Branch
	var address *string
	var deletedAt *time.Time

	if err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.Name,
		&address,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
	); err != nil {
		return repomodel.Branch{}, err
	}

	item.Address = address
	item.DeletedAt = deletedAt
	return item, nil
}
