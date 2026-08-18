// Package handler implements the Vercel serverless function for
// POST /api/v1/matrix/qr: computes the QR factorization of the input
// matrix via Givens rotations, then calls this same deployment's
// /api/v1/stats function (the Node.js statistics API) and returns a
// combined response - the same orchestration flow as go-api's Fiber
// handler, adapted to Vercel's one-function-per-request model.
package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"interseguro/vercel-api/pkg/authutil"
	"interseguro/vercel-api/pkg/httpjson"
	"interseguro/vercel-api/pkg/matrix"
)

type matrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

const serviceTokenTTL = 60 * time.Second

// Handler is the Vercel Go function entrypoint for this file.
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use POST")
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if !isAuthorized(r, secret) {
		httpjson.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing, invalid, or expired token")
		return
	}

	var req matrixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_JSON", `request body must be valid JSON with a "matrix" field`)
		return
	}

	m := matrix.Matrix(req.Matrix)
	if err := validateMatrixInput(m); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	start := time.Now()
	q, rMat, err := matrix.QR(m)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	stats, err := callStats(r, secret, q, rMat)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadGateway, "UPSTREAM_ERROR", "failed to compute statistics: "+err.Error())
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"matrices":   map[string]interface{}{"Q": q, "R": rMat},
		"statistics": stats,
		"meta": map[string]interface{}{
			"inputRows":    m.Rows(),
			"inputCols":    m.Cols(),
			"computedAtMs": time.Since(start).Milliseconds(),
		},
	})
}

func isAuthorized(r *http.Request, secret string) bool {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	_, err := authutil.VerifyToken(secret, strings.TrimPrefix(header, "Bearer "))
	return err == nil
}

// validateMatrixInput checks that m is rectangular, non-empty, and contains
// only finite numbers - mirrors go-api's handler validation exactly.
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

// callStats calls this same Vercel deployment's /api/v1/stats function
// (the Node.js statistics API) over HTTPS, authenticating with a
// short-lived service JWT minted for this single call.
func callStats(r *http.Request, secret string, q, rMat matrix.Matrix) (interface{}, error) {
	base := os.Getenv("VERCEL_URL")
	scheme := "https"
	if base == "" {
		// Fallback for local `vercel dev`, where VERCEL_URL isn't set.
		base = r.Host
		scheme = "http"
	}

	token, err := authutil.MintToken(secret, "qr-function", "service", serviceTokenTTL)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"matrices": map[string]interface{}{"Q": q, "R": rMat},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, scheme+"://"+base+"/api/v1/stats", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading stats response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stats function returned %d: %s", resp.StatusCode, string(respBody))
	}

	var stats interface{}
	if err := json.Unmarshal(respBody, &stats); err != nil {
		return nil, fmt.Errorf("decoding stats response: %w", err)
	}
	return stats, nil
}
