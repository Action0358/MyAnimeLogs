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

	// Get User-Agent and IP address
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	loginResp, err := h.authService.Login(req, userAgent, ipAddress)
	if err != nil {
		// Check if account is locked
		if err == services.ErrAccountLocked {
			h.response.ForbiddenError(c, "アカウントがロックされています。しばらく時間をおいてから再試行してください。")
			return
		}

		if err == services.ErrInvalidCredentials {
			// Try to get remaining attempts for better error message
			// Note: We can't get user ID from email without exposing user existence
			// So we'll just show generic error
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
		h.response.InternalServerError(c, "ユーザー情報の取得に失敗しました")
		return
	}

	// アイコンが設定されていない場合
	if currentUser.Icon == nil || *currentUser.Icon == "" {
		h.response.ValidationError(c, "プロフィール画像が設定されていません")
		return
	}

	oldIconURL := *currentUser.Icon

	// DBからアイコンURLを削除（NULLに設定）
	user, err := h.authService.UpdateUserIcon(userID, "")
	if err != nil {
		h.response.InternalServerError(c, "プロフィール画像の削除に失敗しました")
		return
	}

	// ファイルシステムから画像ファイルを削除
	if err := h.uploadService.DeleteUserIcon(oldIconURL); err != nil {
		// ファイル削除エラーはログに記録するが、処理は継続
		// DBからは既に削除されているため、ユーザーには成功として返す
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

// RefreshToken refreshes an access token using a refresh token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.TokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, "リフレッシュトークンが必要です")
		return
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if err == services.ErrSessionNotFound || err == services.ErrSessionExpired || err == services.ErrInvalidRefreshToken {
			h.response.UnauthorizedError(c, "無効なリフレッシュトークンです")
			return
		}
		h.response.InternalServerError(c, "トークンのリフレッシュに失敗しました")
		return
	}

	h.response.Success(c, response)
}

// Logout logs out a user by deleting their session
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, "リフレッシュトークンが必要です")
		return
	}

	err := h.authService.Logout(req.RefreshToken)
	if err != nil {
		// Even if session not found, return success (idempotent)
		if err != services.ErrSessionNotFound {
			h.logger.Warn("Logout error", "error", err)
		}
	}

	h.response.Success(c, gin.H{"message": "ログアウトしました"})
}

// GetUserSessions returns all active sessions for the current user
func (h *AuthHandler) GetUserSessions(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	// Get current JTI from context to mark current session
	jtiStr := ""
	if jti, exists := c.Get("jti"); exists {
		if jtiValue, ok := jti.(string); ok {
			jtiStr = jtiValue
		}
	}

	sessions, err := h.authService.GetUserSessions(userID, jtiStr)
	if err != nil {
		h.response.InternalServerError(c, "セッション一覧の取得に失敗しました")
		return
	}

	h.response.Success(c, gin.H{"sessions": sessions})
}

// DeleteSession deletes a specific session
func (h *AuthHandler) DeleteSession(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	sessionID, err := h.validator.ParseIDParam(c, "id")
	if err != nil {
		h.response.ValidationError(c, "無効なセッションIDです")
		return
	}

	err = h.authService.DeleteSession(sessionID, userID)
	if err != nil {
		if err == services.ErrSessionNotFound {
			h.response.NotFoundError(c, "セッションが見つかりません")
			return
		}
		h.response.InternalServerError(c, "セッションの削除に失敗しました")
		return
	}

	h.response.Success(c, gin.H{"message": "セッションを削除しました"})
}

// LogoutAllDevices logs out the user from all devices
func (h *AuthHandler) LogoutAllDevices(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	err = h.authService.LogoutAllDevices(userID)
	if err != nil {
		h.response.InternalServerError(c, "全デバイスからのログアウトに失敗しました")
		return
	}

	h.response.Success(c, gin.H{"message": "全てのデバイスからログアウトしました"})
}
