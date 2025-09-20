# デプロイメント設定ガイド

## 環境変数の設定

### 必須環境変数

#### `NEXT_PUBLIC_API_URL`
- **説明**: バックエンドAPIのベースURL
- **開発環境**: `http://localhost:8080`
- **本番環境**: `https://api.yourdomain.com`

```bash
# 開発環境 (.env.local)
NEXT_PUBLIC_API_URL=http://localhost:8080

# 本番環境
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

## ホスティングプラットフォーム別設定

### Cloudflare Pages

1. **環境変数設定**
   - Cloudflare Pages ダッシュボード → Settings → Environment variables
   - `NEXT_PUBLIC_API_URL` を設定

2. **セキュリティヘッダー設定**
   - `public/_headers` ファイルの `connect-src` を更新:
   ```
   Content-Security-Policy: ...; connect-src 'self' https://api.yourdomain.com; ...
   ```

### Vercel

1. **環境変数設定**
   - Vercel ダッシュボード → Settings → Environment Variables
   - `NEXT_PUBLIC_API_URL` を設定

2. **セキュリティヘッダー設定**
   - `vercel.json` ファイルの `connect-src` を更新:
   ```json
   "value": "...; connect-src 'self' https://api.yourdomain.com; ..."
   ```

### Netlify

1. **環境変数設定**
   - Netlify ダッシュボード → Site settings → Environment variables
   - `NEXT_PUBLIC_API_URL` を設定

2. **セキュリティヘッダー設定**
   - `public/_headers.netlify` を `public/_headers` にリネーム
   - `connect-src` を更新:
   ```
   Content-Security-Policy: ...; connect-src 'self' https://api.yourdomain.com; ...
   ```

## セキュリティ設定チェックリスト

### 本番デプロイ前の確認事項

- [ ] `NEXT_PUBLIC_API_URL` が本番APIドメインに設定されている
- [ ] CSPの `connect-src` が本番APIドメインを含んでいる
- [ ] HTTPSが有効になっている
- [ ] セキュリティヘッダーが適切に設定されている

### CSP設定の確認方法

1. ブラウザの開発者ツールを開く
2. Network タブでAPIリクエストを確認
3. Console タブでCSPエラーがないことを確認

### トラブルシューティング

#### API呼び出しがブロックされる場合
- CSPの `connect-src` にAPIドメインが含まれているか確認
- 環境変数 `NEXT_PUBLIC_API_URL` が正しく設定されているか確認

#### セキュリティヘッダーが適用されない場合
- 静的エクスポート (`output: 'export'`) では Next.js の `headers` は無効
- ホスティングプラットフォーム側の設定ファイルを確認

## 例: 完全な本番設定

```bash
# 環境変数
NEXT_PUBLIC_API_URL=https://api.myanimelogs.com

# CSP設定例
connect-src 'self' https://api.myanimelogs.com
```