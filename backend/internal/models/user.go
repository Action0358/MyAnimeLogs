package models

import (
	"time"
)

type User struct {
	ID            int       `json:"id" db:"id"`
	AuthUserID    *string   `json:"auth_user_id,omitempty" db:"auth_user_id"`
	Name          string    `json:"name" db:"name"`
	Email         string    `json:"email" db:"email"`
	Description   *string   `json:"description" db:"description"`
	Icon          *string   `json:"icon" db:"icon"`
	FavoriteAnime *string   `json:"favorite_anime" db:"favorite_anime"`
	SnsLink       *string   `json:"sns_link" db:"sns_link"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type UserUpdateRequest struct {
	Name          *string `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Description   *string `json:"description,omitempty" binding:"omitempty,max=500"`
	FavoriteAnime *string `json:"favorite_anime,omitempty" binding:"omitempty,max=200"`
	SnsLink       *string `json:"sns_link,omitempty" binding:"omitempty,url,max=500"`
}

type UserResponse struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Description   *string   `json:"description"`
	Icon          *string   `json:"icon"`
	FavoriteAnime *string   `json:"favorite_anime"`
	SnsLink       *string   `json:"sns_link"`
	CreatedAt     time.Time `json:"created_at"`
}
