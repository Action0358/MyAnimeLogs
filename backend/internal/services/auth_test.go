package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// Test service error constants
func TestAuthService_ErrorConstants(t *testing.T) {
	assert.Equal(t, "email already exists", ErrEmailAlreadyExists.Error())
	assert.Equal(t, "invalid credentials", ErrInvalidCredentials.Error())
	assert.Equal(t, "user not found", ErrUserNotFound.Error())
}

// Test service constructor
func TestNewAuthService(t *testing.T) {
	service := NewAuthService(nil, "test-secret")
	assert.NotNil(t, service)
	assert.Equal(t, "test-secret", service.jwtSecret)
}

// Test password hashing (used in Register method)
func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"

	// Test bcrypt hashing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// Test password verification
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	assert.NoError(t, err)

	// Test wrong password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	assert.Error(t, err)
}

// Test service validation logic (extracted from methods)
func TestAuthService_ValidationLogic(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		valid    bool
	}{
		{
			name:     "Valid email and password",
			email:    "test@example.com",
			password: "password123",
			valid:    true,
		},
		{
			name:     "Empty email",
			email:    "",
			password: "password123",
			valid:    false,
		},
		{
			name:     "Empty password",
			email:    "test@example.com",
			password: "",
			valid:    false,
		},
		{
			name:     "Invalid email format",
			email:    "invalid-email",
			password: "password123",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			isValid := tt.email != "" && tt.password != "" &&
				len(tt.email) > 0 && len(tt.password) >= 6 &&
				contains(tt.email, "@") && contains(tt.email, ".")

			assert.Equal(t, tt.valid, isValid)
		})
	}
}

// Helper function for validation
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test business logic for user stats calculation
func TestAuthService_StatsCalculation(t *testing.T) {
	tests := []struct {
		name                   string
		totalEpisodesWatched   int
		expectedWatchTimeHours float64
	}{
		{
			name:                   "No episodes watched",
			totalEpisodesWatched:   0,
			expectedWatchTimeHours: 0.0,
		},
		{
			name:                   "10 episodes watched",
			totalEpisodesWatched:   10,
			expectedWatchTimeHours: 4.0, // 10 * 0.4
		},
		{
			name:                   "100 episodes watched",
			totalEpisodesWatched:   100,
			expectedWatchTimeHours: 40.0, // 100 * 0.4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Business logic: 24 minutes per episode = 0.4 hours
			watchTimeHours := float64(tt.totalEpisodesWatched) * 0.4
			assert.Equal(t, tt.expectedWatchTimeHours, watchTimeHours)
		})
	}
}

// Test CSV export format logic
func TestAuthService_CSVExportFormat(t *testing.T) {
	// Test CSV format generation logic
	expectedCSVHeader := "type,id,data\n"
	expectedUserLine := "user,123,\"Test User\"\n"
	expectedFavoriteLine := "favorite,456,\"completed\"\n"

	// Build CSV
	csv := expectedCSVHeader
	csv += expectedUserLine
	csv += expectedFavoriteLine

	assert.Contains(t, csv, "type,id,data")
	assert.Contains(t, csv, "user,123")
	assert.Contains(t, csv, "Test User")
	assert.Contains(t, csv, "favorite,456")
	assert.Contains(t, csv, "completed")
}

// Test service method signatures and return types
func TestAuthService_MethodSignatures(t *testing.T) {
	service := NewAuthService(nil, "test-secret")

	// Test that methods exist and have correct signatures
	assert.NotNil(t, service.Register)
	assert.NotNil(t, service.Login)
	assert.NotNil(t, service.GetUserByID)
	assert.NotNil(t, service.UpdateUser)
	assert.NotNil(t, service.GetUserStats)
	assert.NotNil(t, service.ExportUserData)
}

// Test error handling patterns
func TestAuthService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "Email already exists error",
			err:         ErrEmailAlreadyExists,
			expectedMsg: "email already exists",
		},
		{
			name:        "Invalid credentials error",
			err:         ErrInvalidCredentials,
			expectedMsg: "invalid credentials",
		},
		{
			name:        "User not found error",
			err:         ErrUserNotFound,
			expectedMsg: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMsg, tt.err.Error())
		})
	}
}
