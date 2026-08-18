// Package client contains the outbound HTTP client this API uses to reach
// the Node.js statistics API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"interseguro/go-api/internal/dto"
	"interseguro/go-api/internal/matrix"
	"interseguro/go-api/internal/middleware"
)

// StatsComputer is the abstraction the QR handler depends on to obtain
// statistics for a Q/R pair. The real HTTP-backed NodeClient implements it;
// tests can substitute a stub.
type StatsComputer interface {
	ComputeStatistics(ctx context.Context, q, r matrix.Matrix) (*dto.StatisticsResponse, error)
}

// Sentinel errors so handlers can map a client failure to the right HTTP
// status without string-matching error messages. Wrap these with %w so
// errors.Is still works after fmt.Errorf adds context.
var (
	ErrUpstreamTimeout         = errors.New("node-api request timed out")
	ErrUpstreamUnavailable     = errors.New("node-api is unreachable")
	ErrUpstreamError           = errors.New("node-api returned an error response")
	ErrUpstreamInvalidResponse = errors.New("node-api returned an invalid response")
)

// serviceTokenTTL is intentionally short: this client mints a fresh service
// JWT for every call rather than caching/refreshing one, since HMAC signing
// is cheap and this avoids any token-lifecycle state.
const serviceTokenTTL = 60 * time.Second

// NodeClient calls the Node.js statistics API over HTTP.
type NodeClient struct {
	baseURL    string
	httpClient *http.Client
	jwtSecret  string
	timeout    time.Duration
}

// NewNodeClient constructs a NodeClient targeting baseURL (e.g.
// "http://node-api:4000"), using jwtSecret to mint service tokens and
// timeout as the per-request deadline.
func NewNodeClient(baseURL, jwtSecret string, timeout time.Duration) *NodeClient {
	return &NodeClient{
		baseURL:    baseURL,
		jwtSecret:  jwtSecret,
		timeout:    timeout,
		httpClient: &http.Client{},
	}
}

// ComputeStatistics sends Q and R to the Node.js API's POST /api/v1/stats
// and returns the parsed statistics.
func (c *NodeClient) ComputeStatistics(ctx context.Context, q, r matrix.Matrix) (*dto.StatisticsResponse, error) {
	// Derive an explicit deadline from ctx rather than relying solely on
	// http.Client.Timeout, so ctx.Err() below can reliably distinguish a
	// timeout from any other connection failure.
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	token, err := middleware.MintToken(c.jwtSecret, "go-api-service", "service", serviceTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("minting service token: %w", err)
	}

	reqBody, err := json.Marshal(dto.StatsRequest{
		Matrices: map[string][][]float64{"Q": q, "R": r},
	})
	if err != nil {
		return nil, fmt.Errorf("encoding stats request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/stats", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("building stats request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrUpstreamTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrUpstreamInvalidResponse, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status %d: %s", ErrUpstreamError, resp.StatusCode, string(respBody))
	}

	var stats dto.StatisticsResponse
	if err := json.Unmarshal(respBody, &stats); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstreamInvalidResponse, err)
	}

	return &stats, nil
}
