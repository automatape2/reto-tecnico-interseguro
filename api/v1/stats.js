'use strict';

// Vercel serverless function for POST /api/v1/stats: computes statistics
// over the named matrices sent by the QR function. Reuses the exact same
// validation and stats logic as node-api (../../lib), adapted from Express
// middleware/controllers to Vercel's plain (req, res) function signature.
// Vercel's Node runtime already parses a JSON body into req.body and
// exposes res.status()/res.json(), so the reused logic needs no changes.

const jwt = require('jsonwebtoken');
const { validateStatsRequestBody } = require('../../lib/validators');
const { computeStatistics } = require('../../lib/stats.service');

function requireAuth(req, res) {
  const header = req.headers.authorization || '';
  const [scheme, token] = header.split(' ');

  if (scheme !== 'Bearer' || !token) {
    res.status(401).json({ error: { code: 'UNAUTHORIZED', message: 'missing or malformed Authorization header' } });
    return false;
  }

  try {
    jwt.verify(token, process.env.JWT_SECRET, { algorithms: ['HS256'] });
    return true;
  } catch (err) {
    res.status(401).json({ error: { code: 'UNAUTHORIZED', message: 'invalid or expired token' } });
    return false;
  }
}

module.exports = (req, res) => {
  if (req.method !== 'POST') {
    res.status(405).json({ error: { code: 'METHOD_NOT_ALLOWED', message: 'use POST' } });
    return;
  }

  if (!requireAuth(req, res)) return;

  const validationError = validateStatsRequestBody(req.body);
  if (validationError) {
    res.status(400).json({ error: { code: 'INVALID_INPUT', message: validationError } });
    return;
  }

  res.status(200).json(computeStatistics(req.body.matrices));
};
