'use strict';

// Must be set before requiring ../src/app, since src/config throws at
// require-time if JWT_SECRET is missing (fail-fast startup validation).
process.env.JWT_SECRET = 'test-secret-do-not-use-in-prod';

const test = require('node:test');
const assert = require('node:assert/strict');
const request = require('supertest');

const app = require('../src/app');

test('GET /health returns 200 with service identity', async () => {
  const res = await request(app).get('/health');
  assert.equal(res.status, 200);
  assert.equal(res.body.status, 'ok');
  assert.equal(res.body.service, 'node-api');
  assert.ok(res.body.time);
});
