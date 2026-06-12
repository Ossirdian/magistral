package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ossirdian/magistral/platform/pkg/authn"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	"go.opentelemetry.io/otel/trace"
)

const envBaseURL = "AUDIT_HTTP_BASE_URL"

type Event struct {
	Service    string `json:"service"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Query      string `json:"query,omitempty"`
	StatusCode int    `json:"status_code"`
	DurationMS int64  `json:"duration_ms"`
	UserID     string `json:"user_id,omitempty"`
	UserRole   string `json:"user_role,omitempty"`
	SchoolID   string `json:"school_id,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	TraceID    string `json:"trace_id,omitempty"`
	RemoteAddr string `json:"remote_addr,omitempty"`
}

func Middleware(service string, next http.Handler) http.Handler {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv(envBaseURL)), "/")
	if baseURL == "" {
		return next
	}
	client := &http.Client{Timeout: 2 * time.Second}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)

		event := Event{
			Service:    service,
			Method:     r.Method,
			Path:       r.URL.Path,
			Query:      r.URL.RawQuery,
			StatusCode: recorder.statusCode,
			DurationMS: time.Since(started).Milliseconds(),
			UserID:     strings.TrimSpace(r.Header.Get(authn.HeaderUserID)),
			UserRole:   strings.TrimSpace(r.Header.Get(authn.HeaderUserRole)),
			RemoteAddr: r.RemoteAddr,
			RequestID:  requestID(r),
			TraceID:    traceID(r.Context()),
		}
		if schoolID := strings.TrimSpace(r.Header.Get(authn.HeaderSchoolID)); schoolID != "" {
			event.SchoolID = schoolID
		}
		go send(context.Background(), client, baseURL, event)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func send(ctx context.Context, client *http.Client, baseURL string, event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/audit/events", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set(authn.HeaderUserID, "audit-middleware")
	req.Header.Set(authn.HeaderUserRole, string(authn.RoleSuperAdmin))
	req.Header.Set(authn.HeaderAuthenticated, strconv.FormatBool(true))
	resp, err := client.Do(req)
	if err != nil {
		logger.Global().Warn("failed to send audit event", "error", err)
		return
	}
	_ = resp.Body.Close()
}

func skip(path string) bool {
	return path == "/healthz" || path == "/metrics" || strings.HasPrefix(path, "/audit/events")
}

func requestID(r *http.Request) string {
	for _, header := range []string{"x-request-id", "x-correlation-id"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	return ""
}

func traceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}
