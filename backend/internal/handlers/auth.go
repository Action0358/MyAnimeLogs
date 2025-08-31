package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: err.Error(),
			Code:    models.ErrCodeValidation,
		})
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if err == services.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   models.ErrMsgConflict,
				Message: "このメールアドレスは既に登録されています",
				Code:    models.ErrCodeConflict,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "ユーザー登録に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "ユーザー登録が完了しました。",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: err.Error(),
			Code:    models.ErrCodeValidation,
		})
		return
	}

	loginResp, err := h.authService.Login(req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   models.ErrMsgUnauthorized,
				Message: "メールアドレスまたはパスワードが正しくありません",
				Code:    models.ErrCodeUnauthorized,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "ログインに失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, loginResp)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	user, err := h.authService.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   models.ErrMsgNotFound,
			Message: "ユーザーが見つかりません",
			Code:    models.ErrCodeNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: err.Error(),
			Code:    models.ErrCodeValidation,
		})
		return
	}

	user, err := h.authService.UpdateUser(userID.(int), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "プロフィール更新に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "プロフィールを更新しました。",
		"user":    user,
	})
}

func (h *AuthHandler) GetStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	stats, err := h.authService.GetUserStats(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "統計情報の取得に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AuthHandler) ExportData(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	format := c.DefaultQuery("format", "json")
	if format != "json" && format != "csv" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: "フォーマットはjsonまたはcsvを指定してください",
			Code:    models.ErrCodeValidation,
		})
		return
	}

	data, err := h.authService.ExportUserData(userID.(int), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "データエクスポートに失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=myanimelog_data.csv")
		c.String(http.StatusOK, data.(string))
	} else {
		c.JSON(http.StatusOK, data)
	}
}