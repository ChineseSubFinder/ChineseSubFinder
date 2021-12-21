const fs = require('fs');
const DotEnv = require('dotenv');

const localEnvFile = `.env.${process.env.NODE_ENV}.local`;
const envFile = `.env.${process.env.NODE_ENV}`;

module.exports = function () {
  let parseEnv = null;
  if (fs.existsSync(localEnvFile)) {
    parseEnv = DotEnv.config({ path: localEnvFile }).parsed;
  } else {
    parseEnv = DotEnv.config({ path: envFile }).parsed;
  }
  return parseEnv;
};
