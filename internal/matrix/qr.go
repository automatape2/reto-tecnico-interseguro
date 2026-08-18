package matrix

import "math"

// QR computes the full QR factorization of an m x n matrix A using Givens
// rotations: A = Q * R, where Q is an m x m orthogonal matrix and R is an
// m x n upper trapezoidal matrix.
//
// Givens rotations zero out one sub-diagonal entry of A at a time by
// left-multiplying with a 2x2 rotation acting on a pair of rows. This is the
// classic "rotation-based" method for computing a QR factorization (as
// opposed to Householder reflections or Gram-Schmidt), and is what this
// service uses to satisfy the requirement to compute QR "mediante rotación
// de la matriz".
//
// "Full" QR (Q always square m x m) is used rather than the reduced/economy
// form, so Q is orthogonal and its dimensions are predictable regardless of
// whether the input is tall, wide, or square.
func QR(a Matrix) (q, r Matrix, err error) {
	if err := a.Validate(); err != nil {
		return nil, nil, err
	}

	m, n := a.Rows(), a.Cols()
	r = a.Clone()
	q = Identity(m)

	limit := n
	if m < limit {
		limit = m
	}

	// For each column j, eliminate the entries below the diagonal
	// bottom-up so every rotation's pivot (r[j][j]) reflects all
	// previously applied rotations for this column.
	for j := 0; j < limit; j++ {
		for i := m - 1; i > j; i-- {
			pivot, target := r[j][j], r[i][j]
			if math.Abs(target) <= Epsilon {
				continue // already zero: skip an unnecessary rotation
			}
			c, s, _ := givensRotation(pivot, target)
			applyRotationToRows(r, j, i, c, s)
			applyRotationToQCols(q, j, i, c, s)
		}
	}

	return q, r, nil
}

// givensRotation returns the cosine c and sine s of the rotation that zeroes
// b when applied to the pair (a, b): [[c, s], [-s, c]] * [a, b]^T = [r, 0]^T.
//
// Using math.Hypot to compute r = sqrt(a^2 + b^2) avoids the intermediate
// overflow/underflow that a naive sqrt(a*a+b*b) would suffer for very large
// or very small inputs, so c = a/r and s = b/r remain numerically stable.
func givensRotation(a, b float64) (c, s, r float64) {
	r = math.Hypot(a, b)
	if r == 0 {
		return 1, 0, 0
	}
	return a / r, b / r, r
}

// applyRotationToRows left-multiplies R by the Givens rotation for rows
// (j, i), updating both rows in place. Columns before j are already zero in
// both rows by construction (they were eliminated in earlier iterations of
// the outer loop), so only columns [j, n) need to be touched.
func applyRotationToRows(r Matrix, j, i int, c, s float64) {
	n := r.Cols()
	for k := j; k < n; k++ {
		rj, ri := r[j][k], r[i][k]
		r[j][k] = c*rj + s*ri
		r[i][k] = -s*rj + c*ri
	}
}

// applyRotationToQCols right-multiplies Q by the transpose of the same
// Givens rotation, accumulating Q = G1^T * G2^T * ... so that Q * R
// reconstructs the original matrix once all rotations have been applied.
func applyRotationToQCols(q Matrix, j, i int, c, s float64) {
	m := q.Rows()
	for p := 0; p < m; p++ {
		qj, qi := q[p][j], q[p][i]
		q[p][j] = c*qj + s*qi
		q[p][i] = -s*qj + c*qi
	}
}
