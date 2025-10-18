#!/bin/bash

# マイグレーションスクリプトのテスト
# ローカルのテストデータベースで検証します

set -e

echo "🧪 Testing migration scripts..."
echo ""

# テスト用のデータベース名
TEST_DB="myanimelogs_test_migration"
TEST_USER="${USER}"

echo "📦 Creating test database: $TEST_DB"
createdb "$TEST_DB" 2>/dev/null || {
    echo "⚠️  Database already exists, dropping and recreating..."
    dropdb "$TEST_DB"
    createdb "$TEST_DB"
}

# テスト用のDATABASE_URLを設定
export DATABASE_URL="postgresql://$TEST_USER@localhost/$TEST_DB"

echo "✅ Test database created: $DATABASE_URL"
echo ""

# マイグレーションを実行
echo "📝 Running migrations..."
for migration in migrations/*.sql; do
    echo "  - Testing $(basename $migration)..."
    psql "$DATABASE_URL" -f "$migration" -v ON_ERROR_STOP=1 || {
        echo "❌ Migration failed: $(basename $migration)"
        exit 1
    }
done

echo ""
echo "✅ All migrations applied successfully!"
echo ""

# スキーマを確認
echo "📋 Verifying schema..."
echo ""

echo "Tables:"
psql "$DATABASE_URL" -c "\dt"

echo ""
echo "Animes table structure:"
psql "$DATABASE_URL" -c "\d animes"

echo ""
echo "Episodes table structure:"
psql "$DATABASE_URL" -c "\d episodes"

echo ""
echo "Checking for required columns..."

# 必須カラムのチェック
REQUIRED_COLUMNS=(
    "animes:mal_anime_id"
    "animes:cover_image_large"
    "animes:cover_image_medium"
    "animes:cover_image_updated_at"
    "episodes:duration_minutes"
    "episodes:air_date"
)

ALL_OK=true
for col in "${REQUIRED_COLUMNS[@]}"; do
    TABLE=$(echo $col | cut -d: -f1)
    COLUMN=$(echo $col | cut -d: -f2)
    
    EXISTS=$(psql "$DATABASE_URL" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name='$TABLE' AND column_name='$COLUMN'
    " | xargs)
    
    if [ "$EXISTS" = "1" ]; then
        echo "  ✅ $TABLE.$COLUMN exists"
    else
        echo "  ❌ $TABLE.$COLUMN is missing!"
        ALL_OK=false
    fi
done

echo ""

if [ "$ALL_OK" = true ]; then
    echo "✅ All required columns exist!"
else
    echo "❌ Some required columns are missing!"
    exit 1
fi

echo ""
echo "🧹 Cleaning up test database..."
dropdb "$TEST_DB"

echo ""
echo "✅ Migration test completed successfully!"
echo ""
echo "Next steps:"
echo "  1. Review the migration output above"
echo "  2. If everything looks good, run: ./scripts/check-schema-diff.sh staging"
echo "  3. Then apply to staging: ./scripts/migrate-staging.sh"
