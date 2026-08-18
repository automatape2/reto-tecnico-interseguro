package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"interseguro/go-api/internal/matrix"
)

var testQ = matrix.Matrix{{1, 0}, {0, 1}}
var testR = matrix.Matrix{{2, 3}, {0, 4}}

func TestNodeClient_ComputeStatistics_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got == "" {
			t.Errorf("expected an Authorization header, got none")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"perMatrix": {
				"Q": {"rows":2,"cols":2,"max":1,"min":0,"average":0.5,"sum":2,"isDiagonal":true},
				"R": {"rows":2,"cols":2,"max":4,"min":0,"average":2.25,"sum":9,"isDiagonal":false}
			},
			"overall": {"count":8,"max":4,"min":0,"average":1.375,"sum":11,"anyDiagonal":true,"diagonalMatrices":["Q"]}
		}`))
	}))
	defer server.Close()

	c := NewNodeClient(server.URL, "test-secret", time.Second)
	stats, err := c.ComputeStatistics(context.Background(), testQ, testR)
	if err != nil {
		t.Fatalf("ComputeStatistics() error = %v", err)
	}
	if !stats.Overall.AnyDiagonal {
		t.Errorf("Overall.AnyDiagonal = false, want true")
	}
	if stats.PerMatrix["Q"].IsDiagonal != true {
		t.Errorf("PerMatrix[Q].IsDiagonal = false, want true")
	}
}

func TestNodeClient_ComputeStatistics_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"code":"INTERNAL_ERROR","message":"boom"}}`))
	}))
	defer server.Close()

	c := NewNodeClient(server.URL, "test-secret", time.Second)
	_, err := c.ComputeStatistics(context.Background(), testQ, testR)
	if !errors.Is(err, ErrUpstreamError) {
		t.Errorf("error = %v, want wrapping ErrUpstreamError", err)
	}
}

func TestNodeClient_ComputeStatistics_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	c := NewNodeClient(server.URL, "test-secret", time.Second)
	_, err := c.ComputeStatistics(context.Background(), testQ, testR)
	if !errors.Is(err, ErrUpstreamInvalidResponse) {
		t.Errorf("error = %v, want wrapping ErrUpstreamInvalidResponse", err)
	}
}

func TestNodeClient_ComputeStatistics_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewNodeClient(server.URL, "test-secret", 5*time.Millisecond)
	_, err := c.ComputeStatistics(context.Background(), testQ, testR)
	if !errors.Is(err, ErrUpstreamTimeout) {
		t.Errorf("error = %v, want wrapping ErrUpstreamTimeout", err)
	}
}
