'use strict';

const jwt = require('jsonwebtoken');
const config = require('../config');

// requireAuth verifies the shared-secret JWT minted by go-api (either the
// client-facing token or go-api's own short-lived service token). Claims
// are attached to req.claims but not used for role-based access control -
// that's flagged as a documented simplification, ready to extend later.
function requireAuth(req, res, next) {
  const header = req.get('Authorization') || '';
  const [scheme, token] = header.split(' ');

  if (scheme !== 'Bearer' || !token) {
    return res
      .status(401)
      .json({ error: { code: 'UNAUTHORIZED', message: 'missing or malformed Authorization header' } });
  }

  try {
    req.claims = jwt.verify(token, config.jwtSecret, { algorithms: ['HS256'] });
    return next();
  } catch (err) {
    return res.status(401).json({ error: { code: 'UNAUTHORIZED', message: 'invalid or expired token' } });
  }
}

module.exports = { requireAuth };
