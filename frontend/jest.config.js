const nextJest = require('next/jest')

const createJestConfig = nextJest({
  // Next.jsアプリのパスを指定してテスト環境でnext.config.jsと.envファイルを読み込む
  dir: './',
})

// Jestに渡すカスタム設定を追加
const customJestConfig = {
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  testEnvironment: 'jest-environment-jsdom',
}

// next/jestが非同期のNext.js設定を読み込めるようにこの方法でエクスポート
module.exports = createJestConfig(customJestConfig)
