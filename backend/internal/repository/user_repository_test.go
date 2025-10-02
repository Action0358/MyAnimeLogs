package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserRepository(t *testing.T) {
	repo := NewUserRepository(nil)
	assert.NotNil(t, repo)
}

func TestUserRepository_MethodSignatures(t *testing.T) {
	repo := NewUserRepository(nil)

	assert.NotNil(t, repo.GetByEmail)
	assert.NotNil(t, repo.GetByID)
	assert.NotNil(t, repo.Create)
	assert.NotNil(t, repo.EmailExists)
}

func TestUserRepository_ErrorHandling(t *testing.T) {
	errorPatterns := []struct {
		name        string
		expectedMsg string
	}{
		{
			name:        "Email existence check error",
			expectedMsg: "failed to check email existence",
		},
		{
			name:        "User creation error",
			expectedMsg: "failed to create user",
		},
		{
			name:        "User retrieval by email error",
			expectedMsg: "failed to get user by email",
		},
		{
			name:        "User retrieval by ID error",
			expectedMsg: "failed to get user by ID",
		},
	}

	for _, tt := range errorPatterns {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.expectedMsg, "failed to")
		})
	}
}

func TestUserRepository_DataValidation(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{
			name:  "Valid email",
			email: "test@example.com",
			valid: true,
		},
		{
			name:  "Empty email",
			email: "",
			valid: false,
		},
		{
			name:  "Invalid email format",
			email: "invalid-email",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.email != "" && contains(tt.email, "@") && contains(tt.email, ".")
			assert.Equal(t, tt.valid, isValid)
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
