'use strict';

const { validateStatsRequestBody } = require('../utils/validators');
const { computeStatistics } = require('../services/stats.service');

// POST /api/v1/stats
function postStats(req, res) {
  const validationError = validateStatsRequestBody(req.body);
  if (validationError) {
    return res.status(400).json({ error: { code: 'INVALID_INPUT', message: validationError } });
  }

  const { matrices } = req.body;
  return res.status(200).json(computeStatistics(matrices));
}

module.exports = { postStats };
