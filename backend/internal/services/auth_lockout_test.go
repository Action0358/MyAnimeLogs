package services

import (
	"testing"
	"time"

	"myanimelogs/internal/models"
	"myanimelogs/internal/repository"

	"github.com/stretchr/testify/assert"
)

// TestAccountLockoutConstants verifies the lockout configuration
func TestAccountLockoutConstants(t *testing.T) {
	assert.Equal(t, 5, MaxLoginAttempts, "MaxLoginAttempts should be 5")
	assert.Equal(t, 15*time.Minute, LockoutDuration, "LockoutDuration should be 15 minutes")
}

// TestAccountLockoutErrors verifies error constants
func TestAccountLockoutErrors(t *testing.T) {
	assert.NotNil(t, ErrAccountLocked, "ErrAccountLocked should be defined")
	assert.Equal(t, "account is locked", ErrAccountLocked.Error())
}

// TestCheckAccountLock tests the CheckAccountLock method
func TestCheckAccountLock(t *testing.T) {
	tests := []struct {
		name            string
		userID          int
		mockLocked      bool
		mockLockedUntil *time.Time
		mockError       error
		expectedLocked  bool
		expectedError   bool
	}{
		{
			name:            "Account not locked",
			userID:          1,
			mockLocked:      false,
			mockLockedUntil: nil,
			mockError:       nil,
			expectedLocked:  false,
			expectedError:   false,
		},
		{
			name:       "Account is locked",
			userID:     2,
			mockLocked: true,
			mockLockedUntil: func() *time.Time {
				t := time.Now().Add(10 * time.Minute)
				return &t
			}(),
			mockError:      nil,
			expectedLocked: true,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockRepo.On("IsAccountLocked", tt.userID).Return(tt.mockLocked, tt.mockLockedUntil, tt.mockError)

			service := &AuthService{
				userRepo: mockRepo,
			}

			isLocked, _, err := service.CheckAccountLock(tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLocked, isLocked)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetRemainingAttempts tests the GetRemainingAttempts method
func TestGetRemainingAttempts(t *testing.T) {
	tests := []struct {
		name              string
		userID            int
		mockUser          *models.User
		mockError         error
		expectedRemaining int
		expectedError     error
	}{
		{
			name:              "User not found",
			userID:            999,
			mockUser:          nil,
			mockError:         nil,
			expectedRemaining: 0,
			expectedError:     ErrUserNotFound,
		},
		{
			name:   "No failed attempts",
			userID: 1,
			mockUser: &models.User{
				ID:                  1,
				Email:               "test@example.com",
				FailedLoginAttempts: 0,
			},
			mockError:         nil,
			expectedRemaining: 5,
			expectedError:     nil,
		},
		{
			name:   "Two failed attempts",
			userID: 2,
			mockUser: &models.User{
				ID:                  2,
				Email:               "test2@example.com",
				FailedLoginAttempts: 2,
			},
			mockError:         nil,
			expectedRemaining: 3,
			expectedError:     nil,
		},
		{
			name:   "Max attempts reached",
			userID: 3,
			mockUser: &models.User{
				ID:                  3,
				Email:               "test3@example.com",
				FailedLoginAttempts: 5,
			},
			mockError:         nil,
			expectedRemaining: 0,
			expectedError:     nil,
		},
		{
			name:   "More than max attempts (edge case)",
			userID: 4,
			mockUser: &models.User{
				ID:                  4,
				Email:               "test4@example.com",
				FailedLoginAttempts: 10,
			},
			mockError:         nil,
			expectedRemaining: 0,
			expectedError:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockRepo.On("GetByID", tt.userID).Return(tt.mockUser, tt.mockError)

			service := &AuthService{
				userRepo: mockRepo,
			}

			remaining, err := service.GetRemainingAttempts(tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRemaining, remaining)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUnlockAccount tests the UnlockAccount method
func TestUnlockAccount(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("ResetFailedAttempts", 1).Return(nil)

	service := &AuthService{
		userRepo: mockRepo,
	}

	err := service.UnlockAccount(1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestIncrementFailedAttempts tests the incrementFailedAttempts method
func TestIncrementFailedAttempts(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		mockResult    *repository.LockoutResult
		mockError     error
		expectedError bool
	}{
		{
			name:   "Increment success - not locked",
			userID: 1,
			mockResult: &repository.LockoutResult{
				NewAttempts: 1,
				LockedUntil: nil,
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:   "Increment success - account locked",
			userID: 2,
			mockResult: &repository.LockoutResult{
				NewAttempts: 5,
				LockedUntil: func() *time.Time {
					t := time.Now().Add(15 * time.Minute)
					return &t
				}(),
			},
			mockError:     nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockRepo.On("IncrementFailedAttempts", tt.userID, MaxLoginAttempts, LockoutDuration).Return(tt.mockResult, tt.mockError)

			service := NewAuthServiceWithRepo(mockRepo, nil, "test-secret")

			err := service.incrementFailedAttempts(tt.userID, mockRepo)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
