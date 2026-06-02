# ルート Makefile — ローカル開発環境のオーケストレーション
#
# 主要ターゲット:
#   make local-up           # Supabase + アプリ Docker を全部起動（推奨）
#   make local-down         # アプリ Docker のみ停止（Supabase は残す）
#   make local-stop         # Supabase + アプリ Docker を全停止
#   make local-supabase-up  # Supabase CLI のみ起動 + anon key を frontend/.env.local に反映
#   make local-supabase-down# Supabase CLI のみ停止
#   make sync-anon-key      # Supabase 起動中の anon key を frontend/.env.local に反映

.PHONY: local-up local-down local-stop local-supabase-up local-supabase-down sync-anon-key local-status help

SUPABASE := npx supabase
FRONTEND_ENV := frontend/.env.local

## Supabase 起動 + anon key を frontend/.env.local に反映
local-supabase-up:
	@echo "==> Starting Supabase local stack..."
	$(SUPABASE) start
	@$(MAKE) sync-anon-key
	@echo ""
	@echo "Supabase Studio: http://localhost:54323"
	@echo "Inbucket (mail): http://localhost:54324"

## Supabase 停止
local-supabase-down:
	@echo "==> Stopping Supabase local stack..."
	$(SUPABASE) stop

## supabase status から anon key を取り出し、frontend/.env.local の NEXT_PUBLIC_SUPABASE_ANON_KEY を更新
sync-anon-key:
	@echo "==> Syncing anon key into $(FRONTEND_ENV)..."
	@if [ ! -f $(FRONTEND_ENV) ]; then \
		echo "ERROR: $(FRONTEND_ENV) が存在しません。まず 'cp frontend/.env.example $(FRONTEND_ENV)' を実行してください。"; \
		exit 1; \
	fi
	@ANON_KEY=$$($(SUPABASE) status -o env 2>/dev/null | grep '^ANON_KEY=' | sed 's/^ANON_KEY=//' | tr -d '"'); \
	if [ -z "$$ANON_KEY" ]; then \
		echo "ERROR: anon key を取得できませんでした。'supabase start' が完了しているか確認してください。"; \
		exit 1; \
	fi; \
	if [ "$$(uname)" = "Darwin" ]; then \
		sed -i '' "s|^NEXT_PUBLIC_SUPABASE_ANON_KEY=.*|NEXT_PUBLIC_SUPABASE_ANON_KEY=$$ANON_KEY|" $(FRONTEND_ENV); \
	else \
		sed -i "s|^NEXT_PUBLIC_SUPABASE_ANON_KEY=.*|NEXT_PUBLIC_SUPABASE_ANON_KEY=$$ANON_KEY|" $(FRONTEND_ENV); \
	fi; \
	echo "OK: NEXT_PUBLIC_SUPABASE_ANON_KEY を更新しました"

## Supabase + アプリの Docker を起動（推奨ワンショット）
local-up: local-supabase-up
	@echo "==> Starting application containers..."
	docker compose up -d
	@echo ""
	@echo "Frontend: http://localhost:3000"
	@echo "Backend:  http://localhost:8080"

## アプリの Docker のみ停止（Supabase は維持）
local-down:
	@echo "==> Stopping application containers..."
	docker compose down

## アプリ + Supabase を全停止
local-stop: local-down local-supabase-down

## 各サービスの状態を表示
local-status:
	@echo "==> Supabase status:"
	@$(SUPABASE) status || true
	@echo ""
	@echo "==> Docker compose status:"
	@docker compose ps

help:
	@echo "Local development targets:"
	@echo "  make local-up            - Supabase + application containers up"
	@echo "  make local-down          - Stop application containers only"
	@echo "  make local-stop          - Stop application + Supabase"
	@echo "  make local-supabase-up   - Start Supabase only (and sync anon key)"
	@echo "  make local-supabase-down - Stop Supabase only"
	@echo "  make sync-anon-key       - Re-sync anon key into $(FRONTEND_ENV)"
	@echo "  make local-status        - Show service status"
