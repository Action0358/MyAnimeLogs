package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
)

var (
	ErrAnimeNotFound   = errors.New("anime not found")
	ErrEpisodeNotFound = errors.New("episode not found")
)

type AnimeService struct {
	db *database.DB
}

func NewAnimeService(db *database.DB) *AnimeService {
	return &AnimeService{
		db: db,
	}
}

func (s *AnimeService) SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if params.Keyword != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+params.Keyword+"%")
		argIndex++
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
		SELECT id, title, description, image_url, season, status, created_at, updated_at
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
		err := rows.Scan(&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.CreatedAt, &anime.UpdatedAt)
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
		SELECT id, title, description, image_url, season, status, created_at, updated_at
		FROM animes WHERE id = $1
	`, animeID).Scan(&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.CreatedAt, &anime.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnimeNotFound
		}
		return nil, fmt.Errorf("failed to get anime: %w", err)
	}

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
		SELECT a.id, a.title, a.description, a.image_url, a.season, a.status, a.created_at, a.updated_at,
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
			&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.CreatedAt, &anime.UpdatedAt,
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
			fmt.Printf("Failed to add to watch history: %v\n", err)
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