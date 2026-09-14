# スクリプトガイド

## 📋 このディレクトリについて

データベースマイグレーション、データ同期、デプロイ検証のためのスクリプト集です。

**推奨実行方法**: Fly.io 経由でのマイグレーション実行（環境に依存しない）

### マイグレーション管理の仕組み

このプロジェクトでは、`migrate.go`を使用してマイグレーションを管理しています：

- ✅ **実行履歴の追跡**: `migrations`テーブルで実行済みマイグレーションを記録
- ✅ **冪等性**: 既に実行されたマイグレーションは自動的にスキップ
- ✅ **トランザクション管理**: エラー時は自動ロールバック
- ✅ **一貫性**: 開発環境、ステージング環境、本番環境で同じロジックを使用

**重要**: マイグレーションファイル（`migrations/*.sql`）は冪等である必要はありません。`migrate.go`が実行状態を管理します。

---

## 🎯 1 分で分かるクイックガイド

### マイグレーションの流れ（Fly.io 経由）

```
┌─────────────────────────────────────────────────────────┐
│ 1. ローカルでテスト                                      │
│    ./scripts/test-migration.sh                          │
│    → テスト用DBで安全に確認                              │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 2. ステージング環境に適用（Fly.io経由）                  │
│    flyctl ssh console -a myanimelogs-staging \          │
│      -C "/app/migrate.sh"                               │
│    → 本番前の最終確認                                    │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 3. 本番環境に適用（Fly.io経由）                          │
│    flyctl ssh console -a myanimelogs-production \       │
│      -C "/app/migrate.sh"                               │
│    → 本番環境のデータベースを更新                        │
└─────────────────────────────────────────────────────────┘
```

### データ同期の流れ

```
┌─────────────────────────────────────────────────────────┐
│ ステージング環境                                         │
│    ADMIN_EMAIL=xxx ADMIN_PASSWORD=xxx \                 │
│      ./scripts/sync-staging-data.sh                     │
│    → 現在のシーズン + 人気アニメを取得                   │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ 本番環境                                                 │
│    ADMIN_EMAIL=xxx ADMIN_PASSWORD=xxx \                 │
│      ./scripts/sync-production-data.sh                  │
│    → 確認プロンプト + データ同期                         │
└─────────────────────────────────────────────────────────┘
```

---

## 📊 スクリプト早見表

| スクリプト                | 一言で言うと       | 実行時間 | 危険度        |
| ------------------------- | ------------------ | -------- | ------------- |
| `test-migration.sh`       | ローカルでお試し   | 10 秒    | 🟢 安全       |
| `validate-sql.sh`         | SQL 構文チェック   | 5 秒     | 🟢 安全       |
| `check-schema-diff.sh`    | スキーマ確認       | 3 秒     | 🟢 安全       |
| `manual_full_sync.go`     | データ同期         | 5-15 分  | 🟡 注意       |
| `sync-staging-data.sh`    | アニメデータ取得   | 1 分     | 🟡 注意       |
| `sync-production-data.sh` | 本番データ取得     | 1 分     | 🔴 慎重に     |
| `verify-deployment.sh`    | 動作確認           | 10 秒    | 🟢 安全       |
| `reset-migrations.sh`     | 履歴リセット       | 1 分     | 🔴 緊急時のみ |

---

## 🎯 目的別スクリプト早見表

| やりたいこと                                   | 使うスクリプト                                     | 環境              |
| ---------------------------------------------- | -------------------------------------------------- | ----------------- |
| マイグレーションをテストしたい                 | `test-migration.sh`                                | ローカル          |
| ステージング環境にマイグレーションを適用したい | Fly.io経由: `flyctl ssh console -C "/app/migrate.sh"` | ステージング      |
| 本番環境にマイグレーションを適用したい         | Fly.io経由: `flyctl ssh console -C "/app/migrate.sh"` | 本番              |
| スキーマの差分を確認したい                     | `check-schema-diff.sh`                             | 全環境            |
| アニメデータを同期したい                       | `manual_full_sync.go` / `sync-*-data.sh`           | ステージング/本番 |
| デプロイ後の動作を確認したい                   | `verify-deployment.sh`                             | ステージング/本番 |
| 利用状況（登録数・利用率・検索傾向）を見たい   | `analytics.sql`                                    | 本番/ローカル     |

---

## 📊 利用状況の集計

### `analytics.sql`

新規登録数・利用率・継続・検索傾向を実データから集計する参照専用のクエリ集（#868）。

```bash
# 本番
psql "$PRODUCTION_DATABASE_URL" -f backend/scripts/analytics.sql

# ローカル
docker exec -i supabase_db_myanimelogs-supabase \
  psql -U postgres -d postgres -f /dev/stdin < backend/scripts/analytics.sql
```

全部流すと結果が多いため、普段は必要な節だけコピーして実行するとよい。

**内容**

| 節 | 内容 |
| -- | ---- |
| 1 | 月別の新規登録数 |
| 2 | 利用率（登録者のうち記録機能を使った割合） |
| 3 | 登録月ごとの定着（コホート） |
| 4 | 直近30日に記録操作をしたユーザー数 |
| 5 | 記録の内訳（お気に入り / 視聴ステータス） |
| 6 | 視聴ステータスの分布 |
| 7 | 検索キーワード上位 |
| 8 | 0件だった検索（改善ネタ） |
| 9 | 検索のログイン別内訳 |

**登録率について**

本サービスはログインしなくても使えるため、「来たが登録しなかった人」はDBに痕跡を残さない。
登録率を出すには分母（訪問者数）が必要で、Cloudflare Web Analytics の訪問者数 ÷ 1) の新規登録数 で算出する。

> ⚠️ このファイルは参照専用。`UPDATE` / `DELETE` は書かないこと。

---

## 📁 スクリプト一覧

### 🗄️ データベースマイグレーション



<details>
<summary><strong>migrate.go</strong> - Go言語版マイグレーション（代替手段）</summary>

**いつ使う？**

- 開発環境や CI/CD で使いたいとき
- トランザクション管理が必要なとき

**何をする？**

1. `migrations`テーブルで実行履歴を管理
2. 未実行のマイグレーションだけを実行
3. エラーが起きたら自動ロールバック

**使い方**

```bash
cd backend
go run scripts/migrate.go
```

**注意点**

- ✅ 実行済みマイグレーションは自動スキップ
- ✅ トランザクション内で安全に実行

</details>

---

### 🔍 検証・確認

<details>
<summary><strong>test-migration.sh</strong> - ローカルでマイグレーションをテスト</summary>

**いつ使う？**

- マイグレーションを本番に適用する前
- SQL ファイルが正しく動くか確認したいとき

**何をする？**

1. テスト用データベース`myanimelogs_test_migration`を作成
2. 全マイグレーションを実行
3. 必須カラムが存在するか確認
4. テストが終わったらデータベースを削除

**使い方**

```bash
./scripts/test-migration.sh
```

**結果の見方**

```
✅ animes.mal_anime_id exists
✅ episodes.duration_minutes exists
✅ All required columns exist!
```

</details>

<details>
<summary><strong>validate-sql.sh</strong> - SQLファイルの構文チェック</summary>

**いつ使う？**

- SQL ファイルを作成・編集した後
- 構文エラーがないか確認したいとき

**何をする？**

1. 各 SQL ファイルをテスト実行
2. 構文エラーを検出
3. ロールバックして元に戻す（影響なし）

**使い方**

```bash
./scripts/validate-sql.sh
```

**結果の見方**

```
Validating 000_create_users_table.sql... ✅ OK
Validating 001_create_animes_table.sql... ✅ OK
```

</details>

<details>
<summary><strong>check-schema-diff.sh</strong> - スキーマの差分確認</summary>

**いつ使う？**

- マイグレーション前後でスキーマを確認したいとき
- 開発環境と本番環境の違いを調べたいとき

**何をする？**

1. 指定した環境のデータベースに接続
2. テーブル一覧を表示
3. `animes`と`episodes`テーブルの構造を表示
4. 全カラムのリストを表示

**使い方**

```bash
# ステージング環境を確認
./scripts/check-schema-diff.sh staging

# 本番環境を確認
./scripts/check-schema-diff.sh production
```

</details>

<details>
<summary><strong>verify-deployment.sh</strong> - デプロイ後の動作確認</summary>

**いつ使う？**

- デプロイ後に API が正常に動いているか確認したいとき

**何をする？**

1. ヘルスチェック
2. API エンドポイントの疎通確認
3. Annict API 連携のテスト
4. CORS 設定の確認

**使い方**

```bash
# 両環境を確認
./scripts/verify-deployment.sh

# ステージングのみ
./scripts/verify-deployment.sh staging

# 本番のみ
./scripts/verify-deployment.sh production
```

**結果の見方**

```
✅ Health Check (HTTP 200)
✅ Current Season Anime (Annict)
✅ CORS headers present
```

</details>

---

### 🚢 デプロイ

### 📊 データ同期

<details>
<summary><strong>sync-staging-data.sh</strong> - ステージング環境のアニメデータ同期</summary>

**いつ使う？**

- ステージング環境に最新のアニメデータを取り込みたいとき

**何をする？**

1. 管理者としてログイン
2. 現在のシーズンのアニメを取得
3. 人気アニメを取得
4. データベースに保存

**使い方**

```bash
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=password \
  ./scripts/sync-staging-data.sh
```

**注意点**

- ⚠️ 管理者の認証情報が必要
- ⚠️ レート制限対策で待機時間あり

</details>

<details>
<summary><strong>sync-production-data.sh</strong> - 本番環境のアニメデータ同期</summary>

**いつ使う？**

- 本番環境に最新のアニメデータを取り込みたいとき

**何をする？**

1. 「本当に実行しますか？」と確認
2. 管理者としてログイン
3. 現在のシーズンのアニメを取得
4. 人気アニメを取得
5. データベースに保存

**使い方**

```bash
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=password \
  ./scripts/sync-production-data.sh
```

**注意点**

- ⚠️ 本番環境なので確認プロンプトあり
- ⚠️ 管理者の認証情報が必要
- ⚠️ レート制限対策で待機時間あり

</details>

---

### 🔧 メンテナンス（緊急時のみ）

<details>
<summary><strong>reset-migrations.sh</strong> - マイグレーション履歴のリセット</summary>

**いつ使う？**

- マイグレーション履歴が壊れたとき（めったにない）
- マイグレーションを最初からやり直したいとき

**何をする？**

1. `migrations`テーブルの履歴を全削除
2. アプリケーションを再起動
3. 全マイグレーションを再実行

**使い方**

```bash
./scripts/reset-migrations.sh
```

**⚠️ 重要な注意点**

- ❌ 通常は使わないでください
- ✅ データは削除されません（テーブルは残る）
- ⚠️ 実行前にバックアップを取ること

</details>

<details>
<summary><strong>reset-migrations.go</strong> - Go言語版リセット</summary>

**いつ使う？**

- 開発環境でマイグレーション履歴をリセットしたいとき

**使い方**

```bash
cd backend
go run scripts/reset-migrations.go
```

**注意点**

- ⚠️ 開発環境専用

</details>

---

## 🎓 初めての人向けガイド

### ステップ 1: まずはテスト

```bash
# これを実行すれば、マイグレーションが正しいか分かる
./scripts/test-migration.sh
```

**何が起こる？**

- テスト用のデータベースを作る
- マイグレーションを実行
- 問題がないか確認
- テスト用データベースを削除

**結果**

- ✅ 成功 → 次のステップへ
- ❌ 失敗 → SQL ファイルを修正

### ステップ 2: ステージング環境で試す

```bash
# スキーマを確認
./scripts/check-schema-diff.sh staging

# マイグレーションを実行（Fly.io経由）
flyctl ssh console -a myanimelogs-staging -C "/app/migrate.sh"
```

**何が起こる？**

- ステージング環境のデータベースが更新される
- 新しいカラムやテーブルが追加される

### ステップ 3: 動作確認

```bash
# APIが正常に動いているか確認
./scripts/verify-deployment.sh staging
```

**何が起こる？**

- API エンドポイントをテスト
- 問題があれば教えてくれる

### ステップ 4: 本番環境に適用

```bash
# 本番環境のスキーマを確認
./scripts/check-schema-diff.sh production

# 本番環境にマイグレーションを適用（Fly.io経由）
flyctl ssh console -a myanimelogs-production -C "/app/migrate.sh"
```

**何が起こる？**

- 本番環境のデータベースが更新される
- 新しいカラムやテーブルが追加される

---

## 🚀 よくある使い方

### パターン 1: 新しいマイグレーションを本番に適用

```bash
# 1. ローカルでテスト
./scripts/test-migration.sh

# 2. ステージング環境で確認
./scripts/check-schema-diff.sh staging
flyctl ssh console -a myanimelogs-staging -C "/app/migrate.sh"

# 3. 動作確認
./scripts/verify-deployment.sh staging

# 4. 本番環境に適用
./scripts/check-schema-diff.sh production
flyctl ssh console -a myanimelogs-production -C "/app/migrate.sh"

# 5. 動作確認
./scripts/verify-deployment.sh production
```

### パターン 2: 新しいシーズンのアニメデータを追加

```bash
# ステージング環境で試す
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=password \
  ./scripts/sync-staging-data.sh

# 確認
./scripts/verify-deployment.sh staging

# 本番環境に適用
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=password \
  ./scripts/sync-production-data.sh
```

---

## 💡 よくある質問

### Q: どのスクリプトから使えばいい？

**A**: まずは`test-migration.sh`から始めてください。

### Q: Fly.io 経由とローカル実行、どちらを使うべき？

**A**: **Fly.io 経由を推奨**します。理由：

- ローカル環境に PostgreSQL クライアントが不要
- どこからでも実行可能
- 本番環境と同じ Docker イメージで実行される

ただし、Fly.io がダウンしている場合や緊急時は、ローカルから直接実行できます。

### Q: 本番環境に適用するのが怖い

**A**: 以下の順番で進めれば安全です：

1. `test-migration.sh` でローカルテスト
2. Fly.io 経由でステージングテスト
3. `verify-deployment.sh staging` で動作確認
4. 問題なければ本番環境に適用

### Q: Fly.io の VM が suspended になっている

**A**: マイグレーション前に VM を起動してください：

```bash
flyctl machine start $(flyctl machine list -a myanimelogs-staging --json | jq -r '.[0].id') -a myanimelogs-staging
```

### Q: バックアップはどこに保存される？

**A**: Supabase が自動的にバックアップを作成しています。
Supabase ダッシュボードから復元できます。

### Q: マイグレーションを元に戻したい

**A**: Supabase ダッシュボードからバックアップを復元してください。

### Q: 何度も実行しても大丈夫？

**A**: はい、大丈夫です。マイグレーションは冪等性があります。
既に存在するテーブルやカラムはスキップされます。

---

## 🛠️ トラブルシューティング

### Q: データベースに接続できない

```bash
# 接続テスト
psql $DATABASE_URL -c "SELECT version();"

# 環境変数を確認
echo $DATABASE_URL
```

### Q: テストデータベースが残っている

```bash
# クリーンアップ
dropdb myanimelogs_test_migration
dropdb myanimelogs_sql_validation
```

### Q: マイグレーションが途中で失敗した

```bash
# バックアップから復元（本番環境）
psql $DATABASE_URL < backups/backup_YYYYMMDD_HHMMSS.sql

# 再テスト
./scripts/test-migration.sh
```

### Q: データ同期が失敗する

```bash
# 認証情報を確認
echo $ADMIN_EMAIL
echo $ADMIN_PASSWORD

# APIの状態を確認
./scripts/verify-deployment.sh staging
```

---

## 🆘 緊急時の対応

### マイグレーションが失敗した！

```bash
# 1. 落ち着いて、エラーメッセージを確認
# 2. バックアップから復元
psql $DATABASE_URL < backups/backup_最新のファイル.sql

# 3. 開発チームに連絡
```

### データベースが壊れた！

```bash
# 1. バックアップから復元
psql $DATABASE_URL < backups/backup_最新のファイル.sql

# 2. マイグレーション履歴をリセット（必要な場合）
./scripts/reset-migrations.sh

# 3. 開発チームに連絡
```

---

## ⚠️ 重要な注意事項

### 本番環境への適用前に必ず確認

- [ ] ステージング環境でテスト済み
- [ ] バックアップが作成されている
- [ ] ダウンタイムの計画がある（必要な場合）
- [ ] ロールバック手順を確認済み

### マイグレーションの特性

- ✅ **冪等性**: 何度実行しても安全
- ✅ **自動停止**: エラーが起きたら即座に停止
- ✅ **バックアップ**: 本番環境では自動作成

---

## 🔮 今後の予定

Fly.io 経由でのマイグレーション実行を改善中です。詳細は以下を参照：

- `.kiro/tickets/backend/ISSUE_MIGRATION_AUTOMATION.md`

**実装予定：**

- Docker イメージに`psql`とマイグレーションスクリプトを含める
- `flyctl ssh console`で簡単にマイグレーションを実行できるようにする
- GitHub Actions での自動化

---

## 🔒 運用ルール

### 禁止事項

❌ **本番・ステージング環境で直接 SQL を実行してスキーマを変更**

- データベースに直接接続して`ALTER TABLE`などを実行しない
- 緊急時を除き、必ずマイグレーションファイルを使用する

❌ **マイグレーションファイルを経由しないスキーマ変更**

- スキーマ変更は必ず`migrations/`フォルダ内の SQL ファイルで行う
- 手動での変更は記録が残らず、環境間の不一致を引き起こす

❌ **開発環境のみでマイグレーションをテストせずに本番適用**

- 必ずローカルでテストしてから本番に適用する

### 必須事項

✅ **スキーマ変更は必ずマイグレーションファイルで行う**

```bash
# 新しいマイグレーションファイルを作成
# backend/migrations/010_add_new_column.sql

# ローカルでテスト
./scripts/test-migration.sh

# ステージングに適用（Fly.io経由）
flyctl ssh console -a myanimelogs-staging -C "/app/migrate.sh"

# 本番に適用（Fly.io経由）
flyctl ssh console -a myanimelogs-production -C "/app/migrate.sh"
```

✅ **マイグレーション実行順序を守る**

1. ローカルテスト（`test-migration.sh`）
2. ステージング環境（Fly.io経由: `flyctl ssh console -C "/app/migrate.sh"`）
3. 動作確認（`verify-deployment.sh staging`）
4. 本番環境（Fly.io経由: `flyctl ssh console -C "/app/migrate.sh"`）
5. 動作確認（`verify-deployment.sh production`）

✅ **マイグレーションファイルは必ず Git にコミット**

```bash
git add backend/migrations/010_add_new_column.sql
git commit -m "Add migration: add new column to table"
git push
```

✅ **マイグレーション実行後は動作確認を行う**

- API が正常に動作するか確認
- エラーログを確認
- 必要に応じてロールバック手順を準備

### 推奨事項

✅ **マイグレーション前にバックアップを取得（本番環境）**

- Supabase ダッシュボードから手動バックアップを作成
- または、`pg_dump`でバックアップを取得

✅ **マイグレーション実行ログを記録**

- 実行日時、実行者、結果を記録
- 問題が発生した場合の調査に役立つ

✅ **問題が発生した場合のロールバック手順を準備**

- ロールバック用の SQL を事前に用意
- バックアップからの復元手順を確認

---

## 📞 困ったときは

問題が発生した場合は、開発チームに連絡してください。
