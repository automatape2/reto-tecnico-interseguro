'use strict';

const express = require('express');

const router = express.Router();

// GET /health - unauthenticated liveness check for container/orchestrator probes.
router.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok', service: 'node-api', time: new Date().toISOString() });
});

module.exports = router;
