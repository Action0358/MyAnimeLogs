# MyAnimeLog Backend

MyAnimeLogのバックエンドAPIサーバー（Go + Gin）

## ディレクトリ構成

```
backend/
├── cmd/
│   └── server/          # アプリケーションエントリーポイント
│       └── main.go
├── internal/            # プライベートアプリケーションコード
│   ├── config/          # 設定管理
│   ├── database/        # データベース接続・マイグレーション
│   ├── handlers/        # HTTPハンドラー
│   ├── middleware/      # ミドルウェア
│   ├── models/          # データモデル・構造体
│   ├── router/          # ルーティング設定
│   ├── services/        # ビジネスロジック
│   └── utils/           # ユーティリティ関数
├── .env.example         # 環境変数テンプレート
├── Dockerfile           # Docker設定
├── go.mod              # Go modules
├── go.sum              # 依存関係ロックファイル
└── main.go             # 後方互換性のためのエントリーポイント
```

## 技術スタック

- **言語**: Go 1.21+
- **Webフレームワーク**: Gin
- **データベース**: PostgreSQL (Supabase)
- **認証**: JWT + bcrypt
- **ORM**: database/sql (標準ライブラリ)

## 主要機能

### 認証・ユーザー管理
- ユーザー登録・ログイン
- JWT認証
- プロフィール管理
- 視聴統計
- データエクスポート

### アニメ管理
- アニメ検索・一覧
- アニメ詳細情報
- エピソード管理
- お気に入り機能
- 視聴状態管理
- 評価・レビュー

### セキュリティ
- JWT認証
- パスワードハッシュ化（bcrypt）
- CORS設定
- レート制限
- 入力バリデーション

## 環境変数

`.env.example`を参考に`.env`ファイルを作成してください。

```bash
# データベース
DATABASE_URL=postgresql://user:password@localhost/myanimelog

# JWT設定
JWT_SECRET=your-256-bit-secret
JWT_EXPIRATION=24h

# 外部API
ANNICT_API_KEY=your-annict-api-key

# Redis（キャッシュ）
REDIS_URL=redis://localhost:6379

# サーバー設定
PORT=8080
ENVIRONMENT=development
RATE_LIMIT=5
```

## 開発環境セットアップ

1. 依存関係のインストール
```bash
go mod download
```

2. 環境変数の設定
```bash
cp .env.example .env
# .envファイルを編集
```

3. データベースの準備
- Supabaseプロジェクトを作成
- DATABASE_URLを設定

4. サーバー起動
```bash
# 開発用
go run cmd/server/main.go

# または従来のエントリーポイント
go run main.go
```

## API仕様

### 認証
- `POST /api/register` - ユーザー登録
- `POST /api/login` - ログイン
- `GET /api/users/me` - プロフィール取得
- `PUT /api/users/me` - プロフィール更新

### アニメ
- `GET /api/animes` - アニメ検索・一覧
- `GET /api/animes/:id` - アニメ詳細
- `GET /api/animes/:id/episodes` - エピソード一覧

### お気に入り
- `GET /api/users/me/favorites` - お気に入り一覧
- `PUT /api/users/me/favorites/:anime_id` - お気に入り追加・更新
- `DELETE /api/users/me/favorites/:anime_id` - お気に入り削除

### 視聴管理
- `PUT /api/episodes/:episode_id/watch` - 視聴状態更新
- `GET /api/users/me/recently-watched` - 最近の視聴履歴
- `GET /api/users/me/watchlist` - 視聴予定リスト

### 評価
- `PUT /api/animes/:id/rating` - 評価登録・更新
- `DELETE /api/animes/:id/rating` - 評価削除

## デプロイ

### Fly.io
```bash
# Fly.ioアプリ作成
fly apps create myanimelog-api

# デプロイ
fly deploy
```

### Docker
```bash
# イメージビルド
docker build -t myanimelog-backend .

# コンテナ実行
docker run -p 8080:8080 --env-file .env myanimelog-backend
```

## 開発ガイドライン

### コード構成
- `cmd/`: アプリケーションエントリーポイント
- `internal/`: プライベートコード（外部からimport不可）
- レイヤー分離: handlers → services → database

### エラーハンドリング
- 統一されたエラーレスポンス形式
- 適切なHTTPステータスコード
- 日本語エラーメッセージ

### セキュリティ
- 全てのパスワードはbcryptでハッシュ化
- JWT有効期限は24時間
- レート制限: 1秒間に5リクエスト
- 本番環境ではHTTPS必須

## テスト

```bash
# 全テスト実行
go test ./...

# カバレッジ付き
go test -cover ./...
```

## ライセンス

MIT License