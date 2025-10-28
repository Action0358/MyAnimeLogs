package models

import (
	"time"
)

type User struct {
	ID            int       `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Email         string    `json:"email" db:"email"`
	Password      string    `json:"-" db:"password"` // Never expose password in JSON
	Description   *string   `json:"description" db:"description"`
	Icon          *string   `json:"icon" db:"icon"`
	FavoriteAnime *string   `json:"favorite_anime" db:"favorite_anime"`
	SnsLink       *string   `json:"sns_link" db:"sns_link"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type UserRegisterRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserUpdateRequest struct {
	Name          *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description   *string `json:"description,omitempty" binding:"omitempty,max=10000"`
	FavoriteAnime *string `json:"favorite_anime,omitempty" binding:"omitempty,max=300"`
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

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	User        UserResponse `json:"user"`
}

type UserStats struct {
	TotalEpisodesWatched int      `json:"total_episodes_watched"`
	TotalAnimesCompleted int      `json:"total_animes_completed"`
	TotalWatchTimeHours  float64  `json:"total_watch_time_hours"`
	FavoriteGenres       []string `json:"favorite_genres"`
	ThisMonthEpisodes    int      `json:"this_month_episodes"`
}
