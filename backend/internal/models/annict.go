package models

// AnnictWork represents a single anime work from the Annict REST API response.
type AnnictWork struct {
	ID                int    `json:"id"`
	Title             string `json:"title"`
	TitleKana         string `json:"title_kana"`
	TitleEn           string `json:"title_en"`
	Media             string `json:"media"`
	MediaText         string `json:"media_text"`
	SeasonName        string `json:"season_name"`
	SeasonYear        int    `json:"season_year"`
	ReleasedOn        string `json:"released_on"`
	ReleasedOnAbout   string `json:"released_on_about"`
	OfficialSiteURL   string `json:"official_site_url"`
	WikipediaURL      string `json:"wikipedia_url"`
	TwitterUsername   string `json:"twitter_username"`
	TwitterHashtag    string `json:"twitter_hashtag"`
	MalAnimeID        string `json:"mal_anime_id"`
	ImagesRecommended string `json:"images_recommended"`
	EpisodesCount     int    `json:"episodes_count"`
	WatchersCount     int    `json:"watchers_count"`
	ReviewsCount      int    `json:"reviews_count"`
	NoEpisodes        bool   `json:"no_episodes"`
}

// AnnictWorkResponse represents the REST API response structure
type AnnictWorkResponse struct {
	Works      []AnnictWork `json:"works"`
	TotalCount int          `json:"total_count"`
	NextPage   *int         `json:"next_page"`
	PrevPage   *int         `json:"prev_page"`
}
