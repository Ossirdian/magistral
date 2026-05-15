package httpauth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ossirdian/magistral/platform/pkg/authn"
)

type contextKey string

const (
	validationKey contextKey = "document.auth.validation"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/documents") && !strings.HasPrefix(r.URL.Path, "/document-templates") {
			next.ServeHTTP(w, r)
			return
		}

		if validation, ok := authn.ValidationFromHeaders(r.Header); ok && validation.IsValid && validation.Identity.SchoolID != nil {
			switch validation.Identity.Role {
			case authn.RoleAdmin:
			case authn.RoleOwner, authn.RoleInstructor:
				if r.Method != http.MethodGet {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			default:
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), validationKey, validation)))
			return
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
