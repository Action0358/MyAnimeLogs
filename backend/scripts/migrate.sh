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

# /app/migrateバイナリを使用してマイグレーションを実行
# このバイナリは実行済みマイグレーションを追跡し、未実行のものだけを実行する
if [ -x "/app/migrate" ]; then
    echo "📝 Running migrations using /app/migrate..."
    echo ""
    /app/migrate
elif [ -f "/app/migrate" ]; then
    echo "❌ Error: /app/migrate exists but is not executable"
    echo "   Please check file permissions"
    exit 1
else
    echo "❌ Error: /app/migrate binary not found"
    echo "   Please ensure the Docker image was built correctly"
    exit 1
fi

echo ""
echo "✅ Migration completed successfully!"
echo ""

# スキーマ情報を表示
echo "📋 Current database schema:"
if command -v psql > /dev/null 2>&1; then
    psql "$DATABASE_URL" -c "\dt" || echo "⚠️  Warning: Could not retrieve table list (this is not critical)"
else
    echo "ℹ️  psql not available, skipping schema display"
fi

exit 0
