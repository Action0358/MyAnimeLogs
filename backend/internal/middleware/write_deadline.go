package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ClearWriteDeadline はこのリクエストの接続から書き込み期限を外す。
// http.Server の WriteTimeout を超えて動く処理の応答を届けるために使う。
// 期限が無くなるので、認証と権限の確認より後に置くこと。
// 次のリクエストではサーバーが期限を設定し直すため、効果はこのリクエスト限り。
func ClearWriteDeadline() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 失敗するのは期限を扱えない ResponseWriter のときだけで、その場合は外す期限も無い
		_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})
		c.Next()
	}
}
