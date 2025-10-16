package models

// JikanEpisodeResponse represents the response from Jikan API episodes endpoint
type JikanEpisodeResponse struct {
	Data       []JikanEpisode  `json:"data"`
	Pagination JikanPagination `json:"pagination"`
}

// JikanEpisode represents a single episode from Jikan API
type JikanEpisode struct {
	MalID    int     `json:"mal_id"`
	Title    string  `json:"title"`
	Episode  string  `json:"episode"`
	Duration int     `json:"duration"` // Duration in minutes (not provided by episodes endpoint)
	Aired    *string `json:"aired"`
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
	MalID    int    `json:"mal_id"`
	Title    string `json:"title"`
	Duration string `json:"duration"` // e.g., "24 min per ep"
	Episodes int    `json:"episodes"`
}
