package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
)

var (
	ErrAnimeNotFound   = errors.New("anime not found")
	ErrEpisodeNotFound = errors.New("episode not found")
)

// デフォルト値の定数
var (
	// DefaultPopularSearchKeywords - 検索ログが少ない場合のデフォルト人気検索キーワード
	DefaultPopularSearchKeywords = []string{"進撃の巨人", "鬼滅の刃", "呪術廻戦", "ワンピース", "ナルト"}
)

type AnimeService struct {
	db            *database.DB
	logger        *common.Logger
	anilistClient *AniListClient
}

func NewAnimeService(db *database.DB, logger *common.Logger) *AnimeService {
	return &AnimeService{
		db:            db,
		logger:        logger,
		anilistClient: NewAniListClient(),
	}
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
		SELECT id, title, description, image_url, season, status, mal_anime_id, created_at, updated_at, watchers_count
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
		err := rows.Scan(&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount)
		if err != nil {
			continue
		}

		// AniListから画像を取得
		s.enrichAnimeWithCoverImage(&anime)

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
		SELECT a.id, a.title, a.description, a.image_url, a.season, a.status, a.mal_anime_id, a.created_at, a.updated_at, a.watchers_count
		%s %s
		GROUP BY a.id, a.title, a.description, a.image_url, a.season, a.status, a.mal_anime_id, a.created_at, a.updated_at, a.watchers_count
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
		err := rows.Scan(&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount)
		if err != nil {
			continue
		}

		// AniListから画像を取得
		s.enrichAnimeWithCoverImage(&anime)

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
		SELECT id, title, description, image_url, season, status, mal_anime_id, created_at, updated_at, watchers_count
		FROM animes WHERE id = $1
	`, animeID).Scan(&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount)
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

	// Add user data if authenticated
	if userID != nil {
		// Get user favorite status
		var favorite models.UserFavorite
		err := s.db.QueryRow(`
			SELECT id, user_id, anime_id, status, created_at, updated_at
			FROM user_favorites WHERE user_id = $1 AND anime_id = $2
		`, *userID, animeID).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt)
		if err == nil {
			animeWithData.UserFavorite = &favorite
		}

		// Calculate watch progress
		animeWithData.WatchProgress = s.calculateWatchProgress(*userID, animeID)
	}

	return animeWithData, nil
}

func (s *AnimeService) GetAnimeEpisodes(animeID int, userID *int) ([]models.Episode, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	rows, err := s.db.Query(`
		SELECT id, anime_id, title, episode_number
		FROM episodes WHERE anime_id = $1
		ORDER BY episode_number
	`, animeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", err)
	}
	defer rows.Close()

	episodes := []models.Episode{}
	for rows.Next() {
		var episode models.Episode
		err := rows.Scan(&episode.ID, &episode.AnimeID, &episode.Title, &episode.EpisodeNumber)
		if err != nil {
			continue
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
		SELECT a.id, a.title, a.description, a.image_url, a.season, a.status, a.mal_anime_id, a.created_at, a.updated_at, a.watchers_count,
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
			&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount,
			&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// AniListから画像を取得
		s.enrichAnimeWithCoverImage(&anime)

		animeWithData := models.AnimeWithUserData{
			Anime:        anime,
			UserFavorite: &favorite,
		}

		// Calculate watch progress
		animeWithData.WatchProgress = s.calculateWatchProgress(userID, anime.ID)

		favorites = append(favorites, animeWithData)
	}

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

	// 放送年一覧を取得
	yearRows, err := s.db.Query(`
		SELECT DISTINCT season_year
		FROM animes
		WHERE season_year IS NOT NULL 
		  AND season_year > 1900 
		  AND season_year <= EXTRACT(YEAR FROM CURRENT_DATE) + 1
		ORDER BY season_year DESC
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
		SELECT DISTINCT season
		FROM animes
		WHERE season IS NOT NULL AND season != ''
		ORDER BY 
			CASE season
				WHEN 'spring' THEN 1
				WHEN 'summer' THEN 2
				WHEN 'autumn' THEN 3
				WHEN 'winter' THEN 4
				ELSE 5
			END
	`)
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
		searches = DefaultPopularSearchKeywords
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

	if limit <= 0 {
		limit = 10
	}

	// 検索候補では部分一致が重要なので、FTSと前方一致を組み合わせる
	rows, err := s.db.Query(`
		SELECT title
		FROM animes
		WHERE (
			title ILIKE $1 OR  -- 前方一致（インデックス効率化のため）
			to_tsvector('simple', title) @@ plainto_tsquery('simple', $2)  -- FTS検索
		) AND title IS NOT NULL AND title != ''
		GROUP BY title
		ORDER BY 
			CASE 
				WHEN title ILIKE $1 THEN 1  -- 前方一致を最優先
				WHEN title ILIKE $3 THEN 2  -- 部分一致を次優先
				ELSE 3  -- FTS一致を最後
			END,
			LENGTH(title),
			title
		LIMIT $4
	`, query+"%", query, "%"+query+"%", limit)
	if err != nil {
		// データベースエラーの詳細ログ出力
		s.logger.Error("Failed to get search suggestions", "error", err, "query", query, "limit", limit)
		return []string{}, nil // エラーでも空配列を返す
	}
	defer rows.Close()

	suggestions := []string{}
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {
			s.logger.Error("Failed to scan suggestion title", "error", err)
			continue
		}
		// 空文字列やnullをスキップ
		if strings.TrimSpace(title) != "" {
			suggestions = append(suggestions, title)
		}
	}

	// 結果が0件でも正常レスポンス
	return suggestions, nil
}

// enrichAnimeWithCoverImage - AniListから画像を取得してアニメ情報を拡張
func (s *AnimeService) enrichAnimeWithCoverImage(anime *models.Anime) {
	// MAL IDがない場合はスキップ
	if anime.MalAnimeID == nil || *anime.MalAnimeID == "" {
		return
	}

	// AniListから画像を取得
	coverImage, err := s.anilistClient.GetCoverImageByMALId(*anime.MalAnimeID)
	if err != nil {
		// エラーは無視（フォールバック: Annictの画像を使用）
		s.logger.Debug("Failed to get cover image from AniList",
			"error", err,
			"mal_anime_id", *anime.MalAnimeID,
			"anime_title", anime.Title)
		return
	}

	// 画像URLを設定
	if coverImage != nil {
		anime.CoverImageLarge = coverImage.Large
		anime.CoverImageMedium = coverImage.Medium
	}
}

// enrichAnimesWithCoverImages - 複数のアニメに画像を追加（並列処理）
func (s *AnimeService) enrichAnimesWithCoverImages(animes []models.Anime) {
	// 各アニメに対して画像を取得
	for i := range animes {
		s.enrichAnimeWithCoverImage(&animes[i])
	}
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

	// 年検索
	if params.Year > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("a.season_year = $%d", argIndex))
		args = append(args, params.Year)
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
		return fmt.Sprintf("ORDER BY a.season_year %s, a.season %s", orderDirection, orderDirection)
	case SortByPopularity:
		return fmt.Sprintf("ORDER BY a.watchers_count %s", orderDirection)
	case SortByUpdated:
		return fmt.Sprintf("ORDER BY a.updated_at %s", orderDirection)
	default:
		return fmt.Sprintf("ORDER BY a.title %s", orderDirection)
	}
}
