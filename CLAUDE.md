# MyAnimeLogs

アニメ視聴管理Webアプリ。アニメ検索・お気に入り・視聴進捗管理・評価レビューが可能。

## 技術スタック

- **バックエンド**: Go 1.21+ / Gin / PostgreSQL（Supabase）
- **フロントエンド**: Next.js 14（App Router）/ TypeScript / Tailwind CSS / TanStack Query
- **インフラ**: Fly.io（バックエンド）/ Cloudflare Pages（フロントエンド）/ GitHub Actions

## 開発コマンド

```bash
cd backend && make run          # バックエンド起動
cd frontend && npm run dev      # フロントエンド起動
cd backend && go build ./...    # ビルド確認
cd backend && go test ./...     # テスト実行
cd frontend && npx tsc --noEmit # 型チェック
cd frontend && npm test -- --passWithNoTests
cd backend && make migrate      # マイグレーション実行
```

## 外部API制約（重要）

| API | レート制限 | 用途 |
|-----|-----------|------|
| Annict | 5 req/秒 | アニメ基本情報・エピソードINSERT |
| AniList | 90 req/分 | カバー画像・ジャンル |
| Jikan | 3 req/秒 | エピソード補完UPDATE・放送情報 |
| TMDB | 40 req/10秒 | 日本語あらすじ・配信サービス |

### エピソード同期の責務分離（変更禁止）

- **INSERT**: `CreateEpisodesFromAnnict` のみ（タイトルは常に「第N話」をデフォルト）
- **UPDATE**: `SyncEpisodesFromJikan`（公開メソッド）が `UpdateEpisodesFromJikan`（内部実装）を呼び出す形で補完
- `CreateEpisodesFromJikan` は存在しない

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
