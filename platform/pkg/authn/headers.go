package authn

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	HeaderUserID       = "x-user-id"
	HeaderUserRole     = "x-user-role"
	HeaderSchoolID     = "x-school-id"
	HeaderUserStatus   = "x-user-status"
	HeaderUserActive   = "x-user-active"
	HeaderAuthenticated = "x-authenticated"
)

func ValidationFromHeaders(headers http.Header) (TokenValidation, bool) {
	userID := strings.TrimSpace(headers.Get(HeaderUserID))
	role := Role(strings.TrimSpace(headers.Get(HeaderUserRole)))
	if userID == "" || role == "" {
		return TokenValidation{}, false
	}

	var schoolID *string
	if value := strings.TrimSpace(headers.Get(HeaderSchoolID)); value != "" {
		schoolID = &value
	}

	status := UserStatus(strings.TrimSpace(headers.Get(HeaderUserStatus)))
	if status == "" {
		status = UserStatusActive
	}

	isActive := true
	if value := strings.TrimSpace(headers.Get(HeaderUserActive)); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			isActive = parsed
		}
	}

	isValid := true
	if value := strings.TrimSpace(headers.Get(HeaderAuthenticated)); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			isValid = parsed
		}
	}

	return TokenValidation{
		Identity: Identity{
			UserID:   userID,
			Role:     role,
			SchoolID: schoolID,
		},
		Status:   status,
		IsActive: isActive,
		IsValid:  isValid,
	}, true
}
