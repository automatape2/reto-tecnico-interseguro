// Package authutil provides JWT minting and verification shared by the
// Vercel serverless functions under /api. Unlike go-api's middleware
// package, this has no Fiber dependency: Vercel's Go runtime uses plain
// net/http handlers, so each function calls VerifyToken/MintToken directly
// instead of going through a framework middleware chain.
package authutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT claim set used by both the client-facing tokens minted
// by /api/v1/auth/token and the short-lived service token the QR function
// mints when calling the stats function internally.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// MintToken creates and signs a JWT for the given subject/role pair, valid
// for expiry from now, using HS256 with secret.
func MintToken(secret, subject, role string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// VerifyToken parses and validates tokenString against secret, returning
// the parsed claims on success.
func VerifyToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}
