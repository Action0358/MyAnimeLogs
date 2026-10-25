package common

import "errors"

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    string            `json:"code"`
	Details []ValidationError `json:"details,omitempty"`
}

// Standard error codes based on CSV specification
const (
	ErrCodeValidation          = "VALIDATION_ERROR"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"
	ErrCodeTooManyRequests     = "TOO_MANY_REQUESTS"
	ErrCodeInternalServer      = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeAccountLocked       = "ACCOUNT_LOCKED"
	// ErrCodeMaintenance はメンテナンスモード中の 503。外部API障害などの SERVICE_UNAVAILABLE と
	// フロントで区別し、メンテナンス画面を出すために別コードにしている
	ErrCodeMaintenance = "MAINTENANCE"
)

// Standard error messages based on CSV specification
const (
	ErrMsgValidation          = "バリデーションエラーです"
	ErrMsgUnauthorized        = "認証が必要です"
	ErrMsgForbidden           = "アクセス権限がありません"
	ErrMsgNotFound            = "リソースが見つかりません"
	ErrMsgConflict            = "データの競合が発生しました"
	ErrMsgUnprocessableEntity = "処理できないデータです"
	ErrMsgTooManyRequests     = "リクエスト数が上限を超えました。しばらく待ってから再試行してください。"
	ErrMsgInternalServer      = "サーバー内部エラーが発生しました"
	ErrMsgServiceUnavailable  = "サービスが一時的に利用できません"
	ErrMsgMaintenance         = "メンテナンス中です"
)

// Common application errors
var (
	ErrValidation          = errors.New("validation error")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrInternalServer      = errors.New("internal server error")
	ErrServiceUnavailable  = errors.New("service unavailable")
)
