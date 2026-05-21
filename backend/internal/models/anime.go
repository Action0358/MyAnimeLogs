package models

import (
	"time"
)

type Anime struct {
	ID              int       `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	TitleKana       *string   `json:"title_kana,omitempty" db:"title_kana"`
	TitleEn         *string   `json:"title_en,omitempty" db:"title_en"`
	ImageURL        *string   `json:"image_url" db:"image_url"`
	Season          *string   `json:"season" db:"season"` // Format: "YYYY-season" (e.g., "2025-summer")
	Status          *string   `json:"status" db:"status"`
	Media           *string   `json:"media,omitempty" db:"media"`                         // TV, OVA, Movie, etc.
	Description     *string   `json:"description,omitempty" db:"description"`             // Synopsis from AniList
	ReleasedOn      *string   `json:"released_on,omitempty" db:"released_on"`             // Release date from Annict
	OfficialSiteURL *string   `json:"official_site_url,omitempty" db:"official_site_url"` // Official website
	WatchersCount   int       `json:"watchers_count" db:"watchers_count"`
	MalAnimeID      *string   `json:"mal_anime_id" db:"mal_anime_id"` // MyAnimeList ID
	AnnictID        *int      `json:"annict_id" db:"annict_id"`       // Annict ID
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`

	// AniList から取得する画像（DBにキャッシュ）
	CoverImageExtraLarge string     `json:"cover_image_extra_large,omitempty" db:"cover_image_extra_large"`
	CoverImageLarge      string     `json:"cover_image_large,omitempty" db:"cover_image_large"`
	CoverImageMedium     string     `json:"cover_image_medium,omitempty" db:"cover_image_medium"`
	CoverImageUpdatedAt  *time.Time `json:"-" db:"cover_image_updated_at"` // TTL判定用（APIレスポンスには含めない）

	// シリーズ作品の統合管理用（内部識別用）
	SeriesTitle  *string `json:"series_title,omitempty" db:"series_title"`     // 正規化後のシリーズ名
	SeriesTMDBID *int    `json:"series_tmdb_id,omitempty" db:"series_tmdb_id"` // シリーズ代表のTMDB ID

	// Jikan APIから取得する放送情報
	BroadcastDay      *string `json:"broadcast_day,omitempty" db:"broadcast_day"`           // 放送曜日（例: "Fridays"）
	BroadcastTime     *string `json:"broadcast_time,omitempty" db:"broadcast_time"`         // 放送時間（例: "23:30"）
	BroadcastTimezone *string `json:"broadcast_timezone,omitempty" db:"broadcast_timezone"` // タイムゾーン（例: "Asia/Tokyo"）

	// 映画作品の上映時間（Jikan APIから取得）
	DurationMinutes *int `json:"duration_minutes,omitempty" db:"duration_minutes"` // 映画の上映時間（分）

	// 差分同期用トラッキングフィールド
	LastSyncedAt     *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`         // 最終同期日時
	AnnictUpdatedAt  *time.Time `json:"annict_updated_at,omitempty" db:"annict_updated_at"`   // Annict側の更新日時
	MetadataSyncedAt *time.Time `json:"metadata_synced_at,omitempty" db:"metadata_synced_at"` // メタデータの最終同期日時
	SyncStatus       string     `json:"sync_status" db:"sync_status"`                         // pending/synced/failed
	SyncError        *string    `json:"sync_error,omitempty" db:"sync_error"`                 // エラー内容
}

type AnimeWithUserData struct {
	Anime
	UserFavorite       *UserFavorite   `json:"user_favorite,omitempty"`
	UserRating         *UserRating     `json:"user_rating,omitempty"`
	Genres             []string        `json:"genres,omitempty"`
	AverageRating      *float64        `json:"average_rating,omitempty"`
	RatingCount        int             `json:"rating_count"`
	RatingDistribution map[int]int     `json:"rating_distribution,omitempty"`
	WatchProviders     []WatchProvider `json:"watch_providers,omitempty"`
}

// WatchProvider represents a streaming service provider
type WatchProvider struct {
	ProviderID   int    `json:"provider_id"`
	ProviderName string `json:"provider_name"`
	LogoURL      string `json:"logo_url"`
}

type UserFavorite struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	AnimeID   int       `json:"anime_id" db:"anime_id"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserFavoriteRequest struct {
	Status string `json:"status" binding:"required,oneof=watching completed planned dropped"`
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

