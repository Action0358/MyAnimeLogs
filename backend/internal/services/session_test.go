package services

import (
	"testing"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestSessionService_CreateSession(t *testing.T) {
	// Skip if no database connection
	db, err := database.New("postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := common.NewLogger(common.LogLevelInfo, "json")
	service := NewSessionService(db, logger)

	jti := utils.GenerateJTI()
	session, refreshToken, err := service.CreateSession(1, jti, "test-agent", "127.0.0.1")

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, 1, session.UserID)
	assert.Equal(t, jti, session.AccessTokenJTI)
	assert.True(t, session.ExpiresAt.After(time.Now()))

	// Cleanup
	service.DeleteSession(refreshToken)
}

func TestSessionService_ValidateSession(t *testing.T) {
	// Skip if no database connection
	db, err := database.New("postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := common.NewLogger(common.LogLevelInfo, "json")
	service := NewSessionService(db, logger)

	// Create a session first
	jti := utils.GenerateJTI()
	_, refreshToken, err := service.CreateSession(1, jti, "test-agent", "127.0.0.1")
	assert.NoError(t, err)

	// Validate the session
	session, err := service.ValidateSession(refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, 1, session.UserID)

	// Cleanup
	service.DeleteSession(refreshToken)
}

func TestSessionService_DeleteSession(t *testing.T) {
	// Skip if no database connection
	db, err := database.New("postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := common.NewLogger(common.LogLevelInfo, "json")
	service := NewSessionService(db, logger)

	// Create a session first
	jti := utils.GenerateJTI()
	_, refreshToken, err := service.CreateSession(1, jti, "test-agent", "127.0.0.1")
	assert.NoError(t, err)

	// Delete the session
	err = service.DeleteSession(refreshToken)
	assert.NoError(t, err)

	// Try to validate deleted session
	_, err = service.ValidateSession(refreshToken)
	assert.Error(t, err)
	assert.Equal(t, ErrSessionNotFound, err)
}

func TestGenerateJTI(t *testing.T) {
	jti1 := utils.GenerateJTI()
	jti2 := utils.GenerateJTI()

	assert.NotEmpty(t, jti1)
	assert.NotEmpty(t, jti2)
	assert.NotEqual(t, jti1, jti2)
	assert.Len(t, jti1, 36) // UUID v4 length
}
