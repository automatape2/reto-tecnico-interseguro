'use strict';

function isFiniteNumber(value) {
  return typeof value === 'number' && Number.isFinite(value);
}

// A valid matrix is a non-empty array of non-empty, equal-length rows of
// finite numbers.
function isValidMatrix(matrix) {
  if (!Array.isArray(matrix) || matrix.length === 0) return false;
  if (!Array.isArray(matrix[0]) || matrix[0].length === 0) return false;

  const cols = matrix[0].length;
  return matrix.every(
    (row) => Array.isArray(row) && row.length === cols && row.every(isFiniteNumber)
  );
}

// Returns an error message string if body is not a valid POST /api/v1/stats
// request, or null if it is valid.
function validateStatsRequestBody(body) {
  if (!body || typeof body !== 'object' || Array.isArray(body)) {
    return 'request body must be a JSON object';
  }

  const { matrices } = body;
  if (!matrices || typeof matrices !== 'object' || Array.isArray(matrices)) {
    return '"matrices" must be a non-null object mapping names to matrices';
  }

  const names = Object.keys(matrices);
  if (names.length === 0) {
    return '"matrices" must contain at least one matrix';
  }

  for (const name of names) {
    if (!isValidMatrix(matrices[name])) {
      return `matrix "${name}" must be a non-empty, rectangular array of arrays of finite numbers`;
    }
  }

  return null;
}

module.exports = { isFiniteNumber, isValidMatrix, validateStatsRequestBody };
