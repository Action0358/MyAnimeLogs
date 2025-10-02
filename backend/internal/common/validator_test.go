package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidatorHelper_GetUserIDFromContext(t *testing.T) {
	validator := NewValidatorHelper()

	tests := []struct {
		name        string
		setupFunc   func(*gin.Context)
		expectError bool
		expectedID  int
	}{
		{
			name: "Valid user ID",
			setupFunc: func(c *gin.Context) {
				c.Set("user_id", 123)
			},
			expectError: false,
			expectedID:  123,
		},
		{
			name: "Missing user ID",
			setupFunc: func(c *gin.Context) {
				// Don't set user_id
			},
			expectError: true,
			expectedID:  0,
		},
		{
			name: "Invalid user ID type",
			setupFunc: func(c *gin.Context) {
				c.Set("user_id", "invalid")
			},
			expectError: true,
			expectedID:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupFunc(c)

			userID, err := validator.GetUserIDFromContext(c)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, ErrUnauthorized, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}

func TestValidatorHelper_ParseIDParam(t *testing.T) {
	validator := NewValidatorHelper()

	tests := []struct {
		name        string
		paramValue  string
		expectError bool
		expectedID  int
	}{
		{
			name:        "Valid ID",
			paramValue:  "123",
			expectError: false,
			expectedID:  123,
		},
		{
			name:        "Zero ID",
			paramValue:  "0",
			expectError: false,
			expectedID:  0,
		},
		{
			name:        "Invalid ID - non-numeric",
			paramValue:  "abc",
			expectError: true,
			expectedID:  0,
		},
		{
			name:        "Invalid ID - empty",
			paramValue:  "",
			expectError: true,
			expectedID:  0,
		},
		{
			name:        "Invalid ID - negative",
			paramValue:  "-1",
			expectError: false, // strconv.Atoi allows negative numbers
			expectedID:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			// Setup route with parameter
			router.GET("/test/:id", func(c *gin.Context) {
				id, err := validator.ParseIDParam(c, "id")
				c.JSON(200, gin.H{"id": id, "error": err != nil})
			})

			// Create request with parameter
			req, _ := http.NewRequest("GET", "/test/"+tt.paramValue, nil)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.paramValue}}

			id, err := validator.ParseIDParam(c, "id")

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, ErrValidation, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestValidatorHelper_ParseIntQuery(t *testing.T) {
	validator := NewValidatorHelper()

	tests := []struct {
		name          string
		queryValue    string
		defaultValue  int
		expectedValue int
	}{
		{
			name:          "Valid query value",
			queryValue:    "42",
			defaultValue:  10,
			expectedValue: 42,
		},
		{
			name:          "Empty query value - use default",
			queryValue:    "",
			defaultValue:  10,
			expectedValue: 10,
		},
		{
			name:          "Invalid query value - use default",
			queryValue:    "abc",
			defaultValue:  10,
			expectedValue: 10,
		},
		{
			name:          "Zero query value",
			queryValue:    "0",
			defaultValue:  10,
			expectedValue: 0,
		},
		{
			name:          "Negative query value",
			queryValue:    "-5",
			defaultValue:  10,
			expectedValue: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			// Setup route
			router.GET("/test", func(c *gin.Context) {
				value := validator.ParseIntQuery(c, "param", tt.defaultValue)
				c.JSON(200, gin.H{"value": value})
			})

			// Create request with query parameter
			url := "/test"
			if tt.queryValue != "" {
				url += "?param=" + tt.queryValue
			}
			req, _ := http.NewRequest("GET", url, nil)
			c.Request = req

			value := validator.ParseIntQuery(c, "param", tt.defaultValue)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}
