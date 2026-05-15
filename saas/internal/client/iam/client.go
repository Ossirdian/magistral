package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	svc "github.com/ossirdian/magistral/saas/internal/service"
)

type Client interface {
	CreateSchoolAdmin(ctx context.Context, schoolID, email, password string) (string, error)
}

type httpClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string, client *http.Client) Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" || client == nil {
		return nil
	}

	return &httpClient{baseURL: baseURL, client: client}
}

func (c *httpClient) CreateSchoolAdmin(ctx context.Context, schoolID, email, password string) (string, error) {
	requestPayload := struct {
		SchoolID *string `json:"school_id"`
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Role     string  `json:"role"`
		Status   string  `json:"status"`
	}{
		SchoolID: &schoolID,
		Email:    email,
		Password: password,
		Role:     "admin",
		Status:   "active",
	}
	body, err := json.Marshal(requestPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/admin/users", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Role", "super_admin")
	req.Header.Set("X-Authenticated", "true")
	req.Header.Set("X-User-Id", "saas-bootstrap")

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode == http.StatusBadRequest {
		return "", svc.ErrInvalidArgument
	}
	if res.StatusCode == http.StatusConflict {
		return "", fmt.Errorf("iam create user conflict")
	}
	if res.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return "", fmt.Errorf("iam create user request failed: status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var responsePayload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&responsePayload); err != nil {
		return "", err
	}
	if strings.TrimSpace(responsePayload.Data.ID) == "" {
		return "", errors.New("iam create user response does not contain id")
	}

	return responsePayload.Data.ID, nil
}
