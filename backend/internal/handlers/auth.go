package handlers

import (
	"net/http"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
	"myanimelogs/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	*BaseHandler
	authService   services.AuthServiceInterface
	uploadService *services.UploadService
}

func NewAuthHandler(authService services.AuthServiceInterface, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AuthHandler {
	return &AuthHandler{
		BaseHandler: NewBaseHandlerWithLogger(logger, errorMonitor),
		authService: authService,
	}
}

func NewAuthHandlerWithUpload(authService services.AuthServiceInterface, uploadService *services.UploadService, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AuthHandler {
	return &AuthHandler{
		BaseHandler:   NewBaseHandlerWithLogger(logger, errorMonitor),
		authService:   authService,
		uploadService: uploadService,
	}
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.response.NotFoundError(c, "ユーザーが見つかりません")
		return
	}

	h.response.Success(c, user)
}

func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	if req.SnsLink != nil {
		if err := utils.ValidateURL(*req.SnsLink); err != nil {
			h.response.ValidationError(c, err.Error())
			return
		}
	}

	user, err := h.authService.UpdateUser(userID, req)
	if err != nil {
		h.response.InternalServerError(c, "プロフィール更新に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message": "プロフィールを更新しました。",
		"user":    user,
	})
}

func (h *AuthHandler) UploadIcon(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	if h.uploadService == nil {
		h.response.InternalServerError(c, "アップロードサービスが利用できません")
		return
	}

	file, err := c.FormFile("icon")
	if err != nil {
		h.response.ValidationError(c, "画像ファイルが必要です")
		return
	}

	currentUser, err := h.authService.GetUserByID(userID)
	var oldIconURL string
	if err == nil && currentUser.Icon != nil && *currentUser.Icon != "" {
		oldIconURL = *currentUser.Icon
	}

	iconURL, err := h.uploadService.SaveUserIcon(userID, file)
	if err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	user, err := h.authService.UpdateUserIcon(userID, iconURL)
	if err != nil {
		_ = h.uploadService.DeleteUserIcon(iconURL)
		h.response.InternalServerError(c, "画像の保存に失敗しました")
		return
	}

	if oldIconURL != "" {
		_ = h.uploadService.DeleteUserIcon(oldIconURL)
	}

	h.response.Success(c, gin.H{
		"message": "プロフィール画像を更新しました",
		"user":    user,
	})
}

func (h *AuthHandler) DeleteIcon(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	if h.uploadService == nil {
		h.response.InternalServerError(c, "アップロードサービスが利用できません")
		return
	}

	currentUser, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.response.InternalServerError(c, "ユーザー情報の取得に失敗しました")
		return
	}

	if currentUser.Icon == nil || *currentUser.Icon == "" {
		h.response.ValidationError(c, "プロフィール画像が設定されていません")
		return
	}

	oldIconURL := *currentUser.Icon

	user, err := h.authService.UpdateUserIcon(userID, "")
	if err != nil {
		h.response.InternalServerError(c, "プロフィール画像の削除に失敗しました")
		return
	}

	if err := h.uploadService.DeleteUserIcon(oldIconURL); err != nil {
		h.logger.Warn("Failed to delete icon file from filesystem",
			"user_id", userID,
			"icon_url", oldIconURL,
			"error", err)
	}

	h.response.Success(c, gin.H{
		"message": "プロフィール画像を削除しました",
		"user":    user,
	})
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
