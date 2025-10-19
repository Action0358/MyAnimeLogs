#!/bin/bash
set -e
set -o pipefail

echo "🚀 Starting database migration..."

# DATABASE_URLが設定されているか確認
if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL is not set"
    exit 1
fi

# パスワードをマスクして表示
MASKED_URL=$(echo "$DATABASE_URL" | sed -E 's/(:[^:@]+)@/:*****@/')
echo "📊 Database: $MASKED_URL"
echo ""

# マイグレーションディレクトリの確認
MIGRATION_DIR="/app/migrations"
if [ ! -d "$MIGRATION_DIR" ]; then
    echo "❌ Error: Migration directory not found: $MIGRATION_DIR"
    exit 1
fi

# マイグレーションファイルの数を確認
MIGRATION_COUNT=$(ls -1 "$MIGRATION_DIR"/*.sql 2>/dev/null | wc -l)
if [ "$MIGRATION_COUNT" -eq 0 ]; then
    echo "⚠️  Warning: No migration files found in $MIGRATION_DIR"
    exit 0
fi

echo "📝 Running $MIGRATION_COUNT migration(s)..."
echo ""

# マイグレーションを実行
SUCCESS_COUNT=0
FAILED_COUNT=0

for migration in "$MIGRATION_DIR"/*.sql; do
    if [ -f "$migration" ]; then
        FILENAME=$(basename "$migration")
        echo "  - Applying $FILENAME..."
        
        if psql "$DATABASE_URL" -f "$migration" -v ON_ERROR_STOP=1 > /dev/null 2>&1; then
            echo "    ✅ Success"
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        else
            echo "    ❌ Failed"
            FAILED_COUNT=$((FAILED_COUNT + 1))
            
            # エラーが発生した場合は詳細を表示
            echo ""
            echo "Error details:"
            psql "$DATABASE_URL" -f "$migration" -v ON_ERROR_STOP=1
            exit 1
        fi
    fi
done

echo ""
echo "✅ Migration completed successfully!"
echo "   - Success: $SUCCESS_COUNT"
echo "   - Failed: $FAILED_COUNT"
echo ""

# スキーマ情報を表示
echo "📋 Current database schema:"
psql "$DATABASE_URL" -c "\dt" 2>/dev/null || echo "Could not retrieve table list"

exit 0
