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

func Setup(db *database.DB, cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize structured logger
	logger := common.NewLogger(common.LogLevel(cfg.LogLevel), cfg.LogFormat)

	// Initialize error monitor
	errorMonitor := common.NewErrorMonitor(
		logger,
		cfg.ErrorAlertThreshold,
		time.Duration(cfg.ErrorTimeWindow)*time.Minute,
	)

	r := gin.New() // Use gin.New() instead of gin.Default() for custom middleware

	// Middleware (order matters)
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.CORSMiddleware())

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, time.Second)
	r.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Services
	authService := services.NewAuthService(db, cfg.JWTSecret)
	animeService := services.NewAnimeService(db)
	annictService := services.NewAnnictClientWithDB(cfg.AnnictAPIKey, db) // Annictサービスを初期化

	// Handlers with enhanced error handling
	authHandler := handlers.NewAuthHandler(authService, logger, errorMonitor)
	animeHandler := handlers.NewAnimeHandler(animeService, logger, errorMonitor)
	annictHandler := handlers.NewAnnictHandler(annictService, logger, errorMonitor)
	monitoringHandler := handlers.NewMonitoringHandler(logger, errorMonitor)

	// Health check and monitoring endpoints
	r.GET("/health", monitoringHandler.HealthCheck)
	r.GET("/health/detailed", monitoringHandler.DetailedHealthCheck)

	if gin.IsDebugging() {
		// Hot reload test endpoint
		r.GET("/test-hot-reload", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message":   "Hot reload is working! 🔥",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})
		// Annict API connection test endpoint (dev only)
		r.GET("/annict/test", annictHandler.TestConnection)
	}

	// API routes
	api := r.Group("/api")
	{
		// Public routes (no authentication required)
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		// Anime routes (optional authentication)
		api.GET("/animes", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnimes)
		api.GET("/animes/:id", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnime)
		api.GET("/animes/:id/episodes", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnimeEpisodes)

		// Search routes
		api.GET("/search/animes", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnimes)
		api.GET("/search/advanced", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.AdvancedSearchAnimes)
		api.GET("/search/suggestions", animeHandler.GetSearchSuggestions)
		api.GET("/search/popular", animeHandler.GetPopularSearches)

		// Filter routes
		api.GET("/filters", animeHandler.GetFilterOptions)

		// Annict routes
		api.GET("/annict/works", annictHandler.GetWorksBySeason)
		api.GET("/annict/popular-works", annictHandler.GetPopularWorks)

		// Admin routes for data synchronization and monitoring (protected with authentication)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Data synchronization
			admin.POST("/sync/season", annictHandler.SyncWorksBySeason)
			admin.POST("/sync/popular", annictHandler.SyncPopularWorks)

			// Error monitoring
			admin.GET("/errors/metrics", monitoringHandler.GetErrorMetrics)
			admin.GET("/errors/top", monitoringHandler.GetTopErrors)
			admin.POST("/errors/reset", monitoringHandler.ResetErrorMetrics)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User routes
			protected.GET("/users/me", authHandler.GetMe)
			protected.PUT("/users/me", authHandler.UpdateMe)
			protected.GET("/users/me/stats", authHandler.GetStats)
			protected.GET("/users/me/export", authHandler.ExportData)

			// User favorites
			protected.GET("/users/me/favorites", animeHandler.GetFavorites)
			protected.PUT("/users/me/favorites/:anime_id", animeHandler.UpdateFavorite)
			protected.DELETE("/users/me/favorites/:anime_id", animeHandler.DeleteFavorite)

			// Watch status
			protected.PUT("/episodes/:episode_id/watch", animeHandler.UpdateWatchStatus)

			// Ratings
			protected.PUT("/animes/:id/rating", animeHandler.UpdateRating)
			protected.DELETE("/animes/:id/rating", animeHandler.DeleteRating)

			// Watch history and watchlist
			protected.GET("/users/me/recently-watched", animeHandler.GetRecentlyWatched)
			protected.GET("/users/me/watchlist", animeHandler.GetWatchlist)
		}
	}

	return r
}
