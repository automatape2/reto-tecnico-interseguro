'use strict';

// Must be set before requiring ../src/app, since src/config throws at
// require-time if JWT_SECRET is missing (fail-fast startup validation).
const TEST_SECRET = 'test-secret-do-not-use-in-prod';
process.env.JWT_SECRET = TEST_SECRET;

const test = require('node:test');
const assert = require('node:assert/strict');
const request = require('supertest');
const jwt = require('jsonwebtoken');

const app = require('../src/app');

function validToken() {
  return jwt.sign({ sub: 'client', role: 'client' }, TEST_SECRET, { algorithm: 'HS256', expiresIn: '1h' });
}

function expiredToken() {
  return jwt.sign({ sub: 'client', role: 'client' }, TEST_SECRET, { algorithm: 'HS256', expiresIn: '-1h' });
}

const sampleBody = {
  matrices: {
    Q: [
      [1, 0],
      [0, 1],
    ],
    R: [
      [2, 3],
      [0, 4],
    ],
  },
};

test('POST /api/v1/stats: valid token and body returns 200 with expected shape', async () => {
  const res = await request(app)
    .post('/api/v1/stats')
    .set('Authorization', `Bearer ${validToken()}`)
    .send(sampleBody);

  assert.equal(res.status, 200);
  assert.equal(res.body.perMatrix.Q.isDiagonal, true);
  assert.equal(res.body.perMatrix.R.isDiagonal, false);
  assert.equal(res.body.overall.anyDiagonal, true);
  assert.deepEqual(res.body.overall.diagonalMatrices, ['Q']);
  assert.equal(res.body.overall.count, 8);
});

test('POST /api/v1/stats: missing Authorization header returns 401', async () => {
  const res = await request(app).post('/api/v1/stats').send(sampleBody);
  assert.equal(res.status, 401);
  assert.equal(res.body.error.code, 'UNAUTHORIZED');
});

test('POST /api/v1/stats: invalid token returns 401', async () => {
  const res = await request(app)
    .post('/api/v1/stats')
    .set('Authorization', 'Bearer not-a-real-token')
    .send(sampleBody);
  assert.equal(res.status, 401);
});

test('POST /api/v1/stats: expired token returns 401', async () => {
  const res = await request(app)
    .post('/api/v1/stats')
    .set('Authorization', `Bearer ${expiredToken()}`)
    .send(sampleBody);
  assert.equal(res.status, 401);
});

test('POST /api/v1/stats: missing matrices field returns 400', async () => {
  const res = await request(app)
    .post('/api/v1/stats')
    .set('Authorization', `Bearer ${validToken()}`)
    .send({});
  assert.equal(res.status, 400);
  assert.equal(res.body.error.code, 'INVALID_INPUT');
});

test('POST /api/v1/stats: jagged matrix returns 400', async () => {
  const res = await request(app)
    .post('/api/v1/stats')
    .set('Authorization', `Bearer ${validToken()}`)
    .send({ matrices: { A: [[1, 2], [3]] } });
  assert.equal(res.status, 400);
});

test('POST /api/v1/stats: non-numeric entries return 400', async () => {
  const res = await request(app)
    .post('/api/v1/stats')
    .set('Authorization', `Bearer ${validToken()}`)
    .send({ matrices: { A: [[1, 'x']] } });
  assert.equal(res.status, 400);
});

test('POST /api/v1/stats: malformed JSON body returns 400 INVALID_JSON', async () => {
  const res = await request(app)
    .post('/api/v1/stats')
    .set('Authorization', `Bearer ${validToken()}`)
    .set('Content-Type', 'application/json')
    .send('{not valid json');
  assert.equal(res.status, 400);
  assert.equal(res.body.error.code, 'INVALID_JSON');
});
