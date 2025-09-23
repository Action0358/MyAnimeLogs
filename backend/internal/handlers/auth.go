package handlers

import (
	"net/http"

	"myanimelogs/internal/models"
	"myanimelogs/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	*BaseHandler
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		BaseHandler: NewBaseHandler(),
		authService: authService,
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
		c.String(http.StatusOK, data.(string))
	} else {
		h.response.Success(c, data)
	}
}
