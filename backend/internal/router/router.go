package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
	"myanimelogs/internal/handlers"
	"myanimelogs/internal/middleware"
	"myanimelogs/internal/services"
)

func Setup(db *database.DB, cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware
	r.Use(middleware.CORSMiddleware())
	
	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, time.Second)
	r.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Services
	authService := services.NewAuthService(db, cfg.JWTSecret)
	animeService := services.NewAnimeService(db)
	annictService := services.NewAnnictClient(cfg.AnnictAPIKey) // Annictサービスを初期化

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	animeHandler := handlers.NewAnimeHandler(animeService)
	annictHandler := handlers.NewAnnictHandler(annictService) // Annictハンドラを初期化

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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
		api.GET("/search/animes", middleware.OptionalAuthMiddleware(cfg.JWTSecret), animeHandler.GetAnimes)

		// Annict routes
		api.GET("/annict/works", annictHandler.GetWorksBySeason)
		api.GET("/annict/popular-works", annictHandler.GetPopularWorks)

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