'use strict';

const express = require('express');

const healthRoutes = require('./routes/health.routes');
const statsRoutes = require('./routes/stats.routes');
const errorHandler = require('./middleware/errorHandler.middleware');

// Builds and exports the Express app without binding it to a port, so tests
// can exercise it directly via supertest without opening a real socket.
const app = express();

app.use(express.json());

app.use('/', healthRoutes);
app.use('/api/v1', statsRoutes);

app.use(errorHandler);

module.exports = app;
