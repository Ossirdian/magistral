package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ossirdian/magistral/iam/internal/model"
	repo "github.com/ossirdian/magistral/iam/internal/repository"
	svc "github.com/ossirdian/magistral/iam/internal/service"
)

type service interface {
	CreateUser(ctx context.Context, input model.UserCreate) (model.User, error)
	GetUser(ctx context.Context, userID string) (model.User, error)
	ListUsers(ctx context.Context, filter model.UserFilter) ([]model.User, model.PageMeta, error)
	UpdateUser(ctx context.Context, userID string, patch model.UserPatch) (model.User, error)
	SetUserPassword(ctx context.Context, userID, password string) error
	Login(ctx context.Context, email, password string) (model.Session, error)
	RefreshSession(ctx context.Context, refreshToken string) (model.Session, error)
	ValidateToken(ctx context.Context, accessToken string) (model.TokenValidation, error)
}

type API struct {
	service service
}

func New(service service) (*API, error) {
	if service == nil {
		return nil, errors.New("iam http service is nil")
	}

	return &API{service: service}, nil
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/auth/login", a.handleLogin)
	mux.HandleFunc("/auth/refresh", a.handleRefresh)
	mux.Handle("/auth/me", a.withAuth(http.HandlerFunc(a.handleMe)))
	mux.Handle("/admin/users", a.withSuperAdmin(http.HandlerFunc(a.handleUsers)))
	mux.Handle("/admin/users/", a.withSuperAdmin(http.HandlerFunc(a.handleUserByID)))

	return mux
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type createUserRequest struct {
	SchoolID *string `json:"school_id"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Role     string  `json:"role"`
	Status   string  `json:"status"`
}

type updateUserRequest struct {
	Email    *string `json:"email"`
	Role     *string `json:"role"`
	Status   *string `json:"status"`
	IsActive *bool   `json:"is_active"`
}

type setPasswordRequest struct {
	Password string `json:"password"`
}

type responseEnvelope struct {
	Data any `json:"data"`
}

type sessionResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresAt    string         `json:"expires_at"`
	Identity     identityResult `json:"identity"`
}

type identityResult struct {
	UserID   string  `json:"user_id"`
	Role     string  `json:"role"`
	SchoolID *string `json:"school_id,omitempty"`
}

type userResult struct {
	ID        string  `json:"id"`
	SchoolID  *string `json:"school_id,omitempty"`
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	Status    string  `json:"status"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type listUsersResponse struct {
	Users []userResult `json:"users"`
	Meta  metaResult   `json:"meta"`
}

type metaResult struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	session, err := a.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: sessionToResponse(session)})
}

func (a *API) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	session, err := a.service.RefreshSession(r.Context(), req.RefreshToken)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: sessionToResponse(session)})
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	validation, ok := validationFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: map[string]any{
		"user_id":   validation.Identity.UserID,
		"role":      string(validation.Identity.Role),
		"school_id": validation.Identity.SchoolID,
		"status":    string(validation.Status),
		"is_active": validation.IsActive,
		"is_valid":  validation.IsValid,
	}})
}

func (a *API) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleListUsers(w, r)
	case http.MethodPost:
		a.handleCreateUser(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) handleUserByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/users/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	if strings.HasSuffix(path, "/password") {
		userID := strings.TrimSuffix(path, "/password")
		userID = strings.TrimSuffix(userID, "/")
		a.handleSetPassword(w, r, userID)
		return
	}

	userID := strings.TrimSuffix(path, "/")
	switch r.Method {
	case http.MethodGet:
		a.handleGetUser(w, r, userID)
	case http.MethodPatch:
		a.handleUpdateUser(w, r, userID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) handleListUsers(w http.ResponseWriter, r *http.Request) {
	filter := model.UserFilter{
		Limit:  50,
		Offset: 0,
	}

	values := r.URL.Query()
	if schoolID := strings.TrimSpace(values.Get("school_id")); schoolID != "" {
		filter.SchoolID = &schoolID
	}
	if search := strings.TrimSpace(values.Get("search")); search != "" {
		filter.Search = &search
	}
	if role := strings.TrimSpace(values.Get("role")); role != "" {
		typedRole := model.UserRole(role)
		filter.Role = &typedRole
	}
	if status := strings.TrimSpace(values.Get("status")); status != "" {
		typedStatus := model.UserStatus(status)
		filter.Status = &typedStatus
	}
	if limitValue := strings.TrimSpace(values.Get("limit")); limitValue != "" {
		if limit, err := strconv.Atoi(limitValue); err == nil {
			filter.Limit = limit
		}
	}
	if offsetValue := strings.TrimSpace(values.Get("offset")); offsetValue != "" {
		if offset, err := strconv.Atoi(offsetValue); err == nil {
			filter.Offset = offset
		}
	}

	users, meta, err := a.service.ListUsers(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	payload := listUsersResponse{
		Users: usersToResponse(users),
		Meta:  metaResult{Limit: meta.Limit, Offset: meta.Offset, Count: meta.Count},
	}
	writeJSON(w, http.StatusOK, responseEnvelope{Data: payload})
}

func (a *API) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	user, err := a.service.CreateUser(r.Context(), model.UserCreate{
		SchoolID: req.SchoolID,
		Email:    req.Email,
		Password: req.Password,
		Role:     model.UserRole(req.Role),
		Status:   model.UserStatus(req.Status),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, responseEnvelope{Data: userToResponse(user)})
}

func (a *API) handleGetUser(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := a.service.GetUser(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: userToResponse(user)})
}

func (a *API) handleUpdateUser(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	var role *model.UserRole
	if req.Role != nil {
		value := model.UserRole(strings.TrimSpace(*req.Role))
		role = &value
	}

	var status *model.UserStatus
	if req.Status != nil {
		value := model.UserStatus(strings.TrimSpace(*req.Status))
		status = &value
	}

	user, err := a.service.UpdateUser(r.Context(), userID, model.UserPatch{
		Email:    req.Email,
		Role:     role,
		Status:   status,
		IsActive: req.IsActive,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: userToResponse(user)})
}

func (a *API) handleSetPassword(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req setPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	if err := a.service.SetUserPassword(r.Context(), userID, req.Password); err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: map[string]bool{"updated": true}})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svc.ErrInvalidArgument):
		writeError(w, http.StatusBadRequest, "invalid_argument")
	case errors.Is(err, svc.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, repo.ErrConflict):
		writeError(w, http.StatusConflict, "conflict")
	case errors.Is(err, repo.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	default:
		writeError(w, http.StatusInternalServerError, "internal")
	}
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{
		"error": map[string]string{
			"code":    message,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func sessionToResponse(session model.Session) sessionResponse {
	return sessionResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt.UTC().Format(timeFormat),
		Identity: identityResult{
			UserID:   session.Identity.UserID,
			Role:     string(session.Identity.Role),
			SchoolID: session.Identity.SchoolID,
		},
	}
}

func userToResponse(user model.User) userResult {
	return userResult{
		ID:        user.ID,
		SchoolID:  user.SchoolID,
		Email:     user.Email,
		Role:      string(user.Role),
		Status:    string(user.Status),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.UTC().Format(timeFormat),
		UpdatedAt: user.UpdatedAt.UTC().Format(timeFormat),
	}
}

func usersToResponse(users []model.User) []userResult {
	result := make([]userResult, 0, len(users))
	for _, user := range users {
		result = append(result, userToResponse(user))
	}
	return result
}

const timeFormat = "2006-01-02T15:04:05Z07:00"
