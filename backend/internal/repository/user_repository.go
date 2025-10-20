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
		SELECT id, name, email, password, description, icon, favorite_anime, sns_link, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Description, &user.Icon, &user.FavoriteAnime, &user.SnsLink, &user.CreatedAt, &user.UpdatedAt,
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
		SELECT id, name, email, description, icon, favorite_anime, sns_link, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Description, &user.Icon, &user.FavoriteAnime, &user.SnsLink, &user.CreatedAt, &user.UpdatedAt,
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
	addFieldToUpdate(&setParts, &args, &argIndex, "icon", req.Icon)
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
