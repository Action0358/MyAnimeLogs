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

# マイグレーション実行（必須）
make migrate
# または
go run scripts/migrate.go

# アプリケーション起動
go run cmd/server/main.go
```

⚠️ **重要**: アプリケーション起動前に必ずマイグレーションを実行してください。

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

ファイルベースのマイグレーションシステムを使用：

```bash
# マイグレーション実行
make migrate

# 手動実行
go run scripts/migrate.go
```

**マイグレーションファイル**:
- `000_create_users_table.sql` - ユーザーテーブル
- `001_create_animes_table.sql` - アニメテーブル  
- `002_create_episodes_table.sql` - エピソードテーブル
- `003_create_user_favorites_table.sql` - お気に入りテーブル
- `004_create_watch_statuses_table.sql` - 視聴状況テーブル
- `005_create_user_ratings_table.sql` - 評価テーブル
- `006_create_watch_history_table.sql` - 視聴履歴テーブル
- `007_create_anime_genres_table.sql` - ジャンルテーブル

**特徴**:
- **トランザクション保護**: 各マイグレーションが原子的に実行
- **実行済み追跡**: 重複実行を自動検出・スキップ
- **冪等性**: 安全な重複実行（`CREATE TABLE IF NOT EXISTS`使用）
- **ロールバック**: エラー時の自動ロールバック
- **バージョン管理**: ファイルベースで変更履歴を管理

**安全性**:
- 破壊的な`DROP TABLE`操作を排除
- 各マイグレーションがトランザクション内で実行
- 失敗時の自動ロールバックでデータ整合性を保証

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

### デプロイ後のデータ同期

デプロイ後、AniList APIのレート制限を回避するため、アニメデータをデータベースに事前同期することを推奨します。

#### 自動同期（推奨）

GitHub Actionsで自動的にデータ同期が実行されます。以下のSecretsを設定してください：

```
ADMIN_EMAIL: 管理者のメールアドレス
ADMIN_PASSWORD: 管理者のパスワード
```

#### 手動同期

デプロイ後に手動でデータを同期する場合：

```bash
# Staging環境
ADMIN_EMAIL=your-email@example.com ADMIN_PASSWORD=your-password ./scripts/sync-staging-data.sh

# Production環境
ADMIN_EMAIL=your-email@example.com ADMIN_PASSWORD=your-password ./scripts/sync-production-data.sh
```

または、直接APIを呼び出す：

```bash
# 1. ログインしてトークンを取得
TOKEN=$(curl -s -X POST "https://myanimelogs-staging.fly.dev/api/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"your-email@example.com","password":"your-password"}' | \
  grep -o '"token":"[^"]*' | cut -d'"' -f4)

# 2. 現在のシーズンのアニメを同期
curl -X POST "https://myanimelogs-staging.fly.dev/api/admin/sync/season" \
  -H "Authorization: Bearer $TOKEN"

# 3. 人気アニメを同期
curl -X POST "https://myanimelogs-staging.fly.dev/api/admin/sync/popular" \
  -H "Authorization: Bearer $TOKEN"
```

**データ同期の効果**:
- AniList APIのレート制限エラーを回避
- 画像の表示速度向上（キャッシュから取得）
- ユーザー体験の改善