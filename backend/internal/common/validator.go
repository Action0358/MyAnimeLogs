package common

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ValidatorHelper provides common validation utilities
type ValidatorHelper struct{}

// NewValidatorHelper creates a new validator helper
func NewValidatorHelper() *ValidatorHelper {
	return &ValidatorHelper{}
}

// GetUserIDFromContext extracts and validates user ID from gin context
func (v *ValidatorHelper) GetUserIDFromContext(c *gin.Context) (int, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, ErrUnauthorized
	}
	return userID.(int), nil
}

// ParseIDParam parses and validates ID parameter from URL
func (v *ValidatorHelper) ParseIDParam(c *gin.Context, paramName string) (int, error) {
	idStr := c.Param(paramName)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, ErrValidation
	}
	return id, nil
}

// ParseIntQuery parses integer query parameter with default value
func (v *ValidatorHelper) ParseIntQuery(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// ValidatePagination validates and sets default pagination parameters
func (v *ValidatorHelper) ValidatePagination(page, perPage int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}
