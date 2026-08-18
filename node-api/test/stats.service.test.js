'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');

const {
  isDiagonal,
  computeMatrixStats,
  computeOverallStats,
  computeStatistics,
} = require('../src/services/stats.service');

test('computeMatrixStats: known values', () => {
  const stats = computeMatrixStats([
    [1, 2],
    [3, 4],
  ]);
  assert.equal(stats.rows, 2);
  assert.equal(stats.cols, 2);
  assert.equal(stats.max, 4);
  assert.equal(stats.min, 1);
  assert.equal(stats.sum, 10);
  assert.equal(stats.average, 2.5);
});

test('computeMatrixStats: negative and mixed values', () => {
  const stats = computeMatrixStats([[-5, 2, 0]]);
  assert.equal(stats.max, 2);
  assert.equal(stats.min, -5);
  assert.equal(stats.sum, -3);
  assert.equal(stats.average, -1);
});

test('computeMatrixStats: empty matrix yields nulls, not NaN', () => {
  const stats = computeMatrixStats([]);
  assert.equal(stats.rows, 0);
  assert.equal(stats.cols, 0);
  assert.equal(stats.max, null);
  assert.equal(stats.min, null);
  assert.equal(stats.average, null);
  assert.equal(stats.sum, 0);
  assert.equal(stats.isDiagonal, false);
});

test('isDiagonal: identity matrix is diagonal', () => {
  assert.equal(
    isDiagonal([
      [1, 0],
      [0, 1],
    ]),
    true
  );
});

test('isDiagonal: general diagonal matrix', () => {
  assert.equal(
    isDiagonal([
      [5, 0, 0],
      [0, -2, 0],
      [0, 0, 9],
    ]),
    true
  );
});

test('isDiagonal: non-diagonal square matrix is false', () => {
  assert.equal(
    isDiagonal([
      [1, 2],
      [0, 1],
    ]),
    false
  );
});

test('isDiagonal: non-square matrix is always false', () => {
  assert.equal(
    isDiagonal([
      [1, 0, 0],
      [0, 1, 0],
    ]),
    false
  );
});

test('isDiagonal: 1x1 matrix is trivially diagonal', () => {
  assert.equal(isDiagonal([[42]]), true);
});

test('isDiagonal: empty (0x0) matrix is not diagonal (documented simplification)', () => {
  assert.equal(isDiagonal([]), false);
});

test('isDiagonal: epsilon boundary - negligible noise still counts as diagonal', () => {
  assert.equal(
    isDiagonal([
      [1, 1e-12],
      [1e-12, 1],
    ]),
    true
  );
});

test('isDiagonal: epsilon boundary - values above tolerance are not diagonal', () => {
  assert.equal(
    isDiagonal([
      [1, 1e-6],
      [0, 1],
    ]),
    false
  );
});

test('computeOverallStats: pools values across all matrices', () => {
  const matrices = {
    Q: [
      [1, 0],
      [0, 1],
    ],
    R: [
      [2, 3],
      [0, 4],
    ],
  };
  const perMatrix = {
    Q: computeMatrixStats(matrices.Q),
    R: computeMatrixStats(matrices.R),
  };
  const overall = computeOverallStats(matrices, perMatrix);

  assert.equal(overall.count, 8);
  assert.equal(overall.max, 4);
  assert.equal(overall.min, 0);
  assert.equal(overall.sum, 11);
  assert.equal(overall.average, 11 / 8);
});

test('computeOverallStats: anyDiagonal and diagonalMatrices reflect per-matrix flags', () => {
  const matrices = {
    Q: [
      [1, 0],
      [0, 1],
    ],
    R: [
      [2, 3],
      [0, 4],
    ],
  };
  const perMatrix = {
    Q: computeMatrixStats(matrices.Q),
    R: computeMatrixStats(matrices.R),
  };
  const overall = computeOverallStats(matrices, perMatrix);

  assert.equal(overall.anyDiagonal, true);
  assert.deepEqual(overall.diagonalMatrices, ['Q']);
});

test('computeOverallStats: anyDiagonal is false when no matrix is diagonal', () => {
  const matrices = { A: [[1, 2], [3, 4]] };
  const perMatrix = { A: computeMatrixStats(matrices.A) };
  const overall = computeOverallStats(matrices, perMatrix);

  assert.equal(overall.anyDiagonal, false);
  assert.deepEqual(overall.diagonalMatrices, []);
});

test('computeStatistics: end-to-end shape for named matrices', () => {
  const result = computeStatistics({
    Q: [
      [1, 0],
      [0, 1],
    ],
    R: [
      [2, 3],
      [0, 4],
    ],
  });

  assert.ok(result.perMatrix.Q);
  assert.ok(result.perMatrix.R);
  assert.equal(result.perMatrix.Q.isDiagonal, true);
  assert.equal(result.perMatrix.R.isDiagonal, false);
  assert.equal(result.overall.anyDiagonal, true);
});
