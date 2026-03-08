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

# 本番環境の場合は確認プロンプトを表示
# APP_ENVまたはFLY_APP_NAMEで本番環境を判定
if [ "$APP_ENV" = "production" ] || [[ "$FLY_APP_NAME" == *"-production"* ]] || [[ "$DATABASE_URL" == *"prod"* ]]; then
    echo "⚠️  WARNING: This appears to be a PRODUCTION environment!"
    echo "⚠️  Migration will modify the production database."
    echo ""
    
    # インタラクティブモードの場合のみ確認プロンプトを表示
    if [ -t 0 ]; then
        read -p "Are you sure you want to continue? (yes/no): " confirm
        if [ "$confirm" != "yes" ]; then
            echo "❌ Migration cancelled"
            exit 0
        fi
    else
        echo "ℹ️  Running in non-interactive mode, proceeding with migration..."
    fi
    
    # バックアップの作成を試みる（pg_dumpが利用可能な場合）
    if command -v pg_dump > /dev/null 2>&1; then
        echo ""
        echo "💾 Creating database backup..."
        BACKUP_FILE="/tmp/backup_$(date +%Y%m%d_%H%M%S).sql"
        if pg_dump "$DATABASE_URL" > "$BACKUP_FILE" 2>/dev/null; then
            echo "✅ Backup created: $BACKUP_FILE"
            echo "   (Note: Backup is stored in /tmp and will be lost on container restart)"
        else
            echo "⚠️  Warning: Backup failed, but continuing with migration..."
            echo "   Please ensure you have a recent backup of your database."
        fi
    else
        echo "⚠️  Warning: pg_dump not available, skipping backup"
        echo "   Please ensure you have a recent backup of your database."
    fi
    echo ""
fi

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
