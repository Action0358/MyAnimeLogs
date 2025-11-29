package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
	"myanimelogs/internal/utils"
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// SessionService handles session management
type SessionService struct {
	db        *database.DB
	logger    *common.Logger
	jwtSecret string
}

// NewSessionService creates a new SessionService
func NewSessionService(db *database.DB, logger *common.Logger, jwtSecret string) *SessionService {
	return &SessionService{
		db:        db,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(userID int, jti, userAgent, ipAddress string) (*models.Session, string, error) {
	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(userID, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Calculate expiry (7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Insert session into database
	var session models.Session
	err = s.db.QueryRow(`
		INSERT INTO sessions (user_id, refresh_token, access_token_jti, user_agent, ip_address, expires_at, last_activity_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		RETURNING id, user_id, refresh_token, access_token_jti, user_agent, ip_address, expires_at, created_at, updated_at, last_activity_at
	`, userID, refreshToken, jti, userAgent, ipAddress, expiresAt).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.AccessTokenJTI,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.LastActivityAt,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info("Session created",
		"session_id", session.ID,
		"user_id", userID,
		"expires_at", expiresAt)

	return &session, refreshToken, nil
}

// ValidateSession validates a refresh token and returns the session
func (s *SessionService) ValidateSession(refreshToken string) (*models.Session, error) {
	// Validate refresh token JWT
	claims, err := utils.ValidateRefreshToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get session from database
	var session models.Session
	err = s.db.QueryRow(`
		SELECT id, user_id, refresh_token, access_token_jti, user_agent, ip_address, expires_at, created_at, updated_at, last_activity_at
		FROM sessions
		WHERE refresh_token = $1
	`, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.AccessTokenJTI,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.LastActivityAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		s.DeleteSession(refreshToken)
		return nil, ErrSessionExpired
	}

	// Verify JWT expiry matches database
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// UpdateActivity updates the last activity time for a session
func (s *SessionService) UpdateActivity(sessionID int) error {
	_, err := s.db.Exec(`
		UPDATE sessions
		SET last_activity_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	return nil
}

// UpdateActivityByJTI updates the last activity time using JTI
func (s *SessionService) UpdateActivityByJTI(jti string) error {
	_, err := s.db.Exec(`
		UPDATE sessions
		SET last_activity_at = CURRENT_TIMESTAMP
		WHERE access_token_jti = $1
	`, jti)
	if err != nil {
		return fmt.Errorf("failed to update activity by JTI: %w", err)
	}
	return nil
}

// UpdateSessionJTI updates the access token JTI for a session
func (s *SessionService) UpdateSessionJTI(sessionID int, newJTI string) error {
	_, err := s.db.Exec(`
		UPDATE sessions
		SET access_token_jti = $1, last_activity_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, newJTI, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session JTI: %w", err)
	}
	return nil
}

// DeleteSession deletes a session by refresh token
func (s *SessionService) DeleteSession(refreshToken string) error {
	result, err := s.db.Exec(`
		DELETE FROM sessions WHERE refresh_token = $1
	`, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	s.logger.Info("Session deleted", "refresh_token_prefix", refreshToken[:10]+"...")
	return nil
}

// DeleteSessionByID deletes a session by ID (with user verification)
func (s *SessionService) DeleteSessionByID(sessionID, userID int) error {
	result, err := s.db.Exec(`
		DELETE FROM sessions WHERE id = $1 AND user_id = $2
	`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	s.logger.Info("Session deleted", "session_id", sessionID, "user_id", userID)
	return nil
}

// DeleteUserSessions deletes all sessions for a user
func (s *SessionService) DeleteUserSessions(userID int) error {
	result, err := s.db.Exec(`
		DELETE FROM sessions WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	s.logger.Info("User sessions deleted", "user_id", userID, "count", rowsAffected)
	return nil
}

// GetUserSessions returns all active sessions for a user
func (s *SessionService) GetUserSessions(userID int) ([]models.Session, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, refresh_token, access_token_jti, user_agent, ip_address, expires_at, created_at, updated_at, last_activity_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_activity_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	sessions := []models.Session{}
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&session.AccessTokenJTI,
			&session.UserAgent,
			&session.IPAddress,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.LastActivityAt,
		)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// CleanupExpiredSessions deletes all expired sessions
func (s *SessionService) CleanupExpiredSessions() (int, error) {
	var deletedCount int
	err := s.db.QueryRow(`SELECT delete_expired_sessions()`).Scan(&deletedCount)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	s.logger.Info("Expired sessions cleaned up", "count", deletedCount)
	return deletedCount, nil
}

// GetSessionByJTI gets a session by access token JTI
func (s *SessionService) GetSessionByJTI(jti string) (*models.Session, error) {
	var session models.Session
	err := s.db.QueryRow(`
		SELECT id, user_id, refresh_token, access_token_jti, user_agent, ip_address, expires_at, created_at, updated_at, last_activity_at
		FROM sessions
		WHERE access_token_jti = $1
	`, jti).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.AccessTokenJTI,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.LastActivityAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by JTI: %w", err)
	}

	return &session, nil
}
