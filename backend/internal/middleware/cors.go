package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()

	// 環境変数から許可オリジンを取得
	allowedOrigins := []string{}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		origins := strings.Split(envOrigins, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}

	// 環境変数が設定されていない場合のフォールバック（開発環境用）
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{
			"http://localhost:3000", // Local development
			"http://localhost:3001", // Local development (alternative port)
		}
	}

	// Cloudflare Pages プレビューデプロイの動的許可
	// *.pages.dev ドメインを許可する場合
	config.AllowOriginFunc = func(origin string) bool {
		// 明示的に許可されたオリジン
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}

		// Cloudflare Pages プレビューデプロイを許可
		// 例: https://abc123.myanimelogs.pages.dev
		if strings.HasSuffix(origin, ".myanimelogs.pages.dev") && strings.HasPrefix(origin, "https://") {
			return true
		}

		return false
	}

	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	return cors.New(config)
}
