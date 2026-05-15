package document

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/document/internal/model"
	repo "github.com/ossirdian/magistral/document/internal/repository"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateDocument(ctx context.Context, schoolID string, item model.Document) (model.Document, error) {
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = uuid.NewString()
	}
	item.SchoolID = schoolID
	item.CreatedAt = now
	item.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO documents (
			id, school_id, student_id, type, title, status, object_key, mime_type, size, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		item.ID, item.SchoolID, item.StudentID, item.Type, item.Title, item.Status, item.ObjectKey, item.MimeType, item.Size, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.Document{}, repo.NormalizeError(err)
	}
	return item, nil
}

func (r *Repository) GetDocument(ctx context.Context, schoolID, documentID string) (model.Document, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, school_id, student_id, type, title, status, object_key, mime_type, size, created_at, updated_at
		FROM documents
		WHERE id = $1 AND school_id = $2 AND deleted_at IS NULL
	`, documentID, schoolID)

	item, err := scanDocument(row)
	if err != nil {
		return model.Document{}, err
	}
	return item, nil
}

func (r *Repository) ListDocuments(ctx context.Context, schoolID string, filter model.DocumentFilter) ([]model.Document, model.PageMeta, error) {
	where := []string{"school_id = $1", "deleted_at IS NULL"}
	args := []any{schoolID}
	index := 2

	if filter.StudentID != nil {
		where = append(where, fmt.Sprintf("student_id = $%d", index))
		args = append(args, *filter.StudentID)
		index++
	}
	if filter.Type != nil {
		where = append(where, fmt.Sprintf("type = $%d", index))
		args = append(args, *filter.Type)
		index++
	}
	if filter.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", index))
		args = append(args, *filter.Status)
		index++
	}

	whereClause := strings.Join(where, " AND ")
	limit := filter.Limit
	offset := filter.Offset
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, school_id, student_id, type, title, status, object_key, mime_type, size, created_at, updated_at
		FROM documents
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, index, index+1), append(args, limit, offset)...)
	if err != nil {
		return nil, model.PageMeta{}, repo.NormalizeError(err)
	}
	defer rows.Close()

	items := make([]model.Document, 0)
	for rows.Next() {
		item, err := scanDocument(rows)
		if err != nil {
			return nil, model.PageMeta{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, model.PageMeta{}, repo.NormalizeError(err)
	}

	var count int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(1) FROM documents WHERE %s`, whereClause), args...).Scan(&count); err != nil {
		return nil, model.PageMeta{}, repo.NormalizeError(err)
	}

	return items, model.PageMeta{Limit: limit, Offset: offset, Count: count}, nil
}

func (r *Repository) DeleteDocument(ctx context.Context, schoolID, documentID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE documents
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND school_id = $2 AND deleted_at IS NULL
	`, documentID, schoolID)
	if err != nil {
		return repo.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *Repository) GetTemplate(ctx context.Context, schoolID, templateID string) (model.DocumentTemplate, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, school_id, name, type, content, created_at, updated_at
		FROM document_templates
		WHERE id = $1 AND school_id = $2 AND deleted_at IS NULL
	`, templateID, schoolID)

	item, err := scanTemplate(row)
	if err != nil {
		return model.DocumentTemplate{}, err
	}
	return item, nil
}

func (r *Repository) ListTemplates(ctx context.Context, schoolID string, filter model.DocumentTemplateFilter) ([]model.DocumentTemplate, model.PageMeta, error) {
	where := []string{"school_id = $1", "deleted_at IS NULL"}
	args := []any{schoolID}
	index := 2

	if filter.Type != nil {
		where = append(where, fmt.Sprintf("type = $%d", index))
		args = append(args, *filter.Type)
		index++
	}
	whereClause := strings.Join(where, " AND ")

	limit := filter.Limit
	offset := filter.Offset
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, school_id, name, type, content, created_at, updated_at
		FROM document_templates
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, index, index+1), append(args, limit, offset)...)
	if err != nil {
		return nil, model.PageMeta{}, repo.NormalizeError(err)
	}
	defer rows.Close()

	items := make([]model.DocumentTemplate, 0)
	for rows.Next() {
		item, err := scanTemplate(rows)
		if err != nil {
			return nil, model.PageMeta{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, model.PageMeta{}, repo.NormalizeError(err)
	}

	var count int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(1) FROM document_templates WHERE %s`, whereClause), args...).Scan(&count); err != nil {
		return nil, model.PageMeta{}, repo.NormalizeError(err)
	}

	return items, model.PageMeta{Limit: limit, Offset: offset, Count: count}, nil
}

func (r *Repository) CreateTemplate(ctx context.Context, schoolID string, input model.DocumentTemplateCreate) (model.DocumentTemplate, error) {
	now := time.Now().UTC()
	item := model.DocumentTemplate{
		ID:        uuid.NewString(),
		SchoolID:  schoolID,
		Name:      input.Name,
		Type:      input.Type,
		Content:   input.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO document_templates (id, school_id, name, type, content, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, item.ID, item.SchoolID, item.Name, item.Type, item.Content, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return model.DocumentTemplate{}, repo.NormalizeError(err)
	}
	return item, nil
}

func (r *Repository) UpdateTemplate(ctx context.Context, schoolID, templateID string, patch model.DocumentTemplatePatch) (model.DocumentTemplate, error) {
	sets := make([]string, 0, 4)
	args := []any{templateID, schoolID}
	index := 3

	if patch.NameSet {
		sets = append(sets, fmt.Sprintf("name = $%d", index))
		args = append(args, valueOrEmpty(patch.Name))
		index++
	}
	if patch.TypeSet {
		sets = append(sets, fmt.Sprintf("type = $%d", index))
		args = append(args, valueOrEmpty(patch.Type))
		index++
	}
	if patch.ContentSet {
		sets = append(sets, fmt.Sprintf("content = $%d", index))
		args = append(args, valueOrEmpty(patch.Content))
	}
	if len(sets) == 0 {
		return model.DocumentTemplate{}, repo.ErrConflict
	}
	sets = append(sets, "updated_at = NOW()")

	tag, err := r.pool.Exec(ctx, fmt.Sprintf(`
		UPDATE document_templates
		SET %s
		WHERE id = $1 AND school_id = $2 AND deleted_at IS NULL
	`, strings.Join(sets, ", ")), args...)
	if err != nil {
		return model.DocumentTemplate{}, repo.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.DocumentTemplate{}, repo.ErrNotFound
	}

	return r.GetTemplate(ctx, schoolID, templateID)
}

func (r *Repository) DeleteTemplate(ctx context.Context, schoolID, templateID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE document_templates
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND school_id = $2 AND deleted_at IS NULL
	`, templateID, schoolID)
	if err != nil {
		return repo.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func scanDocument(row interface{ Scan(dest ...any) error }) (model.Document, error) {
	var item model.Document
	if err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.StudentID,
		&item.Type,
		&item.Title,
		&item.Status,
		&item.ObjectKey,
		&item.MimeType,
		&item.Size,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return model.Document{}, repo.ErrNotFound
		}
		return model.Document{}, repo.NormalizeError(err)
	}
	return item, nil
}

func scanTemplate(row interface{ Scan(dest ...any) error }) (model.DocumentTemplate, error) {
	var item model.DocumentTemplate
	if err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.Name,
		&item.Type,
		&item.Content,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return model.DocumentTemplate{}, repo.ErrNotFound
		}
		return model.DocumentTemplate{}, repo.NormalizeError(err)
	}
	return item, nil
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
