package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/ossirdian/magistral/platform/pkg/authn"
	"github.com/ossirdian/magistral/iam/internal/model"
)

type contextKey string

const validationKey contextKey = "iam.http.validation"

func (a *API) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validation, ok := authn.ValidationFromHeaders(r.Header); ok && validation.IsValid {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), validationKey, model.TokenValidation{
				Identity: model.Identity{
					UserID:   validation.Identity.UserID,
					Role:     model.UserRole(validation.Identity.Role),
					SchoolID: validation.Identity.SchoolID,
				},
				Status:   model.UserStatus(validation.Status),
				IsActive: validation.IsActive,
				IsValid:  validation.IsValid,
			})))
			return
		}

		token := bearerToken(r.Header.Get("Authorization"))
		validation, err := a.service.ValidateToken(r.Context(), token)
		if err != nil || !validation.IsValid {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), validationKey, validation)))
	})
}

func (a *API) withSuperAdmin(next http.Handler) http.Handler {
	return a.withAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validation, ok := validationFromContext(r.Context())
		if !ok || validation.Identity.Role != model.UserRoleSuperAdmin {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	}))
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func validationFromContext(ctx context.Context) (model.TokenValidation, bool) {
	value, ok := ctx.Value(validationKey).(model.TokenValidation)
	return value, ok
}
