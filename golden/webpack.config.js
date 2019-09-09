const commonBuilder = require('pulito');
const { resolve } = require('path')

module.exports = (env, argv) => {
  const config = commonBuilder(env, argv, __dirname);
  config.output.publicPath='/dist/';
  config.resolve = config.resolve || {};
  config.resolve.modules = [resolve(__dirname, 'node_modules')];

  return config;
}