#!/bin/bash

# Staging環境のアニメデータを同期するスクリプト
# レート制限対策として、データベースにアニメ情報と画像をキャッシュします

set -e

# 環境変数
STAGING_URL="${STAGING_URL:-https://myanimelogs-staging.fly.dev}"
ADMIN_EMAIL="${ADMIN_EMAIL}"
ADMIN_PASSWORD="${ADMIN_PASSWORD}"

# 色付きログ
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 管理者認証情報のチェック
if [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
    log_error "管理者認証情報が設定されていません"
    echo "使用方法:"
    echo "  ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=password ./scripts/sync-staging-data.sh"
    exit 1
fi

log_info "Staging環境: $STAGING_URL"

# ステップ1: ログインしてJWTトークンを取得
log_info "管理者としてログイン中..."
LOGIN_RESPONSE=$(curl -s -X POST "$STAGING_URL/api/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    log_error "ログインに失敗しました"
    echo "レスポンス: $LOGIN_RESPONSE"
    exit 1
fi

log_info "ログイン成功"

# ステップ2: 現在のシーズンのアニメを同期
log_info "現在のシーズンのアニメを同期中..."
SEASON_RESPONSE=$(curl -s -X POST "$STAGING_URL/api/admin/sync/season" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

echo "$SEASON_RESPONSE" | grep -q '"status":"success"'
if [ $? -eq 0 ]; then
    SEASON_COUNT=$(echo "$SEASON_RESPONSE" | grep -o '"message":"Successfully synced [0-9]* anime works' | grep -o '[0-9]*')
    log_info "シーズンアニメの同期完了: ${SEASON_COUNT}件"
else
    log_warn "シーズンアニメの同期に失敗しました"
    echo "レスポンス: $SEASON_RESPONSE"
fi

# レート制限を避けるため少し待機
sleep 2

# ステップ3: 人気アニメを同期
log_info "人気アニメを同期中..."
POPULAR_RESPONSE=$(curl -s -X POST "$STAGING_URL/api/admin/sync/popular" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

echo "$POPULAR_RESPONSE" | grep -q '"status":"success"'
if [ $? -eq 0 ]; then
    POPULAR_COUNT=$(echo "$POPULAR_RESPONSE" | grep -o '"message":"Successfully synced [0-9]* popular anime works' | grep -o '[0-9]*')
    log_info "人気アニメの同期完了: ${POPULAR_COUNT}件"
else
    log_warn "人気アニメの同期に失敗しました"
    echo "レスポンス: $POPULAR_RESPONSE"
fi

log_info "データ同期が完了しました"
log_info "ブラウザで確認: $STAGING_URL"
