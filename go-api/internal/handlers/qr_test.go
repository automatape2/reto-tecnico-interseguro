package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/client"
	"interseguro/go-api/internal/dto"
	"interseguro/go-api/internal/matrix"
	"interseguro/go-api/internal/middleware"
)

const qrTestSecret = "qr-test-secret"

// stubStatsComputer is a test double for client.StatsComputer.
type stubStatsComputer struct {
	response *dto.StatisticsResponse
	err      error
}

func (s *stubStatsComputer) ComputeStatistics(_ context.Context, _, _ matrix.Matrix) (*dto.StatisticsResponse, error) {
	return s.response, s.err
}

func newQRTestApp(stats client.StatsComputer) *fiber.App {
	app := fiber.New()
	h := NewQRHandler(stats)
	app.Post("/api/v1/matrix/qr", middleware.JWTProtected(qrTestSecret), h.ComputeQR)
	return app
}

func validQRToken(t *testing.T) string {
	t.Helper()
	token, err := middleware.MintToken(qrTestSecret, "client", "client", time.Hour)
	if err != nil {
		t.Fatalf("MintToken() error = %v", err)
	}
	return token
}

func postQR(t *testing.T, app *fiber.App, body string, token string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/qr", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	return resp
}

func TestComputeQR_Success(t *testing.T) {
	stub := &stubStatsComputer{
		response: &dto.StatisticsResponse{
			PerMatrix: map[string]dto.MatrixStats{
				"Q": {Rows: 2, Cols: 2, Max: 1, Min: 0, Average: 0.5, Sum: 2, IsDiagonal: true},
				"R": {Rows: 2, Cols: 2, Max: 4, Min: 0, Average: 2.25, Sum: 9, IsDiagonal: false},
			},
			Overall: dto.OverallStats{Count: 8, Max: 4, Min: 0, Average: 1.375, Sum: 11, AnyDiagonal: true, DiagonalMatrices: []string{"Q"}},
		},
	}
	app := newQRTestApp(stub)

	resp := postQR(t, app, `{"matrix": [[4,3],[6,3]]}`, validQRToken(t))
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body dto.QRResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body.Matrices.Q) != 2 || len(body.Matrices.R) != 2 {
		t.Errorf("Matrices = %+v, want 2x2 Q and R", body.Matrices)
	}
	if !body.Statistics.Overall.AnyDiagonal {
		t.Errorf("Statistics.Overall.AnyDiagonal = false, want true")
	}
	if body.Meta.InputRows != 2 || body.Meta.InputCols != 2 {
		t.Errorf("Meta = %+v, want 2x2 input", body.Meta)
	}
}

func TestComputeQR_InvalidJSON(t *testing.T) {
	app := newQRTestApp(&stubStatsComputer{})
	resp := postQR(t, app, `{not valid json`, validQRToken(t))
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestComputeQR_JaggedMatrix(t *testing.T) {
	app := newQRTestApp(&stubStatsComputer{})
	resp := postQR(t, app, `{"matrix": [[1,2],[3]]}`, validQRToken(t))
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestComputeQR_EmptyMatrix(t *testing.T) {
	app := newQRTestApp(&stubStatsComputer{})
	resp := postQR(t, app, `{"matrix": []}`, validQRToken(t))
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestComputeQR_MissingToken(t *testing.T) {
	app := newQRTestApp(&stubStatsComputer{})
	resp := postQR(t, app, `{"matrix": [[1,2],[3,4]]}`, "")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestComputeQR_ExpiredToken(t *testing.T) {
	app := newQRTestApp(&stubStatsComputer{})
	expired, err := middleware.MintToken(qrTestSecret, "client", "client", -time.Hour)
	if err != nil {
		t.Fatalf("MintToken() error = %v", err)
	}
	resp := postQR(t, app, `{"matrix": [[1,2],[3,4]]}`, expired)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestComputeQR_UpstreamTimeout(t *testing.T) {
	app := newQRTestApp(&stubStatsComputer{err: client.ErrUpstreamTimeout})
	resp := postQR(t, app, `{"matrix": [[1,2],[3,4]]}`, validQRToken(t))
	if resp.StatusCode != fiber.StatusGatewayTimeout {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusGatewayTimeout)
	}
}

func TestComputeQR_UpstreamUnavailable(t *testing.T) {
	app := newQRTestApp(&stubStatsComputer{err: client.ErrUpstreamUnavailable})
	resp := postQR(t, app, `{"matrix": [[1,2],[3,4]]}`, validQRToken(t))
	if resp.StatusCode != fiber.StatusBadGateway {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusBadGateway)
	}
}
