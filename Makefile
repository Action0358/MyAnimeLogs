# ルート Makefile — ローカル開発環境のオーケストレーション
#
# 主要ターゲット:
#   make local-up           # Supabase + アプリ Docker を全部起動（推奨）
#   make local-down         # アプリ Docker のみ停止（Supabase は残す）
#   make local-stop         # Supabase + アプリ Docker を全停止
#   make local-supabase-up  # Supabase CLI のみ起動 + anon key を frontend/.env.local に反映
#   make local-supabase-down# Supabase CLI のみ停止
#   make local-recover      # 停止済み Supabase コンテナを再作成せず起動（Docker Desktop 再起動後の復旧）
#   make sync-anon-key      # Supabase 起動中の anon key を frontend/.env.local に反映

.PHONY: local-up local-down local-stop local-supabase-up local-supabase-down local-recover sync-anon-key local-status help

# package.json でピン留めした版を使う。グローバル CLI を優先すると config.toml の
# スキーマやマイグレーション挙動が版ズレし、「手元だけ動く/動かない」の原因になる。
# node_modules 未インストール時のみ npx にフォールバック。
SUPABASE := $(shell [ -x node_modules/.bin/supabase ] && echo node_modules/.bin/supabase || echo "npx --yes supabase")
FRONTEND_ENV := frontend/.env.local
SUPABASE_PROJECT := myanimelogs-supabase

## Supabase 起動 + anon key を frontend/.env.local に反映
local-supabase-up:
	@echo "==> Starting Supabase local stack..."
	$(SUPABASE) start
	@$(MAKE) sync-anon-key
	@echo ""
	@echo "Supabase Studio: http://localhost:54323"
	@echo "Mailpit (mail):  http://localhost:54324"

## Supabase 停止
local-supabase-down:
	@echo "==> Stopping Supabase local stack..."
	$(SUPABASE) stop

## Docker Desktop 再起動などで Supabase コンテナが落ちた場合の復旧。
## この状態では `supabase start` が「already running」と誤認して復旧できないため、
## 既存コンテナを再作成せずそのまま起動する（DB のデータを保持するため stop/start は使わない）。
local-recover:
	@echo "==> Recovering stopped Supabase containers (data preserved)..."
	@docker start supabase_vector_$(SUPABASE_PROJECT) supabase_db_$(SUPABASE_PROJECT) >/dev/null
	@echo "==> Waiting for database to become healthy..."
	@until [ "$$(docker inspect -f '{{.State.Health.Status}}' supabase_db_$(SUPABASE_PROJECT) 2>/dev/null)" = "healthy" ]; do sleep 2; done
	@echo "==> Starting remaining Supabase services..."
	@docker start $$(docker ps -aq --filter label=com.supabase.cli.project=$(SUPABASE_PROJECT)) >/dev/null
	@echo "==> Starting application containers..."
	docker compose up -d
	@echo ""
	@echo "OK: 復旧しました。ヘルスチェック: curl http://localhost:8080/health"

## supabase status から anon key を取り出し、frontend/.env.local の NEXT_PUBLIC_SUPABASE_ANON_KEY を更新
sync-anon-key:
	@echo "==> Syncing anon key into $(FRONTEND_ENV)..."
	@if [ ! -f $(FRONTEND_ENV) ]; then \
		echo "ERROR: $(FRONTEND_ENV) が存在しません。まず 'cp frontend/env/local.example $(FRONTEND_ENV)' を実行してください。"; \
		exit 1; \
	fi
	@ANON_KEY=$$($(SUPABASE) status -o env | grep '^ANON_KEY=' | sed 's/^ANON_KEY=//' | tr -d '"\r'); \
	if [ -z "$$ANON_KEY" ]; then \
		echo "ERROR: anon key を取得できませんでした。'make local-supabase-up' が完了しているか確認してください。"; \
		echo "       コンテナが停止している場合は 'make local-recover' を試してください。"; \
		exit 1; \
	fi; \
	if grep -q '^NEXT_PUBLIC_SUPABASE_ANON_KEY=' $(FRONTEND_ENV); then \
		if [ "$$(uname)" = "Darwin" ]; then \
			sed -i '' "s|^NEXT_PUBLIC_SUPABASE_ANON_KEY=.*|NEXT_PUBLIC_SUPABASE_ANON_KEY=$$ANON_KEY|" $(FRONTEND_ENV); \
		else \
			sed -i "s|^NEXT_PUBLIC_SUPABASE_ANON_KEY=.*|NEXT_PUBLIC_SUPABASE_ANON_KEY=$$ANON_KEY|" $(FRONTEND_ENV); \
		fi; \
	else \
		echo "" >> $(FRONTEND_ENV); \
		echo "NEXT_PUBLIC_SUPABASE_ANON_KEY=$$ANON_KEY" >> $(FRONTEND_ENV); \
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
	@echo "  make local-recover       - Restart stopped Supabase containers (after Docker restart)"
	@echo "  make sync-anon-key       - Re-sync anon key into $(FRONTEND_ENV)"
	@echo "  make local-status        - Show service status"
