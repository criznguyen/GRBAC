// gen-jwt generates an HS256 JWT for smoke/e2e tests and scripts.
// Env: JWT_SECRET (required), SUB (default: test-actor), TENANT_ID (optional), EXP_HOURS (default: 1).
// Output: JWT string to stdout.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type claims struct {
	jwt.RegisteredClaims
	Tenant string `json:"tenant,omitempty"`
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Fprintln(os.Stderr, "JWT_SECRET is required")
		os.Exit(1)
	}

	sub := os.Getenv("SUB")
	if sub == "" {
		sub = "test-actor"
	}

	tenantID := os.Getenv("TENANT_ID")

	expHours := 1
	if h := os.Getenv("EXP_HOURS"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 {
			expHours = n
		}
	}

	now := time.Now()
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Tenant: tenantID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign jwt: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(signed)
}
