package instructor

import repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"

type scanner interface {
	Scan(dest ...any) error
}

func scanInstructor(row scanner) (repomodel.Instructor, error) {
	var item repomodel.Instructor
	err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.BranchID,
		&item.FullName,
		&item.Phone,
		&item.Email,
		&item.Status,
		&item.AvailabilityStatus,
		&item.CurrentCarID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}
