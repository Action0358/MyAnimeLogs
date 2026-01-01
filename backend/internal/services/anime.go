package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
)

var (
	ErrAnimeNotFound   = errors.New("anime not found")
	ErrEpisodeNotFound = errors.New("episode not found")
)

// GetCurrentSeasonAnimes returns animes for the current season from the database
func (s *AnimeService) GetCurrentSeasonAnimes(limit int) ([]models.Anime, error) {
	if limit <= 0 {
		limit = 12
	}

	// Get current season string (e.g., "2026-winter")
	currentSeason := getSeasonStringFromTime(time.Now())

	return s.getAnimesByQuery(
		`SELECT id, title, image_url, season, status, mal_anime_id, annict_id, created_at, updated_at, watchers_count
		FROM animes
		WHERE season = $1
		ORDER BY watchers_count DESC NULLS LAST, title
		LIMIT $2`,
		[]interface{}{currentSeason, limit},
		"current season animes",
	)
}

// GetPopularAnimes returns popular animes from the database (sorted by watchers_count)
func (s *AnimeService) GetPopularAnimes(limit int) ([]models.Anime, error) {
	if limit <= 0 {
		limit = 12
	}

	return s.getAnimesByQuery(
		`SELECT id, title, image_url, season, status, mal_anime_id, annict_id, created_at, updated_at, watchers_count
		FROM animes
		WHERE watchers_count > 0
		ORDER BY watchers_count DESC, title
		LIMIT $1`,
		[]interface{}{limit},
		"popular animes",
	)
}

// getAnimesByQuery is a helper function to execute anime queries with proper error handling
func (s *AnimeService) getAnimesByQuery(query string, args []interface{}, queryName string) ([]models.Anime, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", queryName, err)
	}
	defer rows.Close()

	animes := []models.Anime{}
	for rows.Next() {
		var anime models.Anime
		err := rows.Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount)
		if err != nil {
			s.logger.Warn("Failed to scan anime row", "query", queryName, "error", err)
			continue
		}
		animes = append(animes, anime)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating anime rows", "query", queryName, "error", err)
		return nil, fmt.Errorf("error iterating %s: %w", queryName, err)
	}

	// Enrich with cover images from AniList
	s.enrichAnimesWithCoverImages(animes)

	return animes, nil
}

// getSeasonStringFromTime returns the season string in format "YYYY-season" for a given time
// This function accepts time.Time as a parameter to improve testability
func getSeasonStringFromTime(t time.Time) string {
	year := t.Year()
	month := t.Month()

	var season string
	switch {
	case month >= 1 && month <= 3:
		season = "winter"
	case month >= 4 && month <= 6:
		season = "spring"
	case month >= 7 && month <= 9:
		season = "summer"
	default:
		season = "autumn"
	}

	return fmt.Sprintf("%d-%s", year, season)
}

type AnimeService struct {
	db            *database.DB
	logger        *common.Logger
	anilistClient *AniListClient
	jikanClient   *JikanClient
	annictClient  *AnnictClient
}

func NewAnimeService(db *database.DB, logger *common.Logger, annictClient *AnnictClient) *AnimeService {
	return &AnimeService{
		db:            db,
		logger:        logger,
		anilistClient: NewAniListClient(),
		jikanClient:   NewJikanClient(logger),
		annictClient:  annictClient,
	}
}

// GetEpisodeSyncStatus returns the sync status for a specific anime's episodes
func (s *AnimeService) GetEpisodeSyncStatus(animeID int) (*models.EpisodeSyncStatus, error) {
	// Get anime info and episode statistics in a single query
	var animeTitle string
	var annictID *int
	var totalEpisodes, defaultTitles, customTitles int

	err := s.db.QueryRow(`
		SELECT 
			a.title,
			a.annict_id,
			COUNT(e.id) as total,
			COUNT(CASE WHEN e.title LIKE '第%話' THEN 1 END) as default_titles,
			COUNT(CASE WHEN e.title NOT LIKE '第%話' THEN 1 END) as custom_titles
		FROM animes a
		LEFT JOIN episodes e ON a.id = e.anime_id
		WHERE a.id = $1
		GROUP BY a.id, a.title, a.annict_id
	`, animeID).Scan(&animeTitle, &annictID, &totalEpisodes, &defaultTitles, &customTitles)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnimeNotFound
		}
		return nil, fmt.Errorf("failed to get anime info and episode statistics: %w", err)
	}

	// Get sample episodes
	rows, err := s.db.Query(`
		SELECT episode_number, title
		FROM episodes
		WHERE anime_id = $1
		ORDER BY episode_number
		LIMIT 5
	`, animeID)

	if err != nil {
		return nil, fmt.Errorf("failed to get sample episodes: %w", err)
	}
	defer rows.Close()

	var sampleEpisodes []string
	for rows.Next() {
		var episodeNumber int
		var title string
		if err := rows.Scan(&episodeNumber, &title); err != nil {
			s.logger.Warn("Failed to scan sample episode", "anime_id", animeID, "error", err)
			continue
		}
		sampleEpisodes = append(sampleEpisodes, fmt.Sprintf("Ep%d: %s", episodeNumber, title))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sample episodes: %w", err)
	}

	syncPercentage := 0.0
	if totalEpisodes > 0 {
		syncPercentage = float64(customTitles) / float64(totalEpisodes) * 100
	}

	return &models.EpisodeSyncStatus{
		AnimeID:        animeID,
		AnimeTitle:     animeTitle,
		AnnictID:       annictID,
		TotalEpisodes:  totalEpisodes,
		DefaultTitles:  defaultTitles,
		CustomTitles:   customTitles,
		SyncPercentage: syncPercentage,
		NeedSync:       defaultTitles > 0,
		SampleEpisodes: sampleEpisodes,
	}, nil
}

func (s *AnimeService) SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if params.Keyword != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(`
			(to_tsvector('simple', COALESCE(title, '') || ' ' || COALESCE(description, '')) @@ plainto_tsquery('simple', $%d)
			 OR title ILIKE $%d OR description ILIKE $%d)
		`, argIndex, argIndex+1, argIndex+1))
		args = append(args, params.Keyword, "%"+params.Keyword+"%")
		argIndex += 2
	}

	if params.Season != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("season = $%d", argIndex))
		args = append(args, params.Season)
		argIndex++
	}

	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	if params.Genre != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("id IN (SELECT anime_id FROM anime_genres WHERE genre = $%d)", argIndex))
		args = append(args, params.Genre)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY title"
	switch params.Sort {
	case "rating":
		orderBy = "ORDER BY (SELECT AVG(rating) FROM user_ratings WHERE anime_id = animes.id) DESC NULLS LAST"
	case "year":
		orderBy = "ORDER BY season DESC"
	}

	// Count total results
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM animes %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count animes: %w", err)
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PerPage
	totalPages := (total + params.PerPage - 1) / params.PerPage

	// Get animes
	query := fmt.Sprintf(`
		SELECT id, title, image_url, season, status, mal_anime_id, annict_id, created_at, updated_at, watchers_count
		FROM animes %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search animes: %w", err)
	}
	defer rows.Close()

	animes := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		err := rows.Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount)
		if err != nil {
			continue
		}

		animeWithData := models.AnimeWithUserData{Anime: anime}

		// Add user data if authenticated
		if userID != nil {
			// Get user favorite status
			var favorite models.UserFavorite
			err := s.db.QueryRow(`
				SELECT id, user_id, anime_id, status, created_at, updated_at
				FROM user_favorites WHERE user_id = $1 AND anime_id = $2
			`, *userID, anime.ID).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt)
			if err == nil {
				animeWithData.UserFavorite = &favorite
			}

			// Calculate watch progress
			animeWithData.WatchProgress = s.calculateWatchProgress(*userID, anime.ID)
		}

		animes = append(animes, animeWithData)
	}

	// AniListから画像を並列で取得
	s.enrichAnimesWithUserDataCoverImages(animes)

	// 検索ログを記録
	s.logBasicSearch(params, userID, total)

	return &models.PaginatedResponse{
		Data: animes,
		Meta: models.PaginationMeta{
			TotalPages:  totalPages,
			CurrentPage: params.Page,
			Total:       total,
		},
	}, nil
}

// AdvancedSearchAnimes - 高度検索機能
func (s *AnimeService) AdvancedSearchAnimes(params models.AdvancedSearchParams, userID *int) (*models.PaginatedResponse, error) {
	// Set defaults
	s.setSearchDefaults(&params)

	// Build query components
	baseQuery := s.buildBaseQuery(params, userID)
	whereClause, args, argIndex := s.buildWhereClause(params, userID)
	orderBy := s.buildOrderByClause(params)

	// Count total results (distinct to handle JOINs)
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT a.id) %s %s", baseQuery, whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count animes: %w", err)
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PerPage
	totalPages := (total + params.PerPage - 1) / params.PerPage

	// Get animes
	query := fmt.Sprintf(`
		SELECT a.id, a.title, a.image_url, a.season, a.status, a.mal_anime_id, a.annict_id, a.created_at, a.updated_at, a.watchers_count
		%s %s
		GROUP BY a.id, a.title, a.image_url, a.season, a.status, a.mal_anime_id, a.annict_id, a.created_at, a.updated_at, a.watchers_count
		%s
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search animes: %w", err)
	}
	defer rows.Close()

	animes := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		err := rows.Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount)
		if err != nil {
			continue
		}

		animeWithData := models.AnimeWithUserData{Anime: anime}

		// Add user data if authenticated
		if userID != nil {
			// Get user favorite status
			var favorite models.UserFavorite
			err := s.db.QueryRow(`
				SELECT id, user_id, anime_id, status, created_at, updated_at
				FROM user_favorites WHERE user_id = $1 AND anime_id = $2
			`, *userID, anime.ID).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt)
			if err == nil {
				animeWithData.UserFavorite = &favorite
			}

			// Calculate watch progress
			animeWithData.WatchProgress = s.calculateWatchProgress(*userID, anime.ID)
		}

		animes = append(animes, animeWithData)
	}

	// AniListから画像を並列で取得
	s.enrichAnimesWithUserDataCoverImages(animes)

	// 検索ログを記録
	s.logAdvancedSearch(params, userID, total)

	return &models.PaginatedResponse{
		Data: animes,
		Meta: models.PaginationMeta{
			TotalPages:  totalPages,
			CurrentPage: params.Page,
			Total:       total,
		},
	}, nil
}

func (s *AnimeService) GetAnimeByID(animeID int, userID *int) (*models.AnimeWithUserData, error) {
	var anime models.Anime
	err := s.db.QueryRow(`
		SELECT id, title, image_url, season, status, mal_anime_id, annict_id, created_at, updated_at, watchers_count
		FROM animes WHERE id = $1
	`, animeID).Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnimeNotFound
		}
		return nil, fmt.Errorf("failed to get anime: %w", err)
	}

	// AniListから画像を取得
	s.enrichAnimeWithCoverImage(&anime)

	animeWithData := &models.AnimeWithUserData{Anime: anime}

	// Get episodes
	episodes, err := s.GetAnimeEpisodes(animeID, userID)
	if err != nil {
		return nil, err
	}
	animeWithData.Episodes = episodes

	// Get rating statistics (average and count)
	// Using database aggregation for better performance and scalability
	var avgRating sql.NullFloat64
	var ratingCount int
	err = s.db.QueryRow(`
		SELECT AVG(rating), COUNT(*)
		FROM user_ratings
		WHERE anime_id = $1
	`, animeID).Scan(&avgRating, &ratingCount)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to get rating statistics",
			"anime_id", animeID,
			"error", err)
	} else if err == nil && avgRating.Valid {
		animeWithData.AverageRating = &avgRating.Float64
		animeWithData.RatingCount = ratingCount
	}

	// Get rating distribution
	// Using database GROUP BY for efficient aggregation
	rows, err := s.db.Query(`
		SELECT FLOOR(rating) as rating_floor, COUNT(*) as count
		FROM user_ratings
		WHERE anime_id = $1
		GROUP BY FLOOR(rating)
		ORDER BY rating_floor
	`, animeID)
	if err != nil {
		s.logger.Warn("Failed to get rating distribution",
			"anime_id", animeID,
			"error", err)
	} else {
		defer rows.Close()
		distribution := make(map[int]int)
		for rows.Next() {
			var rating int
			var count int
			if err := rows.Scan(&rating, &count); err != nil {
				s.logger.Warn("Failed to scan rating distribution row",
					"anime_id", animeID,
					"error", err)
				continue
			}
			distribution[rating] = count
		}

		// Check for errors during iteration
		if err := rows.Err(); err != nil {
			s.logger.Warn("Error during rating distribution iteration",
				"anime_id", animeID,
				"error", err)
		}

		if len(distribution) > 0 {
			animeWithData.RatingDistribution = distribution
		}
	}

	// Add user data if authenticated
	if userID != nil {
		// Get user favorite status
		var favorite models.UserFavorite
		err := s.db.QueryRow(`
			SELECT id, user_id, anime_id, status, created_at, updated_at
			FROM user_favorites WHERE user_id = $1 AND anime_id = $2
		`, *userID, animeID).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt)
		if err != nil && err != sql.ErrNoRows {
			// Log error but don't fail the entire request
			s.logger.Warn("Failed to get user favorite status",
				"user_id", *userID,
				"anime_id", animeID,
				"error", err)
		} else if err == nil {
			animeWithData.UserFavorite = &favorite
		}

		// Get user rating
		var userRating models.UserRating
		err = s.db.QueryRow(`
			SELECT id, user_id, anime_id, rating, comment, created_at, updated_at
			FROM user_ratings WHERE user_id = $1 AND anime_id = $2
		`, *userID, animeID).Scan(&userRating.ID, &userRating.UserID, &userRating.AnimeID, &userRating.Rating, &userRating.Comment, &userRating.CreatedAt, &userRating.UpdatedAt)
		if err != nil && err != sql.ErrNoRows {
			// Log error but don't fail the entire request
			s.logger.Warn("Failed to get user rating",
				"user_id", *userID,
				"anime_id", animeID,
				"error", err)
		} else if err == nil {
			animeWithData.UserRating = &userRating
		}

		// Calculate watch progress
		animeWithData.WatchProgress = s.calculateWatchProgress(*userID, animeID)
	}

	return animeWithData, nil
}

func (s *AnimeService) GetAnimeEpisodes(animeID int, userID *int) ([]models.Episode, error) {
	// Check if anime exists and get MAL ID
	var exists bool
	var malAnimeID *string
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1), mal_anime_id FROM animes WHERE id = $1", animeID).Scan(&exists, &malAnimeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	rows, err := s.db.Query(`
		SELECT id, anime_id, title, episode_number, duration_minutes, air_date, created_at, updated_at
		FROM episodes WHERE anime_id = $1
		ORDER BY episode_number
	`, animeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", err)
	}
	defer rows.Close()

	episodes := []models.Episode{}
	hasMissingDurations := false
	for rows.Next() {
		var episode models.Episode
		err := rows.Scan(&episode.ID, &episode.AnimeID, &episode.Title, &episode.EpisodeNumber, &episode.DurationMinutes, &episode.AirDate, &episode.CreatedAt, &episode.UpdatedAt)
		if err != nil {
			continue
		}

		// Check if duration or air_date is missing
		if episode.DurationMinutes == nil || episode.AirDate == nil {
			hasMissingDurations = true
		}

		// Add watch status if user is authenticated
		if userID != nil {
			var viewed bool
			err := s.db.QueryRow(`
				SELECT viewed FROM watch_statuses 
				WHERE user_id = $1 AND episode_id = $2
			`, *userID, episode.ID).Scan(&viewed)
			if err == nil {
				episode.Viewed = viewed
			}
		}

		episodes = append(episodes, episode)
	}

	// If there are missing durations and we have a MAL ID, try to fetch them in background
	if hasMissingDurations && malAnimeID != nil && *malAnimeID != "" {
		go func() {
			err := s.enrichEpisodesWithDuration(animeID, *malAnimeID)
			if err != nil {
				s.logger.Warn("Failed to enrich episodes with duration",
					"anime_id", animeID,
					"mal_anime_id", *malAnimeID,
					"error", err)
			}
		}()
	}

	return episodes, nil
}

func (s *AnimeService) GetUserFavorites(userID int, status string) ([]models.AnimeWithUserData, error) {
	whereClause := "WHERE uf.user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if status != "" {
		whereClause += fmt.Sprintf(" AND uf.status = $%d", argIndex)
		args = append(args, status)
	}

	query := fmt.Sprintf(`
		SELECT a.id, a.title, a.image_url, a.season, a.status, a.mal_anime_id, a.created_at, a.updated_at, a.watchers_count,
		       uf.id, uf.user_id, uf.anime_id, uf.status, uf.created_at, uf.updated_at
		FROM animes a
		JOIN user_favorites uf ON a.id = uf.anime_id
		%s
		ORDER BY uf.updated_at DESC
	`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user favorites: %w", err)
	}
	defer rows.Close()

	favorites := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		var favorite models.UserFavorite
		err := rows.Scan(
			&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount,
			&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt,
		)
		if err != nil {
			continue
		}

		animeWithData := models.AnimeWithUserData{
			Anime:        anime,
			UserFavorite: &favorite,
		}

		// Calculate watch progress
		animeWithData.WatchProgress = s.calculateWatchProgress(userID, anime.ID)

		favorites = append(favorites, animeWithData)
	}

	// AniListから画像を並列で取得
	s.enrichAnimesWithUserDataCoverImages(favorites)

	return favorites, nil
}

func (s *AnimeService) UpdateFavorite(userID, animeID int, status string) (*models.UserFavorite, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	// Upsert favorite
	var favorite models.UserFavorite
	err = s.db.QueryRow(`
		INSERT INTO user_favorites (user_id, anime_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, anime_id)
		DO UPDATE SET status = $3, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, anime_id, status, created_at, updated_at
	`, userID, animeID, status).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update favorite: %w", err)
	}

	return &favorite, nil
}

func (s *AnimeService) DeleteFavorite(userID, animeID int) error {
	_, err := s.db.Exec(`
		DELETE FROM user_favorites WHERE user_id = $1 AND anime_id = $2
	`, userID, animeID)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}

	return nil
}

func (s *AnimeService) UpdateWatchStatus(userID, episodeID int, viewed bool) (*models.WatchStatus, error) {
	// Check if episode exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM episodes WHERE id = $1)", episodeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check episode existence: %w", err)
	}
	if !exists {
		return nil, ErrEpisodeNotFound
	}

	// Upsert watch status
	var watchStatus models.WatchStatus
	err = s.db.QueryRow(`
		INSERT INTO watch_statuses (user_id, episode_id, viewed, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, episode_id)
		DO UPDATE SET viewed = $3, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, episode_id, viewed, updated_at
	`, userID, episodeID, viewed).Scan(&watchStatus.ID, &watchStatus.UserID, &watchStatus.EpisodeID, &watchStatus.Viewed, &watchStatus.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update watch status: %w", err)
	}

	// Add to watch history if marked as viewed
	if viewed {
		_, err = s.db.Exec(`
			INSERT INTO watch_history (user_id, episode_id, watched_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
		`, userID, episodeID)
		if err != nil {
			// Log error but don't fail the request
			s.logger.Warn("Failed to add to watch history", "error", err, "user_id", userID, "episode_id", episodeID)
		}
	}

	return &watchStatus, nil
}

func (s *AnimeService) UpdateRating(userID, animeID int, rating float64, comment *string) (*models.UserRating, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	// Upsert rating
	var userRating models.UserRating
	err = s.db.QueryRow(`
		INSERT INTO user_ratings (user_id, anime_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, anime_id)
		DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, anime_id, rating, comment, created_at, updated_at
	`, userID, animeID, rating, comment).Scan(&userRating.ID, &userRating.UserID, &userRating.AnimeID, &userRating.Rating, &userRating.Comment, &userRating.CreatedAt, &userRating.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update rating: %w", err)
	}

	return &userRating, nil
}

func (s *AnimeService) DeleteRating(userID, animeID int) error {
	_, err := s.db.Exec(`
		DELETE FROM user_ratings WHERE user_id = $1 AND anime_id = $2
	`, userID, animeID)
	if err != nil {
		return fmt.Errorf("failed to delete rating: %w", err)
	}

	return nil
}

func (s *AnimeService) GetRecentlyWatched(userID, limit int) ([]models.WatchHistory, error) {
	rows, err := s.db.Query(`
		SELECT wh.id, wh.user_id, wh.episode_id, wh.watched_at,
		       a.title as anime_title, e.title as episode_title
		FROM watch_history wh
		JOIN episodes e ON wh.episode_id = e.id
		JOIN animes a ON e.anime_id = a.id
		WHERE wh.user_id = $1
		ORDER BY wh.watched_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently watched: %w", err)
	}
	defer rows.Close()

	history := []models.WatchHistory{}
	for rows.Next() {
		var h models.WatchHistory
		err := rows.Scan(&h.ID, &h.UserID, &h.EpisodeID, &h.WatchedAt, &h.AnimeTitle, &h.EpisodeTitle)
		if err != nil {
			continue
		}
		history = append(history, h)
	}

	return history, nil
}

func (s *AnimeService) calculateWatchProgress(userID, animeID int) float64 {
	var totalEpisodes, watchedEpisodes int

	// Get total episodes
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM episodes WHERE anime_id = $1
	`, animeID).Scan(&totalEpisodes)
	if err != nil || totalEpisodes == 0 {
		return 0.0
	}

	// Get watched episodes
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM watch_statuses ws
		JOIN episodes e ON ws.episode_id = e.id
		WHERE ws.user_id = $1 AND e.anime_id = $2 AND ws.viewed = true
	`, userID, animeID).Scan(&watchedEpisodes)
	if err != nil {
		return 0.0
	}

	return float64(watchedEpisodes) / float64(totalEpisodes) * 100.0
}

// GetFilterOptions - フィルター情報を取得
func (s *AnimeService) GetFilterOptions() (*models.FilterOptions, error) {
	filterOptions := &models.FilterOptions{
		Genres:  []models.GenreInfo{},
		Years:   []int{},
		Seasons: []string{},
		Ratings: models.RatingStats{
			Min:     1.0,
			Max:     5.0,
			Average: 3.0,
			Count:   0,
		},
	}

	// ジャンル一覧を取得（エラーがあっても他の情報は取得を続行）
	genreRows, err := s.db.Query(`
		SELECT genre, COUNT(*) as count
		FROM anime_genres
		WHERE genre IS NOT NULL AND genre != ''
		GROUP BY genre
		ORDER BY count DESC, genre ASC
	`)
	if err != nil {
		s.logger.Error("Failed to get genres", "error", err)
	} else {
		defer genreRows.Close()
		genres := []models.GenreInfo{}
		for genreRows.Next() {
			var genre models.GenreInfo
			err := genreRows.Scan(&genre.Name, &genre.Count)
			if err != nil {
				s.logger.Error("Failed to scan genre", "error", err)
				continue
			}
			if strings.TrimSpace(genre.Name) != "" {
				genres = append(genres, genre)
			}
		}
		filterOptions.Genres = genres
	}

	// 放送年一覧を取得（season から年を抽出）
	yearRows, err := s.db.Query(`
		SELECT DISTINCT CAST(SUBSTRING(season FROM 1 FOR 4) AS INTEGER) as year
		FROM animes
		WHERE season IS NOT NULL 
		  AND season ~ '^\d{4}-'
		  AND CAST(SUBSTRING(season FROM 1 FOR 4) AS INTEGER) > 1900
		  AND CAST(SUBSTRING(season FROM 1 FOR 4) AS INTEGER) <= EXTRACT(YEAR FROM CURRENT_DATE) + 1
		ORDER BY year DESC
	`)
	if err != nil {
		s.logger.Error("Failed to get years", "error", err)
	} else {
		defer yearRows.Close()
		years := []int{}
		for yearRows.Next() {
			var year int
			err := yearRows.Scan(&year)
			if err != nil {
				s.logger.Error("Failed to scan year", "error", err)
				continue
			}
			years = append(years, year)
		}
		filterOptions.Years = years
	}

	// シーズン一覧を取得
	seasonRows, err := s.db.Query(`
		SELECT season
		FROM animes
		WHERE season IS NOT NULL AND season != ''
		GROUP BY season
		ORDER BY 
			CASE season
				WHEN $1 THEN 1
				WHEN $2 THEN 2
				WHEN $3 THEN 3
				WHEN $4 THEN 4
				ELSE 5
			END
	`, constants.SeasonSpring, constants.SeasonSummer, constants.SeasonAutumn, constants.SeasonWinter)
	if err != nil {
		s.logger.Error("Failed to get seasons", "error", err)
	} else {
		defer seasonRows.Close()
		seasons := []string{}
		for seasonRows.Next() {
			var season string
			err := seasonRows.Scan(&season)
			if err != nil {
				s.logger.Error("Failed to scan season", "error", err)
				continue
			}
			if strings.TrimSpace(season) != "" {
				seasons = append(seasons, season)
			}
		}
		filterOptions.Seasons = seasons
	}

	// 評価統計を取得
	var ratingStats models.RatingStats
	err = s.db.QueryRow(`
		SELECT 
			COALESCE(MIN(rating), 1.0) as min_rating,
			COALESCE(MAX(rating), 5.0) as max_rating,
			COALESCE(AVG(rating), 3.0) as avg_rating,
			COUNT(*) as count
		FROM user_ratings
		WHERE rating IS NOT NULL AND rating >= 1.0 AND rating <= 5.0
	`).Scan(&ratingStats.Min, &ratingStats.Max, &ratingStats.Average, &ratingStats.Count)
	if err != nil {
		s.logger.Error("Failed to get rating stats", "error", err)
		// デフォルト値はすでに設定済み
	} else {
		filterOptions.Ratings = ratingStats
	}

	return filterOptions, nil
}

// logSearch - 検索ログを記録（非同期）
func (s *AnimeService) logSearch(query string, filters map[string]interface{}, userID *int, resultCount int) {
	// 安全にJSONに変換
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		// JSON変換エラーは無視（メイン機能に影響しない）
		s.logger.Warn("Failed to marshal search filters", "error", err, "filters", filters)
		return
	}

	_, err = s.db.Exec(`
		INSERT INTO search_logs (user_id, query, filters, result_count, searched_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`, userID, query, string(filtersJSON), resultCount)
	if err != nil {
		// ログ記録エラーは無視（メイン機能に影響しない）
		s.logger.Warn("Failed to log search", "error", err, "query", query, "result_count", resultCount)
	}
}

// logAdvancedSearch - 高度検索のログを記録（非同期）
func (s *AnimeService) logAdvancedSearch(params models.AdvancedSearchParams, userID *int, resultCount int) {
	filters := map[string]interface{}{
		"genres":      params.Genres,
		"year":        params.Year,
		"season":      params.Season,
		"min_rating":  params.MinRating,
		"max_rating":  params.MaxRating,
		"status":      params.Status,
		"user_status": params.UserStatus,
		"sort":        params.Sort,
		"order":       params.Order,
	}
	go s.logSearch(params.Query, filters, userID, resultCount)
}

// logBasicSearch - 基本検索のログを記録（非同期）
func (s *AnimeService) logBasicSearch(params models.AnimeSearchParams, userID *int, resultCount int) {
	filters := map[string]interface{}{
		"genre":  params.Genre,
		"season": params.Season,
		"status": params.Status,
		"sort":   params.Sort,
	}
	go s.logSearch(params.Keyword, filters, userID, resultCount)
}

// GetPopularSearches - 人気検索キーワードを取得
func (s *AnimeService) GetPopularSearches(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}

	// まず検索ログが存在するかチェック
	var logCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM search_logs").Scan(&logCount)
	if err != nil {
		s.logger.Error("Failed to count search logs", "error", err)
		return []string{}, nil
	}

	// ログが少ない場合は期間を拡張
	intervalDays := 30
	if logCount < 10 {
		intervalDays = 365 // 1年に拡張
	}

	rows, err := s.db.Query(`
		SELECT query, COUNT(*) as search_count
		FROM search_logs
		WHERE query IS NOT NULL 
		  AND query != '' 
		  AND LENGTH(TRIM(query)) > 0
		  AND searched_at > CURRENT_TIMESTAMP - ($1 || ' days')::INTERVAL
		GROUP BY query
		ORDER BY search_count DESC, query ASC
		LIMIT $2
	`, intervalDays, limit)
	if err != nil {
		s.logger.Error("Failed to get popular searches", "error", err, "limit", limit)
		return []string{}, nil // エラーでも空配列を返す
	}
	defer rows.Close()

	searches := []string{}
	for rows.Next() {
		var query string
		var count int
		err := rows.Scan(&query, &count)
		if err != nil {
			s.logger.Error("Failed to scan popular search", "error", err)
			continue
		}
		// 空文字列やnullをスキップ
		if strings.TrimSpace(query) != "" {
			searches = append(searches, query)
		}
	}

	// 検索ログが少ない場合はデフォルトの人気検索を返す
	if len(searches) == 0 {
		// グローバルスライスを変更しないようにコピーを作成
		searches = make([]string, len(constants.DefaultPopularSearches))
		copy(searches, constants.DefaultPopularSearches)
		if limit < len(searches) {
			searches = searches[:limit]
		}
	}

	return searches, nil
}

// GetSearchSuggestions - 検索候補を取得
func (s *AnimeService) GetSearchSuggestions(query string, limit int) ([]string, error) {
	// 空のクエリに対する適切な処理
	if strings.TrimSpace(query) == "" {
		return []string{}, nil
	}

	// limitのデフォルト値設定
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Query(`
		SELECT title
		FROM animes
		WHERE title ILIKE $1
		  AND TRIM(title) <> ''
		GROUP BY title
		ORDER BY MAX(watchers_count) DESC NULLS LAST, title
		LIMIT $2
	`, query+"%", limit)

	if err != nil {
		s.logger.Error("Failed to get search suggestions", "error", err, "query", query, "limit", limit)
		return []string{}, nil // エラーでも空配列を返す
	}
	defer rows.Close()

	suggestions := []string{}
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			s.logger.Warn("Failed to scan suggestion title", "error", err)
			continue
		}
		suggestions = append(suggestions, title)
	}

	return suggestions, nil
}

// enrichAnimeWithCoverImage - AniListから画像を取得してアニメ情報を拡張
func (s *AnimeService) enrichAnimeWithCoverImage(anime *models.Anime) {
	// Use background context for image enrichment
	ctx := context.Background()

	var coverImage *CoverImage
	var err error

	// MAL IDがある場合は、まずMAL IDで検索
	if anime.MalAnimeID != nil && *anime.MalAnimeID != "" {
		coverImage, err = s.anilistClient.GetCoverImageByMALId(ctx, *anime.MalAnimeID)
		if err != nil {
			s.logger.Debug("Failed to get cover image by MAL ID, trying title fallback",
				"error", err,
				"mal_anime_id", *anime.MalAnimeID,
				"anime_title", anime.Title)
		}
	}

	// MAL IDがないか、MAL IDでの取得に失敗した場合は、タイトルで検索（フォールバック）
	if coverImage == nil && anime.Title != "" {
		coverImage, err = s.anilistClient.GetCoverImageByTitle(ctx, anime.Title)
		if err != nil {
			// フォールバックも失敗した場合は、Annictの画像を使用
			s.logger.Debug("Failed to get cover image by title",
				"error", err,
				"anime_title", anime.Title)
			return
		}
	}

	// 画像URLを設定
	if coverImage != nil {
		anime.CoverImageLarge = coverImage.Large
		anime.CoverImageMedium = coverImage.Medium
	}
}

// enrichAnimesWithCoverImages - 複数のアニメに画像を追加（並列処理）
func (s *AnimeService) enrichAnimesWithCoverImages(animes []models.Anime) {
	var wg sync.WaitGroup

	// 各アニメに対して並列で画像を取得
	for i := range animes {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			s.enrichAnimeWithCoverImage(&animes[index])
		}(i)
	}

	// すべてのゴルーチンが完了するまで待機
	wg.Wait()
}

// enrichAnimesWithUserDataCoverImages - AnimeWithUserDataのスライスに画像を追加（並列処理）
func (s *AnimeService) enrichAnimesWithUserDataCoverImages(animes []models.AnimeWithUserData) {
	var wg sync.WaitGroup

	// 各アニメに対して並列で画像を取得
	for i := range animes {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			s.enrichAnimeWithCoverImage(&animes[index].Anime)
		}(i)
	}

	// すべてのゴルーチンが完了するまで待機
	wg.Wait()
}

// AdvancedSearchAnimes のヘルパー関数

// setSearchDefaults - 検索パラメータのデフォルト値を設定
func (s *AnimeService) setSearchDefaults(params *models.AdvancedSearchParams) {
	const (
		SortByTitle = "title"
		OrderAsc    = "asc"
	)

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}
	if params.Sort == "" {
		params.Sort = SortByTitle
	}
	if params.Order == "" {
		params.Order = OrderAsc
	}
}

// buildBaseQuery - ベースクエリとJOINを構築
func (s *AnimeService) buildBaseQuery(params models.AdvancedSearchParams, userID *int) string {
	const SortByRating = "rating"

	baseQuery := "FROM animes a"

	// Track which JOINs are needed
	needGenreJoin := len(params.Genres) > 0
	needRatingJoin := params.MinRating > 0 || params.MaxRating > 0 || params.Sort == SortByRating
	needUserFavoriteJoin := userID != nil && params.UserStatus != ""

	// Add JOINs only when needed
	if needGenreJoin {
		baseQuery += "\n\t\tLEFT JOIN anime_genres ag ON a.id = ag.anime_id"
	}
	if needRatingJoin {
		baseQuery += `
		LEFT JOIN (
			SELECT anime_id, AVG(rating) as avg_rating
			FROM user_ratings
			GROUP BY anime_id
		) AS avg_ratings ON a.id = avg_ratings.anime_id`
	}
	if needUserFavoriteJoin {
		baseQuery += `
		INNER JOIN user_favorites uf ON a.id = uf.anime_id`
	}

	return baseQuery
}

// buildWhereClause - WHERE句と引数を構築
func (s *AnimeService) buildWhereClause(params models.AdvancedSearchParams, userID *int) (string, []interface{}, int) {
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	// 全文検索（日本語対応のためFTSとILIKEを併用）
	if params.Query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(`
			(to_tsvector('simple', COALESCE(a.title, '') || ' ' || COALESCE(a.description, '')) @@ plainto_tsquery('simple', $%d)
			 OR a.title ILIKE $%d OR a.description ILIKE $%d)
		`, argIndex, argIndex+1, argIndex+1))
		args = append(args, params.Query, "%"+params.Query+"%")
		argIndex += 2
	}

	// ジャンル検索（複数対応）
	if len(params.Genres) > 0 {
		genrePlaceholders := make([]string, len(params.Genres))
		for i, genre := range params.Genres {
			genrePlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, genre)
			argIndex++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("ag.genre IN (%s)", strings.Join(genrePlaceholders, ",")))
	}

	// 年検索（LIKE演算子を使用してインデックスを活用）
	if params.Year > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("a.season LIKE $%d", argIndex))
		args = append(args, fmt.Sprintf("%d-%%", params.Year))
		argIndex++
	}

	// シーズン検索
	if params.Season != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.season = $%d", argIndex))
		args = append(args, params.Season)
		argIndex++
	}

	// ステータス検索
	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	// 評価範囲検索
	if params.MinRating > 0 || params.MaxRating > 0 {
		if params.MinRating > 0 && params.MaxRating > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("avg_ratings.avg_rating BETWEEN $%d AND $%d", argIndex, argIndex+1))
			args = append(args, params.MinRating, params.MaxRating)
			argIndex += 2
		} else if params.MinRating > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("avg_ratings.avg_rating >= $%d", argIndex))
			args = append(args, params.MinRating)
			argIndex++
		} else if params.MaxRating > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("avg_ratings.avg_rating <= $%d", argIndex))
			args = append(args, params.MaxRating)
			argIndex++
		}
	}

	// ユーザーステータス検索（認証時のみ）
	// JOINを使用するため、WHERE句で条件を追加
	if userID != nil && params.UserStatus != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("uf.user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
		whereClauses = append(whereClauses, fmt.Sprintf("uf.status = $%d", argIndex))
		args = append(args, params.UserStatus)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	return whereClause, args, argIndex
}

// buildOrderByClause - ORDER BY句を構築
func (s *AnimeService) buildOrderByClause(params models.AdvancedSearchParams) string {
	const (
		SortByTitle      = "title"
		SortByRating     = "rating"
		SortByYear       = "year"
		SortByPopularity = "popularity"
		SortByUpdated    = "updated"
		OrderDesc        = "desc"
	)

	orderDirection := "ASC"
	if params.Order == OrderDesc {
		orderDirection = "DESC"
	}

	switch params.Sort {
	case SortByTitle:
		return fmt.Sprintf("ORDER BY a.title %s", orderDirection)
	case SortByRating:
		return fmt.Sprintf("ORDER BY avg_ratings.avg_rating %s NULLS LAST", orderDirection)
	case SortByYear:
		// season から年を抽出してソート
		// NOTE: SUBSTRING関数を使用しているため、インデックスが利用されません
		// 大規模データセットでパフォーマンスが問題になる場合は、
		// CREATE INDEX idx_animes_season_year ON animes (SUBSTRING(season FROM 1 FOR 4))
		// のような関数ベースのインデックスを作成することを検討してください
		return fmt.Sprintf("ORDER BY SUBSTRING(a.season FROM 1 FOR 4) %s, a.season %s", orderDirection, orderDirection)
	case SortByPopularity:
		return fmt.Sprintf("ORDER BY a.watchers_count %s", orderDirection)
	case SortByUpdated:
		return fmt.Sprintf("ORDER BY a.updated_at %s", orderDirection)
	default:
		return fmt.Sprintf("ORDER BY a.title %s", orderDirection)
	}
}

// enrichEpisodesWithDuration fetches episode durations from Jikan API and updates the database
func (s *AnimeService) enrichEpisodesWithDuration(animeID int, malAnimeID string) error {
	// Fetch episodes from Jikan API
	jikanEpisodes, err := s.jikanClient.GetAnimeEpisodes(malAnimeID)
	if err != nil {
		return fmt.Errorf("failed to fetch episodes from Jikan API: %w", err)
	}

	if len(jikanEpisodes) == 0 {
		s.logger.Debug("No episodes found in Jikan API", "mal_anime_id", malAnimeID)
		return nil
	}

	// Get existing episodes from database
	rows, err := s.db.Query(`
		SELECT id, episode_number, duration_minutes, air_date
		FROM episodes
		WHERE anime_id = $1
		ORDER BY episode_number
	`, animeID)
	if err != nil {
		return fmt.Errorf("failed to get episodes from database: %w", err)
	}
	defer rows.Close()

	// Create a map of episode number to episode ID
	episodeMap := make(map[int]struct {
		id              int
		durationMinutes *int
		airDate         *time.Time
	})
	for rows.Next() {
		var id, episodeNumber int
		var durationMinutes *int
		var airDate *time.Time
		if err := rows.Scan(&id, &episodeNumber, &durationMinutes, &airDate); err != nil {
			continue
		}
		episodeMap[episodeNumber] = struct {
			id              int
			durationMinutes *int
			airDate         *time.Time
		}{id, durationMinutes, airDate}
	}

	// Update episodes with duration from Jikan API
	updatedCount := 0
	for _, jikanEp := range jikanEpisodes {
		// Use mal_id as episode number (Jikan API's mal_id corresponds to episode number)
		episodeNumber := jikanEp.MalID

		if episodeNumber <= 0 {
			s.logger.Debug("Invalid episode number from Jikan API",
				"mal_id", jikanEp.MalID,
				"mal_anime_id", malAnimeID)
			continue
		}

		// Check if episode exists in database
		dbEpisode, exists := episodeMap[episodeNumber]
		if !exists {
			s.logger.Debug("Episode not found in database",
				"episode_number", episodeNumber,
				"mal_anime_id", malAnimeID)
			continue
		}

		// Check if we need to update this episode
		// Skip only if both duration and air_date already exist
		hasDuration := dbEpisode.durationMinutes != nil
		hasAirDate := dbEpisode.airDate != nil

		if hasDuration && hasAirDate {
			// Both fields are already set, skip this episode
			continue
		}

		// Parse air_date from Jikan API
		// Only update if not already set
		var airDate *time.Time
		if dbEpisode.airDate == nil && jikanEp.Aired != nil && *jikanEp.Aired != "" {
			// Try ISO 8601 format first (e.g., "2025-07-06T00:00:00+00:00")
			parsedDate, err := time.Parse(time.RFC3339, *jikanEp.Aired)
			if err != nil {
				// Try simple date format (e.g., "2025-07-06")
				parsedDate, err = time.Parse("2006-01-02", *jikanEp.Aired)
			}
			if err == nil {
				airDate = &parsedDate
			} else {
				s.logger.Debug("Failed to parse air_date from Jikan API",
					"episode_number", episodeNumber,
					"aired", *jikanEp.Aired,
					"error", err)
			}
		}

		// Validate duration (should be between 1 and 180 minutes)
		// Only update if not already set
		var duration *int
		if dbEpisode.durationMinutes == nil {
			if jikanEp.Duration >= 1 && jikanEp.Duration <= 180 {
				duration = &jikanEp.Duration
			} else if jikanEp.Duration != 0 {
				s.logger.Warn("Invalid duration value from Jikan API",
					"episode_number", episodeNumber,
					"duration", jikanEp.Duration,
					"mal_anime_id", malAnimeID)
			}
		}

		// Update episode duration and air_date
		err = s.updateEpisodeMetadata(dbEpisode.id, duration, airDate)
		if err != nil {
			s.logger.Warn("Failed to update episode metadata",
				"episode_id", dbEpisode.id,
				"episode_number", episodeNumber,
				"error", err)
			continue
		}

		updatedCount++
	}

	s.logger.Info("Successfully enriched episodes with metadata",
		"anime_id", animeID,
		"mal_anime_id", malAnimeID,
		"updated_count", updatedCount,
		"total_jikan_episodes", len(jikanEpisodes))

	return nil
}

// updateEpisodeMetadata updates the duration and air_date of a single episode
func (s *AnimeService) updateEpisodeMetadata(episodeID int, duration *int, airDate *time.Time) error {
	// Build dynamic UPDATE query based on which fields need updating
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if duration != nil {
		updates = append(updates, fmt.Sprintf("duration_minutes = $%d", argIndex))
		args = append(args, *duration)
		argIndex++
	}

	if airDate != nil {
		updates = append(updates, fmt.Sprintf("air_date = $%d", argIndex))
		args = append(args, *airDate)
		argIndex++
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, episodeID)

	query := fmt.Sprintf(`
		UPDATE episodes
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argIndex)

	_, err := s.db.Exec(query, args...)
	return err
}

// SyncEpisodeDurations synchronizes episode durations for a specific anime or all animes
func (s *AnimeService) SyncEpisodeDurations(animeID *int) (int, int, int, int, []string, error) {
	var animes []struct {
		id         int
		malAnimeID string
	}

	var totalAnimes int
	var skippedCount int

	// Get total anime count and skipped count
	if animeID != nil {
		// Sync specific anime - check if it has MAL ID
		var malAnimeID sql.NullString
		err := s.db.QueryRow(`
			SELECT mal_anime_id
			FROM animes
			WHERE id = $1
		`, *animeID).Scan(&malAnimeID)

		if err != nil {
			if err == sql.ErrNoRows {
				return 0, 0, 0, 0, nil, ErrAnimeNotFound
			}
			return 0, 0, 0, 0, nil, fmt.Errorf("failed to get anime: %w", err)
		}

		totalAnimes = 1

		// If no MAL ID, skip this anime
		if !malAnimeID.Valid || malAnimeID.String == "" {
			skippedCount = 1
			s.logger.Info("Anime skipped (no MAL ID)",
				"anime_id", *animeID)
			return totalAnimes, 0, 0, skippedCount, nil, nil
		}

		// Add to sync list
		animes = append(animes, struct {
			id         int
			malAnimeID string
		}{*animeID, malAnimeID.String})
	} else {
		// Sync all animes - get total count
		err := s.db.QueryRow(`SELECT COUNT(*) FROM animes`).Scan(&totalAnimes)
		if err != nil {
			return 0, 0, 0, 0, nil, fmt.Errorf("failed to count animes: %w", err)
		}

		// Get animes with MAL ID
		rows, err := s.db.Query(`
			SELECT id, mal_anime_id
			FROM animes
			WHERE mal_anime_id IS NOT NULL AND mal_anime_id != ''
			ORDER BY id
		`)
		if err != nil {
			return 0, 0, 0, 0, nil, fmt.Errorf("failed to get animes: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var malAnimeID string
			if err := rows.Scan(&id, &malAnimeID); err != nil {
				continue
			}
			animes = append(animes, struct {
				id         int
				malAnimeID string
			}{id, malAnimeID})
		}

		// Calculate skipped count (animes without MAL ID)
		skippedCount = totalAnimes - len(animes)
	}

	// Sync each anime
	successCount := 0
	failureCount := 0
	errors := []string{}

	for _, anime := range animes {
		err := s.enrichEpisodesWithDuration(anime.id, anime.malAnimeID)
		if err != nil {
			failureCount++
			errorMsg := fmt.Sprintf("Anime ID %d: %s", anime.id, err.Error())
			errors = append(errors, errorMsg)
			s.logger.Warn("Failed to sync episode durations",
				"anime_id", anime.id,
				"mal_anime_id", anime.malAnimeID,
				"error", err)
		} else {
			successCount++
		}
	}

	s.logger.Info("Episode duration sync completed",
		"total_animes", totalAnimes,
		"success_count", successCount,
		"failure_count", failureCount,
		"skipped_count", skippedCount)

	return totalAnimes, successCount, failureCount, skippedCount, errors, nil
}

// CreateEpisodesForAllAnimes creates episodes for all animes that have missing episodes
func (s *AnimeService) CreateEpisodesForAllAnimes() (int, int, error) {
	// Get all animes where episodes_count doesn't match actual episode count
	query := `
		SELECT a.id, a.title, a.episodes_count, COALESCE(COUNT(e.id), 0) as actual_count
		FROM animes a
		LEFT JOIN episodes e ON a.id = e.anime_id
		WHERE a.episodes_count > 0
		GROUP BY a.id, a.title, a.episodes_count
		HAVING a.episodes_count > COALESCE(COUNT(e.id), 0)
		ORDER BY a.id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	type animeInfo struct {
		id            int
		title         string
		episodesCount int
		actualCount   int
	}

	var animes []animeInfo
	for rows.Next() {
		var anime animeInfo
		if err := rows.Scan(&anime.id, &anime.title, &anime.episodesCount, &anime.actualCount); err != nil {
			return 0, 0, fmt.Errorf("failed to scan anime: %w", err)
		}
		animes = append(animes, anime)
	}

	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("error iterating animes: %w", err)
	}

	s.logger.Info("Found animes with missing episodes", "count", len(animes))

	successCount := 0
	failureCount := 0

	for _, anime := range animes {
		missingCount := anime.episodesCount - anime.actualCount
		s.logger.Info("Creating missing episodes for anime",
			"anime_id", anime.id,
			"title", anime.title,
			"expected_count", anime.episodesCount,
			"actual_count", anime.actualCount,
			"missing_count", missingCount)

		err := s.createEpisodesForAnime(anime.id, anime.episodesCount, anime.actualCount)
		if err != nil {
			failureCount++
			s.logger.Error("Failed to create episodes",
				"anime_id", anime.id,
				"error", err)
		} else {
			successCount++
		}
	}

	s.logger.Info("Episode creation completed",
		"success_count", successCount,
		"failure_count", failureCount)

	return successCount, failureCount, nil
}

// createEpisodesForAnime creates episode records for a specific anime using bulk insert
// existingCount is passed from the caller to avoid redundant database query
func (s *AnimeService) createEpisodesForAnime(animeID int, count int, existingCount int) error {
	// If we already have all episodes, skip
	if existingCount >= count {
		return nil
	}

	// Use generate_series for bulk insert - much more efficient than N individual inserts
	// This creates all episodes in a single database call
	_, err := s.db.Exec(`
		INSERT INTO episodes (anime_id, title, episode_number, created_at, updated_at)
		SELECT 
			$1,
			'第' || n || '話',
			n,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
		FROM generate_series(1, $2) AS n
		ON CONFLICT (anime_id, episode_number) DO NOTHING
	`, animeID, count)

	if err != nil {
		return fmt.Errorf("failed to create episodes: %w", err)
	}

	s.logger.Info("Created episodes for anime",
		"anime_id", animeID,
		"episode_count", count,
		"existing_count", existingCount)

	return nil
}

// EnrichEpisodesWithTitles enriches episodes with titles from Annict API
func (s *AnimeService) EnrichEpisodesWithTitles(animeID *int) (int, int, []string, error) {
	// Constants for rate limit handling
	const (
		progressLogInterval      = 10              // Log progress every N animes
		maxConsecutiveRateLimits = 3               // Stop after N consecutive rate limit errors
		rateLimitWaitDuration    = 5 * time.Second // Wait duration when rate limit is hit
	)

	// アニメ一覧を取得
	var animes []struct {
		id       int
		annictID int
	}

	query := `
		SELECT a.id, a.annict_id
		FROM animes a
		WHERE a.annict_id IS NOT NULL
		AND EXISTS (
			SELECT 1 FROM episodes e 
			WHERE e.anime_id = a.id
		)
	`
	args := []interface{}{}

	if animeID != nil {
		query += " AND a.id = $1"
		args = append(args, *animeID)
	}

	// Add ORDER BY to prioritize recent seasons (descending order)
	// This ensures newer anime are processed first
	query += " ORDER BY a.season DESC NULLS LAST, a.id DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var anime struct {
			id       int
			annictID int
		}
		if err := rows.Scan(&anime.id, &anime.annictID); err != nil {
			return 0, 0, nil, fmt.Errorf("failed to scan anime: %w", err)
		}
		animes = append(animes, anime)
	}

	s.logger.Info("Starting episode title enrichment",
		"total_animes", len(animes))

	// エピソードタイトルを更新
	successCount := 0
	failureCount := 0
	skippedCount := 0
	errorMessages := []string{}
	consecutiveRateLimitErrors := 0

	for i, anime := range animes {
		// Log progress at regular intervals
		if i > 0 && i%progressLogInterval == 0 {
			s.logger.Info("Episode title enrichment progress",
				"processed", i,
				"total", len(animes),
				"success", successCount,
				"failure", failureCount,
				"skipped", skippedCount)
		}

		err := s.annictClient.UpdateEpisodeTitlesFromAnnict(anime.id, anime.annictID)
		if err != nil {
			// Check if this is a rate limit error using errors.Is
			if errors.Is(err, ErrRateLimitExceeded) {
				consecutiveRateLimitErrors++
				skippedCount++

				// If we hit rate limit multiple times in a row, stop processing
				if consecutiveRateLimitErrors >= maxConsecutiveRateLimits {
					s.logger.Warn("Hit rate limit multiple times consecutively, stopping episode title enrichment",
						"processed", i,
						"total", len(animes),
						"success", successCount,
						"failure", failureCount,
						"skipped", skippedCount,
						"max_consecutive_rate_limits", maxConsecutiveRateLimits)
					break
				}

				// Wait before continuing
				s.logger.Warn("Rate limit hit, waiting before continuing",
					"anime_id", anime.id,
					"consecutive_errors", consecutiveRateLimitErrors,
					"wait_duration", rateLimitWaitDuration.String())
				time.Sleep(rateLimitWaitDuration)
				continue
			}

			// Reset consecutive rate limit counter on other errors
			consecutiveRateLimitErrors = 0
			failureCount++
			errorMsg := fmt.Sprintf("Anime ID %d: %s", anime.id, err.Error())
			errorMessages = append(errorMessages, errorMsg)
			s.logger.Warn("Failed to update episode titles",
				"anime_id", anime.id,
				"error", err)
		} else {
			consecutiveRateLimitErrors = 0
			successCount++
		}

		// Note: Rate limiting is now handled internally by AnnictClient
		// No need to call applyRateLimit() here
	}

	s.logger.Info("Episode title enrichment finished",
		"total_animes", len(animes),
		"success_count", successCount,
		"failure_count", failureCount,
		"skipped_count", skippedCount)

	return successCount, failureCount, errorMessages, nil
}

// SyncGenresForAllAnimes syncs genre information for all animes that have MAL ID
// Only syncs animes that haven't been synced in the last 30 days to reduce API calls
// Returns: success count, failure count, error messages, error
func (s *AnimeService) SyncGenresForAllAnimes() (int, int, []string, error) {
	return s.syncGenresForAnimesWithFilter(30, 0) // 30 days filter, no limit
}

// SyncGenresWithLimit syncs genre information for a limited number of animes
// Useful for batch processing to avoid API rate limits
// Returns: success count, failure count, error messages, error
func (s *AnimeService) SyncGenresWithLimit(limit int) (int, int, []string, error) {
	return s.syncGenresForAnimesWithFilter(30, limit) // 30 days filter, with limit
}

// syncGenresForAnimesWithFilter syncs genres for animes that haven't been synced in the specified number of days
// If daysThreshold is 0, syncs all animes regardless of last sync time
// If limit > 0, only syncs up to that many animes (for batch processing)
func (s *AnimeService) syncGenresForAnimesWithFilter(daysThreshold int, limit int) (int, int, []string, error) {
	s.logger.Info("Starting genre sync for animes",
		"days_threshold", daysThreshold,
		"limit", limit)

	// Build query to get animes with MAL ID that need genre sync
	query := `
		SELECT id, mal_anime_id
		FROM animes
		WHERE mal_anime_id IS NOT NULL AND mal_anime_id != ''
	`
	args := []interface{}{}
	argIndex := 1

	// Add filter for animes that haven't been synced recently
	if daysThreshold > 0 {
		query += fmt.Sprintf(`
		AND (genres_synced_at IS NULL OR genres_synced_at < CURRENT_TIMESTAMP - ($%d || ' days')::INTERVAL)
		`, argIndex)
		args = append(args, daysThreshold)
		argIndex++
	}

	query += " ORDER BY id"

	// Add limit if specified
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	successCount := 0
	failureCount := 0
	var errorMessages []string
	processedCount := 0
	logInterval := 100 // Log progress every 100 animes
	consecutiveErrors := 0
	maxConsecutiveErrors := 10 // Stop if 10 consecutive errors occur

	// Stream processing: process each anime one by one without loading all into memory
	for rows.Next() {
		var animeID int
		var malAnimeID string

		if err := rows.Scan(&animeID, &malAnimeID); err != nil {
			s.logger.Error("Failed to scan anime row", "error", err)
			failureCount++
			consecutiveErrors++
			errorMessages = append(errorMessages, fmt.Sprintf("Anime scan error: %s", err.Error()))
		} else {
			// Sync genres for this anime
			err := s.syncGenresForAnime(animeID, malAnimeID)
			if err != nil {
				failureCount++
				consecutiveErrors++
				errorMsg := fmt.Sprintf("Anime ID %d (MAL ID %s): %s", animeID, malAnimeID, err.Error())
				errorMessages = append(errorMessages, errorMsg)
				s.logger.Warn("Failed to sync genres for anime",
					"anime_id", animeID,
					"mal_anime_id", malAnimeID,
					"error", err,
					"consecutive_errors", consecutiveErrors)
			} else {
				successCount++
				consecutiveErrors = 0 // Reset on success
			}

			processedCount++

			// Rate limiting is handled by AniListClient's internal rate limiter
			// No need for manual sleep here

			// Log progress periodically
			if processedCount%logInterval == 0 {
				s.logger.Info("Genre sync progress",
					"processed", processedCount,
					"success", successCount,
					"failure", failureCount,
					"error_rate", fmt.Sprintf("%.2f%%", float64(failureCount)/float64(processedCount)*100))
			}
		}

		// Check for too many consecutive errors (circuit breaker)
		if consecutiveErrors >= maxConsecutiveErrors {
			s.logger.Error("Too many consecutive errors, stopping genre sync",
				"consecutive_errors", consecutiveErrors,
				"processed", processedCount)
			break
		}
	}

	if err := rows.Err(); err != nil {
		return 0, 0, nil, fmt.Errorf("error iterating anime rows: %w", err)
	}

	s.logger.Info("Genre sync completed",
		"success_count", successCount,
		"failure_count", failureCount,
		"total_processed", processedCount,
		"days_threshold", daysThreshold,
		"limit", limit)

	return successCount, failureCount, errorMessages, nil
}

// syncGenresForAnime syncs genre information for a single anime using transaction
// Uses AniList API for genre information (recommended over Jikan API)
func (s *AnimeService) syncGenresForAnime(animeID int, malAnimeID string) error {
	// Use background context for batch processing
	// In a real-world scenario, you might want to pass context from the caller
	ctx := context.Background()

	// Fetch genres from AniList API (returns Japanese genre names)
	genres, err := s.anilistClient.GetGenresByMALId(ctx, malAnimeID)
	if err != nil {
		return fmt.Errorf("failed to fetch genres from AniList: %w", err)
	}

	// Save genres to database using shared function
	if err := saveGenresToDatabase(s.db, s.logger, animeID, genres); err != nil {
		return err
	}

	s.logger.Debug("Synced genres for anime",
		"anime_id", animeID,
		"mal_anime_id", malAnimeID,
		"genre_count", len(genres))

	return nil
}

// SyncEpisodeDetailsForAllAnimes syncs episode details (air_date and duration) for all animes
// Returns: success count, failure count, error messages, error
func (s *AnimeService) SyncEpisodeDetailsForAllAnimes() (int, int, []string, error) {
	s.logger.Info("Starting episode details sync for all animes")

	// Get total count for progress reporting (only animes with episodes missing duration or air_date)
	countQuery := `
		SELECT COUNT(DISTINCT a.id)
		FROM animes a
		INNER JOIN episodes e ON a.id = e.anime_id
		WHERE a.mal_anime_id IS NOT NULL AND a.mal_anime_id != ''
		AND (e.duration_minutes IS NULL OR e.air_date IS NULL)
	`
	var totalCount int
	err := s.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to count animes: %w", err)
	}

	s.logger.Info("Starting episode details sync", "total_animes", totalCount)

	// Stream processing: get animes with MAL ID that have episodes missing duration or air_date
	query := `
		SELECT DISTINCT a.id, a.mal_anime_id
		FROM animes a
		INNER JOIN episodes e ON a.id = e.anime_id
		WHERE a.mal_anime_id IS NOT NULL AND a.mal_anime_id != ''
		AND (e.duration_minutes IS NULL OR e.air_date IS NULL)
		ORDER BY a.id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	successCount := 0
	failureCount := 0
	var errorMessages []string
	processedCount := 0
	logInterval := 100 // Log progress every 100 animes

	// Stream processing: process each anime one by one without loading all into memory
	for rows.Next() {
		var animeID int
		var malAnimeID string

		if err := rows.Scan(&animeID, &malAnimeID); err != nil {
			s.logger.Error("Failed to scan anime row", "error", err)
			failureCount++
			errorMsg := fmt.Sprintf("Anime scan error: %s", err.Error())
			errorMessages = append(errorMessages, errorMsg)
			processedCount++
			continue
		}

		// Sync episode details for this anime
		err := s.syncEpisodeDetailsForAnime(animeID, malAnimeID)
		if err != nil {
			failureCount++
			errorMsg := fmt.Sprintf("Anime ID %d (MAL ID %s): %s", animeID, malAnimeID, err.Error())
			errorMessages = append(errorMessages, errorMsg)
			s.logger.Warn("Failed to sync episode details for anime",
				"anime_id", animeID,
				"mal_anime_id", malAnimeID,
				"error", err)
		} else {
			successCount++
		}

		processedCount++

		// Log progress periodically
		if processedCount%logInterval == 0 {
			s.logger.Info("Episode details sync progress",
				"processed", processedCount,
				"total", totalCount,
				"success", successCount,
				"failure", failureCount,
				"progress", fmt.Sprintf("%.1f%%", float64(processedCount)/float64(totalCount)*100))
		}
	}

	if err := rows.Err(); err != nil {
		return 0, 0, nil, fmt.Errorf("error iterating anime rows: %w", err)
	}

	s.logger.Info("Episode details sync completed",
		"success_count", successCount,
		"failure_count", failureCount,
		"total_processed", processedCount,
		"total_animes", totalCount)

	return successCount, failureCount, errorMessages, nil
}

// syncEpisodeDetailsForAnime syncs episode details (air_date and duration) for a single anime
// Uses a single transaction to ensure atomicity and batch updates for better performance
func (s *AnimeService) syncEpisodeDetailsForAnime(animeID int, malAnimeID string) error {
	// Fetch episodes from Jikan API
	episodes, err := s.jikanClient.GetAnimeEpisodes(malAnimeID)
	if err != nil {
		if errors.Is(err, ErrJikanNotFound) {
			s.logger.Debug("Anime not found in Jikan", "anime_id", animeID, "mal_anime_id", malAnimeID)
			return nil // Not an error, just no data available
		}
		return fmt.Errorf("failed to fetch episodes from Jikan: %w", err)
	}

	if len(episodes) == 0 {
		return nil // No episodes to sync
	}

	// Begin transaction for atomic updates
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Prepare update statement for better performance
	updateQuery := `
		UPDATE episodes
		SET duration_minutes = COALESCE($1, duration_minutes),
		    air_date = COALESCE($2, air_date),
		    updated_at = CURRENT_TIMESTAMP
		WHERE anime_id = $3 AND episode_number = $4
	`
	stmt, err := tx.Prepare(updateQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	updatedCount := 0
	for _, jikanEp := range episodes {
		// Use mal_id as episode number (Jikan API's mal_id corresponds to episode number)
		episodeNumber := jikanEp.MalID

		if episodeNumber <= 0 {
			s.logger.Debug("Invalid episode number from Jikan API",
				"anime_id", animeID,
				"mal_id", jikanEp.MalID)
			continue
		}

		// Prepare duration (nullable)
		var duration *int
		if jikanEp.Duration > 0 {
			duration = &jikanEp.Duration
		}

		// Parse air_date to time.Time (nullable)
		var airDate *time.Time
		if jikanEp.Aired != nil && *jikanEp.Aired != "" {
			// Try ISO 8601 format first (e.g., "2025-07-06T00:00:00+00:00")
			parsedDate, err := time.Parse(time.RFC3339, *jikanEp.Aired)
			if err != nil {
				// Try simple date format (e.g., "2025-07-06")
				parsedDate, err = time.Parse("2006-01-02", *jikanEp.Aired)
			}
			if err == nil {
				airDate = &parsedDate
			} else {
				s.logger.Debug("Failed to parse air_date",
					"anime_id", animeID,
					"episode_number", episodeNumber,
					"aired", *jikanEp.Aired,
					"error", err)
			}
		}

		// Execute update with prepared statement
		_, err = stmt.Exec(duration, airDate, animeID, episodeNumber)
		if err != nil {
			s.logger.Warn("Failed to update episode details",
				"anime_id", animeID,
				"episode_number", episodeNumber,
				"error", err)
			continue
		}

		updatedCount++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Synced episode details for anime",
		"anime_id", animeID,
		"mal_anime_id", malAnimeID,
		"updated_count", updatedCount,
		"total_episodes", len(episodes))

	return nil
}
