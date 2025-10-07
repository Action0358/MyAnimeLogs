package handlers

import (
	"myanimelogs/internal/common"
)

// BaseHandler すべてのハンドラーに共通の機能を提供
type BaseHandler struct {
	response  *common.ResponseHandler
	validator *common.ValidatorHelper
	logger    *common.Logger
}

// NewBaseHandler 新しいベースハンドラーを作成
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		response:  common.NewResponseHandler(),
		validator: common.NewValidatorHelper(),
	}
}

// NewBaseHandlerWithLogger ロガー付きの新しいベースハンドラーを作成
func NewBaseHandlerWithLogger(logger *common.Logger, errorMonitor *common.ErrorMonitor) *BaseHandler {
	return &BaseHandler{
		response:  common.NewResponseHandlerWithMonitoring(logger, errorMonitor),
		validator: common.NewValidatorHelper(),
		logger:    logger,
	}
}
