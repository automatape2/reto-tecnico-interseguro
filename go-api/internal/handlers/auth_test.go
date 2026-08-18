package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"interseguro/go-api/internal/dto"
	"interseguro/go-api/internal/middleware"
)

const (
	authTestPresharedKey = "test-preshared-key"
	authTestJWTSecret    = "test-jwt-secret"
)

func newAuthTestApp() *fiber.App {
	app := fiber.New()
	h := NewAuthHandler(authTestPresharedKey, authTestJWTSecret, time.Hour)
	app.Post("/api/v1/auth/token", h.IssueToken)
	return app
}

func TestIssueToken_ValidKey(t *testing.T) {
	app := newAuthTestApp()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", nil)
	req.Header.Set("X-API-Key", authTestPresharedKey)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body dto.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Token == "" {
		t.Error("expected a non-empty token")
	}
	if body.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", body.TokenType, "Bearer")
	}
	if body.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", body.ExpiresIn)
	}

	claims := &middleware.Claims{}
	_, err = jwt.ParseWithClaims(body.Token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(authTestJWTSecret), nil
	})
	if err != nil {
		t.Fatalf("issued token does not parse/verify: %v", err)
	}
	if claims.Subject != "client" || claims.Role != "client" {
		t.Errorf("claims = %+v, want subject/role = client", claims)
	}
}

func TestIssueToken_InvalidKey(t *testing.T) {
	app := newAuthTestApp()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", nil)
	req.Header.Set("X-API-Key", "wrong-key")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestIssueToken_MissingKey(t *testing.T) {
	app := newAuthTestApp()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}
