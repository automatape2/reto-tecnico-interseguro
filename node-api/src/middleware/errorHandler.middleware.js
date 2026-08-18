'use strict';

// Catches malformed-JSON bodies (thrown by express.json()) and anything
// else unhandled, so every error response - not just validation errors -
// uses the shared { error: { code, message } } envelope.
// eslint-disable-next-line no-unused-vars
function errorHandler(err, req, res, next) {
  if (err.type === 'entity.parse.failed') {
    return res
      .status(400)
      .json({ error: { code: 'INVALID_JSON', message: 'request body must be valid JSON' } });
  }

  console.error(err);
  return res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'unexpected server error' } });
}

module.exports = errorHandler;
