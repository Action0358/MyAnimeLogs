//go:build ignore
// +build ignore

// 既存アニメのあらすじを一括で整形・日本語化するバックフィルスクリプト（一回限り実行用）。
//
// 使い方:
//
//	# 件数・コスト概算のみ（DB更新・翻訳APIなし）
//	go run scripts/backfill_descriptions.go -env=staging -dry-run
//
//	# 少数サンプルで動作検証（5件だけ実翻訳）
//	go run scripts/backfill_descriptions.go -env=staging -limit=5
//
//	# 本番フルバックフィル（DATABASE_URL/ANTHROPIC_API_KEY を本番向けに設定して実行）
//	go run scripts/backfill_descriptions.go -env=production
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/services"

	"github.com/joho/godotenv"
)

// translationUnitCostUSD は dry-run のコスト概算に使う1件あたりの目安（Haiku 4.5）。
const translationUnitCostUSD = 0.003

func main() {
	env := flag.String("env", "local", "環境: local, staging, production")
	dryRun := flag.Bool("dry-run", false, "集計のみ（DB更新・翻訳API呼び出しなし）")
	limit := flag.Int("limit", 0, "先頭N件のみ処理（0=全件）")
	throttleMs := flag.Int("throttle-ms", 200, "翻訳API呼び出し間の待機ミリ秒")
	flag.Parse()

	logger := common.NewLogger(common.LogLevelInfo, "text")

	var envFile string
	switch *env {
	case "local":
		envFile = ".env.local"
	case "staging":
		envFile = ".env.staging"
	case "production":
		envFile = ".env.production"
	default:
		logger.Error("invalid env (use: local, staging, production)", "env", *env)
		os.Exit(1)
	}

	if err := godotenv.Load(envFile); err != nil {
		logger.Warn("環境ファイルが見つかりません。システム環境変数を使用します", "file", envFile)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	db, err := database.New(dbURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	animeService := services.NewAnimeService(db, logger)

	// Ctrl+C / SIGTERM で安全に中断し、ここまでの結果を出力できるようにする
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("Backfill start", "env", *env, "dry_run", *dryRun, "limit", *limit, "throttle_ms", *throttleMs)

	result, err := animeService.BackfillDescriptions(ctx, services.BackfillOptions{
		DryRun:   *dryRun,
		Limit:    *limit,
		Throttle: time.Duration(*throttleMs) * time.Millisecond,
	})
	if err != nil {
		logger.Error("backfill failed", "error", err)
		os.Exit(1)
	}
	if result.Processed < result.Total {
		logger.Warn("処理が中断されました（全件未処理）", "processed", result.Processed, "total", result.Total)
	}

	fmt.Println("\n==== Backfill Result ====")
	fmt.Printf("対象(非空description): %d 件 / 処理済み: %d 件\n", result.Total, result.Processed)
	fmt.Printf("翻訳候補(英語あらすじ): %d 件\n", result.PendingTranslation)
	if *dryRun {
		fmt.Printf("[dry-run] 翻訳見込み: %d 件 / 概算コスト: 約 $%.2f\n",
			result.PendingTranslation, float64(result.PendingTranslation)*translationUnitCostUSD)
		fmt.Printf("[dry-run] 書式のみ更新見込み(日本語): %d 件\n", result.FormattedOnly)
		fmt.Println("※ DBは更新していません。実行するには -dry-run を外してください。")
	} else {
		fmt.Printf("翻訳実施: %d / 書式のみ更新: %d / 変更なし: %d / 失敗: %d\n",
			result.Translated, result.FormattedOnly, result.AlreadyClean, result.Failed)
	}
}
