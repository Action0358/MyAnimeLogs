#!/bin/bash

# Staging環境のマイグレーションスクリプト
# Usage: ./scripts/migrate-staging.sh

set -e

echo "🚀 Starting migration for STAGING environment..."

# Load staging environment variables
if [ -f .env.staging ]; then
    export $(cat .env.staging | grep -v '^#' | xargs)
else
    echo "❌ Error: .env.staging file not found"
    exit 1
fi

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL is not set in .env.staging"
    exit 1
fi

echo "📊 Database: $DATABASE_URL"
echo ""

# Run migrations
echo "📝 Running migrations..."
for migration in migrations/*.sql; do
    echo "  - Applying $(basename $migration)..."
    psql "$DATABASE_URL" -f "$migration" -v ON_ERROR_STOP=1
done

echo ""
echo "✅ Migration completed successfully for STAGING!"
echo ""

# Show current schema version
echo "📋 Current database schema:"
psql "$DATABASE_URL" -c "\dt" -c "\d animes" -c "\d episodes"
