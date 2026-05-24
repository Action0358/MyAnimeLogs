package services

import (
	"errors"
	"fmt"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
	"myanimelogs/internal/repository"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type AuthService struct {
	userRepo UserRepositoryInterface
	db       *database.DB
	logger   *common.Logger
}

func NewAuthService(db *database.DB) *AuthService {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	return &AuthService{
		db:     db,
		logger: logger,
	}
}

func (s *AuthService) getRepository() UserRepositoryInterface {
	if s.userRepo != nil {
		return s.userRepo
	}
	return repository.NewUserRepository(s.db)
}

func (s *AuthService) GetUserByID(userID int) (*models.UserResponse, error) {
	user, err := s.getRepository().GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return toUserResponse(user), nil
}

func (s *AuthService) UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error) {
	user, err := s.getRepository().UpdatePartial(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return toUserResponse(user), nil
}

func (s *AuthService) UpdateUserIcon(userID int, iconURL string) (*models.UserResponse, error) {
	user, err := s.getRepository().UpdateIcon(userID, iconURL)
	if err != nil {
		return nil, fmt.Errorf("failed to update user icon: %w", err)
	}
	return toUserResponse(user), nil
}

func (s *AuthService) ExportUserData(userID int, format string) (interface{}, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	favorites := []models.UserFavorite{}
	rows, err := s.db.Query(`
		SELECT id, user_id, anime_id, status, created_at, updated_at
		FROM user_favorites WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorites: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var fav models.UserFavorite
		if err := rows.Scan(&fav.ID, &fav.UserID, &fav.AnimeID, &fav.Status, &fav.CreatedAt, &fav.UpdatedAt); err != nil {
			continue
		}
		favorites = append(favorites, fav)
	}

	ratings := []models.UserRating{}
	rows, err = s.db.Query(`
		SELECT id, user_id, anime_id, rating, comment, created_at, updated_at
		FROM user_ratings WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var rating models.UserRating
		if err := rows.Scan(&rating.ID, &rating.UserID, &rating.AnimeID, &rating.Rating, &rating.Comment, &rating.CreatedAt, &rating.UpdatedAt); err != nil {
			continue
		}
		ratings = append(ratings, rating)
	}

	data := map[string]interface{}{
		"user_info":   user,
		"favorites":   favorites,
		"ratings":     ratings,
		"exported_at": time.Now(),
	}

	if format == "csv" {
		csv := "type,id,data\n"
		csv += fmt.Sprintf("user,%d,\"%s\"\n", user.ID, user.Name)
		for _, fav := range favorites {
			csv += fmt.Sprintf("favorite,%d,\"%s\"\n", fav.AnimeID, fav.Status)
		}
		return csv, nil
	}

	return data, nil
}

func toUserResponse(user *models.User) *models.UserResponse {
	return &models.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Description:   user.Description,
		Icon:          user.Icon,
		FavoriteAnime: user.FavoriteAnime,
		SnsLink:       user.SnsLink,
		CreatedAt:     user.CreatedAt,
	}
}
