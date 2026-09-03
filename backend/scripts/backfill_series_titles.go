//go:build ignore
// +build ignore

// 既存アニメの series_title を一括で設定するバックフィルスクリプト（一回限り実行用）。
// #755: 詳細ページの「同シリーズの他シーズン」表示の前提となるデータを揃える。
// 外部APIを呼ばない純粋な文字列処理（NormalizeAnimeTitle）のため、翻訳バックフィルと
// 異なりコスト・レート制限の考慮は不要。
//
// 使い方:
//
//	# 件数のみ確認（DB更新なし）
//	go run scripts/backfill_series_titles.go -env=staging -dry-run
//
//	# 本番フルバックフィル（DATABASE_URLを本番向けに設定して実行）
//	go run scripts/backfill_series_titles.go -env=production
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	env := flag.String("env", "local", "環境: local, staging, production")
	dryRun := flag.Bool("dry-run", false, "集計のみ（DB更新なし）")
	limit := flag.Int("limit", 0, "先頭N件のみ処理（0=全件）")
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
	defer animeService.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("Backfill start", "env", *env, "dry_run", *dryRun, "limit", *limit)

	result, err := animeService.BackfillSeriesTitles(ctx, *dryRun, *limit)

	if result != nil {
		fmt.Println("\n==== Backfill Result ====")
		fmt.Printf("対象(series_title未設定): %d 件\n", result.Total)
		if *dryRun {
			fmt.Println("※ DryRunのためDBは更新していません。実行するには -dry-run を外してください。")
		} else {
			fmt.Printf("更新: %d / 失敗: %d\n", result.Updated, result.Failed)
		}
	}

	if err != nil {
		logger.Error("backfill aborted", "error", err)
		os.Exit(1)
	}
}
