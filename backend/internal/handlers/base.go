package handlers

import (
	"myanimelogs/internal/common"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	response  *common.ResponseHandler
	validator *common.ValidatorHelper
}

// NewBaseHandler creates a new base handler
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		response:  common.NewResponseHandler(),
		validator: common.NewValidatorHelper(),
	}
}
