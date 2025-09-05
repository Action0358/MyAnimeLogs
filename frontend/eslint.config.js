const nextPlugin = require('@next/eslint-plugin-next');
const prettierRecommended = require('eslint-plugin-prettier/recommended');

module.exports = [
  nextPlugin.configs.recommended,
  nextPlugin.configs['core-web-vitals'],
  prettierRecommended,
];
