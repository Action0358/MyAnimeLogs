# MyAnimeLogs

アニメの視聴管理・記録を行う Web アプリケーション

## 概要

MyAnimeLogs は、アニメファンが視聴したアニメを記録・管理し、お気に入りのアニメを整理できる Web アプリケーションです。視聴履歴の管理、評価・レビューの投稿、視聴統計の確認などの機能を提供します。

## 主要機能

- **ユーザー認証**: 新規登録・ログイン・プロフィール管理
- **アニメ検索**: タイトル・ジャンル・シーズンでの検索
- **視聴管理**: エピソード単位での視聴状態管理
- **お気に入り**: 視聴中・完了・予定・中断のステータス管理
- **評価・レビュー**: 5 段階評価とコメント投稿
- **視聴統計**: 視聴時間・完了作品数などの統計情報
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
- **PostgreSQL** (データベース)
- **JWT** (認証)
- **bcrypt** (パスワードハッシュ化)

### インフラ・開発環境

- **Docker & Docker Compose**
- **pgAdmin** (データベース管理)
- **OpenAPI 3.0** (API 仕様)

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
├── .kiro/                 # Kiro設定・仕様書
│   └── specs/
└── docker-compose.yml     # 開発環境設定
```

## 開発環境セットアップ

### 前提条件

- Docker & Docker Compose
- Git

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd myanimelogs
```

### 2. 環境変数の設定

```bash
# バックエンド
cp backend/.env.example backend/.env
# frontend/.env.example が存在する場合
cp frontend/.env.example frontend/.env
```

### 3. Docker 環境の起動

```bash
docker compose up -d
```

### 4. アプリケーションへのアクセス

- **フロントエンド**: http://localhost:3000
- **バックエンド API**: http://localhost:8080
- **pgAdmin**: http://localhost:5050

## API 仕様

OpenAPI 3.0 形式で定義された API 仕様書は `openapi/openapi.yaml` にあります。

### 主要エンドポイント

#### 認証

- `POST /api/register` - ユーザー登録
- `POST /api/login` - ログイン
- `GET /api/users/me` - プロフィール取得

#### アニメ

- `GET /api/animes` - アニメ検索・一覧
- `GET /api/animes/{id}` - アニメ詳細
- `GET /api/animes/{id}/episodes` - エピソード一覧

#### 視聴管理

- `GET /api/users/me/favorites` - お気に入り一覧
- `PUT /api/users/me/favorites/{anime_id}` - お気に入り追加・更新
- `PUT /api/episodes/{episode_id}/watch` - 視聴状態更新

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
- **Port**: 5432
- **Database**: myanimelogs
- **管理ツール**: pgAdmin (http://localhost:5050)

### 主要テーブル

- `users` - ユーザー情報
- `animes` - アニメ情報
- `episodes` - エピソード情報
- `favorites` - お気に入り管理
- `ratings` - 評価・レビュー
- `watch_history` - 視聴履歴

## デプロイ

### 本番環境

- **フロントエンド**: Vercel / Netlify
- **バックエンド**: Fly.io / Railway
- **データベース**: Supabase / Railway

### Docker 本番ビルド

```bash
# 本番用イメージビルド
docker compose -f docker-compose.prod.yml build

# 本番環境起動
docker compose -f docker-compose.prod.yml up -d
```

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
