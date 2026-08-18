'use strict';

// Fail fast: throwing here means a missing JWT_SECRET breaks the process at
// startup (require-time), not on the first request that needs it.
function required(name) {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} is required`);
  }
  return value;
}

module.exports = {
  port: process.env.PORT || '4000',
  jwtSecret: required('JWT_SECRET'),
};
