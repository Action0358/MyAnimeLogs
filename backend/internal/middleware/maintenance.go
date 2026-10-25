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

// MaintenanceStatusPath はメンテナンス状態を返すエンドポイント。
// フロントのエッジ（Cloudflare Pages Functions）が、HTML を返す前にここを見て
// メンテナンス画面に差し替えるかを決める。
const MaintenanceStatusPath = "/api/maintenance"

// isMaintenanceExempt はメンテナンス中も応答するリクエストか。
//   - /health: Fly.io のヘルスチェックと監視。503 を返すと機械が再起動を繰り返す
//   - /api/maintenance: メンテナンス状態そのものを返すため
//   - /api/admin/: 管理者（認証必須）が同期などの作業を続けられるように
//   - OPTIONS: CORS のプリフライト（CORSMiddleware が応答する）
func isMaintenanceExempt(c *gin.Context) bool {
	if c.Request.Method == http.MethodOptions {
		return true
	}
	path := c.Request.URL.Path
	return path == "/health" || path == MaintenanceStatusPath || strings.HasPrefix(path, "/api/admin/")
}

// MaintenanceStatusHandler はメンテナンス状態を 200 で返す（メンテナンス中も応答する）。
// 状態は起動時の設定で決まるため、切り替えは fly secrets（再起動を伴う）で行う。
func MaintenanceStatusHandler(enabled bool, endsAt *time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := gin.H{"enabled": enabled}
		if enabled && endsAt != nil {
			body["ends_at"] = endsAt.Format(time.RFC3339)
		}
		// キャッシュの長さはエッジ側で決める。経路上で古い状態が残らないようにする
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, body)
	}
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
