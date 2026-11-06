package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
)

// クライアントから受け取る X-Request-ID の形式。英数字とハイフンだけを、64文字まで許す。
// ログやレスポンスヘッダーに任意の文字列が入らないようにするため
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

// RequestIDMiddleware generates and adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader("X-Request-ID")

		// 指定がないか、形式に合わない場合は新しく生成する
		if !validRequestID.MatchString(requestID) {
			requestID = generateRequestID()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Failure to read from crypto/rand indicates a serious system problem
		// with the entropy source. The application should not continue.
		// Wrap the error to preserve structure before converting to string for panic
		wrappedErr := fmt.Errorf("failed to generate random bytes for request ID: %w", err)
		panic(wrappedErr.Error())
	}
	return hex.EncodeToString(bytes)
}
