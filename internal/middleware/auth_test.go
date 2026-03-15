package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractBearer(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{"missing", "", "", true},
		{"empty value", "Bearer ", "", true},
		{"wrong prefix", "Basic abc", "", true},
		{"no space", "Bearertoken", "", true},
		{"valid", "Bearer my-token-123", "my-token-123", false},
		{"valid with spaces", "Bearer  token ", "token", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("Authorization", tt.header)
			}
			got, err := extractBearer(r)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuth_MissingToken(t *testing.T) {
	cfg := AuthConfig{JWTSecret: "test-secret"}
	auth := Auth(cfg)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	auth(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "error")
	assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
}

func TestAuth_InvalidToken(t *testing.T) {
	cfg := AuthConfig{JWTSecret: "test-secret"}
	auth := Auth(cfg)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	auth(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid token")
}

func TestAuth_ValidToken(t *testing.T) {
	secret := "test-secret-for-hs256"
	cfg := AuthConfig{JWTSecret: secret}
	auth := Auth(cfg)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "user-123"},
		Tenant:           "tenant-456",
	})
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	var capturedSub, capturedTenant string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSub = SubjectID(r.Context())
		capturedTenant = TenantClaim(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	auth(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user-123", capturedSub)
	assert.Equal(t, "tenant-456", capturedTenant)
}

func TestAuth_ValidTokenNoTenantClaim(t *testing.T) {
	secret := "test-secret"
	cfg := AuthConfig{JWTSecret: secret}
	auth := Auth(cfg)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "user-only"},
	})
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	var capturedSub, capturedTenant string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSub = SubjectID(r.Context())
		capturedTenant = TenantClaim(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	auth(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user-only", capturedSub)
	assert.Empty(t, capturedTenant)
}

func TestAuth_MissingSubject(t *testing.T) {
	secret := "test-secret"
	cfg := AuthConfig{JWTSecret: secret}
	auth := Auth(cfg)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		Tenant:           "tenant-1",
	})
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	auth(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing subject")
}
