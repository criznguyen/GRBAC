//go:build integration

package integration

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// testClaims holds JWT claims for integration tests.
// Matches middleware.Claims: sub (required), tenant (optional).
type testClaims struct {
	jwt.RegisteredClaims
	Tenant string `json:"tenant,omitempty"`
}

// NewTestJWT generates an HS256 JWT for testing with sub and optional tenant claims.
// Expiry is 1 hour from now. Uses the given secret for signing.
func NewTestJWT(sub string, tenantID string, secret string) (string, error) {
	claims := testClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Tenant: tenantID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}
