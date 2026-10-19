package services

import (
	"errors"
	"fmt"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
	"myanimelogs/internal/repository"
)

var (
	ErrUserNotFound = errors.New("user not found")
	// ErrAdminAPIUnavailable は service_role キー未設定の環境で返る。
	// 設定漏れを 500 で潰さず、原因が分かる形で表面化させる。
	ErrAdminAPIUnavailable = errors.New("supabase admin api is not configured")
)

type AuthService struct {
	userRepo UserRepositoryInterface
	db       *database.DB
	logger   *common.Logger
	// アカウント削除でのみ使用する依存。未設定でも他の機能は動作する。
	admin *SupabaseAdminClient
}

func NewAuthService(db *database.DB) *AuthService {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	return &AuthService{
		db:     db,
		logger: logger,
	}
}

// NewAuthServiceWithAdmin builds an AuthService that can delete accounts.
// admin が nil または未設定の場合、DeleteAccount は ErrAdminAPIUnavailable を返す。
func NewAuthServiceWithAdmin(db *database.DB, admin *SupabaseAdminClient) *AuthService {
	s := NewAuthService(db)
	s.admin = admin
	return s
}

func (s *AuthService) getRepository() UserRepositoryInterface {
	if s.userRepo != nil {
		return s.userRepo
	}
	return repository.NewUserRepository(s.db)
}

// DeleteAccount permanently removes the user's account.
//
// auth.users を消すと fk_users_auth_user_id の ON DELETE CASCADE により
// public.users が消え、user_favorites / user_watch_status も連鎖削除される。
// search_logs も ON DELETE CASCADE により連鎖削除される（migration 054）。
// したがってアプリ側で個別に DELETE を発行する必要はない。
//
// アイコンはプリセットIDを users.icon に持つだけでファイル実体がないため、
// 連鎖削除以外の後始末は不要（#763 でアップロードを廃止した）。
func (s *AuthService) DeleteAccount(userID int, authUserID string) error {
	if s.admin == nil || !s.admin.IsConfigured() {
		return ErrAdminAPIUnavailable
	}

	if err := s.admin.DeleteUser(authUserID); err != nil {
		return fmt.Errorf("failed to delete auth user: %w", err)
	}

	return nil
}

func (s *AuthService) GetUserByID(userID int, avatarURL string) (*models.UserResponse, error) {
	user, err := s.getRepository().GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return toUserResponse(user, avatarURL), nil
}

func toUserResponse(user *models.User, avatarURL string) *models.UserResponse {
	var avatarURLPtr *string
	if avatarURL != "" {
		avatarURLPtr = &avatarURL
	}
	return &models.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: avatarURLPtr,
		CreatedAt: user.CreatedAt,
	}
}
