const { fixupPluginRules } = require('@eslint/compat')
const globals = require('globals')
const tseslint = require('typescript-eslint')
const nextPlugin = require('@next/eslint-plugin-next')
const reactHooks = require('eslint-plugin-react-hooks')
const prettierRecommended = require('eslint-plugin-prettier/recommended')

// Next 14 系のプラグインは configs が旧形式（plugins: ['@next/next']）のため、
// flat config には plugins とルールを取り出して組み立てる。
// また ESLint 9 で廃止された API（context.getAncestors など）を使うため、
// fixupPluginRules で包む。Next 15 へ上げたら外せる。
module.exports = tseslint.config(
  {
    ignores: [
      '.next/**',
      'out/**',
      'storybook-static/**',
      'coverage/**',
      'next-env.d.ts',
      // msw が生成する Service Worker
      'public/mockServiceWorker.js',
      // orval の生成物。手で直しても再生成で戻る
      'src/lib/api/generated/**',
    ],
  },
  {
    files: ['**/*.{js,jsx,mjs,cjs,ts,tsx}'],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: { ecmaFeatures: { jsx: true } },
      globals: { ...globals.browser, ...globals.node, ...globals.jest },
    },
    plugins: {
      '@typescript-eslint': tseslint.plugin,
      '@next/next': fixupPluginRules(nextPlugin),
      'react-hooks': reactHooks,
    },
    rules: {
      ...nextPlugin.configs.recommended.rules,
      ...nextPlugin.configs['core-web-vitals'].rules,
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
    },
  },
  prettierRecommended
)
