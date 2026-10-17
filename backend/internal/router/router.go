package router

import (
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
	"myanimelogs/internal/handlers"
	"myanimelogs/internal/middleware"
	"myanimelogs/internal/services"

	"github.com/gin-gonic/gin"
)

func Setup(db *database.DB, cfg *config.Config, syncScheduler handlers.SyncScheduler, logger *common.Logger, annictClient *services.AnnictClient) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	errorMonitor := common.NewErrorMonitor(
		logger,
		cfg.ErrorAlertThreshold,
		time.Duration(cfg.ErrorTimeWindow)*time.Minute,
	)

	r := gin.New()

	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.CORSMiddleware())

	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, time.Second)
	r.Use(middleware.RateLimitMiddleware(rateLimiter))

	animeService := services.NewAnimeService(db, logger)
	// service_role キー未設定の環境でも起動は妨げない。
	// アカウント削除エンドポイントのみが 503 を返す。
	supabaseAdmin := services.NewSupabaseAdminClient(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey)
	if !supabaseAdmin.IsConfigured() {
		logger.Warn("SUPABASE_SERVICE_ROLE_KEY is not set; account deletion is disabled")
	}
	authService := services.NewAuthServiceWithAdmin(db, supabaseAdmin)

	authHandler := handlers.NewAuthHandler(authService, logger, errorMonitor)
	animeHandler := handlers.NewAnimeHandler(animeService, logger, errorMonitor)
	annictHandler := handlers.NewAnnictHandler(annictClient, logger, errorMonitor)
	adminHandler := handlers.NewAdminHandler(animeService, logger, errorMonitor, syncScheduler)
	monitoringHandler := handlers.NewMonitoringHandler(logger, errorMonitor)

	r.GET("/health", monitoringHandler.HealthCheck)
	r.GET("/health/detailed", monitoringHandler.DetailedHealthCheck)

	if gin.IsDebugging() {
		r.GET("/test-hot-reload", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message":   "Hot reload is working!",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})
		r.GET("/annict/test", annictHandler.TestConnection)
	}

	// Supabase JWT ミドルウェア
	authRequired := middleware.SupabaseAuthMiddleware(db, cfg.SupabaseJWTSecret, logger)
	authOptional := middleware.OptionalSupabaseAuthMiddleware(db, cfg.SupabaseJWTSecret, logger)

	api := r.Group("/api")
	{
		// アニメルート（認証オプション）
		api.GET("/animes", authOptional, animeHandler.GetAnimes)
		api.GET("/animes/:id", authOptional, animeHandler.GetAnime)

		// 検索ルート
		api.GET("/search/animes", authOptional, animeHandler.GetAnimes)
		api.GET("/search/advanced", authOptional, animeHandler.AdvancedSearchAnimes)
		api.GET("/search/suggestions", animeHandler.GetSearchSuggestions)
		api.GET("/search/popular", animeHandler.GetPopularSearches)

		// フィルタールート
		api.GET("/filters", animeHandler.GetFilterOptions)

		// 今期アニメ・人気アニメルート
		api.GET("/animes/current", animeHandler.GetCurrentSeasonAnimes)
		api.GET("/animes/popular", animeHandler.GetPopularAnimes)

		// Annictルート
		api.GET("/annict/works", annictHandler.GetWorksBySeason)
		api.GET("/annict/popular-works", annictHandler.GetPopularWorks)

		// 管理者ルート（認証 + 管理者ロール必須）
		admin := api.Group("/admin")
		admin.Use(authRequired, middleware.RequireAdmin())
		{
			admin.POST("/sync/season", annictHandler.SyncWorksBySeason)
			admin.POST("/sync/popular", annictHandler.SyncPopularWorks)
			admin.POST("/sync/genres", adminHandler.SyncGenresForAllAnimes)
			admin.POST("/sync/statuses", adminHandler.SyncAnimeStatuses)
			admin.POST("/sync/released-on", adminHandler.SyncReleasedOn)
			admin.GET("/sync/status", adminHandler.GetSyncStatus)
			admin.POST("/cache/clear", adminHandler.ClearCache)
			admin.GET("/errors/metrics", monitoringHandler.GetErrorMetrics)
			admin.GET("/errors/top", monitoringHandler.GetTopErrors)
			admin.POST("/errors/reset", monitoringHandler.ResetErrorMetrics)
		}

		// 保護されたルート（認証必須）
		protected := api.Group("/")
		protected.Use(authRequired)
		{
			// ユーザールート
			protected.GET("/users/me", authHandler.GetMe)
			// アカウントの完全削除（復旧不可）
			protected.DELETE("/users/me", authHandler.DeleteMe)

			// ユーザーのお気に入り
			protected.GET("/users/me/favorites", animeHandler.GetFavorites)
			protected.PUT("/users/me/favorites/:anime_id", animeHandler.UpdateFavorite)
			protected.DELETE("/users/me/favorites/:anime_id", animeHandler.DeleteFavorite)

			// 視聴ステータス
			protected.GET("/users/me/watch-status", animeHandler.GetWatchStatuses)
			protected.PUT("/users/me/watch-status/:anime_id", animeHandler.UpdateWatchStatus)
			protected.DELETE("/users/me/watch-status/:anime_id", animeHandler.DeleteWatchStatus)

			// 検索履歴
			protected.GET("/search/history", animeHandler.GetSearchHistory)
			protected.DELETE("/search/history", animeHandler.DeleteSearchHistory)
			protected.DELETE("/search/history/:id", animeHandler.DeleteSearchHistoryItem)
		}
	}

	// アイコンはプリセット方式へ移行し、フロントエンドの静的アセットとして
	// 配信するため、バックエンドでのファイル配信は廃止した（#763）
	return r
}
