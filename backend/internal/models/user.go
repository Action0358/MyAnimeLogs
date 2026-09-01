package models

import (
	"time"
)

type User struct {
	ID         int       `json:"id" db:"id"`
	AuthUserID *string   `json:"auth_user_id,omitempty" db:"auth_user_id"`
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// AvatarURL はGoogleアカウントのプロフィール画像URL。DBには保持せず、
	// 認証時にSupabase JWT（user_metadata.avatar_url/picture）から都度取得する（#779）。
	AvatarURL *string   `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}
