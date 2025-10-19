#!/bin/bash
set -e
set -o pipefail

# Production環境のマイグレーション履歴をリセットして001から再実行
# 
# ⚠️ 警告: このスクリプトはProduction環境のマイグレーション履歴を完全にリセットします
# 
# 使用目的:
# - マイグレーションファイルを修正した後、改善版を適用する場合
# - マイグレーション履歴が破損した場合の復旧
# 
# 注意:
# - 通常の運用では使用しないでください
# - データは保持されますが、マイグレーション履歴は削除されます
# - 実行前に必ずバックアップを取ってください
# 
# 最終使用日: 2025-10-07 (Production DB同期問題の解決時)

set -e

echo "=== Production Migration Reset ==="
echo ""
echo "⚠️  このスクリプトは以下を実行します："
echo "  1. migrationsテーブルの全レコードを削除"
echo "  2. アプリケーションを再起動して全マイグレーションを再実行"
echo ""
echo "データは保持されます（テーブルは削除されません）"
echo ""
read -p "続行しますか？ (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "キャンセルしました"
    exit 0
fi

echo ""
echo "📊 現在のマイグレーション履歴を確認中..."
flyctl ssh console -a myanimelogs-production << 'EOF'
psql $DATABASE_URL -c "SELECT id, filename, executed_at FROM migrations ORDER BY id;"
EOF

echo ""
echo "🗑️  マイグレーション履歴をリセット中..."
flyctl ssh console -a myanimelogs-production << 'EOF'
psql $DATABASE_URL -c "TRUNCATE TABLE migrations;"
echo "✅ マイグレーション履歴をリセットしました"
EOF

echo ""
echo "🔄 アプリケーションを再起動してマイグレーションを再実行..."
flyctl apps restart myanimelogs-production

echo ""
echo "⏳ 10秒待機中..."
sleep 10

echo ""
echo "📋 マイグレーションログを確認中..."
flyctl logs -a myanimelogs-production -n | grep -E "(migration|Migration)" | tail -30

echo ""
echo "✅ リセット完了"
echo ""
echo "次のステップ:"
echo "1. マイグレーションが正常に実行されたか確認"
echo "2. データ同期をテスト"
echo "   cd backend && echo 'y' | ADMIN_EMAIL=your-email ADMIN_PASSWORD=your-password ./scripts/sync-production-data.sh"
