package router

import (
	"path/filepath"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
	"myanimelogs/internal/handlers"
	"myanimelogs/internal/middleware"
	"myanimelogs/internal/services"

	"github.com/gin-gonic/gin"
)

func Setup(db *database.DB, cfg *config.Config) *gin.Engine {
	// Ginモードを設定
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 構造化ロガーを初期化
	logger := common.NewLogger(common.LogLevel(cfg.LogLevel), cfg.LogFormat)

	// エラーモニターを初期化
	errorMonitor := common.NewErrorMonitor(
		logger,
		cfg.ErrorAlertThreshold,
		time.Duration(cfg.ErrorTimeWindow)*time.Minute,
	)

	r := gin.New() // カスタムミドルウェアのためgin.Default()の代わりにgin.New()を使用

	// ミドルウェア（順序が重要）
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.CORSMiddleware())

	// レート制限
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, time.Second)
	r.Use(middleware.RateLimitMiddleware(rateLimiter))

	// サービス
	authService := services.NewAuthService(db, cfg.JWTSecret)
	annictService := services.NewAnnictClientWithDB(cfg.AnnictAPIKey, db, logger) // Annictサービスを初期化
	animeService := services.NewAnimeService(db, logger, annictService)
	uploadService := services.NewUploadService(cfg.UploadDir, cfg.UploadPublicPath, cfg.MaxUploadSize, logger)

	// 強化されたエラーハンドリングを持つハンドラー
	authHandler := handlers.NewAuthHandlerWithUpload(authService, uploadService, logger, errorMonitor)
	animeHandler := handlers.NewAnimeHandler(animeService, logger, errorMonitor)
	annictHandler := handlers.NewAnnictHandler(annictService, logger, errorMonitor)
	adminHandler := handlers.NewAdminHandler(animeService, logger, errorMonitor)
	monitoringHandler := handlers.NewMonitoringHandler(logger, errorMonitor)

	// ヘルスチェックと監視エンドポイント
	r.GET("/health", monitoringHandler.HealthCheck)
	r.GET("/health/detailed", monitoringHandler.DetailedHealthCheck)

	if gin.IsDebugging() {
		// ホットリロードテストエンドポイント
		r.GET("/test-hot-reload", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message":   "Hot reload is working! 🔥",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})
		// Annict API接続テストエンドポイント（開発環境のみ）
		r.GET("/annict/test", annictHandler.TestConnection)
	}

	// APIルート
	api := r.Group("/api")
	{
		// 公開ルート（認証不要）
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		// アニメルート（認証オプション）
		api.GET("/animes", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnimes)
		api.GET("/animes/:id", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnime)
		api.GET("/animes/:id/episodes", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnimeEpisodes)

		// 検索ルート
		api.GET("/search/animes", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnimes)
		api.GET("/search/advanced", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.AdvancedSearchAnimes)
		api.GET("/search/suggestions", animeHandler.GetSearchSuggestions)
		api.GET("/search/popular", animeHandler.GetPopularSearches)

		// フィルタールート
		api.GET("/filters", animeHandler.GetFilterOptions)

		// Annictルート
		api.GET("/annict/works", annictHandler.GetWorksBySeason)
		api.GET("/annict/popular-works", annictHandler.GetPopularWorks)

		// データ同期と監視のための管理者ルート（認証で保護）
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// データ同期
			admin.POST("/sync/season", annictHandler.SyncWorksBySeason)
			admin.POST("/sync/popular", annictHandler.SyncPopularWorks)
			admin.POST("/sync/episode-durations", adminHandler.SyncEpisodeDurations)
			admin.POST("/create-episodes", adminHandler.CreateEpisodesForAllAnimes)
			admin.POST("/enrich/episode-titles", adminHandler.EnrichEpisodeTitles)

			// エラー監視
			admin.GET("/errors/metrics", monitoringHandler.GetErrorMetrics)
			admin.GET("/errors/top", monitoringHandler.GetTopErrors)
			admin.POST("/errors/reset", monitoringHandler.ResetErrorMetrics)
		}

		// 保護されたルート（認証必須）
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// ユーザールート
			protected.GET("/users/me", authHandler.GetMe)
			protected.PUT("/users/me", authHandler.UpdateMe)
			protected.POST("/users/me/icon", authHandler.UploadIcon)
			protected.GET("/users/me/stats", authHandler.GetStats)
			protected.GET("/users/me/export", authHandler.ExportData)

			// ユーザーのお気に入り
			protected.GET("/users/me/favorites", animeHandler.GetFavorites)
			protected.PUT("/users/me/favorites/:anime_id", animeHandler.UpdateFavorite)
			protected.DELETE("/users/me/favorites/:anime_id", animeHandler.DeleteFavorite)

			// 視聴ステータス
			protected.PUT("/episodes/:episode_id/watch", animeHandler.UpdateWatchStatus)

			// 評価
			protected.PUT("/animes/:id/rating", animeHandler.UpdateRating)
			protected.DELETE("/animes/:id/rating", animeHandler.DeleteRating)

			// 視聴履歴とウォッチリスト
			protected.GET("/users/me/recently-watched", animeHandler.GetRecentlyWatched)
			protected.GET("/users/me/watchlist", animeHandler.GetWatchlist)
		}
	}

	// 静的ファイル配信（アップロードされた画像）
	// 設定から公開パスとファイルシステムパスを取得
	uploadRootDir := filepath.Dir(cfg.UploadDir) // "./uploads/icons" -> "./uploads"
	r.Static(cfg.UploadPublicPath, uploadRootDir)

	return r
}
