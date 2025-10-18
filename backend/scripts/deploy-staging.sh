#!/bin/bash

# ステージング環境への完全デプロイスクリプト
# マイグレーション + アプリケーションデプロイ

set -e

echo "🚀 Starting deployment to STAGING..."
echo ""

# Step 1: マイグレーション前の確認
echo "📋 Step 1: Checking current schema..."
./scripts/check-schema-diff.sh staging

echo ""
read -p "Continue with migration? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "❌ Deployment cancelled"
    exit 0
fi

# Step 2: データベースマイグレーション
echo ""
echo "📝 Step 2: Running database migrations..."
./scripts/migrate-staging.sh

# Step 3: アプリケーションのビルド
echo ""
echo "🔨 Step 3: Building application..."
go build -o bin/server cmd/server/main.go

# Step 4: テスト実行
echo ""
echo "🧪 Step 4: Running tests..."
go test ./internal/handlers ./internal/services

# Step 5: アプリケーションのデプロイ
echo ""
echo "🚢 Step 5: Deploying application..."
# ここに実際のデプロイコマンドを追加
# 例: scp, rsync, docker push, kubectl apply など

echo ""
echo "✅ Deployment to STAGING completed successfully!"
echo ""
echo "Next steps:"
echo "  1. Verify the application is running"
echo "  2. Run smoke tests"
echo "  3. Monitor logs for errors"
