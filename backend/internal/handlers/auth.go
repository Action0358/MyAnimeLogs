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
		BaseHandler:   NewBaseHandlerWithLogger(logger, errorMonitor),
		authService:   authService,
		uploadService: nil,
	}
}

func NewAuthHandlerWithUpload(authService services.AuthServiceInterface, uploadService *services.UploadService, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AuthHandler {
	return &AuthHandler{
		BaseHandler:   NewBaseHandlerWithLogger(logger, errorMonitor),
		authService:   authService,
		uploadService: uploadService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if err == services.ErrEmailAlreadyExists {
			h.response.ConflictError(c, "このメールアドレスは既に登録されています")
			return
		}
		h.response.InternalServerError(c, "ユーザー登録に失敗しました")
		return
	}

	h.response.Created(c, gin.H{
		"message": "ユーザー登録が完了しました。",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	loginResp, err := h.authService.Login(req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			h.response.UnauthorizedError(c, "メールアドレスまたはパスワードが正しくありません")
			return
		}
		h.response.InternalServerError(c, "ログインに失敗しました")
		return
	}

	h.response.Success(c, loginResp)
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

	// SNSリンクのバリデーション
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

	// ファイルを取得
	file, err := c.FormFile("icon")
	if err != nil {
		h.response.ValidationError(c, "画像ファイルが必要です")
		return
	}

	// 現在のユーザー情報を取得（古い画像のURLを保持）
	currentUser, err := h.authService.GetUserByID(userID)
	var oldIconURL string
	if err == nil && currentUser.Icon != nil && *currentUser.Icon != "" {
		oldIconURL = *currentUser.Icon
	}

	// 新しい画像を保存
	iconURL, err := h.uploadService.SaveUserIcon(userID, file)
	if err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	// DBを更新
	user, err := h.authService.UpdateUserIcon(userID, iconURL)
	if err != nil {
		// DB更新に失敗した場合、アップロードした新しい画像を削除
		_ = h.uploadService.DeleteUserIcon(iconURL)
		h.response.InternalServerError(c, "画像の保存に失敗しました")
		return
	}

	// 全ての処理が成功した後に古い画像を削除
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

	// 現在のユーザー情報を取得
	currentUser, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.response.NotFoundError(c, "ユーザーが見つかりません")
		return
	}

	// 古い画像URLを保持
	var oldIconURL string
	if currentUser.Icon != nil && *currentUser.Icon != "" {
		oldIconURL = *currentUser.Icon
	}

	// DBを更新（iconをNULLに）
	user, err := h.authService.UpdateUserIcon(userID, "")
	if err != nil {
		h.response.InternalServerError(c, "画像の削除に失敗しました")
		return
	}

	// DB更新が成功した後に画像ファイルを削除
	if oldIconURL != "" {
		_ = h.uploadService.DeleteUserIcon(oldIconURL)
	}

	h.response.Success(c, gin.H{
		"message": "プロフィール画像を削除しました",
		"user":    user,
	})
}

func (h *AuthHandler) GetStats(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	stats, err := h.authService.GetUserStats(userID)
	if err != nil {
		h.response.InternalServerError(c, "統計情報の取得に失敗しました")
		return
	}

	h.response.Success(c, stats)
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

		// 安全な型アサーション
		if csvData, ok := data.(string); ok {
			c.String(http.StatusOK, csvData)
		} else {
			h.response.InternalServerError(c, "CSVデータの形式が正しくありません")
		}
	} else {
		h.response.Success(c, data)
	}
}
