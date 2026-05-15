package httpauth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ossirdian/magistral/platform/pkg/authn"
)

type contextKey string

const validationKey contextKey = "saas.auth.validation"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/schools") && !strings.HasPrefix(r.URL.Path, "/branches") && !strings.HasPrefix(r.URL.Path, "/admin/") {
			next.ServeHTTP(w, r)
			return
		}

		if validation, ok := authn.ValidationFromHeaders(r.Header); ok && validation.IsValid {
			if validation.Identity.Role == authn.RoleSuperAdmin {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), validationKey, validation)))
				return
			}

			if canReadSchoolBranches(r, validation) {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), validationKey, validation)))
				return
			}

			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func canReadSchoolBranches(r *http.Request, validation authn.TokenValidation) bool {
	if r.Method != http.MethodGet || r.URL.Path != "/branches" || validation.Identity.SchoolID == nil {
		return false
	}

	switch validation.Identity.Role {
	case authn.RoleAdmin, authn.RoleOwner, authn.RoleInstructor:
	default:
		return false
	}

	return strings.TrimSpace(r.URL.Query().Get("school_id")) == *validation.Identity.SchoolID
}
