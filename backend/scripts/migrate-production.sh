#!/bin/bash

# Production環境のマイグレーションスクリプト
# Usage: ./scripts/migrate-production.sh

set -e

echo "🚀 Starting migration for PRODUCTION environment..."
echo "⚠️  WARNING: This will modify the PRODUCTION database!"
echo ""
read -p "Are you sure you want to continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "❌ Migration cancelled"
    exit 0
fi

# Load production environment variables
if [ -f .env.production ]; then
    export $(cat .env.production | grep -v '^#' | xargs)
else
    echo "❌ Error: .env.production file not found"
    exit 1
fi

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL is not set in .env.production"
    exit 1
fi

echo "📊 Database: $DATABASE_URL"
echo ""

# Create backup
echo "💾 Creating backup..."
BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
pg_dump "$DATABASE_URL" > "backups/$BACKUP_FILE" || {
    echo "⚠️  Warning: Backup failed, but continuing..."
}

# Run migrations
echo "📝 Running migrations..."
for migration in migrations/*.sql; do
    echo "  - Applying $(basename $migration)..."
    psql "$DATABASE_URL" -f "$migration" -v ON_ERROR_STOP=1
done

echo ""
echo "✅ Migration completed successfully for PRODUCTION!"
echo ""

# Show current schema version
echo "📋 Current database schema:"
psql "$DATABASE_URL" -c "\dt" -c "\d animes" -c "\d episodes"
