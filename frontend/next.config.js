/** @type {import('next').NextConfig} */
const nextConfig = {
  // 開発環境でのホットリロード最適化
  webpack: (config, { dev, isServer }) => {
    if (dev && !isServer) {
      // ファイル監視の最適化
      config.watchOptions = {
        poll: 1000,
        aggregateTimeout: 300,
        ignored: ['**/node_modules', '**/.git', '**/.next'],
      }
    }
    return config
  },

  // 開発サーバー設定
  ...(process.env.NODE_ENV === 'development' && {
    // Fast Refresh設定
    experimental: {
      // 高速リフレッシュの最適化
      optimizeCss: false,
      // TypeScript設定
      typedRoutes: true,
    },
  }),

  // 本番ビルド最適化
  ...(process.env.NODE_ENV === 'production' && {
    output: 'standalone',
    compress: true,
    poweredByHeader: false,
  }),

  // 環境変数設定
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY,
  },

  // 画像最適化
  images: {
    domains: ['localhost'],
    unoptimized: process.env.NODE_ENV === 'development',
  },

  // セキュリティヘッダー
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
        ],
      },
    ]
  },
}

module.exports = nextConfig
