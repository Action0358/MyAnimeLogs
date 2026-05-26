package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Cloudflare Pages のホスティング先（本番・staging）を動的に許可するための基底ドメイン。
// apex (https://example.pages.dev) とサブドメイン (https://abc.example.pages.dev) の両方を許可する。
var cloudflarePagesDomains = []string{
	"myanimelogs.pages.dev",
	"myanimelogs-staging.pages.dev",
}

const httpsPrefix = "https://"

// isOriginAllowed は CORS の許可判定を行う。
// 明示的に許可されたオリジンの完全一致、または Cloudflare Pages の apex/サブドメインを許可する。
// 全リクエストパスで呼ばれるため、文字列結合を避けてゼロアロケーションで判定する。
func isOriginAllowed(origin string, allowedOrigins, pagesDomains []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	if !strings.HasPrefix(origin, httpsPrefix) {
		return false
	}
	for _, domain := range pagesDomains {
		if !strings.HasSuffix(origin, domain) {
			continue
		}
		// 1) apex 一致: "https://" + domain ぴったり
		if len(origin) == len(httpsPrefix)+len(domain) {
			return true
		}
		// 2) サブドメイン一致: 直前が '.' であること。
		//    "https://evilmyanimelogs.pages.dev" のような偽装は直前が '.' でないため弾かれる。
		sepIdx := len(origin) - len(domain) - 1
		if sepIdx >= len(httpsPrefix) && origin[sepIdx] == '.' {
			return true
		}
	}
	return false
}

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

	config.AllowOriginFunc = func(origin string) bool {
		return isOriginAllowed(origin, allowedOrigins, cloudflarePagesDomains)
	}

	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	return cors.New(config)
}
