//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "http://localhost:8080"

// APIClient is an HTTP client wrapper for the RBAC Admin API.
// Automatically adds Authorization Bearer token and X-Tenant-ID where needed.
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	token      string
	tenantID   string
}

// NewAPIClient creates an APIClient with the given token and optional tenant ID.
// BaseURL defaults to http://localhost:8080 if empty.
func NewAPIClient(token string, tenantID string) *APIClient {
	base := defaultBaseURL
	return &APIClient{
		BaseURL:    base,
		HTTPClient: &http.Client{},
		token:      token,
		tenantID:   tenantID,
	}
}

// WithBaseURL sets the base URL (e.g. httptest.Server.URL).
func (c *APIClient) WithBaseURL(url string) *APIClient {
	c.BaseURL = url
	return c
}

func (c *APIClient) do(method, path string, body any, tenantID string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	if tenantID != "" {
		req.Header.Set("X-Tenant-ID", tenantID)
	} else if c.tenantID != "" {
		req.Header.Set("X-Tenant-ID", c.tenantID)
	}

	return c.HTTPClient.Do(req)
}

// CreateTenant creates a tenant. Does not require X-Tenant-ID.
func (c *APIClient) CreateTenant(name string) (*http.Response, error) {
	return c.do(http.MethodPost, "/api/v1/tenants", map[string]string{"name": name}, "")
}

// GetTenant retrieves a tenant by ID. X-Tenant-ID must match the requested tenant (pass id).
func (c *APIClient) GetTenant(id string) (*http.Response, error) {
	return c.do(http.MethodGet, "/api/v1/tenants/"+id, nil, id)
}

// CreateRole creates a role in the current tenant.
func (c *APIClient) CreateRole(name, description, status string) (*http.Response, error) {
	body := map[string]string{"name": name}
	if description != "" {
		body["description"] = description
	}
	if status != "" {
		body["status"] = status
	}
	return c.do(http.MethodPost, "/api/v1/roles", body, c.tenantID)
}

// GetRole retrieves a role by ID.
func (c *APIClient) GetRole(id string) (*http.Response, error) {
	return c.do(http.MethodGet, "/api/v1/roles/"+id, nil, c.tenantID)
}

// ListRoles lists roles with optional limit and cursor.
func (c *APIClient) ListRoles(limit int, cursor string) (*http.Response, error) {
	path := "/api/v1/roles"
	if limit > 0 || cursor != "" {
		path += "?"
		if limit > 0 {
			path += fmt.Sprintf("limit=%d", limit)
		}
		if cursor != "" {
			if limit > 0 {
				path += "&"
			}
			path += "cursor=" + cursor
		}
	}
	return c.do(http.MethodGet, path, nil, c.tenantID)
}

// UpdateRole updates a role.
func (c *APIClient) UpdateRole(id, name, description, status string) (*http.Response, error) {
	body := map[string]string{"name": name}
	if description != "" {
		body["description"] = description
	}
	if status != "" {
		body["status"] = status
	}
	return c.do(http.MethodPut, "/api/v1/roles/"+id, body, c.tenantID)
}

// DeleteRole deletes a role.
func (c *APIClient) DeleteRole(id string) (*http.Response, error) {
	return c.do(http.MethodDelete, "/api/v1/roles/"+id, nil, c.tenantID)
}

// ListRolePermissions lists permissions for a role.
func (c *APIClient) ListRolePermissions(roleID string) (*http.Response, error) {
	return c.do(http.MethodGet, "/api/v1/roles/"+roleID+"/permissions", nil, c.tenantID)
}

// ReplaceRolePermissions replaces all permissions for a role.
func (c *APIClient) ReplaceRolePermissions(roleID string, permissions []string) (*http.Response, error) {
	return c.do(http.MethodPut, "/api/v1/roles/"+roleID+"/permissions", map[string][]string{"permissions": permissions}, c.tenantID)
}

// AddRolePermissions adds permissions to a role.
func (c *APIClient) AddRolePermissions(roleID string, permissions []string) (*http.Response, error) {
	return c.do(http.MethodPatch, "/api/v1/roles/"+roleID+"/permissions", map[string][]string{"permissions": permissions}, c.tenantID)
}

// RemoveRolePermissions removes permissions from a role.
func (c *APIClient) RemoveRolePermissions(roleID string, permissions []string) (*http.Response, error) {
	return c.do(http.MethodDelete, "/api/v1/roles/"+roleID+"/permissions", map[string][]string{"permissions": permissions}, c.tenantID)
}

// Health checks the /health endpoint (no auth).
func (c *APIClient) Health() (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/health", nil)
	if err != nil {
		return nil, err
	}
	return c.HTTPClient.Do(req)
}

// DecodeJSON decodes the response body into v.
func DecodeJSON(resp *http.Response, v any) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}
