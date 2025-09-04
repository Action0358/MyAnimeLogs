const nextPlugin = require('@next/eslint-plugin-next');
const prettierConfig = require('eslint-config-prettier');

module.exports = [
  nextPlugin.configs.recommended,
  nextPlugin.configs['core-web-vitals'],
  prettierConfig,
];
