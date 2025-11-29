package models

import "time"

// Session represents a user session with refresh token
type Session struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"user_id" db:"user_id"`
	RefreshToken   string    `json:"-" db:"refresh_token"` // Never expose in JSON
	AccessTokenJTI string    `json:"-" db:"access_token_jti"`
	UserAgent      *string   `json:"user_agent" db:"user_agent"`
	IPAddress      *string   `json:"ip_address" db:"ip_address"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	LastActivityAt time.Time `json:"last_activity_at" db:"last_activity_at"`
}

// SessionResponse is the public representation of a session
type SessionResponse struct {
	ID             int       `json:"id"`
	UserAgent      *string   `json:"user_agent"`
	IPAddress      *string   `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsCurrent      bool      `json:"is_current"`
}

// ToResponse converts Session to SessionResponse
func (s *Session) ToResponse(currentJTI string) SessionResponse {
	return SessionResponse{
		ID:             s.ID,
		UserAgent:      s.UserAgent,
		IPAddress:      s.IPAddress,
		CreatedAt:      s.CreatedAt,
		LastActivityAt: s.LastActivityAt,
		ExpiresAt:      s.ExpiresAt,
		IsCurrent:      s.AccessTokenJTI == currentJTI,
	}
}
