/** @type {import('next').NextConfig} */

// セキュリティヘッダーの定義（本番環境ではホスティングプラットフォーム側で設定必要）
const getSecurityHeaders = () => {
  // API URLを環境変数から取得（デフォルトは localhost:8080）
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  
  return [
    {
      key: 'X-Frame-Options',
      value: 'DENY',
    },
    {
      key: 'X-Content-Type-Options',
      value: 'nosniff',
    },
    {
      key: 'Referrer-Policy',
      value: 'strict-origin-when-cross-origin',
    },
    {
      key: 'Content-Security-Policy',
      value: [
        "default-src 'self'",
        "script-src 'self' 'unsafe-eval' 'unsafe-inline'", // Next.jsに必要
        "style-src 'self' 'unsafe-inline'", // Tailwind CSSに必要
        "img-src 'self' data: https://cdn.myanimelist.net",
        "font-src 'self'",
        `connect-src 'self' ${apiUrl} ${apiUrl}/*`, // 環境変数からAPI URLを取得（全てのサブパスを許可）
        "frame-ancestors 'none'", // X-Frame-Optionsと同等
      ].join('; '),
    },
  ];
};

// 環境別設定の構築
const buildConfig = () => {
  const isDevelopment = process.env.NODE_ENV === 'development';
  const isProduction = process.env.NODE_ENV === 'production';

  const config = {
    // 開発環境でのホットリロード最適化
    webpack: (config, { dev, isServer }) => {
      if (dev && !isServer) {
        config.watchOptions = {
          poll: 1000,
          aggregateTimeout: 300,
          ignored: ['**/node_modules', '**/.git', '**/.next'],
        };
      }
      return config;
    },

    // 画像最適化設定
    images: {
      unoptimized: isProduction, // 静的エクスポートでは画像最適化を無効化
      remotePatterns: [
        {
          protocol: 'http',
          hostname: 'localhost',
        },
        // Anilist CDN（メイン画像ソース）
        {
          protocol: 'https',
          hostname: 's4.anilist.co',
        },
        {
          protocol: 'https',
          hostname: '*.anilist.co',
        },
        // MyAnimeList CDN
        {
          protocol: 'https',
          hostname: 'cdn.myanimelist.net',
        },
        // Twitter画像（プロフィール画像など）
        {
          protocol: 'https',
          hostname: 'pbs.twimg.com',
        },
        {
          protocol: 'https',
          hostname: '*.twimg.com',
        },
      ],
      dangerouslyAllowSVG: false,
    },
  };

  // 開発環境固有の設定
  if (isDevelopment) {
    config.experimental = {
      optimizeCss: false,
      typedRoutes: true,
    };

    // 開発環境でもセキュアなCSPを適用
    config.headers = async () => [
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
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
          {
            key: 'Content-Security-Policy',
            value: [
              "default-src 'self'",
              "script-src 'self' 'unsafe-eval' 'unsafe-inline'",
              "style-src 'self' 'unsafe-inline'",
              // 画像ソースを厳格に制限（必要なドメインのみ）
              "img-src 'self' data: blob: https://*.anilist.co https://cdn.myanimelist.net https://*.twimg.com",
              "font-src 'self'",
              "connect-src 'self' http://localhost:* ws://localhost:*",
              "frame-ancestors 'none'",
            ].join('; '),
          },
        ],
      },
    ];
  }

  // 本番環境固有の設定
  if (isProduction) {
    config.output = 'export';
    config.compress = true;
    config.poweredByHeader = false;
    config.trailingSlash = true;
    config.distDir = 'out';
    
    // 本番環境では静的エクスポートのためheadersは無効
    // 重要: ホスティングプラットフォーム側でセキュリティヘッダーを設定してください
    console.warn(
      '⚠️  本番環境では静的エクスポートのため、セキュリティヘッダーはNext.jsで設定できません。\n' +
      '   Cloudflare Pages、Vercel、Netlifyなどのホスティングプラットフォーム側で\n' +
      '   以下のセキュリティヘッダーを必ず設定してください:\n' +
      '   - X-Frame-Options: DENY\n' +
      '   - X-Content-Type-Options: nosniff\n' +
      '   - Referrer-Policy: strict-origin-when-cross-origin\n' +
      '   - Content-Security-Policy: (適切なCSP設定)\n'
    );
  }

  return config;
};

const nextConfig = buildConfig();

module.exports = nextConfig
