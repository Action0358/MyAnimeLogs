package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, name, email, password, description, icon, favorite_anime, sns_link,
		       failed_login_attempts, last_failed_login_at, locked_until,
		       created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Description, &user.Icon, &user.FavoriteAnime, &user.SnsLink,
		&user.FailedLoginAttempts, &user.LastFailedLoginAt, &user.LockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, name, email, description, icon, favorite_anime, sns_link,
		       failed_login_attempts, last_failed_login_at, locked_until,
		       created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Description, &user.Icon, &user.FavoriteAnime, &user.SnsLink,
		&user.FailedLoginAttempts, &user.LastFailedLoginAt, &user.LockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Create(user *models.User) error {
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`, user.Name, user.Email, user.Password).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(user *models.User) error {
	err := r.db.QueryRow(`
		UPDATE users SET name = $1, description = $2, icon = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, name, email, description, icon, created_at, updated_at
	`, user.Name, user.Description, user.Icon, user.ID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// addFieldToUpdate is a helper function to add a field to the update query
func addFieldToUpdate(setParts *[]string, args *[]interface{}, argIndex *int, fieldName string, value *string) {
	if value != nil {
		*setParts = append(*setParts, fmt.Sprintf("%s = $%d", fieldName, *argIndex))
		*args = append(*args, *value)
		*argIndex++
	}
}

func (r *UserRepository) UpdatePartial(userID int, req models.UserUpdateRequest) (*models.User, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	// Add fields to update using helper function
	addFieldToUpdate(&setParts, &args, &argIndex, "name", req.Name)
	addFieldToUpdate(&setParts, &args, &argIndex, "description", req.Description)
	addFieldToUpdate(&setParts, &args, &argIndex, "favorite_anime", req.FavoriteAnime)
	addFieldToUpdate(&setParts, &args, &argIndex, "sns_link", req.SnsLink)

	if len(setParts) == 0 {
		// No fields to update, just return current user
		return r.GetByID(userID)
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, name, email, description, icon, favorite_anime, sns_link, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var user models.User
	err := r.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon, &user.FavoriteAnime, &user.SnsLink, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}

// UpdateIcon updates only the user's icon field
func (r *UserRepository) UpdateIcon(userID int, iconURL string) (*models.User, error) {
	var user models.User
	var iconValue interface{}

	if iconURL == "" {
		iconValue = nil
	} else {
		iconValue = iconURL
	}

	err := r.db.QueryRow(`
		UPDATE users SET icon = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, name, email, description, icon, favorite_anime, sns_link, created_at, updated_at
	`, iconValue, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon,
		&user.FavoriteAnime, &user.SnsLink, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user icon: %w", err)
	}

	return &user, nil
}

// LockoutResult contains the result of an atomic lockout operation
type LockoutResult struct {
	NewAttempts int
	LockedUntil *time.Time
}

// IncrementFailedAttempts atomically increments the failed login attempts counter
// and locks the account if max attempts is reached. Returns the new attempt count
// and lock time (if locked).
func (r *UserRepository) IncrementFailedAttempts(userID int, maxAttempts int, lockoutDuration time.Duration) (*LockoutResult, error) {
	now := time.Now()
	lockTime := now.Add(lockoutDuration)

	// Atomic update with RETURNING to get the new values
	// This prevents race conditions by doing everything in a single query
	var newAttempts int
	var lockedUntil *time.Time

	err := r.db.QueryRow(`
		UPDATE users 
		SET failed_login_attempts = failed_login_attempts + 1,
		    last_failed_login_at = $1,
		    locked_until = CASE 
		        WHEN failed_login_attempts + 1 >= $2 THEN $3
		        ELSE locked_until
		    END
		WHERE id = $4
		RETURNING failed_login_attempts, locked_until
	`, now, maxAttempts, lockTime, userID).Scan(&newAttempts, &lockedUntil)

	if err != nil {
		return nil, fmt.Errorf("failed to increment failed login attempts: %w", err)
	}

	return &LockoutResult{
		NewAttempts: newAttempts,
		LockedUntil: lockedUntil,
	}, nil
}

// ResetFailedAttempts resets the failed login attempts counter and unlocks the account
func (r *UserRepository) ResetFailedAttempts(userID int) error {
	_, err := r.db.Exec(`
		UPDATE users 
		SET failed_login_attempts = 0,
		    last_failed_login_at = NULL,
		    locked_until = NULL
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to reset failed login attempts: %w", err)
	}
	return nil
}

// IsAccountLocked checks if the account is currently locked
func (r *UserRepository) IsAccountLocked(userID int) (bool, *time.Time, error) {
	var lockedUntil *time.Time
	err := r.db.QueryRow(`
		SELECT locked_until FROM users WHERE id = $1
	`, userID).Scan(&lockedUntil)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check account lock status: %w", err)
	}

	if lockedUntil != nil && lockedUntil.After(time.Now()) {
		return true, lockedUntil, nil
	}
	return false, nil, nil
}
