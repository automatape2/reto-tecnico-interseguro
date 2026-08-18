package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandler_ValidKey(t *testing.T) {
	os.Setenv("GO_API_PRESHARED_KEY", "test-key")
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("GO_API_PRESHARED_KEY")
	defer os.Unsetenv("JWT_SECRET")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", nil)
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"tokenType":"Bearer"`) {
		t.Errorf("body missing tokenType: %s", rec.Body.String())
	}
}

func TestHandler_InvalidKey(t *testing.T) {
	os.Setenv("GO_API_PRESHARED_KEY", "test-key")
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("GO_API_PRESHARED_KEY")
	defer os.Unsetenv("JWT_SECRET")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", nil)
	req.Header.Set("X-API-Key", "wrong")
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/token", nil)
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
