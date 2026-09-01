package handlers

import (
	"errors"
	"net/http"

	"myanimelogs/internal/common"
	"myanimelogs/internal/middleware"
	"myanimelogs/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	*BaseHandler
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AuthHandler {
	return &AuthHandler{
		BaseHandler: NewBaseHandlerWithLogger(logger, errorMonitor),
		authService: authService,
	}
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	avatarURL, _ := c.Get("user_avatar_url")
	avatarURLStr, _ := avatarURL.(string)

	user, err := h.authService.GetUserByID(userID, avatarURLStr)
	if err != nil {
		h.response.NotFoundError(c, "ユーザーが見つかりません")
		return
	}

	h.response.Success(c, user)
}


// DeleteMe permanently deletes the authenticated user's account.
//
// 削除は復旧不可。利用規約 第9条 の記載に合わせた挙動とする。
// 呼び出し側（フロントエンド）は成功後、直ちに signOut すること。
// JWT は最大1時間有効なまま残るが、users 行の再作成は外部キー制約により
// 失敗するため、データが復活することはない。
func (h *AuthHandler) DeleteMe(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	authUserID, ok := c.Get("auth_user_id")
	authUserIDStr, isString := authUserID.(string)
	if !ok || !isString || authUserIDStr == "" {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	if err := h.authService.DeleteAccount(userID, authUserIDStr); err != nil {
		if errors.Is(err, services.ErrAdminAPIUnavailable) {
			// 設定漏れを 500 に紛れ込ませず、原因が分かる形で返す
			h.logger.Error("Account deletion requested but admin API is not configured", "user_id", userID)
			h.response.ServiceUnavailableError(c, "アカウント削除機能が利用できません。時間をおいて再度お試しください")
			return
		}
		h.logger.Error("Failed to delete account", "user_id", userID, "error", err)
		h.response.InternalServerError(c, "アカウントの削除に失敗しました")
		return
	}

	// 削除済みユーザーの内部IDを返し続けないよう、プロセス内キャッシュを落とす
	middleware.InvalidateUserIDCache(authUserIDStr)

	h.logger.Info("Account deleted", "user_id", userID)
	h.response.Success(c, gin.H{"message": "アカウントを削除しました"})
}

func (h *AuthHandler) ExportData(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	format := c.DefaultQuery("format", "json")
	if format != "json" && format != "csv" {
		h.response.ValidationError(c, "フォーマットはjsonまたはcsvを指定してください")
		return
	}

	data, err := h.authService.ExportUserData(userID, format)
	if err != nil {
		h.response.InternalServerError(c, "データエクスポートに失敗しました")
		return
	}

	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=myanimelog_data.csv")
		if csvData, ok := data.(string); ok {
			c.String(http.StatusOK, csvData)
		} else {
			h.response.InternalServerError(c, "CSVデータの形式が正しくありません")
		}
	} else {
		h.response.Success(c, data)
	}
}
