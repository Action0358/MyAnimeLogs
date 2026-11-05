package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware はハンドラー内の panic を 500 の応答に変える。
// panic の内容とスタックはログだけに残し、応答には含めない。
//
// CORS より前に置いても、CORS が設定したヘッダーは応答に残る。
// LoggingMiddleware より後ろに置き、panic したリクエストも 500 として記録されるようにする。
func RecoveryMiddleware(errorMonitor *common.ErrorMonitor) gin.HandlerFunc {
	// ログを RecordError にまとめるため、Gin 自身の出力は止める(nil)。
	// 切断された接続への書き込みによる panic は、Gin がここを呼ばずに処理する
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		errorMonitor.RecordError(c.Request.Context(), common.ErrCodeInternalServer,
			fmt.Errorf("panic: %v", recovered),
			map[string]interface{}{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"request_id": c.GetString("request_id"),
				"stack":      string(debug.Stack()),
			})

		// 応答を書き始めた後は 500 に変えられない。正常に終えると途中までの本文が
		// 完全な応答に見えるので、接続を切って失敗をクライアントに伝える
		if c.Writer.Written() {
			panic(http.ErrAbortHandler)
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, common.ErrorResponse{
			Error:   common.ErrMsgInternalServer,
			Message: common.ErrMsgInternalServer,
			Code:    common.ErrCodeInternalServer,
		})
	})
}
