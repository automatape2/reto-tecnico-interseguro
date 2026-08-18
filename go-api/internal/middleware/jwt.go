// Package middleware provides Fiber middleware shared across handlers. The
// only middleware currently needed is JWT verification.
package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"interseguro/go-api/internal/dto"
)

// Claims is the JWT claim set used by this service, for both the
// client-facing tokens minted by /api/v1/auth/token and the short-lived
// service tokens this API mints for itself when calling the Node.js API.
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

// JWTProtected returns Fiber middleware that requires a valid
// "Authorization: Bearer <token>" header signed with secret. On success the
// parsed claims are stored in the request context under "claims".
func JWTProtected(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return writeUnauthorized(c, "missing or malformed Authorization header")
		}
		tokenString := strings.TrimPrefix(header, prefix)

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil || !token.Valid {
			return writeUnauthorized(c, "invalid or expired token")
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}

func writeUnauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
		Error: dto.ErrorBody{Code: "UNAUTHORIZED", Message: message},
	})
}
