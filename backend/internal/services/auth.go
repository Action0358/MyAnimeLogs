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
		SELECT id, user_id, anime_id, created_at, updated_at
		FROM user_favorites WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorites: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var fav models.UserFavorite
		if err := rows.Scan(&fav.ID, &fav.UserID, &fav.AnimeID, &fav.CreatedAt, &fav.UpdatedAt); err != nil {
			continue
		}
		favorites = append(favorites, fav)
	}

	watchStatuses := []models.UserWatchStatus{}
	rows, err = s.db.Query(`
		SELECT id, user_id, anime_id, status, created_at, updated_at
		FROM user_watch_status WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get watch statuses: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ws models.UserWatchStatus
		if err := rows.Scan(&ws.ID, &ws.UserID, &ws.AnimeID, &ws.Status, &ws.CreatedAt, &ws.UpdatedAt); err != nil {
			continue
		}
		watchStatuses = append(watchStatuses, ws)
	}

	data := map[string]interface{}{
		"user_info":      user,
		"favorites":      favorites,
		"watch_statuses": watchStatuses,
		"exported_at":    time.Now(),
	}

	if format == "csv" {
		csv := "type,id,data\n"
		csv += fmt.Sprintf("user,%d,\"%s\"\n", user.ID, user.Name)
		for _, fav := range favorites {
			csv += fmt.Sprintf("favorite,%d,\"\"\n", fav.AnimeID)
		}
		for _, ws := range watchStatuses {
			csv += fmt.Sprintf("watch_status,%d,\"%s\"\n", ws.AnimeID, ws.Status)
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
