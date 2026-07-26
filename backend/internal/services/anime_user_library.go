package services

import (
	"fmt"
	"strings"

	"myanimelogs/internal/models"
)

// GetUserFavorites はお気に入り（ハート）一覧を返す。視聴ステータスは持たない。
func (s *AnimeService) GetUserFavorites(userID int) ([]models.AnimeWithUserData, error) {
	// VIEWを使用してステータスを動的に計算（GetAnimeByID / SearchAnimes と一貫性を保つ）
	query := `
		SELECT ` + aliasedAnimeSelectColumns + `,
		       uf.id, uf.user_id, uf.anime_id, uf.created_at, uf.updated_at
		FROM animes_with_computed_status a
		JOIN user_favorites uf ON a.id = uf.anime_id
		WHERE uf.user_id = $1
		ORDER BY uf.updated_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user favorites: %w", err)
	}
	defer rows.Close()

	favorites := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		var favorite models.UserFavorite
		err := rows.Scan(
			&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status,
			&anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount,
			&anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium,
			&anime.ReleasedOn, &anime.Media,
			&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.CreatedAt, &favorite.UpdatedAt,
		)
		if err != nil {
			s.logger.Warn("Failed to scan favorite anime row", "error", err)
			continue
		}

		animeWithData := models.AnimeWithUserData{
			Anime:        anime,
			UserFavorite: &favorite,
		}

		favorites = append(favorites, animeWithData)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user favorites: %w", err)
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithUserDataCoverImagesIfNeeded(favorites)

	return favorites, nil
}

// UpdateFavorite はアニメをお気に入り（ハート）に追加する。既に存在する場合はそのまま返す。
func (s *AnimeService) UpdateFavorite(userID, animeID int) (*models.UserFavorite, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	// Upsert favorite（ステータスを持たないため、既存行は updated_at のみ更新）
	var favorite models.UserFavorite
	err = s.db.QueryRow(`
		INSERT INTO user_favorites (user_id, anime_id, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, anime_id)
		DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, anime_id, created_at, updated_at
	`, userID, animeID).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.CreatedAt, &favorite.UpdatedAt)
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

// GetUserWatchStatuses は視聴ステータス一覧を返す。status を指定するとそのステータスで絞り込む。
func (s *AnimeService) GetUserWatchStatuses(userID int, status string) ([]models.AnimeWithUserData, error) {
	whereClause := "WHERE ws.user_id = $1"
	args := []interface{}{userID}

	if status != "" {
		whereClause += " AND ws.status = $2"
		args = append(args, status)
	}

	// VIEWを使用してステータスを動的に計算（GetAnimeByID / SearchAnimes と一貫性を保つ）
	query := fmt.Sprintf(`
		SELECT `+aliasedAnimeSelectColumns+`,
		       ws.id, ws.user_id, ws.anime_id, ws.status, ws.created_at, ws.updated_at
		FROM animes_with_computed_status a
		JOIN user_watch_status ws ON a.id = ws.anime_id
		%s
		ORDER BY ws.updated_at DESC
	`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user watch statuses: %w", err)
	}
	defer rows.Close()

	statuses := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		var watchStatus models.UserWatchStatus
		err := rows.Scan(
			&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status,
			&anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount,
			&anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium,
			&anime.ReleasedOn, &anime.Media,
			&watchStatus.ID, &watchStatus.UserID, &watchStatus.AnimeID, &watchStatus.Status, &watchStatus.CreatedAt, &watchStatus.UpdatedAt,
		)
		if err != nil {
			s.logger.Warn("Failed to scan watch status anime row", "error", err)
			continue
		}

		statuses = append(statuses, models.AnimeWithUserData{
			Anime:           anime,
			UserWatchStatus: &watchStatus,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user watch statuses: %w", err)
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithUserDataCoverImagesIfNeeded(statuses)

	return statuses, nil
}

// UpdateWatchStatus は視聴ステータスを設定・更新する。
func (s *AnimeService) UpdateWatchStatus(userID, animeID int, status string) (*models.UserWatchStatus, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	// Upsert watch status
	var watchStatus models.UserWatchStatus
	err = s.db.QueryRow(`
		INSERT INTO user_watch_status (user_id, anime_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, anime_id)
		DO UPDATE SET status = $3, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, anime_id, status, created_at, updated_at
	`, userID, animeID, status).Scan(&watchStatus.ID, &watchStatus.UserID, &watchStatus.AnimeID, &watchStatus.Status, &watchStatus.CreatedAt, &watchStatus.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update watch status: %w", err)
	}

	return &watchStatus, nil
}

func (s *AnimeService) DeleteWatchStatus(userID, animeID int) error {
	_, err := s.db.Exec(`
		DELETE FROM user_watch_status WHERE user_id = $1 AND anime_id = $2
	`, userID, animeID)
	if err != nil {
		return fmt.Errorf("failed to delete watch status: %w", err)
	}

	return nil
}

// batchGetUserFavorites - 複数アニメのお気に入り情報を一括取得（N+1クエリ解消）
func (s *AnimeService) batchGetUserFavorites(userID int, animeIDs []int) (map[int]*models.UserFavorite, error) {
	if len(animeIDs) == 0 {
		return make(map[int]*models.UserFavorite), nil
	}

	// プレースホルダーを動的に生成
	placeholders := make([]string, len(animeIDs))
	args := make([]interface{}, len(animeIDs)+1)
	args[0] = userID
	for i, id := range animeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, anime_id, created_at, updated_at
		FROM user_favorites
		WHERE user_id = $1 AND anime_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get user favorites: %w", err)
	}
	defer rows.Close()

	favorites := make(map[int]*models.UserFavorite)
	for rows.Next() {
		var fav models.UserFavorite
		if err := rows.Scan(&fav.ID, &fav.UserID, &fav.AnimeID, &fav.CreatedAt, &fav.UpdatedAt); err != nil {
			s.logger.Warn("Failed to scan user favorite", "error", err)
			continue
		}
		favorites[fav.AnimeID] = &fav
	}

	return favorites, nil
}

// enrichAnimesWithUserData - アニメリストにユーザーデータを一括追加（N+1クエリ解消版）
// 注意: この関数はユーザーデータ取得に失敗してもエラーを返さず、警告ログのみ出力します。
// これは意図的な設計で、ユーザーデータ（お気に入り・視聴進捗）が取得できなくても
// アニメ一覧自体は表示できるようにするためです。クライアント側では、
// UserFavoriteやWatchProgressがnilまたは0の場合を適切にハンドリングしてください。
func (s *AnimeService) enrichAnimesWithUserData(animes []models.AnimeWithUserData, userID int) {
	if len(animes) == 0 {
		return
	}

	// アニメIDのリストを作成
	animeIDs := make([]int, len(animes))
	for i, anime := range animes {
		animeIDs[i] = anime.ID
	}

	// お気に入り情報を一括取得
	// エラー時は空のmapを使用し、処理を続行（グレースフルデグラデーション）
	favorites, err := s.batchGetUserFavorites(userID, animeIDs)
	if err != nil {
		s.logger.Warn("Failed to batch get user favorites", "error", err)
		favorites = make(map[int]*models.UserFavorite)
	}

	// 各アニメにユーザーデータを設定
	for i := range animes {
		animeID := animes[i].ID
		if fav, ok := favorites[animeID]; ok {
			animes[i].UserFavorite = fav
		}
	}
}
