package common

import (
	"net/http"

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
	case ErrValidation:
		r.ValidationError(c, err.Error())
	case ErrUnauthorized:
		r.UnauthorizedError(c, "認証が必要です")
	case ErrNotFound:
		r.NotFoundError(c, "リソースが見つかりません")
	case ErrConflict:
		r.ConflictError(c, "リソースが既に存在します")
	default:
		r.InternalServerError(c, "内部サーバーエラーが発生しました")
	}
}

// ValidationError sends a validation error response
func (r *ResponseHandler) ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   ErrMsgValidation,
		Message: message,
		Code:    ErrCodeValidation,
	})
}

// UnauthorizedError sends an unauthorized error response
func (r *ResponseHandler) UnauthorizedError(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error:   ErrMsgUnauthorized,
		Message: message,
		Code:    ErrCodeUnauthorized,
	})
}

// NotFoundError sends a not found error response
func (r *ResponseHandler) NotFoundError(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error:   ErrMsgNotFound,
		Message: message,
		Code:    ErrCodeNotFound,
	})
}

// ConflictError sends a conflict error response
func (r *ResponseHandler) ConflictError(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, ErrorResponse{
		Error:   ErrMsgConflict,
		Message: message,
		Code:    ErrCodeConflict,
	})
}

// InternalServerError sends an internal server error response
func (r *ResponseHandler) InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   ErrMsgInternalServer,
		Message: message,
		Code:    ErrCodeInternalServer,
	})
}

// UnprocessableEntityError sends an unprocessable entity error response
func (r *ResponseHandler) UnprocessableEntityError(c *gin.Context, message string) {
	c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
		Error:   ErrMsgUnprocessableEntity,
		Message: message,
		Code:    ErrCodeUnprocessableEntity,
	})
}
