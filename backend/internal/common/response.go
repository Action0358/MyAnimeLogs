package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResponseHandler handles common response patterns
type ResponseHandler struct {
	logger       *Logger
	errorMonitor *ErrorMonitor
}

// NewResponseHandler creates a new response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// NewResponseHandlerWithMonitoring creates a new response handler with logging and monitoring
func NewResponseHandlerWithMonitoring(logger *Logger, errorMonitor *ErrorMonitor) *ResponseHandler {
	return &ResponseHandler{
		logger:       logger,
		errorMonitor: errorMonitor,
	}
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
		r.ValidationError(c, "バリデーションエラーが発生しました")
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
	r.recordError(c, ErrCodeValidation, ErrValidation, map[string]interface{}{
		"validation_message": message,
	})

	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   ErrMsgValidation,
		Message: message,
		Code:    ErrCodeValidation,
	})
}

// ValidationErrorWithErr sends a validation error response with structured error information
func (r *ResponseHandler) ValidationErrorWithErr(c *gin.Context, err error) {
	message := err.Error()
	r.recordError(c, ErrCodeValidation, err, map[string]interface{}{
		"validation_message": message,
	})

	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   ErrMsgValidation,
		Message: message,
		Code:    ErrCodeValidation,
	})
}

// UnauthorizedError sends an unauthorized error response
func (r *ResponseHandler) UnauthorizedError(c *gin.Context, message string) {
	r.recordError(c, ErrCodeUnauthorized, ErrUnauthorized, map[string]interface{}{
		"auth_message": message,
	})

	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error:   ErrMsgUnauthorized,
		Message: message,
		Code:    ErrCodeUnauthorized,
	})
}

// NotFoundError sends a not found error response
func (r *ResponseHandler) NotFoundError(c *gin.Context, message string) {
	r.recordError(c, ErrCodeNotFound, ErrNotFound, map[string]interface{}{
		"not_found_message": message,
	})

	c.JSON(http.StatusNotFound, ErrorResponse{
		Error:   ErrMsgNotFound,
		Message: message,
		Code:    ErrCodeNotFound,
	})
}

// ConflictError sends a conflict error response
func (r *ResponseHandler) ConflictError(c *gin.Context, message string) {
	r.recordError(c, ErrCodeConflict, ErrConflict, map[string]interface{}{
		"conflict_message": message,
	})

	c.JSON(http.StatusConflict, ErrorResponse{
		Error:   ErrMsgConflict,
		Message: message,
		Code:    ErrCodeConflict,
	})
}

// InternalServerError sends an internal server error response
func (r *ResponseHandler) InternalServerError(c *gin.Context, message string) {
	r.recordError(c, ErrCodeInternalServer, ErrInternalServer, map[string]interface{}{
		"internal_message": message,
	})

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   ErrMsgInternalServer,
		Message: message,
		Code:    ErrCodeInternalServer,
	})
}

// UnprocessableEntityError sends an unprocessable entity error response
func (r *ResponseHandler) UnprocessableEntityError(c *gin.Context, message string) {
	r.recordError(c, ErrCodeUnprocessableEntity, ErrUnprocessableEntity, map[string]interface{}{
		"unprocessable_message": message,
	})

	c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
		Error:   ErrMsgUnprocessableEntity,
		Message: message,
		Code:    ErrCodeUnprocessableEntity,
	})
}

// recordError records error for monitoring if monitoring is enabled
func (r *ResponseHandler) recordError(c *gin.Context, errorCode string, err error, metadata map[string]interface{}) {
	if r.errorMonitor == nil || r.logger == nil {
		return
	}

	// Add request context to metadata
	requestMetadata := make(map[string]interface{})
	for k, v := range metadata {
		requestMetadata[k] = v
	}

	// Add request information
	requestMetadata["method"] = c.Request.Method
	requestMetadata["path"] = c.Request.URL.Path
	requestMetadata["client_ip"] = c.ClientIP()
	requestMetadata["user_agent"] = c.Request.UserAgent()

	if requestID := c.GetString("request_id"); requestID != "" {
		requestMetadata["request_id"] = requestID
	}

	if userID, exists := c.Get("user_id"); exists {
		requestMetadata["user_id"] = userID
	}

	// Record the error
	r.errorMonitor.RecordError(c.Request.Context(), errorCode, err, requestMetadata)
}

// ForbiddenError sends a forbidden error response
func (r *ResponseHandler) ForbiddenError(c *gin.Context, message string) {
	r.recordError(c, ErrCodeForbidden, ErrForbidden, map[string]interface{}{
		"forbidden_message": message,
	})

	c.JSON(http.StatusForbidden, ErrorResponse{
		Error:   ErrMsgForbidden,
		Message: message,
		Code:    ErrCodeForbidden,
	})
}

// TooManyRequestsError sends a too many requests error response
func (r *ResponseHandler) TooManyRequestsError(c *gin.Context, message string) {
	r.recordError(c, ErrCodeTooManyRequests, ErrTooManyRequests, map[string]interface{}{
		"rate_limit_message": message,
	})

	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Error:   ErrMsgTooManyRequests,
		Message: message,
		Code:    ErrCodeTooManyRequests,
	})
}

// ServiceUnavailableError sends a service unavailable error response
func (r *ResponseHandler) ServiceUnavailableError(c *gin.Context, message string) {
	r.recordError(c, ErrCodeServiceUnavailable, ErrServiceUnavailable, map[string]interface{}{
		"service_unavailable_message": message,
	})

	c.JSON(http.StatusServiceUnavailable, ErrorResponse{
		Error:   ErrMsgServiceUnavailable,
		Message: message,
		Code:    ErrCodeServiceUnavailable,
	})
}
