package v1

import (
	"context"
	"errors"
	"strconv"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/ossirdian/magistral/iam/internal/model"
	"github.com/ossirdian/magistral/platform/pkg/authn"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

type service interface {
	ValidateToken(ctx context.Context, accessToken string) (model.TokenValidation, error)
}

type api struct {
	authv3.UnimplementedAuthorizationServer
	service service
}

func New(service service) (*api, error) {
	if service == nil {
		return nil, errors.New("envoy auth service is nil")
	}

	return &api{service: service}, nil
}

func (a *api) Check(ctx context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	token := accessTokenFromRequest(req)
	if token == "" {
		return denied(codes.Unauthenticated, "missing credentials"), nil
	}

	validation, err := a.service.ValidateToken(ctx, token)
	if err != nil || !validation.IsValid || !validation.IsActive {
		return denied(codes.Unauthenticated, "unauthorized"), nil
	}

	return allowed(validation), nil
}

func accessTokenFromRequest(req *authv3.CheckRequest) string {
	httpRequest := req.GetAttributes().GetRequest().GetHttp()
	if httpRequest == nil {
		return ""
	}

	headers := httpRequest.GetHeaders()
	if token := bearerToken(headerValue(headers, "authorization")); token != "" {
		return token
	}

	for _, key := range []string{"session-uuid", "x-session-uuid"} {
		if value := strings.TrimSpace(headerValue(headers, key)); value != "" {
			return value
		}
	}

	return ""
}

func headerValue(headers map[string]string, key string) string {
	for headerKey, value := range headers {
		if strings.EqualFold(headerKey, key) {
			return value
		}
	}
	return ""
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

func allowed(validation model.TokenValidation) *authv3.CheckResponse {
	headers := []*corev3.HeaderValueOption{
		header(authn.HeaderAuthenticated, "true"),
		header(authn.HeaderUserID, validation.Identity.UserID),
		header(authn.HeaderUserRole, string(validation.Identity.Role)),
		header(authn.HeaderUserStatus, string(validation.Status)),
		header(authn.HeaderUserActive, strconv.FormatBool(validation.IsActive)),
	}
	if validation.Identity.SchoolID != nil && strings.TrimSpace(*validation.Identity.SchoolID) != "" {
		headers = append(headers, header(authn.HeaderSchoolID, *validation.Identity.SchoolID))
	}

	return &authv3.CheckResponse{
		Status: &statuspb.Status{Code: int32(codes.OK)},
		HttpResponse: &authv3.CheckResponse_OkResponse{
			OkResponse: &authv3.OkHttpResponse{
				Headers:         headers,
				HeadersToRemove: []string{"authorization"},
			},
		},
	}
}

func denied(code codes.Code, message string) *authv3.CheckResponse {
	return &authv3.CheckResponse{
		Status: &statuspb.Status{Code: int32(code), Message: message},
		HttpResponse: &authv3.CheckResponse_DeniedResponse{
			DeniedResponse: &authv3.DeniedHttpResponse{
				Status: &typev3.HttpStatus{Code: typev3.StatusCode_Unauthorized},
				Headers: []*corev3.HeaderValueOption{
					header("content-type", "application/json"),
					header("www-authenticate", "Bearer"),
				},
				Body: `{"error":{"code":"unauthorized","message":"unauthorized"}}`,
			},
		},
	}
}

func header(key, value string) *corev3.HeaderValueOption {
	return &corev3.HeaderValueOption{
		Header: &corev3.HeaderValue{
			Key:   key,
			Value: value,
		},
	}
}

var _ authv3.AuthorizationServer = (*api)(nil)
