package models

import (
	"time"
)

type Anime struct {
	ID            int       `json:"id" db:"id"`
	Title         string    `json:"title" db:"title"`
	ImageURL      *string   `json:"image_url" db:"image_url"`
	Season        *string   `json:"season" db:"season"` // Format: "YYYY-season" (e.g., "2025-summer")
	Status        *string   `json:"status" db:"status"`
	WatchersCount int       `json:"watchers_count" db:"watchers_count"`
	MalAnimeID    *string   `json:"mal_anime_id" db:"mal_anime_id"` // MyAnimeList ID
	AnnictID      *int      `json:"annict_id" db:"annict_id"`       // Annict ID
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// AniList から取得する画像（DBにキャッシュ）
	CoverImageLarge  string `json:"cover_image_large,omitempty" db:"cover_image_large"`
	CoverImageMedium string `json:"cover_image_medium,omitempty" db:"cover_image_medium"`
}

type AnimeWithUserData struct {
	Anime
	UserFavorite       *UserFavorite `json:"user_favorite,omitempty"`
	UserRating         *UserRating   `json:"user_rating,omitempty"`
	WatchProgress      float64       `json:"watch_progress"`
	Episodes           []Episode     `json:"episodes,omitempty"`
	AverageRating      *float64      `json:"average_rating,omitempty"`
	RatingCount        int           `json:"rating_count"`
	RatingDistribution map[int]int   `json:"rating_distribution,omitempty"`
}

type Episode struct {
	ID              int        `json:"id" db:"id"`
	AnimeID         int        `json:"anime_id" db:"anime_id"`
	Title           string     `json:"title" db:"title"`
	EpisodeNumber   int        `json:"episode_number" db:"episode_number"`
	AirDate         *time.Time `json:"air_date,omitempty" db:"air_date"`       // null if not yet fetched from Jikan API
	DurationMinutes *int       `json:"duration_minutes" db:"duration_minutes"` // null if not yet fetched from Jikan API
	Viewed          bool       `json:"viewed"`                                 // false if not authenticated or not viewed
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type UserFavorite struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	AnimeID   int       `json:"anime_id" db:"anime_id"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// EpisodeSyncStatus represents the sync status for an anime's episodes
type EpisodeSyncStatus struct {
	AnimeID        int      `json:"anime_id"`
	AnimeTitle     string   `json:"anime_title"`
	AnnictID       *int     `json:"annict_id"`
	TotalEpisodes  int      `json:"total_episodes"`
	DefaultTitles  int      `json:"default_titles"`
	CustomTitles   int      `json:"custom_titles"`
	SyncPercentage float64  `json:"sync_percentage"`
	NeedSync       bool     `json:"need_sync"`
	SampleEpisodes []string `json:"sample_episodes,omitempty"`
}

type UserFavoriteRequest struct {
	Status string `json:"status" binding:"required,oneof=watching completed planned dropped"`
}

type WatchStatus struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	EpisodeID int       `json:"episode_id" db:"episode_id"`
	Viewed    bool      `json:"viewed" db:"viewed"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type WatchStatusRequest struct {
	Viewed bool `json:"viewed" binding:"required"`
}

type UserRating struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	AnimeID   int       `json:"anime_id" db:"anime_id"`
	Rating    float64   `json:"rating" db:"rating"`
	Comment   *string   `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserRatingRequest struct {
	Rating  float64 `json:"rating" binding:"required,min=1.0,max=5.0"`
	Comment *string `json:"comment,omitempty" binding:"omitempty,max=1000"`
}

type WatchHistory struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	EpisodeID    int       `json:"episode_id" db:"episode_id"`
	WatchedAt    time.Time `json:"watched_at" db:"watched_at"`
	AnimeTitle   string    `json:"anime_title,omitempty"`
	EpisodeTitle string    `json:"episode_title,omitempty"`
}

type AnimeGenre struct {
	ID      int    `json:"id" db:"id"`
	AnimeID int    `json:"anime_id" db:"anime_id"`
	Genre   string `json:"genre" db:"genre"`
}

type AnimeSearchParams struct {
	Keyword string `form:"keyword"`
	Season  string `form:"season"`
	Status  string `form:"status"`
	Genre   string `form:"genre"`
	Year    int    `form:"year"`
	Sort    string `form:"sort"`
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
}

// AdvancedSearchParams - 高度検索用のパラメータ
type AdvancedSearchParams struct {
	Query      string   `form:"q"`
	Genres     []string `form:"genre"`
	Year       int      `form:"year"`
	Season     string   `form:"season"`
	MinRating  float64  `form:"min_rating"`
	MaxRating  float64  `form:"max_rating"`
	Status     string   `form:"status"`
	UserStatus string   `form:"user_status"` // watching, completed, etc.
	Sort       string   `form:"sort"`
	Order      string   `form:"order"`
	Page       int      `form:"page"`
	PerPage    int      `form:"per_page"`
}

// FilterOptions - フィルター情報
type FilterOptions struct {
	Genres  []GenreInfo `json:"genres"`
	Years   []int       `json:"years"`
	Seasons []string    `json:"seasons"`
	Ratings RatingStats `json:"ratings"`
}

// GenreInfo - ジャンル情報
type GenreInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// RatingStats - 評価統計
type RatingStats struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

// SearchLog - 検索ログ
type SearchLog struct {
	ID          int       `json:"id" db:"id"`
	UserID      *int      `json:"user_id,omitempty" db:"user_id"`
	Query       string    `json:"query" db:"query"`
	Filters     string    `json:"filters" db:"filters"`
	ResultCount int       `json:"result_count" db:"result_count"`
	ClickedID   *int      `json:"clicked_id,omitempty" db:"clicked_anime_id"`
	SearchedAt  time.Time `json:"searched_at" db:"searched_at"`
}

type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	TotalPages  int `json:"total_pages"`
	CurrentPage int `json:"current_page"`
	Total       int `json:"total"`
}
