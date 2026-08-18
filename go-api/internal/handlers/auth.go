package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/dto"
	"interseguro/go-api/internal/middleware"
)

// AuthHandler issues client-facing JWTs in exchange for a pre-shared API
// key. This stands in for a real identity provider, which is out of scope
// for this challenge; the endpoint exists so the QR endpoint can still
// demonstrate JWT-protected access end to end.
type AuthHandler struct {
	presharedKey string
	jwtSecret    string
	jwtExpiry    time.Duration
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(presharedKey, jwtSecret string, jwtExpiry time.Duration) *AuthHandler {
	return &AuthHandler{presharedKey: presharedKey, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

// IssueToken handles POST /api/v1/auth/token.
func (h *AuthHandler) IssueToken(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" || apiKey != h.presharedKey {
		return writeError(c, fiber.StatusUnauthorized, "INVALID_API_KEY", "invalid or missing X-API-Key header")
	}

	token, err := middleware.MintToken(h.jwtSecret, "client", "client", h.jwtExpiry)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to issue token")
	}

	return c.JSON(dto.TokenResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int(h.jwtExpiry.Seconds()),
	})
}
