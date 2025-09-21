# デプロイメントガイド

## 環境別デプロイ設定

### ステージング環境
```bash
# Netlifyのビルド設定
Build command: npm run build:staging
Environment variables:
- NEXT_PUBLIC_API_URL=https://myanimelogs-staging.fly.dev
```

### 本番環境
```bash
# Netlifyのビルド設定
Build command: npm run build:production
Environment variables:
- NEXT_PUBLIC_API_URL=https://myanimelogs-production.fly.dev
```

## 手動設定の場合

### ステージング環境
1. `frontend/public/_headers.staging` を `frontend/public/_headers` にコピー
2. 環境変数 `NEXT_PUBLIC_API_URL=https://myanimelogs-staging.fly.dev` を設定
3. デプロイ

### 本番環境
1. `frontend/public/_headers.production` を `frontend/public/_headers` にコピー
2. 本番APIドメインを設定
3. 環境変数 `NEXT_PUBLIC_API_URL=https://your-production-domain.com` を設定
4. デプロイ

## 環境変数一覧

| 変数名 | ステージング | 本番 |
|--------|-------------|------|
| NEXT_PUBLIC_API_URL | https://myanimelogs-staging.fly.dev | https://your-production-domain.com |
| NODE_ENV | production | production |

## CSP設定について

各環境のCSP設定は `connect-src` ディレクティブで適切なAPIドメインを許可しています：
- ステージング: `https://myanimelogs-staging.fly.dev`
- 本番: 実際の本番APIドメイン

## トラブルシューティング

### CSPエラーが発生する場合
1. 環境変数 `NEXT_PUBLIC_API_URL` が正しく設定されているか確認
2. `_headers` ファイルの `connect-src` ディレクティブが正しいドメインを指しているか確認
3. ブラウザのキャッシュをクリア