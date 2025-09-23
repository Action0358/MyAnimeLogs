package common

import (
	"net/http"

	"myanimelogs/internal/models"

	"github.com/gin-gonic/gin"
)

// ResponseHandler handles common response patterns
type ResponseHandler struct{}

// NewResponseHandler creates a new response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// Success sends a successful response
func (r *ResponseHandler) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// Created sends a created response
func (r *ResponseHandler) Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// Error sends an error response based on the error type
func (r *ResponseHandler) Error(c *gin.Context, err error) {
	switch err {
	case models.ErrValidation:
		r.ValidationError(c, err.Error())
	case models.ErrUnauthorized:
		r.UnauthorizedError(c, "認証が必要です")
	case models.ErrNotFound:
		r.NotFoundError(c, "リソースが見つかりません")
	case models.ErrConflict:
		r.ConflictError(c, "リソースが既に存在します")
	default:
		r.InternalServerError(c, "内部サーバーエラーが発生しました")
	}
}

// ValidationError sends a validation error response
func (r *ResponseHandler) ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, models.ErrorResponse{
		Error:   models.ErrMsgValidation,
		Message: message,
		Code:    models.ErrCodeValidation,
	})
}

// UnauthorizedError sends an unauthorized error response
func (r *ResponseHandler) UnauthorizedError(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, models.ErrorResponse{
		Error:   models.ErrMsgUnauthorized,
		Message: message,
		Code:    models.ErrCodeUnauthorized,
	})
}

// NotFoundError sends a not found error response
func (r *ResponseHandler) NotFoundError(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, models.ErrorResponse{
		Error:   models.ErrMsgNotFound,
		Message: message,
		Code:    models.ErrCodeNotFound,
	})
}

// ConflictError sends a conflict error response
func (r *ResponseHandler) ConflictError(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, models.ErrorResponse{
		Error:   models.ErrMsgConflict,
		Message: message,
		Code:    models.ErrCodeConflict,
	})
}

// InternalServerError sends an internal server error response
func (r *ResponseHandler) InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error:   models.ErrMsgInternalServer,
		Message: message,
		Code:    models.ErrCodeInternalServer,
	})
}

// UnprocessableEntityError sends an unprocessable entity error response
func (r *ResponseHandler) UnprocessableEntityError(c *gin.Context, message string) {
	c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{
		Error:   models.ErrMsgUnprocessableEntity,
		Message: message,
		Code:    models.ErrCodeUnprocessableEntity,
	})
}
