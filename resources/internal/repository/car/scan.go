package car

import repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"

type scanner interface {
	Scan(dest ...any) error
}

func scanCar(row scanner) (repomodel.Car, error) {
	var item repomodel.Car
	err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.BranchID,
		&item.InstructorID,
		&item.PlateNumber,
		&item.Brand,
		&item.Model,
		&item.Color,
		&item.Transmission,
		&item.Status,
		&item.AvailabilityStatus,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}
