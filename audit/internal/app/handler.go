package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ossirdian/magistral/audit/internal/store"
	"github.com/ossirdian/magistral/platform/pkg/authn"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var event store.EventCreate
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.store.Create(r.Context(), event); err != nil {
		http.Error(w, "failed to create audit event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"data":{"status":"accepted"}}`))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	validation, ok := authn.ValidationFromHeaders(r.Header)
	if !ok || !validation.IsValid || validation.Identity.Role != authn.RoleSuperAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	filter := store.EventFilter{
		Service:  strings.TrimSpace(r.URL.Query().Get("service")),
		UserID:   strings.TrimSpace(r.URL.Query().Get("user_id")),
		SchoolID: strings.TrimSpace(r.URL.Query().Get("school_id")),
		Limit:    intQuery(r, "limit", 100),
		Offset:   intQuery(r, "offset", 0),
	}
	if value := strings.TrimSpace(r.URL.Query().Get("date_from")); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			filter.DateFrom = &parsed
		}
	}
	if value := strings.TrimSpace(r.URL.Query().Get("date_to")); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			filter.DateTo = &parsed
		}
	}

	items, count, err := h.store.List(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to list audit events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{
			"limit":  filter.Limit,
			"offset": filter.Offset,
			"count":  count,
		},
	})
}

func intQuery(r *http.Request, key string, fallback int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || value < 0 {
		return fallback
	}
	if key == "limit" && (value == 0 || value > 500) {
		return fallback
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
