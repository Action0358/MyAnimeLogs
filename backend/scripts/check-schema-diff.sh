#!/bin/bash

# スキーマの差分をチェックするスクリプト
# Usage: ./scripts/check-schema-diff.sh [staging|production]

set -e

ENV=${1:-staging}

echo "🔍 Checking schema differences for $ENV environment..."
echo ""

# Load environment variables
if [ -f ".env.$ENV" ]; then
    export $(cat ".env.$ENV" | grep -v '^#' | xargs)
else
    echo "❌ Error: .env.$ENV file not found"
    exit 1
fi

if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL is not set"
    exit 1
fi

echo "📊 Checking tables..."
psql "$DATABASE_URL" -c "\dt" || {
    echo "❌ Error: Could not connect to database"
    exit 1
}

echo ""
echo "📋 Checking animes table structure..."
psql "$DATABASE_URL" -c "\d animes"

echo ""
echo "📋 Checking episodes table structure..."
psql "$DATABASE_URL" -c "\d episodes"

echo ""
echo "📋 Checking for missing columns..."

# Check for specific columns that should exist
echo "Checking animes table columns:"
psql "$DATABASE_URL" -t -c "
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'animes' 
ORDER BY ordinal_position;
"

echo ""
echo "Checking episodes table columns:"
psql "$DATABASE_URL" -t -c "
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'episodes' 
ORDER BY ordinal_position;
"

echo ""
echo "✅ Schema check completed!"
