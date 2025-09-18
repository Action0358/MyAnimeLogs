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
    output: 'export',
    compress: true,
    poweredByHeader: false,
  }),

  

  // 画像最適化 - セキュリティ強化版
  images: {
    remotePatterns: [
      // 開発環境用（localhost）
      {
        protocol: 'http',
        hostname: 'localhost',
      },
      // MyAnimeList CDN（Jikan API用）
      {
        protocol: 'https',
        hostname: 'cdn.myanimelist.net',
      },
      // 必要に応じて信頼できるドメインのみ追加
      // 現在は画像表示を無効化しているため、最小限の設定
    ],
    unoptimized: process.env.NODE_ENV === 'development',
    // 画像機能を一時的に制限（現在プレースホルダーのみ使用）
    dangerouslyAllowSVG: false,
    contentSecurityPolicy: "default-src 'self'; script-src 'none'; sandbox;",
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
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
          // 画像読み込み制限のためのCSP
          {
            key: 'Content-Security-Policy',
            value: "img-src 'self' data: https://cdn.myanimelist.net;",
          },
        ],
      },
    ]
  },
}

module.exports = nextConfig
