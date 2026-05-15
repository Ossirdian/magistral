package v1

import (
	"context"
	"testing"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/ossirdian/magistral/iam/internal/model"
	svc "github.com/ossirdian/magistral/iam/internal/service"
	"github.com/ossirdian/magistral/platform/pkg/authn"
	"google.golang.org/grpc/codes"
)

type stubService struct {
	validate func(ctx context.Context, accessToken string) (model.TokenValidation, error)
}

func (s stubService) ValidateToken(ctx context.Context, accessToken string) (model.TokenValidation, error) {
	return s.validate(ctx, accessToken)
}

func TestCheckAllowsAuthorizedRequest(t *testing.T) {
	t.Parallel()

	schoolID := "school-1"
	api, err := New(stubService{
		validate: func(_ context.Context, accessToken string) (model.TokenValidation, error) {
			if accessToken != "token-123" {
				t.Fatalf("unexpected token: %s", accessToken)
			}

			return model.TokenValidation{
				Identity: model.Identity{
					UserID:   "user-1",
					Role:     model.UserRoleOwner,
					SchoolID: &schoolID,
				},
				Status:   model.UserStatusActive,
				IsActive: true,
				IsValid:  true,
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := api.Check(context.Background(), &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Headers: map[string]string{
						"authorization": "Bearer token-123",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if got := codes.Code(resp.GetStatus().GetCode()); got != codes.OK {
		t.Fatalf("status code = %v, want %v", got, codes.OK)
	}

	okResponse := resp.GetOkResponse()
	if okResponse == nil {
		t.Fatal("expected ok response")
	}

	headers := map[string]string{}
	for _, item := range okResponse.GetHeaders() {
		headers[item.GetHeader().GetKey()] = item.GetHeader().GetValue()
	}

	if headers[authn.HeaderUserID] != "user-1" {
		t.Fatalf("x-user-id = %q", headers[authn.HeaderUserID])
	}
	if headers[authn.HeaderUserRole] != string(model.UserRoleOwner) {
		t.Fatalf("x-user-role = %q", headers[authn.HeaderUserRole])
	}
	if headers[authn.HeaderSchoolID] != schoolID {
		t.Fatalf("x-school-id = %q", headers[authn.HeaderSchoolID])
	}
	if headers[authn.HeaderAuthenticated] != "true" {
		t.Fatalf("x-authenticated = %q", headers[authn.HeaderAuthenticated])
	}
	if len(okResponse.GetHeadersToRemove()) != 1 || okResponse.GetHeadersToRemove()[0] != "authorization" {
		t.Fatalf("headers_to_remove = %v", okResponse.GetHeadersToRemove())
	}
}

func TestCheckDeniesMissingCredentials(t *testing.T) {
	t.Parallel()

	api, err := New(stubService{
		validate: func(_ context.Context, _ string) (model.TokenValidation, error) {
			t.Fatal("ValidateToken should not be called")
			return model.TokenValidation{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := api.Check(context.Background(), &authv3.CheckRequest{})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if got := codes.Code(resp.GetStatus().GetCode()); got != codes.Unauthenticated {
		t.Fatalf("status code = %v, want %v", got, codes.Unauthenticated)
	}
	if resp.GetDeniedResponse() == nil {
		t.Fatal("expected denied response")
	}
}

func TestCheckAcceptsSessionUUIDFallback(t *testing.T) {
	t.Parallel()

	api, err := New(stubService{
		validate: func(_ context.Context, accessToken string) (model.TokenValidation, error) {
			if accessToken != "session-uuid-1" {
				t.Fatalf("unexpected fallback token: %s", accessToken)
			}

			return model.TokenValidation{
				Identity: model.Identity{
					UserID: "user-2",
					Role:   model.UserRoleAdmin,
				},
				Status:   model.UserStatusActive,
				IsActive: true,
				IsValid:  true,
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := api.Check(context.Background(), &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Headers: map[string]string{
						"session-uuid": "session-uuid-1",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if got := codes.Code(resp.GetStatus().GetCode()); got != codes.OK {
		t.Fatalf("status code = %v, want %v", got, codes.OK)
	}
}

func TestCheckDeniesInvalidToken(t *testing.T) {
	t.Parallel()

	api, err := New(stubService{
		validate: func(_ context.Context, _ string) (model.TokenValidation, error) {
			return model.TokenValidation{}, svc.ErrUnauthorized
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := api.Check(context.Background(), &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Headers: map[string]string{
						"authorization": "Bearer broken",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if got := codes.Code(resp.GetStatus().GetCode()); got != codes.Unauthenticated {
		t.Fatalf("status code = %v, want %v", got, codes.Unauthenticated)
	}
}
