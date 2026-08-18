package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestApp(secret string) *fiber.App {
	app := fiber.New()
	app.Get("/protected", JWTProtected(secret), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func doRequest(t *testing.T, app *fiber.App, authHeader string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	return resp
}

func TestJWTProtected_ValidToken(t *testing.T) {
	app := newTestApp(testSecret)
	token, err := MintToken(testSecret, "client", "client", time.Hour)
	if err != nil {
		t.Fatalf("MintToken() error = %v", err)
	}

	resp := doRequest(t, app, "Bearer "+token)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

func TestJWTProtected_MissingHeader(t *testing.T) {
	app := newTestApp(testSecret)
	resp := doRequest(t, app, "")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestJWTProtected_MalformedHeader(t *testing.T) {
	app := newTestApp(testSecret)
	resp := doRequest(t, app, "NotBearer sometoken")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestJWTProtected_ExpiredToken(t *testing.T) {
	app := newTestApp(testSecret)
	token, err := MintToken(testSecret, "client", "client", -time.Hour)
	if err != nil {
		t.Fatalf("MintToken() error = %v", err)
	}

	resp := doRequest(t, app, "Bearer "+token)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestJWTProtected_WrongSecret(t *testing.T) {
	app := newTestApp(testSecret)
	token, err := MintToken("a-different-secret", "client", "client", time.Hour)
	if err != nil {
		t.Fatalf("MintToken() error = %v", err)
	}

	resp := doRequest(t, app, "Bearer "+token)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestJWTProtected_GarbageToken(t *testing.T) {
	app := newTestApp(testSecret)
	resp := doRequest(t, app, "Bearer not-a-real-jwt")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}
