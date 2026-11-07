package middleware

import "github.com/gin-gonic/gin"

// SecurityHeadersMiddleware はすべての応答にセキュリティ用のヘッダーを付ける。
//   - X-Content-Type-Options: ブラウザが JSON を別の形式と推測して扱わないようにする
//   - Cache-Control: 応答をブラウザや途中のキャッシュに残さない。
//     認証付きの応答だけでなく、ログイン状態で内容が変わる公開APIの応答もあるため、一律に付ける
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
