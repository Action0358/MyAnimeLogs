# MyAnimeLogs

アニメ視聴管理Webアプリ。アニメ検索・お気に入り・視聴進捗管理が可能。

> 注: 評価(rating)機能は2026-06に撤去済み（commit 943c4db2 / migration 043_drop_user_ratings_table）。

## 技術スタック

- **バックエンド**: Go 1.24+ / Gin / PostgreSQL（Supabase）
- **フロントエンド**: Next.js 14（App Router）/ TypeScript / Tailwind CSS / TanStack Query
- **インフラ**: Fly.io（バックエンド）/ Cloudflare Pages（フロントエンド）/ GitHub Actions

## 開発コマンド

```bash
# バックエンド
cd backend && make run          # 起動
cd backend && go build ./...    # ビルド確認
cd backend && go test ./...     # テスト実行
cd backend && make migrate      # マイグレーション実行

# フロントエンド
cd frontend && npm run dev      # 起動
cd frontend && npx tsc --noEmit # 型チェック
cd frontend && npm test         # テスト実行
```

### コンテナの再ビルド（重要）

バックエンドのGoコードを変更した場合、またはマイグレーションを含む変更をpullした場合は、
**必ず再ビルドする**。`restart` では古いバイナリのままで変更が反映されない。

```bash
docker compose up -d --build backend
```

マイグレーションは起動時に自動実行されるため、再ビルドを怠るとDBだけが新しく、
バイナリが古い不整合状態になる。この状態ではログイン自体が失敗する
（例: 削除済みカラムをSELECTして `pq: column "..." does not exist` → `/api/users/me` が500）。
STG・本番はデプロイのたびにビルドし直すため、この問題はローカル固有。

## 外部API制約（重要）

| API | レート制限 | 用途 |
|-----|-----------|------|
| Annict | 5 req/秒 | アニメ基本情報 |
| AniList | 90 req/分 | カバー画像・ジャンル |
| Jikan | 3 req/秒 | 放送情報 |
| TMDB | 40 req/10秒 | 日本語あらすじ・配信サービス |

### ステータス値（DB値）

`airing` / `finished` / `upcoming` のみ使用。`currently_airing` `not_yet_aired` は不正。

### MAL IDの役割

AniList・Jikan・TMDB の横断検索キー。Annictから取得。

## スキルの自動適用

以下の変更を行った際は、対応するスキルを**自動的に実行**してください。

| 変更内容 | 実行するスキル |
|---------|--------------|
| 機能実装・バグ修正・リファクタリング完了時、またはコミット前 | `/post-implementation-verification` |
| `backend/internal/router/router.go` / `backend/internal/handlers/*.go` / `backend/internal/models/*.go` の変更 | `/api-consistency-check` |
| `backend/internal/scheduler/` / `backend/internal/services/` の sync・annict・jikan 関連コードの変更 | `/sync-logic-review` |
| `.github/workflows/` / `backend/fly.toml` / `frontend/public/_headers` / `**/Dockerfile` の変更 | `/deploy-consistency-check` |
| `backend/migrations/` へのファイル追加・変更 | `/migration-safety-check` |
