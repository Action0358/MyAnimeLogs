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
- [Supabase CLI](https://supabase.com/docs/guides/local-development/cli/getting-started)（`brew install supabase/tap/supabase` または同梱の `npx supabase`）

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

手動で行いたい場合:

```bash
supabase start                  # 出力された "anon key" を控える
# frontend/.env.local の NEXT_PUBLIC_SUPABASE_ANON_KEY を書き換え
docker compose up -d
```

### 4. 主要なローカルサービス

| サービス | URL | 用途 |
|---|---|---|
| Frontend | http://localhost:3000 | Next.js dev サーバ |
| Backend API | http://localhost:8080 | Go / Gin API |
| Supabase API | http://localhost:54321 | Auth / REST / Storage |
| Supabase Studio | http://localhost:54323 | DB 管理 GUI |
| **Inbucket（メールテスト）** | **http://localhost:54324** | サインアップ確認・パスワードリセット等のメール受信 |

### 5. 認証フローのローカル検証

ローカル Supabase は `enable_confirmations = true` のため、本番同様に「サインアップ後はメール確認しないとログイン不可」の挙動になります。実際のメールは送信されず、すべて **Inbucket** で受信できます。

- **サインアップ**: 確認メールが Inbucket に届く → リンクを踏むとログイン可能になる
- **パスワードリセット**: リセット申請後にメールが届く → リンクから `/auth/update-password` に遷移し新パスワードを設定

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
