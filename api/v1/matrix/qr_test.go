package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"interseguro/vercel-api/pkg/authutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func validToken(t *testing.T) string {
	t.Helper()
	token, err := authutil.MintToken(testSecret, "client", "client", time.Hour)
	if err != nil {
		t.Fatalf("MintToken() error = %v", err)
	}
	return token
}

// TestHandler_Success exercises the real internal HTTP call to
// /api/v1/stats: a stub server stands in for the Node function, and
// VERCEL_URL is left unset so callStats falls back to r.Host (the same
// fallback path used by `vercel dev` locally), pointed at the stub via
// req.Host.
func TestHandler_Success(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got == "" {
			t.Errorf("expected an Authorization header on the internal stats call, got none")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"perMatrix":{"Q":{"isDiagonal":true}},"overall":{"anyDiagonal":true}}`))
	}))
	defer stub.Close()

	os.Setenv("JWT_SECRET", testSecret)
	os.Unsetenv("VERCEL_URL")
	t.Cleanup(func() { os.Unsetenv("JWT_SECRET") })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/qr", strings.NewReader(`{"matrix":[[4,3],[6,3]]}`))
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	req.Host = stub.Listener.Addr().String()

	rec := httptest.NewRecorder()
	Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"anyDiagonal":true`) {
		t.Errorf("response missing embedded statistics: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"matrices"`) {
		t.Errorf("response missing matrices: %s", rec.Body.String())
	}
}

// TestHandler_StatsFunctionError guards against the bug where a non-2xx
// response from the internal /api/v1/stats call (e.g. an auth failure) was
// silently decoded and returned to the client as if it were valid
// statistics, instead of surfacing as an error.
func TestHandler_StatsFunctionError(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}`))
	}))
	defer stub.Close()

	os.Setenv("JWT_SECRET", testSecret)
	os.Unsetenv("VERCEL_URL")
	t.Cleanup(func() { os.Unsetenv("JWT_SECRET") })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/qr", strings.NewReader(`{"matrix":[[4,3],[6,3]]}`))
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	req.Host = stub.Listener.Addr().String()

	rec := httptest.NewRecorder()
	Handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadGateway, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"matrices"`) {
		t.Errorf("a failed stats call must not return a 200-shaped body: %s", rec.Body.String())
	}
}

func TestHandler_MissingToken(t *testing.T) {
	os.Setenv("JWT_SECRET", testSecret)
	t.Cleanup(func() { os.Unsetenv("JWT_SECRET") })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/qr", strings.NewReader(`{"matrix":[[1,2],[3,4]]}`))
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_JaggedMatrix(t *testing.T) {
	os.Setenv("JWT_SECRET", testSecret)
	t.Cleanup(func() { os.Unsetenv("JWT_SECRET") })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/qr", strings.NewReader(`{"matrix":[[1,2],[3]]}`))
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	os.Setenv("JWT_SECRET", testSecret)
	t.Cleanup(func() { os.Unsetenv("JWT_SECRET") })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/qr", strings.NewReader(`{not valid`))
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
