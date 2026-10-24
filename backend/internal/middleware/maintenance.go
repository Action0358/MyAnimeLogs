package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

// 通常時に Retry-After を返せない（終了予定なし）ときの目安秒数
const defaultMaintenanceRetryAfterSeconds = 300

// isMaintenanceExempt はメンテナンス中も応答するリクエストか。
//   - /health: Fly.io のヘルスチェックと監視。503 を返すと機械が再起動を繰り返す
//   - /api/admin/: 管理者（認証必須）が同期などの作業を続けられるように
//   - OPTIONS: CORS のプリフライト（CORSMiddleware が応答する）
func isMaintenanceExempt(c *gin.Context) bool {
	if c.Request.Method == http.MethodOptions {
		return true
	}
	path := c.Request.URL.Path
	return path == "/health" || strings.HasPrefix(path, "/api/admin/")
}

// MaintenanceMiddleware はメンテナンスモード中、免除対象以外に 503 + code=MAINTENANCE を返す。
// enabled=false のときは何もしない。endsAt があれば ends_at と Retry-After に反映する。
func MaintenanceMiddleware(enabled bool, endsAt *time.Time) gin.HandlerFunc {
	if !enabled {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if isMaintenanceExempt(c) {
			c.Next()
			return
		}

		retryAfter := defaultMaintenanceRetryAfterSeconds
		body := gin.H{
			"error":   common.ErrMsgMaintenance,
			"message": common.ErrMsgMaintenance,
			"code":    common.ErrCodeMaintenance,
		}
		if endsAt != nil {
			body["ends_at"] = endsAt.Format(time.RFC3339)
			if remaining := int(time.Until(*endsAt).Seconds()); remaining > 0 {
				retryAfter = remaining
			}
		}

		c.Header("Retry-After", strconv.Itoa(retryAfter))
		// 解除後に古い 503 が残らないよう、ブラウザ・CDNにキャッシュさせない
		c.Header("Cache-Control", "no-store")
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, body)
	}
}
