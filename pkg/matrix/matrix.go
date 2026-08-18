// Package matrix provides a minimal dense-matrix type and the Givens-rotation
// QR factorization used by the QR endpoint. It has no HTTP/JSON knowledge so
// it can be tested and reasoned about in isolation.
package matrix

import "errors"

// Matrix is a dense matrix stored as rows of float64. The wire format for
// this API is already an array of arrays, so this representation avoids a
// conversion layer at every handler boundary. For very large matrices a flat
// []float64 with explicit dimensions would be more cache-friendly; that
// trade-off isn't worth the extra indirection at this challenge's scale.
type Matrix [][]float64

// ErrEmptyMatrix is returned when a matrix has zero rows or zero columns.
var ErrEmptyMatrix = errors.New("matrix must have at least one row and one column")

// ErrRaggedMatrix is returned when rows have differing lengths.
var ErrRaggedMatrix = errors.New("matrix rows must all have the same length")

// Epsilon is the tolerance used throughout this package when a value should
// be treated as zero (e.g. an already-zeroed sub-diagonal entry, or a
// negligible sine in a Givens rotation). QR factorization is done in
// floating point, so exact equality to zero is never the right test.
const Epsilon = 1e-9

// Rows returns the number of rows in m.
func (m Matrix) Rows() int {
	return len(m)
}

// Cols returns the number of columns in m, based on the first row. Callers
// should validate rectangularity first via IsRectangular.
func (m Matrix) Cols() int {
	if len(m) == 0 {
		return 0
	}
	return len(m[0])
}

// IsRectangular reports whether m is non-empty and every row has the same,
// non-zero length.
func (m Matrix) IsRectangular() bool {
	if len(m) == 0 || len(m[0]) == 0 {
		return false
	}
	cols := len(m[0])
	for _, row := range m {
		if len(row) != cols {
			return false
		}
	}
	return true
}

// Validate returns a descriptive error if m is not a well-formed,
// rectangular, non-empty matrix.
func (m Matrix) Validate() error {
	if len(m) == 0 || len(m[0]) == 0 {
		return ErrEmptyMatrix
	}
	cols := len(m[0])
	for _, row := range m {
		if len(row) != cols {
			return ErrRaggedMatrix
		}
	}
	return nil
}

// Clone returns a deep copy of m, so mutating the result never affects the
// original matrix.
func (m Matrix) Clone() Matrix {
	out := make(Matrix, len(m))
	for i, row := range m {
		out[i] = make([]float64, len(row))
		copy(out[i], row)
	}
	return out
}

// Identity returns the n x n identity matrix.
func Identity(n int) Matrix {
	out := make(Matrix, n)
	for i := range out {
		out[i] = make([]float64, n)
		out[i][i] = 1
	}
	return out
}
