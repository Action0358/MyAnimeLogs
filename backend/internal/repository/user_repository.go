package repository

import (
	"database/sql"
	"fmt"
	"strings"

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
		SELECT id, name, email, description, icon, favorite_anime, sns_link,
		       auth_user_id, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Description, &user.Icon, &user.FavoriteAnime, &user.SnsLink,
		&user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
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
		       auth_user_id, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Description, &user.Icon, &user.FavoriteAnime, &user.SnsLink,
		&user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetOrCreateByAuthUserID returns an existing user or creates a new one for the given Supabase auth UUID.
func (r *UserRepository) GetOrCreateByAuthUserID(authUserID, email, name string) (*models.User, error) {
	var user models.User

	// Try INSERT ... ON CONFLICT to atomically get or create.
	// The partial index (WHERE auth_user_id IS NOT NULL) must be repeated in the
	// ON CONFLICT clause so PostgreSQL can infer it as the conflict target.
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, auth_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (auth_user_id) WHERE auth_user_id IS NOT NULL
		DO UPDATE SET auth_user_id = EXCLUDED.auth_user_id
		RETURNING id, name, email, description, icon, favorite_anime, sns_link,
		          auth_user_id, created_at, updated_at
	`, name, email, authUserID).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Description, &user.Icon, &user.FavoriteAnime, &user.SnsLink,
		&user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create user by auth_user_id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Create(user *models.User) error {
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, auth_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`, user.Name, user.Email, user.AuthUserID).Scan(
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

func addFieldToUpdate(setParts *[]string, args *[]interface{}, argIndex *int, fieldName string, value *string) {
	if value != nil {
		*setParts = append(*setParts, fmt.Sprintf("%s = $%d", fieldName, *argIndex))
		*args = append(*args, *value)
		*argIndex++
	}
}

func (r *UserRepository) UpdatePartial(userID int, req models.UserUpdateRequest) (*models.User, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	addFieldToUpdate(&setParts, &args, &argIndex, "name", req.Name)
	addFieldToUpdate(&setParts, &args, &argIndex, "description", req.Description)
	addFieldToUpdate(&setParts, &args, &argIndex, "favorite_anime", req.FavoriteAnime)
	addFieldToUpdate(&setParts, &args, &argIndex, "sns_link", req.SnsLink)

	if len(setParts) == 0 {
		return r.GetByID(userID)
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, name, email, description, icon, favorite_anime, sns_link,
		          auth_user_id, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var user models.User
	err := r.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon,
		&user.FavoriteAnime, &user.SnsLink, &user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// UpdateIcon updates only the user's icon field.
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
		RETURNING id, name, email, description, icon, favorite_anime, sns_link,
		          auth_user_id, created_at, updated_at
	`, iconValue, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon,
		&user.FavoriteAnime, &user.SnsLink, &user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user icon: %w", err)
	}

	return &user, nil
}
