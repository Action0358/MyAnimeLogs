package common

import "errors"

// Common application errors
var (
	ErrValidation     = errors.New("validation error")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrInternalServer = errors.New("internal server error")
)
