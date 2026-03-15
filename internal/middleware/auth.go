// Package middleware provides HTTP middleware: Auth (Bearer JWT), tenant validation, etc.
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"

	"github.com/grbac/grbac/internal/config"
)

var (
	ErrMissingToken            = errors.New("missing or invalid bearer token")
	ErrNoJWTConfig             = errors.New("neither JWT_SECRET nor JWKS_URL is configured")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	contextKeySubjectID   contextKey = "subject_id"
	contextKeyTenantClaim contextKey = "tenant_claim"
)

// Claims holds JWT claims we extract (sub required, tenant optional).
type Claims struct {
	jwt.RegisteredClaims
	Tenant string `json:"tenant,omitempty"`
}

// AuthConfig holds JWT validation configuration.
type AuthConfig struct {
	JWTSecret       string // HS256 secret; empty if using JWKS
	JWKSURL         string // OIDC JWKS endpoint for RS256
	JWTClaimsIssuer string // Expected iss claim (optional)
}

// AuthConfigFrom returns AuthConfig from app config.
func AuthConfigFrom(cfg *config.Config) AuthConfig {
	return AuthConfig{
		JWTSecret:       cfg.JWTSecret,
		JWKSURL:         cfg.JWKSURL,
		JWTClaimsIssuer: cfg.JWTClaimsIssuer,
	}
}

// Auth returns middleware that extracts and validates Bearer JWT, then injects subject_id and tenant_claim into context.
// Uses JWT_SECRET (HS256) if set, else JWKS_URL (RS256). Returns 401 if token is missing or invalid.
func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	keyFunc, err := buildKeyFunc(cfg)
	if err != nil {
		slog.Error("auth middleware: failed to build JWT key func", "error", err)
		// Continue with nil keyFunc; validation will fail with 401
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearer(r)
			if err != nil {
				respondAuthError(w, r, "UNAUTHORIZED", "missing or invalid authorization header", http.StatusUnauthorized)
				return
			}

			if keyFunc == nil {
				respondAuthError(w, r, "UNAUTHORIZED", "JWT validation not configured", http.StatusUnauthorized)
				return
			}

			var claims Claims
			opts := []jwt.ParserOption{jwt.WithValidMethods([]string{"HS256", "RS256"})}
			if cfg.JWTClaimsIssuer != "" {
				opts = append(opts, jwt.WithIssuer(cfg.JWTClaimsIssuer))
			}
			token, err := jwt.ParseWithClaims(tokenStr, &claims, keyFunc, opts...)
			if err != nil {
				slog.Debug("auth: JWT parse failed", "error", err)
				respondAuthError(w, r, "UNAUTHORIZED", "invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				respondAuthError(w, r, "UNAUTHORIZED", "invalid token", http.StatusUnauthorized)
				return
			}

			sub := claims.Subject
			if sub == "" {
				respondAuthError(w, r, "UNAUTHORIZED", "missing subject claim", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, contextKeySubjectID, sub)
			if claims.Tenant != "" {
				ctx = context.WithValue(ctx, contextKeyTenantClaim, claims.Tenant)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SubjectID returns the subject ID from the request context (set by Auth middleware).
func SubjectID(ctx context.Context) string {
	v, ok := ctx.Value(contextKeySubjectID).(string)
	if !ok {
		return ""
	}
	return v
}

// TenantClaim returns the optional tenant claim from the JWT (set by Auth middleware).
func TenantClaim(ctx context.Context) string {
	v, ok := ctx.Value(contextKeyTenantClaim).(string)
	if !ok {
		return ""
	}
	return v
}

func extractBearer(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", ErrMissingToken
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", ErrMissingToken
	}
	token := strings.TrimSpace(h[len(prefix):])
	if token == "" {
		return "", ErrMissingToken
	}
	return token, nil
}

func buildKeyFunc(cfg AuthConfig) (jwt.Keyfunc, error) {
	if cfg.JWKSURL != "" {
		jwks, err := keyfunc.Get(cfg.JWKSURL, keyfunc.Options{})
		if err != nil {
			return nil, err
		}
		return jwks.Keyfunc, nil
	}
	if cfg.JWTSecret != "" {
		secret := []byte(cfg.JWTSecret)
		return func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrUnexpectedSigningMethod
			}
			return secret, nil
		}, nil
	}
	return nil, ErrNoJWTConfig
}

// respondAuthError writes an error response in api-spec format.
func respondAuthError(w http.ResponseWriter, r *http.Request, code, message string, status int) {
	reqID := middleware.GetReqID(r.Context())
	if reqID == "" {
		reqID = "unknown"
	}
	body := map[string]any{
		"error": map[string]string{
			"code":       code,
			"message":    message,
			"request_id": reqID,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to encode error response", "error", err)
	}
}
