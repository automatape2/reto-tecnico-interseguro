package matrix

import (
	"math"
	"testing"
)

const testTol = 1e-9

// assertOrthogonal fails the test unless Q^T * Q is (approximately) the
// identity matrix.
func assertOrthogonal(t *testing.T, q Matrix) {
	t.Helper()
	n := q.Rows()
	if q.Cols() != n {
		t.Fatalf("Q is not square: %dx%d", q.Rows(), q.Cols())
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			var dot float64
			for k := 0; k < n; k++ {
				dot += q[k][i] * q[k][j]
			}
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(dot-want) > testTol {
				t.Errorf("Q^T*Q[%d][%d] = %v, want %v (Q not orthogonal)", i, j, dot, want)
			}
		}
	}
}

// assertUpperTrapezoidal fails the test unless every entry below the
// diagonal of R is (approximately) zero.
func assertUpperTrapezoidal(t *testing.T, r Matrix) {
	t.Helper()
	for i := 0; i < r.Rows(); i++ {
		for j := 0; j < r.Cols() && j < i; j++ {
			if math.Abs(r[i][j]) > testTol {
				t.Errorf("R[%d][%d] = %v, want ~0 (R not upper trapezoidal)", i, j, r[i][j])
			}
		}
	}
}

// assertReconstructs fails the test unless Q * R (approximately) equals A.
func assertReconstructs(t *testing.T, a, q, r Matrix) {
	t.Helper()
	m, n := a.Rows(), a.Cols()
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			var sum float64
			for k := 0; k < q.Cols(); k++ {
				sum += q[i][k] * r[k][j]
			}
			if math.Abs(sum-a[i][j]) > testTol {
				t.Errorf("(Q*R)[%d][%d] = %v, want %v (does not reconstruct A)", i, j, sum, a[i][j])
			}
		}
	}
}

func TestQR(t *testing.T) {
	cases := []struct {
		name string
		a    Matrix
	}{
		{"square 2x2", Matrix{{4, 3}, {6, 3}}},
		{"square 3x3", Matrix{{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}}},
		{"tall 4x2", Matrix{{1, 2}, {3, 4}, {5, 6}, {7, 8}}},
		{"wide 2x4", Matrix{{1, 2, 3, 4}, {5, 6, 7, 8}}},
		{"1x1", Matrix{{7}}},
		{"pre-zero subdiagonal", Matrix{{2, 1}, {0, 3}}},
		{"zero column", Matrix{{0, 1}, {0, 2}, {0, 3}}},
		{"negative entries", Matrix{{-4, -3}, {6, -3}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, r, err := QR(tc.a)
			if err != nil {
				t.Fatalf("QR() returned unexpected error: %v", err)
			}
			if q.Rows() != tc.a.Rows() || q.Cols() != tc.a.Rows() {
				t.Errorf("Q dims = %dx%d, want %dx%d (square, m x m)", q.Rows(), q.Cols(), tc.a.Rows(), tc.a.Rows())
			}
			if r.Rows() != tc.a.Rows() || r.Cols() != tc.a.Cols() {
				t.Errorf("R dims = %dx%d, want %dx%d (same shape as A)", r.Rows(), r.Cols(), tc.a.Rows(), tc.a.Cols())
			}
			assertOrthogonal(t, q)
			assertUpperTrapezoidal(t, r)
			assertReconstructs(t, tc.a, q, r)
		})
	}
}

func TestQR_TallMatrixHasZeroTrailingRows(t *testing.T) {
	a := Matrix{{1, 2}, {3, 4}, {5, 6}, {7, 8}}
	_, r, err := QR(a)
	if err != nil {
		t.Fatalf("QR() returned unexpected error: %v", err)
	}
	// For a tall (m > n) input, the bottom (m - n) rows of R must be zero.
	for i := a.Cols(); i < a.Rows(); i++ {
		for j := 0; j < a.Cols(); j++ {
			if math.Abs(r[i][j]) > testTol {
				t.Errorf("R[%d][%d] = %v, want 0 (trailing row of tall R should be zero)", i, j, r[i][j])
			}
		}
	}
}

func TestQR_RejectsRaggedMatrix(t *testing.T) {
	_, _, err := QR(Matrix{{1, 2}, {3}})
	if err != ErrRaggedMatrix {
		t.Errorf("QR() on ragged matrix returned %v, want ErrRaggedMatrix", err)
	}
}

func TestQR_RejectsEmptyMatrix(t *testing.T) {
	_, _, err := QR(Matrix{})
	if err != ErrEmptyMatrix {
		t.Errorf("QR() on empty matrix returned %v, want ErrEmptyMatrix", err)
	}
}

func TestGivensRotation(t *testing.T) {
	cases := []struct{ a, b float64 }{
		{4, 3}, {-4, 3}, {4, -3}, {-4, -3}, {0, 5}, {5, 0}, {0, 0}, {1e10, 1e-10},
	}
	for _, tc := range cases {
		c, s, r := givensRotation(tc.a, tc.b)
		if math.Abs(c*c+s*s-1) > testTol {
			t.Errorf("givensRotation(%v, %v): c^2+s^2 = %v, want 1", tc.a, tc.b, c*c+s*s)
		}
		gotR := c*tc.a + s*tc.b
		if math.Abs(gotR-r) > testTol {
			t.Errorf("givensRotation(%v, %v): c*a+s*b = %v, want r = %v", tc.a, tc.b, gotR, r)
		}
		gotZero := -s*tc.a + c*tc.b
		if math.Abs(gotZero) > testTol {
			t.Errorf("givensRotation(%v, %v): -s*a+c*b = %v, want ~0", tc.a, tc.b, gotZero)
		}
	}
}
