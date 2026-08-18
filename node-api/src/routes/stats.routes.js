'use strict';

const express = require('express');
const { requireAuth } = require('../middleware/auth.middleware');
const { postStats } = require('../controllers/stats.controller');

const router = express.Router();

router.post('/stats', requireAuth, postStats);

module.exports = router;
