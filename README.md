# MyAnimeLogs

アニメの視聴管理・記録を行う Web アプリケーション

## 概要

MyAnimeLogs は、アニメファンが視聴したアニメを記録・管理し、お気に入りのアニメを整理できる Web アプリケーションです。お気に入り管理、評価・レビューの投稿などの機能を提供します。

## 主要機能

- **ユーザー認証**: 新規登録・ログイン・プロフィール管理
- **アニメ検索**: タイトル・ジャンル・シーズンでの検索
- **お気に入り**: 視聴中・完了・予定・中断のステータス管理
- **評価・レビュー**: 5 段階評価とコメント投稿
- **データエクスポート**: JSON/CSV 形式でのデータ出力

## 技術スタック

### フロントエンド

- **Next.js 14** (App Router)
- **React 18**
- **TypeScript 5**
- **Tailwind CSS**
- **Storybook** (コンポーネント開発)
- **MSW** (モック)
- **Orval** (OpenAPI クライアント生成)

### バックエンド

- **Go 1.21**
- **Gin** (Web フレームワーク)
- **JWT** (認証)
- **bcrypt** (パスワードハッシュ化)

### インフラストラクチャ & 開発環境

- **Frontend**: Cloudflare Pages
- **Backend**: Fly.io
- **Database**: Supabase (PostgreSQL)
- **開発環境**: Docker & Docker Compose
- **API 仕様**: OpenAPI 3.0

## プロジェクト構成

```
myanimelogs/
├── backend/                 # Go APIサーバー
│   ├── cmd/server/         # アプリケーションエントリーポイント
│   ├── internal/           # プライベートコード
│   │   ├── config/         # 設定管理
│   │   ├── database/       # データベース接続
│   │   ├── handlers/       # HTTPハンドラー
│   │   ├── middleware/     # ミドルウェア
│   │   ├── models/         # データモデル
│   │   ├── router/         # ルーティング
│   │   ├── services/       # ビジネスロジック
│   │   └── utils/          # ユーティリティ
│   └── Dockerfile
├── frontend/               # Next.jsアプリケーション
│   ├── src/
│   │   ├── app/           # App Router (ページ)
│   │   ├── features/      # 機能別コンポーネント
│   │   ├── lib/           # ライブラリ・API
│   │   ├── hooks/         # カスタムフック
│   │   └── types/         # TypeScript型定義
│   └── Dockerfile
├── docs/                   # 設計ドキュメント
├── openapi/               # API仕様書
│   └── openapi.yaml
└── docker-compose.yml     # 開発環境設定
```

## 開発環境セットアップ

### 前提条件

- Docker & Docker Compose
- Git
- Node.js（`npm install` でルートの依存を取得）

[Supabase CLI](https://supabase.com/docs/guides/local-development/cli/getting-started) は `package.json` の依存として同梱しているため、**グローバルインストールは不要**です。Makefile が `node_modules/.bin/supabase` を使うので、版が `package.json` にピン留めされ、環境間で挙動が揃います（brew 等でグローバル導入すると版ズレの原因になります）。

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd myanimelogs
```

### 2. 環境変数の設定

各サービスの `env/` ディレクトリにテンプレート (`local.example` / `staging.example` / `production.example`) があります。ローカル開発では `local.example` をコピーして使います。

```bash
cp backend/env/local.example backend/.env.local
cp frontend/env/local.example frontend/.env.local
```

外部 API キー（`ANNICT_API_KEY` 等）は `backend/.env.local` に手動で記入してください。Supabase 関連はローカル CLI のデフォルト値なので変更不要です。

### 3. ローカル Supabase の起動 + anon key 反映

ルート Makefile に一括起動ターゲットを用意しています。

```bash
make local-up
```

これは内部で以下を行います:

1. `supabase start`（初回は Docker イメージの pull で数分かかります）
2. `supabase status -o env` から `ANON_KEY` を抽出して `frontend/.env.local` の `NEXT_PUBLIC_SUPABASE_ANON_KEY` を自動更新
3. `docker compose up -d` でアプリを起動

手動で行いたい場合（CLI は同梱版を使うため `npx supabase` で呼びます）:

```bash
npx supabase start              # 出力された "anon key" を控える
# frontend/.env.local の NEXT_PUBLIC_SUPABASE_ANON_KEY を書き換え
docker compose up -d
```

### 3-1. Docker Desktop 再起動後に backend が起動しない場合

Docker Desktop を終了すると Supabase の全コンテナが SIGKILL で停止します。この状態で `docker compose up -d` すると、backend が起動時マイグレーションで DB（`localhost:54322`）に接続できず、リトライ後 `exit status 1` で落ちます。

```
Failed to ping database (attempt 1/5): dial tcp ...:54322: connect: connection refused
```

この場合 `supabase start` は「already running」と誤認して復旧できません。次のターゲットを使ってください（既存コンテナをそのまま起動するため、DB のデータは保持されます）。

```bash
make local-recover
```

### 3-2. Google OAuth の設定（初回のみ）

本サービスの認証は **Google ソーシャルログインのみ**です。ローカルでログインを動かすには Google の OAuth クライアントが必要です。

1. [Google Cloud Console](https://console.cloud.google.com/apis/credentials) で「OAuth 2.0 クライアント ID」を作成（種類: ウェブアプリケーション）
2. 「承認済みのリダイレクト URI」に以下を登録

   | 環境 | リダイレクト URI |
   |---|---|
   | ローカル | `http://localhost:54321/auth/v1/callback` |
   | STG / 本番 | `https://<supabase-project-ref>.supabase.co/auth/v1/callback` |

3. 発行された client ID / client secret を**シェルの環境変数**として渡す（`supabase/config.toml` が `env(...)` で参照します）

```bash
export SUPABASE_AUTH_EXTERNAL_GOOGLE_CLIENT_ID="xxxxx.apps.googleusercontent.com"
export SUPABASE_AUTH_EXTERNAL_GOOGLE_SECRET="xxxxx"
make local-supabase-down && make local-up   # 設定反映のため Supabase を再作成
```

> **シークレットは絶対にコミットしないでください。** `supabase/config.toml` には `env(...)` 参照のみを記述しています。`.env.local` ではなくシェル環境変数（`~/.zshrc` や direnv 等）で渡してください。
>
> **`127.0.0.1` と `localhost` は別物として扱われます。** Supabase は既定で `api_url`（`http://127.0.0.1`）からコールバック URL を組み立てますが、Google の承認済みリダイレクト URI は完全一致判定のため、そのままでは `redirect_uri_mismatch` になります。`[auth.external.google]` の `redirect_uri` で `localhost` を明示して回避しています。
>
> 設定が反映されたかは以下で確認できます。
>
> ```bash
> curl -sI "http://localhost:54321/auth/v1/authorize?provider=google&redirect_to=http%3A%2F%2Flocalhost%3A3000%2Fauth%2Fcallback" | grep -i location
> ```
>
> STG / 本番では Supabase ダッシュボードの Authentication → Providers → Google に同じ値を設定します。

### 4. 主要なローカルサービス

| サービス | URL | 用途 |
|---|---|---|
| Frontend | http://localhost:3000 | Next.js dev サーバ |
| Backend API | http://localhost:8080 | Go / Gin API |
| Supabase API | http://localhost:54321 | Auth / REST / Storage |
| Supabase Studio | http://localhost:54323 | DB 管理 GUI |
| Mailpit（メールテスト） | http://localhost:54324 | Supabase が送信するメールの受信確認 |

### 5. 認証フローのローカル検証

認証手段は **Google ソーシャルログインのみ**です（メール＋パスワード認証は廃止しました）。事前に「3-2. Google OAuth の設定」を完了させてください。

1. http://localhost:3000/login を開く
2. 「Googleでログイン」を押す → Google の同意画面へリダイレクト
3. 同意後 `/auth/callback` に戻り、セッションが確立してトップページへ遷移する
4. 初回ログイン時はバックエンドが `users` 行を自動生成し、表示名に Google のプロフィール名が入る

メール確認やパスワードリセットのフローは存在しないため、Mailpit を使う場面は通常ありません。

> `supabase/config.toml` のセクション名は歴史的経緯で `[inbucket]` のままですが、現行 Supabase CLI が起動するのは Mailpit です。

### 6. 停止

```bash
make local-stop          # アプリ + Supabase を全停止
make local-down          # アプリのみ停止（Supabase は残す）
make local-supabase-down # Supabase のみ停止
```

## API 仕様

OpenAPI 3.0 形式で定義された API 仕様書は `openapi/openapi.yaml` にあります。

### 主要エンドポイント

> 認証は Supabase Auth が担当し、サインアップ・ログインはフロントエンドから Supabase に対して直接行います。バックエンド API は Supabase 発行の JWT を検証して認可します。

#### ユーザー

- `GET /api/users/me` - プロフィール取得
- `PUT /api/users/me` - プロフィール更新
- `GET /api/users/me/export` - データエクスポート

#### アニメ

- `GET /api/animes` - アニメ検索・一覧
- `GET /api/animes/{id}` - アニメ詳細
- `GET /api/animes/current` - 今期アニメ
- `GET /api/animes/popular` - 人気アニメ

#### お気に入り・評価

- `GET /api/users/me/favorites` - お気に入り一覧
- `PUT /api/users/me/favorites/{anime_id}` - お気に入り追加・更新
- `DELETE /api/users/me/favorites/{anime_id}` - お気に入り削除
- `PUT /api/animes/{id}/rating` - 評価投稿・更新
- `DELETE /api/animes/{id}/rating` - 評価削除

## 開発コマンド

### フロントエンド

```bash
# 開発サーバー起動
docker compose exec frontend npm run dev

# API クライアント生成
docker compose exec frontend npm run generate:api

# Storybook起動
docker compose exec frontend npm run storybook

# テスト実行
docker compose exec frontend npm test

# リント・フォーマット
docker compose exec frontend npm run lint
docker compose exec frontend npm run format
```

### バックエンド

```bash
# 開発サーバー起動
docker compose exec backend go run cmd/server/main.go

# テスト実行
docker compose exec backend go test ./...

# 依存関係更新
docker compose exec backend go mod tidy
```

## データベース

### 接続情報

- **Host**: localhost
- **Port**: 54322
- **Database**: postgres
- **管理ツール**: Supabase Studio (http://localhost:54323)

### 主要テーブル

- `users` - ユーザー情報
- `animes` - アニメ情報
- `user_favorites` - お気に入り管理
- `user_ratings` - 評価・レビュー

## デプロイ

本プロジェクトは、`main`ブランチを本番環境、`staging`ブランチをステージング環境として、それぞれのブランチへのプッシュをトリガーに自動でデプロイされます。

### 本番環境 (Production)

- **Frontend**: Cloudflare Pages (`main` ブランチ)
- **Backend**: Fly.io (`myanimelogs-production` アプリ)
- **Database**: Supabase (本番用プロジェクト)
- **URL**: `https://myanimelogs.com` (または設定した本番ドメイン)

### ステージング環境 (Staging)

- **Frontend**: Cloudflare Pages (`staging` ブランチ, 別プロジェクト)
- **Backend**: Fly.io (`myanimelogs-staging` アプリ)
- **Database**: Supabase (ステージング用プロジェクト)
- **URL**: `https://stg.myanimelogs.com`
- **アクセス制限**: Cloudflare Accessにより、許可されたユーザーのみ閲覧可能です。

## 貢献

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## ライセンス

MIT License

## 関連ドキュメント

- [バックエンド README](./backend/README.md)
- [フロントエンド README](./frontend/README.md)
- [API 仕様書](./openapi/openapi.yaml)
- [設計ドキュメント](./docs/)
