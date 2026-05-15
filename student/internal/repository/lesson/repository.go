package lesson

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CompletedHoursByStudentIDs(ctx context.Context, schoolID string, studentIDs []string) (map[string]float64, error) {
	result := make(map[string]float64, len(studentIDs))
	if len(studentIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+1)
	args = append(args, schoolID)
	for i, studentID := range studentIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d::uuid", i+2))
		args = append(args, studentID)
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT
			student_id::text,
			COALESCE(SUM(EXTRACT(EPOCH FROM (end_at - start_at)) / 3600), 0)::double precision
		FROM lessons
		WHERE school_id = $1
		  AND deleted_at IS NULL
		  AND status = 'completed'
		  AND student_id IN (%s)
		GROUP BY student_id
	`, strings.Join(placeholders, ", ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var studentID string
		var completedHours float64
		if err := rows.Scan(&studentID, &completedHours); err != nil {
			return nil, err
		}
		result[studentID] = completedHours
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
