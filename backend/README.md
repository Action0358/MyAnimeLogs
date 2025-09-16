# MyAnimeLogs Backend

MyAnimeLogs バックエンドAPI - Go + PostgreSQL + JWT認証

## 🚀 セットアップ

### 1. 環境設定

環境設定ファイルをコピーして設定：

```bash
# 開発環境
cp .env.example .env.local

# テスト環境  
cp .env.example .env.test

# 本番環境
cp .env.example .env.production
```

### 2. 必要な環境変数

各環境ファイルで以下の値を設定してください：

#### 必須設定
- `POSTGRES_*`: データベース接続情報
- `JWT_SECRET`: JWT署名用秘密鍵（本番では256bit以上）
- `ANNICT_API_KEY`: Annict API キー

#### オプション設定
- `DB_MAX_OPEN_CONNS`: 最大接続数（デフォルト: 25）
- `RATE_LIMIT`: レート制限（デフォルト: 100req/min）

### 3. データベースセットアップ

```bash
# Docker Composeでデータベース起動
docker compose up -d postgres

# マイグレーション実行（アプリ起動時に自動実行）
go run cmd/server/main.go
```

### 4. アプリケーション起動

```bash
# 開発環境
ENVIRONMENT=development go run cmd/server/main.go

# テスト環境
ENVIRONMENT=test go run cmd/server/main.go

# 本番環境
ENVIRONMENT=production go run cmd/server/main.go
```

## 🧪 テスト

```bash
# APIテスト実行
cd ../api-tests
# VS Code REST Client で test.http を実行
```

## 📊 データベース

### 接続プール設定

環境別に最適化された接続プール設定：

- **Development**: MaxOpen=25, MaxIdle=5, Lifetime=5min
- **Test**: MaxOpen=50, MaxIdle=10, Lifetime=2min
- **Production**: MaxOpen=50, MaxIdle=10, Lifetime=10min

### マイグレーション

アプリケーション起動時に自動実行：
- users テーブル
- animes テーブル  
- episodes テーブル
- user_favorites テーブル
- watch_statuses テーブル
- user_ratings テーブル
- watch_history テーブル
- anime_genres テーブル

## 🔐 セキュリティ

- JWT認証（24時間有効期限）
- bcryptパスワードハッシュ化
- レート制限
- CORS設定
- SQL インジェクション対策

## 📝 API仕様

### 認証API
- `POST /api/register` - 新規ユーザー登録
- `POST /api/login` - ログイン・JWT取得
- `GET /api/users/me` - プロフィール取得
- `PUT /api/users/me` - プロフィール更新

### 統計・エクスポートAPI
- `GET /api/users/me/stats` - 視聴統計取得
- `GET /api/users/me/export` - データエクスポート

詳細なAPI仕様は `../api-tests/test.http` を参照してください。

## 🛠️ 開発

### ディレクトリ構造

```
backend/
├── cmd/server/          # アプリケーションエントリーポイント
├── internal/
│   ├── config/         # 設定管理
│   ├── database/       # データベース接続・マイグレーション
│   ├── handlers/       # HTTPハンドラー
│   ├── middleware/     # ミドルウェア
│   ├── models/         # データモデル
│   ├── router/         # ルーティング
│   ├── services/       # ビジネスロジック
│   └── utils/          # ユーティリティ
├── .env.example        # 環境設定テンプレート
└── README.md
```

### 依存関係

- **Web Framework**: Gin
- **Database**: PostgreSQL + database/sql
- **Authentication**: JWT (golang-jwt/jwt)
- **Password**: bcrypt
- **Configuration**: godotenv
- **External API**: Annict API

## 🚀 デプロイ

GitHub Actions による自動デプロイ設定済み：
- **Staging**: `staging` ブランチ → Fly.io staging
- **Production**: `main` ブランチ → Fly.io production

環境変数は各デプロイ環境で設定してください。