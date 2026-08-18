'use strict';

// Tolerance used whenever a value should be treated as zero. The matrices
// this service receives come from a floating-point QR factorization, so
// off-diagonal entries that are mathematically zero can arrive as tiny
// non-zero noise (e.g. 1e-16) — exact equality would wrongly call every
// diagonal matrix "not diagonal".
const EPSILON = 1e-9;

function flatten(matrix) {
  return matrix.flat();
}

// A matrix is diagonal if every off-diagonal entry is (within EPSILON)
// zero. This is only meaningful for square matrices: a non-square matrix is
// never considered diagonal, by definition, rather than raising an error.
// A 1x1 matrix is trivially diagonal (no off-diagonal entries to check); a
// 0x0 matrix is treated as not diagonal, a deliberate simplification for
// predictability that is unreachable from the real go-api -> node-api flow
// (go-api always sends a non-empty Q).
function isDiagonal(matrix, epsilon = EPSILON) {
  const rows = matrix.length;
  if (rows === 0) return false;

  const cols = matrix[0].length;
  if (rows !== cols) return false;

  for (let i = 0; i < rows; i++) {
    for (let j = 0; j < cols; j++) {
      if (i !== j && Math.abs(matrix[i][j]) > epsilon) return false;
    }
  }
  return true;
}

// computeMatrixStats returns max/min/average/sum/isDiagonal for a single
// matrix. Math.max/min use a reduce rather than `Math.max(...values)` so a
// very large matrix can't blow the call stack via argument spreading.
function computeMatrixStats(matrix) {
  const rows = matrix.length;
  const cols = rows ? matrix[0].length : 0;
  const values = flatten(matrix);

  if (values.length === 0) {
    return { rows, cols, max: null, min: null, average: null, sum: 0, isDiagonal: false };
  }

  const sum = values.reduce((a, b) => a + b, 0);
  return {
    rows,
    cols,
    max: values.reduce((a, b) => Math.max(a, b)),
    min: values.reduce((a, b) => Math.min(a, b)),
    average: sum / values.length,
    sum,
    isDiagonal: isDiagonal(matrix),
  };
}

// computeOverallStats pools every value across all named matrices, and
// answers the spec's literal question - "is any of the matrices diagonal" -
// as a single anyDiagonal boolean plus the list of matrices that qualify.
// Diagonality itself is not pooled (it's a per-matrix structural property,
// meaningless once values are merged).
function computeOverallStats(matricesByName, perMatrix) {
  const allValues = Object.values(matricesByName).flatMap(flatten);
  const diagonalMatrices = Object.entries(perMatrix)
    .filter(([, stats]) => stats.isDiagonal)
    .map(([name]) => name);

  if (allValues.length === 0) {
    return {
      count: 0,
      max: null,
      min: null,
      average: null,
      sum: 0,
      anyDiagonal: diagonalMatrices.length > 0,
      diagonalMatrices,
    };
  }

  const sum = allValues.reduce((a, b) => a + b, 0);
  return {
    count: allValues.length,
    max: allValues.reduce((a, b) => Math.max(a, b)),
    min: allValues.reduce((a, b) => Math.min(a, b)),
    average: sum / allValues.length,
    sum,
    anyDiagonal: diagonalMatrices.length > 0,
    diagonalMatrices,
  };
}

// computeStatistics is the orchestrating entry point: given { name: matrix }
// pairs, it returns { perMatrix, overall } as documented in the API
// contract.
function computeStatistics(matricesByName) {
  const perMatrix = {};
  for (const [name, matrix] of Object.entries(matricesByName)) {
    perMatrix[name] = computeMatrixStats(matrix);
  }
  const overall = computeOverallStats(matricesByName, perMatrix);
  return { perMatrix, overall };
}

module.exports = {
  EPSILON,
  isDiagonal,
  computeMatrixStats,
  computeOverallStats,
  computeStatistics,
};
