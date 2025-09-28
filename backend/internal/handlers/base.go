package handlers

import (
	"myanimelogs/internal/common"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	response  *common.ResponseHandler
	validator *common.ValidatorHelper
	logger    *common.Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		response:  common.NewResponseHandler(),
		validator: common.NewValidatorHelper(),
	}
}

// NewBaseHandlerWithLogger creates a new base handler with logger
func NewBaseHandlerWithLogger(logger *common.Logger, errorMonitor *common.ErrorMonitor) *BaseHandler {
	return &BaseHandler{
		response:  common.NewResponseHandlerWithMonitoring(logger, errorMonitor),
		validator: common.NewValidatorHelper(),
		logger:    logger,
	}
}
