#!/bin/bash
set -e
set -o pipefail

# SQLファイルの構文チェック
# PostgreSQLの--dry-runオプションを使用して構文を検証

set -e

echo "🔍 Validating SQL migration files..."
echo ""

# テスト用のデータベース
TEST_DB="myanimelogs_sql_validation"
TEST_USER="${USER}"

# 一時的なテストデータベースを作成
createdb "$TEST_DB" 2>/dev/null || {
    dropdb "$TEST_DB" 2>/dev/null
    createdb "$TEST_DB"
}

export DATABASE_URL="postgresql://$TEST_USER@localhost/$TEST_DB"

ALL_VALID=true

for migration in migrations/*.sql; do
    echo -n "Validating $(basename $migration)... "
    
    # SQLファイルをトランザクション内で実行し、ロールバック
    # psqlコマンドの終了コードを直接評価
    if psql "$DATABASE_URL" -v ON_ERROR_STOP=1 > /dev/null 2>&1 <<EOF
BEGIN;
\i $migration
ROLLBACK;
EOF
    then
        echo "✅ OK"
    else
        echo "❌ FAILED"
        echo ""
        echo "Error details:"
        # エラー詳細を表示するため再実行（|| true で失敗を許容）
        psql "$DATABASE_URL" -v ON_ERROR_STOP=1 <<EOF || true
BEGIN;
\i $migration
ROLLBACK;
EOF
        echo ""
        ALL_VALID=false
    fi
done

# クリーンアップ
dropdb "$TEST_DB" 2>/dev/null

echo ""
if [ "$ALL_VALID" = true ]; then
    echo "✅ All SQL files are valid!"
    exit 0
else
    echo "❌ Some SQL files have errors!"
    exit 1
fi
