package middleware

import (
	"net/http"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

// MaxRequestBodyBytes はリクエスト本文の上限。
// 本文を受け取るAPIは視聴ステータスの更新だけで、数十バイトしか使わない。
const MaxRequestBodyBytes int64 = 64 << 10

// BodyLimitMiddleware はリクエスト本文の大きさを制限する。
// Content-Length で超過が分かるものはここで 413 を返す。
// チャンク転送は MaxBytesReader が読み取り時にエラーを返すため、本文を読むハンドラーで処理する。
func BodyLimitMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, common.ErrorResponse{
				Error:   common.ErrMsgRequestTooLarge,
				Message: common.ErrMsgRequestTooLarge,
				Code:    common.ErrCodeRequestTooLarge,
			})
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
