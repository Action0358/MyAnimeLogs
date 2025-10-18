package models

// AnnictWork represents a single anime work from the Annict REST API response.
type AnnictWork struct {
	ID              int               `json:"id"`
	Title           string            `json:"title"`
	TitleKana       string            `json:"title_kana"`
	TitleEn         string            `json:"title_en"`
	Media           string            `json:"media"`
	MediaText       string            `json:"media_text"`
	SeasonName      string            `json:"season_name"` // Format: "YYYY-season" (e.g., "2025-summer")
	ReleasedOn      string            `json:"released_on"`
	ReleasedOnAbout string            `json:"released_on_about"`
	OfficialSiteURL string            `json:"official_site_url"`
	WikipediaURL    string            `json:"wikipedia_url"`
	TwitterUsername string            `json:"twitter_username"`
	TwitterHashtag  string            `json:"twitter_hashtag"`
	MalAnimeID      *string           `json:"mal_anime_id,omitempty"`
	Images          *AnnictWorkImages `json:"images,omitempty"`
	EpisodesCount   int               `json:"episodes_count"`
	WatchersCount   int               `json:"watchers_count"`
	ReviewsCount    int               `json:"reviews_count"`
	NoEpisodes      bool              `json:"no_episodes"`
	// Anilist API から取得した画像（リアルタイム取得）
	CoverImageLarge  string `json:"cover_image_large,omitempty"`
	CoverImageMedium string `json:"cover_image_medium,omitempty"`
}

// AnnictWorkImages represents the images object in Annict API response
type AnnictWorkImages struct {
	RecommendedURL string                    `json:"recommended_url,omitempty"`
	Facebook       *AnnictWorkImagesFacebook `json:"facebook,omitempty"`
	Twitter        *AnnictWorkImagesTwitter  `json:"twitter,omitempty"`
}

// AnnictWorkImagesFacebook represents Facebook image data
type AnnictWorkImagesFacebook struct {
	OgImageURL string `json:"og_image_url,omitempty"`
}

// AnnictWorkImagesTwitter represents Twitter image data
type AnnictWorkImagesTwitter struct {
	MiniAvatarURL     string `json:"mini_avatar_url,omitempty"`
	NormalAvatarURL   string `json:"normal_avatar_url,omitempty"`
	BiggerAvatarURL   string `json:"bigger_avatar_url,omitempty"`
	OriginalAvatarURL string `json:"original_avatar_url,omitempty"`
	ImageURL          string `json:"image_url,omitempty"`
}

// AnnictWorkResponse represents the REST API response structure
type AnnictWorkResponse struct {
	Works      []AnnictWork `json:"works"`
	TotalCount int          `json:"total_count"`
	NextPage   *int         `json:"next_page"`
	PrevPage   *int         `json:"prev_page"`
}
