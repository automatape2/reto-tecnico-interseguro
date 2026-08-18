// Package handler implements the Vercel serverless function for
// POST /api/v1/auth/token: exchanges a pre-shared API key for a JWT.
// This is the same trade-off documented in go-api - a stand-in for a real
// identity provider, out of scope for this challenge.
package handler

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"interseguro/vercel-api/internal/authutil"
	"interseguro/vercel-api/internal/httpjson"
)

type tokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"tokenType"`
	ExpiresIn int    `json:"expiresIn"`
}

// Handler is the Vercel Go function entrypoint for this file.
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use POST")
		return
	}

	presharedKey := os.Getenv("GO_API_PRESHARED_KEY")
	apiKey := r.Header.Get("X-API-Key")
	if presharedKey == "" || apiKey == "" || apiKey != presharedKey {
		httpjson.WriteError(w, http.StatusUnauthorized, "INVALID_API_KEY", "invalid or missing X-API-Key header")
		return
	}

	expiryMinutes := 60
	if v := os.Getenv("JWT_EXPIRY_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			expiryMinutes = n
		}
	}
	expiry := time.Duration(expiryMinutes) * time.Minute

	token, err := authutil.MintToken(os.Getenv("JWT_SECRET"), "client", "client", expiry)
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to issue token")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, tokenResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int(expiry.Seconds()),
	})
}
