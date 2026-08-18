// Package dto holds the JSON request/response shapes shared by the HTTP
// handlers and the outbound client to the Node.js statistics API, so both
// sides of that contract are defined in exactly one place.
package dto

// MatrixRequest is the body of POST /api/v1/matrix/qr.
type MatrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

// MatricesQR names the two matrices produced by the QR factorization.
type MatricesQR struct {
	Q [][]float64 `json:"Q"`
	R [][]float64 `json:"R"`
}

// Meta carries small, non-essential metadata about how a request was
// processed.
type Meta struct {
	InputRows    int   `json:"inputRows"`
	InputCols    int   `json:"inputCols"`
	ComputedAtMs int64 `json:"computedAtMs"`
}

// QRResponse is the body returned by POST /api/v1/matrix/qr: the QR
// factorization plus the statistics computed by the Node.js API over Q and
// R, embedded verbatim so there is a single source of truth for that shape.
type QRResponse struct {
	Matrices   MatricesQR         `json:"matrices"`
	Statistics StatisticsResponse `json:"statistics"`
	Meta       Meta               `json:"meta"`
}

// StatsRequest is the body sent to the Node.js API's POST /api/v1/stats.
// Matrices is name -> matrix so the callee can report diagonality per named
// matrix rather than by positional index.
type StatsRequest struct {
	Matrices map[string][][]float64 `json:"matrices"`
}

// MatrixStats mirrors the per-matrix statistics object returned by the
// Node.js API.
type MatrixStats struct {
	Rows       int     `json:"rows"`
	Cols       int     `json:"cols"`
	Max        float64 `json:"max"`
	Min        float64 `json:"min"`
	Average    float64 `json:"average"`
	Sum        float64 `json:"sum"`
	IsDiagonal bool    `json:"isDiagonal"`
}

// OverallStats mirrors the pooled-across-all-matrices statistics object
// returned by the Node.js API, including a direct answer to "is any of the
// matrices diagonal".
type OverallStats struct {
	Count            int      `json:"count"`
	Max              float64  `json:"max"`
	Min              float64  `json:"min"`
	Average          float64  `json:"average"`
	Sum              float64  `json:"sum"`
	AnyDiagonal      bool     `json:"anyDiagonal"`
	DiagonalMatrices []string `json:"diagonalMatrices"`
}

// StatisticsResponse is the body returned by the Node.js API's
// POST /api/v1/stats, and is re-used verbatim as the "statistics" field of
// QRResponse.
type StatisticsResponse struct {
	PerMatrix map[string]MatrixStats `json:"perMatrix"`
	Overall   OverallStats           `json:"overall"`
}

// TokenResponse is the body returned by POST /api/v1/auth/token. There is no
// corresponding request DTO: the API key travels in the X-API-Key header,
// not the request body.
type TokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"tokenType"`
	ExpiresIn int    `json:"expiresIn"`
}

// ErrorBody is the shape of the "error" field in every error response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the top-level shape of every non-2xx JSON response
// returned by this API.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// HealthResponse is the body returned by GET /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}
