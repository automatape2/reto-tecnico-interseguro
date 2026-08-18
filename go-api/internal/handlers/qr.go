package handlers

import (
	"errors"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/client"
	"interseguro/go-api/internal/dto"
	"interseguro/go-api/internal/matrix"
)

// QRHandler handles POST /api/v1/matrix/qr: it computes the QR
// factorization of the input matrix via Givens rotations, then enriches the
// response with statistics obtained from the Node.js API over Q and R.
type QRHandler struct {
	stats client.StatsComputer
}

// NewQRHandler constructs a QRHandler. stats is injected as an interface so
// tests can substitute a stub for the real HTTP-backed NodeClient.
func NewQRHandler(stats client.StatsComputer) *QRHandler {
	return &QRHandler{stats: stats}
}

// ComputeQR handles POST /api/v1/matrix/qr.
func (h *QRHandler) ComputeQR(c *fiber.Ctx) error {
	var req dto.MatrixRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "INVALID_JSON", `request body must be valid JSON with a "matrix" field`)
	}

	m := matrix.Matrix(req.Matrix)
	if err := validateMatrixInput(m); err != nil {
		return writeError(c, fiber.StatusBadRequest, "INVALID_INPUT", err.Error())
	}

	start := time.Now()
	q, r, err := matrix.QR(m)
	if err != nil {
		// Defensive: validateMatrixInput already checked this, but QR
		// re-validates internally too, so surface any mismatch the same way.
		return writeError(c, fiber.StatusBadRequest, "INVALID_INPUT", err.Error())
	}

	stats, err := h.stats.ComputeStatistics(c.Context(), q, r)
	if err != nil {
		return mapStatsError(c, err)
	}

	return c.JSON(dto.QRResponse{
		Matrices:   dto.MatricesQR{Q: q, R: r},
		Statistics: *stats,
		Meta: dto.Meta{
			InputRows:    m.Rows(),
			InputCols:    m.Cols(),
			ComputedAtMs: time.Since(start).Milliseconds(),
		},
	})
}

// validateMatrixInput checks that m is rectangular, non-empty, and contains
// only finite numbers. Go's JSON decoder cannot itself produce NaN/Inf from
// valid input text, but the check is cheap and guards against a future
// change to how the body is parsed.
func validateMatrixInput(m matrix.Matrix) error {
	if err := m.Validate(); err != nil {
		return err
	}
	for _, row := range m {
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return errors.New("matrix entries must be finite numbers")
			}
		}
	}
	return nil
}

// mapStatsError translates a failure from the Node.js client into the
// appropriate HTTP status and error code, without leaking upstream error
// details the caller can't act on.
func mapStatsError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, client.ErrUpstreamTimeout):
		return writeError(c, fiber.StatusGatewayTimeout, "UPSTREAM_TIMEOUT", "node-api did not respond in time")
	case errors.Is(err, client.ErrUpstreamUnavailable):
		return writeError(c, fiber.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "node-api is unreachable")
	case errors.Is(err, client.ErrUpstreamInvalidResponse):
		return writeError(c, fiber.StatusBadGateway, "UPSTREAM_INVALID_RESPONSE", "node-api returned an invalid response")
	case errors.Is(err, client.ErrUpstreamError):
		return writeError(c, fiber.StatusBadGateway, "UPSTREAM_ERROR", "node-api returned an error response")
	default:
		return writeError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "unexpected error computing statistics")
	}
}
