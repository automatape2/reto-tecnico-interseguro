// Package httpjson provides tiny JSON response helpers shared by the
// Vercel Go functions under /api, which use plain net/http (no framework),
// so there's no built-in c.JSON()/writeError() to reuse.
package httpjson

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

// WriteJSON writes v as a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes the shared { "error": { "code", "message" } } envelope
// used by every non-2xx response across this API.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}
