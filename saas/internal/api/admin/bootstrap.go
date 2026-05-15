package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ossirdian/magistral/saas/internal/model"
	repo "github.com/ossirdian/magistral/saas/internal/repository"
	svc "github.com/ossirdian/magistral/saas/internal/service"
	schoolservice "github.com/ossirdian/magistral/saas/internal/service/school"
)

type schoolBootstrapService interface {
	BootstrapWithAdmin(ctx context.Context, input model.SchoolCreate, admin model.SchoolAdminBootstrap) (model.School, string, error)
}

type schoolBootstrapHandler struct {
	service schoolBootstrapService
}

type bootstrapSchoolRequest struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	AdminEmail string `json:"admin_email"`
	Password   string `json:"password"`
}

type bootstrapSchoolResponse struct {
	Data bootstrapSchoolResult `json:"data"`
}

type bootstrapSchoolResult struct {
	SchoolID    string `json:"school_id"`
	SchoolName  string `json:"school_name"`
	SchoolState string `json:"school_status"`
	AdminUserID string `json:"admin_user_id"`
}

func NewSchoolBootstrapHandler(service schoolBootstrapService) (http.Handler, error) {
	if service == nil {
		return nil, errors.New("school bootstrap service is nil")
	}

	return &schoolBootstrapHandler{service: service}, nil
}

func (h *schoolBootstrapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req bootstrapSchoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	var status *model.SchoolStatus
	if req.Status != "" {
		value := model.SchoolStatus(req.Status)
		status = &value
	}

	school, adminUserID, err := h.service.BootstrapWithAdmin(r.Context(), model.SchoolCreate{
		Name:   req.Name,
		Status: status,
	}, model.SchoolAdminBootstrap{
		Email:    req.AdminEmail,
		Password: req.Password,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, bootstrapSchoolResponse{
		Data: bootstrapSchoolResult{
			SchoolID:    school.ID,
			SchoolName:  school.Name,
			SchoolState: string(school.Status),
			AdminUserID: adminUserID,
		},
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svc.ErrInvalidArgument):
		writeError(w, http.StatusBadRequest, "invalid_argument")
	case errors.Is(err, repo.ErrConflict):
		writeError(w, http.StatusConflict, "conflict")
	case errors.Is(err, repo.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	case errors.Is(err, schoolservice.ErrIAMClientUnavailable):
		writeError(w, http.StatusServiceUnavailable, "iam_unavailable")
	default:
		writeError(w, http.StatusInternalServerError, "internal")
	}
}

func writeError(w http.ResponseWriter, statusCode int, code string) {
	writeJSON(w, statusCode, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": code,
		},
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
