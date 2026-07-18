package middleware

import (
	"net/http"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

// RequireAdmin ensures the authenticated user has the admin role.
//
// It must run after SupabaseAuthMiddleware, which validates the JWT and sets
// "is_admin" in the context based on the token's app_metadata.role. Requests
// from non-admin (or unauthenticated) users are rejected with 403 Forbidden.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("is_admin") {
			c.JSON(http.StatusForbidden, common.ErrorResponse{
				Error:   common.ErrMsgForbidden,
				Message: "管理者権限が必要です",
				Code:    common.ErrCodeForbidden,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
