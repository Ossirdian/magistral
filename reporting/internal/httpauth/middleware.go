package httpauth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ossirdian/magistral/platform/pkg/authn"
)

type contextKey string

const validationKey contextKey = "reporting.auth.validation"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/reports") {
			next.ServeHTTP(w, r)
			return
		}

		if validation, ok := authn.ValidationFromHeaders(r.Header); ok && validation.IsValid && validation.Identity.SchoolID != nil {
			switch validation.Identity.Role {
			case authn.RoleAdmin, authn.RoleOwner, authn.RoleInstructor:
				ctx := context.WithValue(r.Context(), validationKey, validation)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			default:
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func SchoolIDFromContext(ctx context.Context) (string, bool) {
	validation, ok := ctx.Value(validationKey).(authn.TokenValidation)
	if !ok || validation.Identity.SchoolID == nil || *validation.Identity.SchoolID == "" {
		return "", false
	}
	return *validation.Identity.SchoolID, true
}
