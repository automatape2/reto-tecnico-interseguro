'use strict';

const app = require('./app');
const config = require('./config');

app.listen(config.port, () => {
  console.log(`node-api listening on :${config.port}`);
});
