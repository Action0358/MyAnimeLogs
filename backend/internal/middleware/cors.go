package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()

	// デフォルトの許可オリジン
	defaultOrigins := []string{
		"http://localhost:3000",           // React development server
		"http://localhost:3001",           // React development server (alternative port)
		"https://stg.myanimelogs.com",     // Staging environment
		"https://myanimelogs.com",         // Production environment
		"https://www.myanimelogs.com",     // Production environment (www)
		"https://myanimelogs.pages.dev",   // Cloudflare Pages (if used)
		"https://myanimelogs.netlify.app", // Netlify (if used)
	}

	// 環境変数から追加のオリジンを取得
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		additionalOrigins := strings.Split(envOrigins, ",")
		for _, origin := range additionalOrigins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				defaultOrigins = append(defaultOrigins, origin)
			}
		}
	}

	config.AllowOrigins = defaultOrigins
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	return cors.New(config)
}
