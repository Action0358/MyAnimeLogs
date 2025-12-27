package services

import (
	"errors"
	"testing"
	"time"

	"myanimelogs/internal/models"
	"myanimelogs/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// TestLoginFailureIncrementsAttempts tests that failed login attempts are tracked
func TestLoginFailureIncrementsAttempts(t *testing.T) {
	// Setup
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	mockUser := &models.User{
		ID:                  1,
		Email:               "test@example.com",
		Password:            string(hashedPassword),
		Name:                "Test User",
		FailedLoginAttempts: 0,
	}

	mockRepo := new(MockUserRepository)
	mockSessionService := new(MockSessionService)

	// User exists and is not locked
	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)
	mockRepo.On("IsAccountLocked", 1).Return(false, (*time.Time)(nil), nil)

	// Expect failed attempts to be incremented (wrong password)
	mockRepo.On("IncrementFailedAttempts", 1, MaxLoginAttempts, LockoutDuration).Return(&repository.LockoutResult{
		NewAttempts: 1,
		LockedUntil: nil,
	}, nil)

	service := NewAuthServiceWithDeps(mockRepo, nil, mockSessionService, "test-secret")

	// Attempt login with wrong password
	_, err := service.Login(models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}, "test-agent", "127.0.0.1")

	// Should fail with invalid credentials
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)

	// Verify IncrementFailedAttempts was called
	mockRepo.AssertCalled(t, "IncrementFailedAttempts", 1, MaxLoginAttempts, LockoutDuration)
}

// TestAccountLocksAfterMaxAttempts tests that account locks after max failed attempts
func TestAccountLocksAfterMaxAttempts(t *testing.T) {
	// Setup
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	mockUser := &models.User{
		ID:                  1,
		Email:               "test@example.com",
		Password:            string(hashedPassword),
		Name:                "Test User",
		FailedLoginAttempts: 4, // Already at 4 attempts
	}

	mockRepo := new(MockUserRepository)
	mockSessionService := new(MockSessionService)

	// User exists and is not locked (yet)
	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)
	mockRepo.On("IsAccountLocked", 1).Return(false, (*time.Time)(nil), nil)

	// This will be the 10th attempt - account should be locked
	lockTime := time.Now().Add(LockoutDuration)
	mockRepo.On("IncrementFailedAttempts", 1, MaxLoginAttempts, LockoutDuration).Return(&repository.LockoutResult{
		NewAttempts: 10,
		LockedUntil: &lockTime,
	}, nil)

	service := NewAuthServiceWithDeps(mockRepo, nil, mockSessionService, "test-secret")

	// Attempt login with wrong password (10th attempt)
	_, err := service.Login(models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}, "test-agent", "127.0.0.1")

	// Should fail with invalid credentials
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)

	// Verify IncrementFailedAttempts was called and returned locked status
	mockRepo.AssertCalled(t, "IncrementFailedAttempts", 1, MaxLoginAttempts, LockoutDuration)
}

// TestLockedAccountRejectsLogin tests that locked accounts cannot login
func TestLockedAccountRejectsLogin(t *testing.T) {
	// Setup
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	lockTime := time.Now().Add(10 * time.Minute)
	mockUser := &models.User{
		ID:                  1,
		Email:               "test@example.com",
		Password:            string(hashedPassword),
		Name:                "Test User",
		FailedLoginAttempts: 5,
		LockedUntil:         &lockTime,
	}

	mockRepo := new(MockUserRepository)
	mockSessionService := new(MockSessionService)

	// User exists and IS locked
	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)
	mockRepo.On("IsAccountLocked", 1).Return(true, &lockTime, nil)

	service := NewAuthServiceWithDeps(mockRepo, nil, mockSessionService, "test-secret")

	// Attempt login with correct password
	_, err := service.Login(models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "correctpassword",
	}, "test-agent", "127.0.0.1")

	// Should fail with account locked error
	assert.Error(t, err)
	assert.Equal(t, ErrAccountLocked, err)

	// IncrementFailedAttempts should NOT be called (login blocked before password check)
	mockRepo.AssertNotCalled(t, "IncrementFailedAttempts", mock.Anything, mock.Anything, mock.Anything)
}

// TestSuccessfulLoginResetsAttempts tests that successful login resets failed attempts
func TestSuccessfulLoginResetsAttempts(t *testing.T) {
	// Setup
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	mockUser := &models.User{
		ID:                  1,
		Email:               "test@example.com",
		Password:            string(hashedPassword),
		Name:                "Test User",
		FailedLoginAttempts: 3, // Has some failed attempts
	}

	mockRepo := new(MockUserRepository)
	mockSessionService := new(MockSessionService)

	// User exists and is not locked
	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)
	mockRepo.On("IsAccountLocked", 1).Return(false, (*time.Time)(nil), nil)

	// Expect failed attempts to be reset on successful login
	mockRepo.On("ResetFailedAttempts", 1).Return(nil)

	// Session creation
	mockSessionService.On("CreateSession", 1, mock.AnythingOfType("string"), "test-agent", "127.0.0.1").
		Return(&models.Session{ID: 1, UserID: 1}, "refresh-token", nil)

	service := NewAuthServiceWithDeps(mockRepo, nil, mockSessionService, "test-secret")

	// Attempt login with correct password
	response, err := service.Login(models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "correctpassword",
	}, "test-agent", "127.0.0.1")

	// Should succeed
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)

	// Verify ResetFailedAttempts was called
	mockRepo.AssertCalled(t, "ResetFailedAttempts", 1)
}

// TestLockExpirationAllowsLogin tests that expired locks allow login
func TestLockExpirationAllowsLogin(t *testing.T) {
	// Setup - lock has expired
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	expiredLockTime := time.Now().Add(-1 * time.Minute) // Lock expired 1 minute ago
	mockUser := &models.User{
		ID:                  1,
		Email:               "test@example.com",
		Password:            string(hashedPassword),
		Name:                "Test User",
		FailedLoginAttempts: 5,
		LockedUntil:         &expiredLockTime,
	}

	mockRepo := new(MockUserRepository)
	mockSessionService := new(MockSessionService)

	// User exists and lock has expired (IsAccountLocked returns false)
	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)
	mockRepo.On("IsAccountLocked", 1).Return(false, (*time.Time)(nil), nil)

	// Expect failed attempts to be reset on successful login
	mockRepo.On("ResetFailedAttempts", 1).Return(nil)

	// Session creation
	mockSessionService.On("CreateSession", 1, mock.AnythingOfType("string"), "test-agent", "127.0.0.1").
		Return(&models.Session{ID: 1, UserID: 1}, "refresh-token", nil)

	service := NewAuthServiceWithDeps(mockRepo, nil, mockSessionService, "test-secret")

	// Attempt login with correct password
	response, err := service.Login(models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "correctpassword",
	}, "test-agent", "127.0.0.1")

	// Should succeed (lock has expired)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)

	// Verify ResetFailedAttempts was called
	mockRepo.AssertCalled(t, "ResetFailedAttempts", 1)
}

// TestIsAccountLockedReturnsErrorForMissingUser tests that IsAccountLocked returns error for missing user
func TestIsAccountLockedReturnsErrorForMissingUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockSessionService := new(MockSessionService)

	// User exists in GetByEmail but not in IsAccountLocked (race condition simulation)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	mockUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Name:     "Test User",
	}

	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)
	// Simulate user not found in IsAccountLocked (data inconsistency)
	mockRepo.On("IsAccountLocked", 1).Return(false, (*time.Time)(nil), errors.New("user not found (id=1): possible data inconsistency"))

	service := NewAuthServiceWithDeps(mockRepo, nil, mockSessionService, "test-secret")

	// Attempt login
	_, err := service.Login(models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password",
	}, "test-agent", "127.0.0.1")

	// Should fail with error (not silently continue)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check account lock status")
}

// TestConcurrentLoginAttempts tests that concurrent login attempts are handled correctly
func TestConcurrentLoginAttempts(t *testing.T) {
	// This test verifies that the atomic IncrementFailedAttempts prevents race conditions
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	mockUser := &models.User{
		ID:                  1,
		Email:               "test@example.com",
		Password:            string(hashedPassword),
		Name:                "Test User",
		FailedLoginAttempts: 0,
	}

	mockRepo := new(MockUserRepository)
	mockSessionService := new(MockSessionService)

	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)
	mockRepo.On("IsAccountLocked", 1).Return(false, (*time.Time)(nil), nil)

	// Simulate atomic increment returning sequential values
	mockRepo.On("IncrementFailedAttempts", 1, MaxLoginAttempts, LockoutDuration).Return(&repository.LockoutResult{
		NewAttempts: 1,
		LockedUntil: nil,
	}, nil).Once()

	service := NewAuthServiceWithDeps(mockRepo, nil, mockSessionService, "test-secret")

	// First failed attempt
	_, err := service.Login(models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}, "test-agent", "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)

	// Verify atomic increment was called
	mockRepo.AssertCalled(t, "IncrementFailedAttempts", 1, MaxLoginAttempts, LockoutDuration)
}
