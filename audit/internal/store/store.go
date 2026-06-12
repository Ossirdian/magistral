package store

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

type Event struct {
	ID         string    `json:"id"`
	Service    string    `json:"service"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Query      string    `json:"query,omitempty"`
	StatusCode int       `json:"status_code"`
	DurationMS int64     `json:"duration_ms"`
	UserID     string    `json:"user_id,omitempty"`
	UserRole   string    `json:"user_role,omitempty"`
	SchoolID   string    `json:"school_id,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	TraceID    string    `json:"trace_id,omitempty"`
	RemoteAddr string    `json:"remote_addr,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type EventCreate struct {
	Service    string `json:"service"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Query      string `json:"query"`
	StatusCode int    `json:"status_code"`
	DurationMS int64  `json:"duration_ms"`
	UserID     string `json:"user_id"`
	UserRole   string `json:"user_role"`
	SchoolID   string `json:"school_id"`
	RequestID  string `json:"request_id"`
	TraceID    string `json:"trace_id"`
	RemoteAddr string `json:"remote_addr"`
}

type EventFilter struct {
	Service  string
	UserID   string
	SchoolID string
	DateFrom *time.Time
	DateTo   *time.Time
	Limit    int
	Offset   int
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    service text NOT NULL,
    method text NOT NULL,
    path text NOT NULL,
    query text NOT NULL DEFAULT '',
    status_code integer NOT NULL,
    duration_ms bigint NOT NULL,
    user_id text NOT NULL DEFAULT '',
    user_role text NOT NULL DEFAULT '',
    school_id uuid NULL,
    request_id text NOT NULL DEFAULT '',
    trace_id text NOT NULL DEFAULT '',
    remote_addr text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS audit_events_created_at_idx ON audit_events (created_at DESC);
CREATE INDEX IF NOT EXISTS audit_events_service_created_idx ON audit_events (service, created_at DESC);
CREATE INDEX IF NOT EXISTS audit_events_school_created_idx ON audit_events (school_id, created_at DESC);
`)
	return err
}

func (s *Store) Create(ctx context.Context, event EventCreate) error {
	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO audit_events
		 (service, method, path, query, status_code, duration_ms, user_id, user_role, school_id, request_id, trace_id, remote_addr)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, '')::uuid, $10, $11, $12)`,
		trim(event.Service),
		trim(event.Method),
		trim(event.Path),
		trim(event.Query),
		event.StatusCode,
		event.DurationMS,
		trim(event.UserID),
		trim(event.UserRole),
		trim(event.SchoolID),
		trim(event.RequestID),
		trim(event.TraceID),
		trim(event.RemoteAddr),
	)
	return err
}

func (s *Store) List(ctx context.Context, filter EventFilter) ([]Event, int, error) {
	where, args := whereClause(filter)
	countQuery := "SELECT count(*) FROM audit_events" + where
	var count int
	if err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&count); err != nil {
		return nil, 0, err
	}

	args = append(args, filter.Limit, filter.Offset)
	rows, err := s.pool.Query(ctx, `SELECT id::text, service, method, path, query, status_code, duration_ms,
       user_id, user_role, COALESCE(school_id::text, ''), request_id, trace_id, remote_addr, created_at
FROM audit_events`+where+` ORDER BY created_at DESC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Event, 0, filter.Limit)
	for rows.Next() {
		var item Event
		if err := rows.Scan(&item.ID, &item.Service, &item.Method, &item.Path, &item.Query, &item.StatusCode, &item.DurationMS, &item.UserID, &item.UserRole, &item.SchoolID, &item.RequestID, &item.TraceID, &item.RemoteAddr, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, count, rows.Err()
}

func whereClause(filter EventFilter) (string, []any) {
	clauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	add := func(clause string, arg any) {
		args = append(args, arg)
		clauses = append(clauses, clause+" $"+itoa(len(args)))
	}
	if filter.Service != "" {
		add("service =", filter.Service)
	}
	if filter.UserID != "" {
		add("user_id =", filter.UserID)
	}
	if filter.SchoolID != "" {
		args = append(args, filter.SchoolID)
		clauses = append(clauses, "school_id = NULLIF($"+itoa(len(args))+", '')::uuid")
	}
	if filter.DateFrom != nil {
		add("created_at >=", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		add("created_at <=", *filter.DateTo)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func trim(value string) string {
	return strings.TrimSpace(value)
}

func itoa(value int) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	buf := make([]byte, 0, 8)
	for value > 0 {
		buf = append([]byte{digits[value%10]}, buf...)
		value /= 10
	}
	return string(buf)
}
