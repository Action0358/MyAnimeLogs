# MyAnimeLogs Backend

MyAnimeLogs バックエンド API - Go + PostgreSQL + JWT 認証

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
- `JWT_SECRET`: JWT 署名用秘密鍵（本番では 256bit 以上）
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

- JWT 認証（24 時間有効期限）
- bcrypt パスワードハッシュ化
- レート制限
- CORS 設定
- SQL インジェクション対策

## 📝 API 仕様

### 認証 API

- `POST /api/register` - 新規ユーザー登録
- `POST /api/login` - ログイン・JWT 取得
- `GET /api/users/me` - プロフィール取得
- `PUT /api/users/me` - プロフィール更新

### 統計・エクスポート API

- `GET /api/users/me/stats` - 視聴統計取得
- `GET /api/users/me/export` - データエクスポート

詳細な API 仕様は `../api-tests/test.http` を参照してください。

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

デプロイ後、AniList API のレート制限を回避するため、アニメデータをデータベースに事前同期することを推奨します。

#### 自動同期（推奨）

GitHub Actions で自動的にデータ同期が実行されます。以下の Secrets を設定してください：

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

または、直接 API を呼び出す：

```bash
# 1. ログインしてトークンを取得
TOKEN=$(curl -s -X POST "https://myanimelogs-staging.fly.dev/api/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"your-email@example.com","password":"your-password"}' | \
  jq -r '.token')

# 2. 現在のシーズンのアニメを同期
curl -X POST "https://myanimelogs-staging.fly.dev/api/admin/sync/season" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# 3. 人気アニメを同期
curl -X POST "https://myanimelogs-staging.fly.dev/api/admin/sync/popular" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

**データ同期の効果**:

- AniList API のレート制限エラーを回避
- 画像の表示速度向上（キャッシュから取得）
- ユーザー体験の改善

### Production 環境のデータベース問題のトラブルシューティング

#### 基本的な確認手順

```bash
# 1. エラーログを確認
flyctl logs -a myanimelogs-production -n | grep -E "(Failed|failed|error)"

# 2. データ同期を実行
echo "y" | ADMIN_EMAIL=your-email@example.com ADMIN_PASSWORD=your-password ./scripts/sync-production-data.sh

# 3. アニメ数を確認
curl -s "https://myanimelogs-production.fly.dev/api/animes" | jq '.meta.total'
```

#### よくある問題

**問題: データ同期が 0 件になる**

症状: `"Successfully synced 0 anime works"`

確認事項:

1. エラーログを確認: `flyctl logs -a myanimelogs-production -n | grep -E "(Failed|failed)"`
2. Annict API からデータ取得を確認: `curl -s "https://myanimelogs-production.fly.dev/api/annict/works" | jq '.works | length'`

解決方法:

- エラーメッセージに基づいて適切なマイグレーションを作成
- 例: `pq: null value in column "id"` → シーケンスの問題
- 例: `ON CONFLICT` エラー → UNIQUE 制約の問題

**問題: マイグレーションが実行されない**

```bash
# アプリを再起動
flyctl apps restart myanimelogs-production

# ログを確認
flyctl logs -a myanimelogs-production -n | grep -i migration
```

#### マイグレーション履歴のリセット（緊急時のみ）

⚠️ **注意: 通常の運用では使用しないでください**

マイグレーションファイルを修正した後、改善版を適用する必要がある場合のみ使用:

```bash
# スクリプトを使用
./scripts/reset-migrations.sh

# または手動で
flyctl ssh console -a myanimelogs-production
# データベースでTRUNCATE TABLE migrations;を実行
# アプリを再起動してマイグレーションを再実行
```

詳細は `scripts/reset-migrations.sh` のコメントを参照してください。


## 🎬 Jikan API 連携

MyAnimeLogs は、エピソードの視聴時間データを取得するために Jikan API（MyAnimeList の非公式 API）を使用しています。

### 概要

- **目的**: エピソードごとの視聴時間（分）を取得し、ユーザーの総視聴時間を計算
- **データソース**: Jikan API v4 (https://api.jikan.moe/v4)
- **認証**: 不要（無料で利用可能）
- **レート制限**: 1 秒間に 3 リクエストまで

### 使用方法

#### 1. エピソード視聴時間の自動取得

エピソード情報を取得する際、視聴時間データが未取得の場合、バックグラウンドで自動的に Jikan API から取得されます。

```bash
# エピソード情報を取得（視聴時間が自動的に取得される）
GET /api/animes/{anime_id}/episodes
```

#### 2. 管理者による手動同期

管理者は、特定のアニメまたは全アニメのエピソード視聴時間を手動で同期できます。

```bash
# 特定のアニメの視聴時間を同期
curl -X POST "http://localhost:8080/api/admin/sync/episode-durations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"anime_id": 123}'

# 全アニメの視聴時間を同期
curl -X POST "http://localhost:8080/api/admin/sync/episode-durations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

レスポンス例：

```json
{
  "total_animes": 100,
  "success_count": 95,
  "failure_count": 3,
  "skipped_count": 2,
  "errors": [
    "Anime ID 456: MAL ID not found",
    "Anime ID 789: anime not found (404)"
  ]
}
```

### 技術仕様

#### レート制限

- **制限**: 1 秒間に 3 リクエスト
- **実装**: `golang.org/x/time/rate` パッケージを使用
- **リトライ**: 429 エラー時は指数バックオフでリトライ（最大 3 回）

#### エラーハンドリング

| エラーコード | 説明                 | 対応                                  |
| ------------ | -------------------- | ------------------------------------- |
| 404          | アニメが見つからない | ログに記録し、スキップ                |
| 429          | レート制限超過       | 指数バックオフでリトライ（最大 3 回） |
| 500          | サーバーエラー       | ログに記録し、スキップ                |
| Timeout      | タイムアウト         | リトライ（最大 3 回）                 |

#### データ検証

- 視聴時間は 1〜180 分の範囲でのみ保存
- 異常値の場合はログに記録し、保存しない

### 統計機能

視聴時間データを使用して、以下の統計情報を提供します：

```bash
# ユーザー統計を取得
GET /api/users/me/stats
```

レスポンス例：

```json
{
  "total_watched_episodes": 250,
  "total_watch_time_minutes": 6000,
  "completed_animes": 15,
  "monthly_watch_time": 480
}
```

### 注意事項

- Jikan API は非公式 API のため、将来的に仕様変更の可能性があります
- MAL ID が存在しないアニメは視聴時間データを取得できません
- エピソード数が多いアニメ（100 話以上）の場合、同期に時間がかかる可能性があります

### トラブルシューティング

#### 視聴時間が表示されない

1. MAL ID が設定されているか確認
2. 管理者 API で手動同期を実行
3. ログでエラーメッセージを確認

```bash
# ログを確認
flyctl logs -a myanimelogs-production | grep -i jikan
```

#### レート制限エラー

Jikan API のレート制限に達した場合、自動的にリトライされます。大量のアニメを同期する場合は、時間をおいて実行してください。

#### エピソードデータの同期確認

エピソードの `duration_minutes` や `air_date` が更新されない場合：

```bash
# 1. 特定のアニメで同期を実行
curl -X POST "http://localhost:8080/api/admin/sync/episode-durations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"anime_id": 1}'

# 2. データベースで確認
docker exec myanimelogs_postgres psql -U user -d myanimelogs \
  -c "SELECT id, episode_number, duration_minutes, air_date FROM episodes WHERE anime_id = 1;"

# 3. Jikan API から直接確認
curl -s "https://api.jikan.moe/v4/anime/{mal_anime_id}/episodes?page=1" | jq '.data[0]'
```
