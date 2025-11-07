package models

// JikanEpisodeResponse represents the response from Jikan API episodes endpoint
type JikanEpisodeResponse struct {
	Data       []JikanEpisode  `json:"data"`
	Pagination JikanPagination `json:"pagination"`
}

// JikanEpisode represents a single episode from Jikan API
type JikanEpisode struct {
	MalID         int     `json:"mal_id"`
	URL           string  `json:"url"` // Episode URL (used to extract episode number)
	Title         string  `json:"title"`
	TitleJapanese *string `json:"title_japanese"` // Japanese title
	TitleRomanji  *string `json:"title_romanji"`  // Romanji title
	Aired         *string `json:"aired"`
	Score         float64 `json:"score"`     // Episode rating
	Filler        bool    `json:"filler"`    // Is filler episode
	Recap         bool    `json:"recap"`     // Is recap episode
	ForumURL      string  `json:"forum_url"` // Forum URL
	Duration      int     `json:"-"`         // Duration in minutes (set from anime detail endpoint, not in JSON)
}

// JikanPagination represents pagination info from Jikan API
type JikanPagination struct {
	LastVisiblePage int  `json:"last_visible_page"`
	HasNextPage     bool `json:"has_next_page"`
}

// JikanAnimeResponse represents the response from Jikan API anime detail endpoint
type JikanAnimeResponse struct {
	Data JikanAnimeData `json:"data"`
}

// JikanAnimeData represents anime details from Jikan API
type JikanAnimeData struct {
	MalID    int          `json:"mal_id"`
	Title    string       `json:"title"`
	Duration string       `json:"duration"` // e.g., "24 min per ep"
	Episodes int          `json:"episodes"`
	Genres   []JikanGenre `json:"genres"`
}

// JikanGenre represents a genre from Jikan API
type JikanGenre struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}
