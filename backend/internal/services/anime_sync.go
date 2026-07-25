package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"
)

// SyncAnimeStatus - TMDBからステータスを取得して更新
func (s *AnimeService) SyncAnimeStatus(ctx context.Context, animeID int, clearCache bool) error {
	// アニメ情報を取得（status, season, released_onを含める）
	var currentStatus, releasedOn sql.NullString
	var season *string
	err := s.db.QueryRowContext(ctx, `
		SELECT season, released_on, status
		FROM animes
		WHERE id = $1
	`, animeID).Scan(&season, &releasedOn, &currentStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAnimeNotFound
		}
		return fmt.Errorf("failed to get anime for status sync: %w", err)
	}

	var releasedOnPtr *string
	if releasedOn.Valid {
		releasedOnPtr = &releasedOn.String
	}

	// season + released_on ベースでステータスを判定
	status := determineStatusBySeason(season, releasedOnPtr)

	// 既存のステータスと比較して変更がなければ更新しない
	if currentStatus.Valid && currentStatus.String == status {
		return nil
	}

	// ステータスを更新
	_, err = s.db.ExecContext(ctx, `
		UPDATE animes
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, status, animeID)
	if err != nil {
		return fmt.Errorf("failed to update anime status for anime %d to %s: %w", animeID, status, err)
	}

	if clearCache {
		s.ClearAnimeCache()
	}
	s.logger.Debug("Anime status updated", "anime_id", animeID, "new_status", status)

	return nil
}

// SyncAllAnimeStatuses - 全アニメのステータスを更新
// season + released_on ベースで判定するため、外部API呼び出しなしで高速に処理可能
func (s *AnimeService) SyncAllAnimeStatuses(ctx context.Context) (int, int, error) {
	// 差分フィルタ: 放送中・未放送・ステータス未設定の作品のみ対象
	// 完了作品はステータスが変わらないため除外
	rows, err := s.db.QueryContext(ctx, `
        SELECT id FROM animes
        WHERE season IS NOT NULL
        AND (
            status IN ($1, $2)
            OR status IS NULL
        )
    `, constants.AnimeStatusAiring.String(), constants.AnimeStatusUpcoming.String())
	if err != nil {
		s.logger.Error("Failed to query animes for status sync", "error", err)
		return 0, 0, fmt.Errorf("failed to query animes for status sync: %w", err)
	}
	defer rows.Close()

	successCount := 0
	failureCount := 0

	var animeIDs []int
	for rows.Next() {
		var animeID int
		if err := rows.Scan(&animeID); err != nil {
			s.logger.Warn("Failed to scan anime ID for status sync", "error", err)
			failureCount++
			continue
		}
		animeIDs = append(animeIDs, animeID)
	}

	totalAnimes := len(animeIDs)
	if totalAnimes == 0 {
		s.logger.Info("No animes found requiring status sync.")
		return 0, 0, nil
	}

	s.logger.Info("Starting status sync for animes", "total_animes", totalAnimes)

	for i, animeID := range animeIDs {
		// コンテキストキャンセルチェック
		if ctx.Err() != nil {
			s.logger.Warn("Status sync cancelled", "progress", fmt.Sprintf("%d/%d", i+1, totalAnimes), "reason", ctx.Err())
			return successCount, failureCount, ctx.Err()
		}

		// 進行状況のロギング
		if i > 0 && i%100 == 0 {
			s.logger.Info("Processing anime status sync", "progress", fmt.Sprintf("%d/%d", i+1, totalAnimes))
		}

		if err := s.SyncAnimeStatus(ctx, animeID, false); err != nil {
			s.logger.Warn("Failed to sync anime status", "anime_id", animeID, "error", err)
			failureCount++
		} else {
			successCount++
		}
	}

	// 全ての更新が完了した後に一度だけキャッシュをクリア
	if successCount > 0 {
		s.ClearAnimeCache()
	}

	s.logger.Info("Anime status sync completed", "success", successCount, "failure", failureCount)
	return successCount, failureCount, nil
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
		AND (genres_synced_at IS NULL OR genres_synced_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * $%d)
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
// Uses AniList API for genre information
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
	if err := SaveGenresToDatabase(s.db, s.logger, animeID, genres); err != nil {
		return err
	}

	s.logger.Debug("Synced genres for anime",
		"anime_id", animeID,
		"mal_anime_id", malAnimeID,
		"genre_count", len(genres))

	return nil
}

// SyncCoverImagesForAllAnimes - 全アニメのカバー画像をAniListから事前取得してDBにキャッシュ
// スケジューラーから呼び出され、画像未取得のアニメに対してバッチ処理を行う
// 戻り値: 成功数, 失敗数, エラーメッセージリスト, エラー
func (s *AnimeService) SyncCoverImagesForAllAnimes() (int, int, []string, error) {
	// 画像未取得のアニメを取得（MAL IDがあるもののみ）
	rows, err := s.db.Query(`
		SELECT id, title, mal_anime_id
		FROM animes
		WHERE mal_anime_id IS NOT NULL
		  AND mal_anime_id != ''
		  AND (cover_image_large IS NULL OR cover_image_large = '')
		ORDER BY watchers_count DESC NULLS LAST
	`)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to query animes without cover images: %w", err)
	}
	defer rows.Close()

	type animeToSync struct {
		ID    int
		Title string
		MalID string
	}

	var animesToSync []animeToSync
	for rows.Next() {
		var a animeToSync
		if err := rows.Scan(&a.ID, &a.Title, &a.MalID); err != nil {
			s.logger.Warn("Failed to scan anime for cover image sync", "error", err)
			continue
		}
		animesToSync = append(animesToSync, a)
	}

	if len(animesToSync) == 0 {
		s.logger.Info("No animes need cover image sync")
		return 0, 0, nil, nil
	}

	s.logger.Info("Starting cover image sync", "total_animes", len(animesToSync))

	// 結果を集約するための構造体
	type syncResult struct {
		success bool
		errMsg  string
	}

	results := make(chan syncResult, len(animesToSync))
	ctx := context.Background()

	// ワーカープールで並列処理（AniListClientの内部レートリミッターが制御）
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, constants.CoverImageSyncWorkerPoolSize)

	for i, anime := range animesToSync {
		// 進捗ログ（100件ごと）
		if (i+1)%100 == 0 {
			s.logger.Info("Cover image sync progress",
				"queued", i+1,
				"total", len(animesToSync))
		}

		wg.Add(1)
		go func(anime animeToSync) {
			defer wg.Done()
			semaphore <- struct{}{}        // セマフォ取得
			defer func() { <-semaphore }() // セマフォ解放

			// AniListから画像を取得
			coverImage, apiErr := s.anilistClient.GetCoverImageByMALId(ctx, anime.MalID)

			// MAL IDで取得失敗またはnot foundの場合、タイトルでフォールバック
			if apiErr != nil || coverImage == nil {
				coverImage, apiErr = s.anilistClient.GetCoverImageByTitle(ctx, anime.Title)
			}

			if apiErr != nil || coverImage == nil {
				errMsg := fmt.Sprintf("anime_id=%d (%s): failed to get cover image: %v", anime.ID, anime.Title, apiErr)
				results <- syncResult{success: false, errMsg: errMsg}
				return
			}

			// DBに保存
			_, saveErr := s.db.Exec(`
				UPDATE animes 
				SET cover_image_extra_large = $1,
					cover_image_large = $2,
					cover_image_medium = $3,
					cover_image_updated_at = NOW()
				WHERE id = $4
			`, coverImage.ExtraLarge, coverImage.Large, coverImage.Medium, anime.ID)

			if saveErr != nil {
				errMsg := fmt.Sprintf("anime_id=%d (%s): failed to save to DB: %v", anime.ID, anime.Title, saveErr)
				results <- syncResult{success: false, errMsg: errMsg}
				return
			}

			results <- syncResult{success: true}
		}(anime)
	}

	// 全ワーカーの完了を待ってからチャネルを閉じる
	go func() {
		wg.Wait()
		close(results)
	}()

	// 結果を集計
	successCount := 0
	failureCount := 0
	var errors []string

	for result := range results {
		if result.success {
			successCount++
		} else {
			failureCount++
			if len(errors) < constants.BatchSyncMaxErrors {
				errors = append(errors, result.errMsg)
			}
		}
	}

	s.logger.Info("Cover image sync completed",
		"success", successCount,
		"failure", failureCount,
		"total", len(animesToSync))

	return successCount, failureCount, errors, nil
}

// IsTMDBEnabled reports whether the TMDB client is configured and usable.
func (s *AnimeService) IsTMDBEnabled() bool {
	return s.tmdbClient != nil && s.tmdbClient.IsEnabled()
}

// SyncDescriptionsFromTMDB syncs Japanese descriptions (and tmdb_id) from TMDB for all
// animes that need them. Thin wrapper over SyncTMDBMetadata with no per-run limit,
// used by the manual full-sync script.
// Returns: success count, not found count, failure count, error
func (s *AnimeService) SyncDescriptionsFromTMDB() (int, int, int, error) {
	return s.SyncTMDBMetadata(context.Background(), 0)
}

// tmdbAnimeInfo holds the minimal anime fields needed for TMDB metadata resolution.
type tmdbAnimeInfo struct {
	ID    int
	Title string
	MalID *string
}

// loadAnimesNeedingTMDBMetadata returns up to `limit` animes that need TMDB metadata
// (missing description, or an English one). The DB connection is released before the
// caller's long TMDB processing loop (rows are read fully here), and rows.Close() is
// deferred so the connection is always returned to the pool even on early return/panic.
func (s *AnimeService) loadAnimesNeedingTMDBMetadata(ctx context.Context, limit int) ([]tmdbAnimeInfo, error) {
	// Animes that need description update:
	// 1. No description (NULL or empty)
	// 2. English description (HTML tags like <i>, <br>, or ASCII start without Japanese punctuation)
	// Ordered by popularity so resolvable (popular) works are filled first; bounded by
	// limit per run to control TMDB API load.
	query := `
		SELECT id, title, mal_anime_id
		FROM animes
		WHERE description IS NULL
		   OR description = ''
		   OR description LIKE '%<i>%'
		   OR description LIKE '%<br>%'
		   OR (description ~ '^[A-Za-z]' AND description NOT LIKE '%。%' AND description NOT LIKE '%、%')
		ORDER BY watchers_count DESC NULLS LAST, id`
	var args []interface{}
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get animes: %w", err)
	}
	defer rows.Close()

	var animes []tmdbAnimeInfo
	for rows.Next() {
		var a tmdbAnimeInfo
		if err := rows.Scan(&a.ID, &a.Title, &a.MalID); err != nil {
			s.logger.Warn("Failed to scan anime", "error", err)
			continue
		}
		animes = append(animes, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating animes: %w", err)
	}

	return animes, nil
}

// SyncTMDBMetadata resolves tmdb_id and Japanese descriptions from TMDB for animes that
// need them (missing description, or an English one detected by HTML tags / lack of
// Japanese punctuation). Resolving the description also persists tmdb_id / series_tmdb_id
// as a side effect, so this doubles as the TMDB ID backfill.
//
// It honors the TMDB rate limit (~40 req/10s) via batching, stops early when ctx is
// cancelled (graceful shutdown / timeout), and processes at most `limit` animes per run
// (limit <= 0 means unlimited). Returns: success, not-found, failure counts.
func (s *AnimeService) SyncTMDBMetadata(ctx context.Context, limit int) (int, int, int, error) {
	s.logger.Info("Starting TMDB metadata sync", "limit", limit)

	// Check if TMDB client is enabled
	if s.tmdbClient == nil || !s.tmdbClient.IsEnabled() {
		return 0, 0, 0, fmt.Errorf("TMDB client is not configured")
	}

	// 対象作品を取得（接続は読み取り後すぐ解放し、長時間のTMDB処理中は保持しない）
	animes, err := s.loadAnimesNeedingTMDBMetadata(ctx, limit)
	if err != nil {
		return 0, 0, 0, err
	}

	totalAnimes := len(animes)
	if totalAnimes == 0 {
		s.logger.Info("All animes already have TMDB metadata")
		return 0, 0, 0, nil
	}

	s.logger.Info("Found animes needing TMDB metadata", "count", totalAnimes)

	// TMDB rate limit: ~40 requests/10 seconds
	const batchSize = 40
	const batchInterval = 11 * time.Second

	successCount := 0
	notFoundCount := 0
	failureCount := 0

	// Process in batches
	for i := 0; i < totalAnimes; i += batchSize {
		// タイムアウト/シャットダウン時は中断（未処理分は次回バッチで処理）
		if ctx.Err() != nil {
			s.logger.Info("TMDB metadata sync interrupted (context cancelled)",
				"processed", successCount+notFoundCount+failureCount)
			break
		}

		end := i + batchSize
		if end > totalAnimes {
			end = totalAnimes
		}
		batch := animes[i:end]

		s.logger.Info("Processing TMDB batch",
			"progress", fmt.Sprintf("%d/%d", end, totalAnimes))

		for _, anime := range batch {
			if ctx.Err() != nil {
				break
			}

			itemCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

			malID := ""
			if anime.MalID != nil {
				malID = *anime.MalID
			}

			// AniListからメディアタイプを取得（MOVIE, TV等）
			mediaTypeHint := ""
			if malID != "" {
				format, formatErr := s.anilistClient.GetMediaFormatByMALId(itemCtx, malID)
				if formatErr == nil {
					mediaTypeHint = format
				}
			}

			overview, tmdbID, err := s.tmdbClient.GetOverviewByMALIDOrTitle(itemCtx, malID, anime.Title, mediaTypeHint)
			cancel()

			// 親ctxがキャンセル/タイムアウトした場合は、フォールバックDBクエリや
			// notFound計上をせず即中断する（err は context.Canceled 等になる）。
			// itemCtx の10秒タイムアウト（親は生存）の場合は ctx.Err()==nil のため
			// 通常どおり not-found 扱いで継続する。
			if ctx.Err() != nil {
				s.logger.Info("TMDB metadata sync interrupted during fetch (context cancelled)",
					"processed", successCount+notFoundCount+failureCount)
				break
			}

			if err != nil || overview == "" {
				// TMDBから直接取得できなかった場合、シリーズフォールバックを試行
				seriesTitle := NormalizeAnimeTitle(anime.Title)
				var seriesDesc string
				var seriesTMDBID sql.NullInt64
				seriesErr := s.db.QueryRow(`
					SELECT description, series_tmdb_id
					FROM animes
					WHERE series_title = $1
					  AND description IS NOT NULL
					  AND description != ''
					  AND description NOT LIKE '%<i>%'
					  AND description NOT LIKE '%<br>%'
					ORDER BY series_tmdb_id IS NOT NULL DESC, id ASC
					LIMIT 1
				`, seriesTitle).Scan(&seriesDesc, &seriesTMDBID)

				if seriesErr == nil && seriesDesc != "" {
					// シリーズフォールバックで取得成功
					var seriesTMDBIDVal int
					if seriesTMDBID.Valid {
						seriesTMDBIDVal = int(seriesTMDBID.Int64)
					}
					s.updateAnimeSeriesInfo(context.Background(), anime.ID, seriesDesc, 0, seriesTitle, seriesTMDBIDVal)
					successCount++
					continue
				}

				notFoundCount++
				continue
			}

			// Save to database with series_title and series_tmdb_id
			seriesTitle := NormalizeAnimeTitle(anime.Title)
			s.updateAnimeSeriesInfo(context.Background(), anime.ID, overview, tmdbID, seriesTitle, tmdbID)
			successCount++
		}

		// Wait before next batch (except for last batch); abort early on cancellation
		if end < totalAnimes {
			if err := sleepWithContext(ctx, batchInterval); err != nil {
				s.logger.Info("TMDB metadata sync interrupted during backoff",
					"processed", successCount+notFoundCount+failureCount)
				break
			}
		}
	}

	s.logger.Info("TMDB metadata sync completed",
		"success", successCount,
		"not_found", notFoundCount,
		"failure", failureCount)

	return successCount, notFoundCount, failureCount, nil
}

// SyncReleasedOnFromJikan syncs release dates from Jikan API for animes without released_on
// Returns: success count, not found count, failure count, error
func (s *AnimeService) SyncReleasedOnFromJikan() (int, int, int, error) {
	s.logger.Info("Starting Jikan released_on sync")

	// Get animes that need released_on update:
	// 1. No released_on (NULL or empty string)
	// 2. Has MAL ID (required for Jikan API)
	rows, err := s.db.Query(`
		SELECT id, title, mal_anime_id 
		FROM animes 
		WHERE (released_on IS NULL OR released_on = '')
		  AND mal_anime_id IS NOT NULL AND mal_anime_id != ''
		ORDER BY watchers_count DESC NULLS LAST, id
	`)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get animes: %w", err)
	}

	type animeInfo struct {
		ID    int
		Title string
		MalID string
	}
	var animes []animeInfo

	for rows.Next() {
		var a animeInfo
		// SQL query already filters for mal_anime_id IS NOT NULL AND mal_anime_id != ''
		// so we can scan directly into the string field
		if err := rows.Scan(&a.ID, &a.Title, &a.MalID); err != nil {
			s.logger.Warn("Failed to scan anime", "error", err)
			continue
		}
		animes = append(animes, a)
	}
	rows.Close()

	totalAnimes := len(animes)
	if totalAnimes == 0 {
		s.logger.Info("All animes already have released_on")
		return 0, 0, 0, nil
	}

	s.logger.Info("Found animes without released_on", "count", totalAnimes)

	// Jikan rate limit: 3 requests/second
	// Process in batches of 30 with 11 second intervals for safety
	const batchSize = 30
	const batchInterval = 11 * time.Second

	successCount := 0
	notFoundCount := 0
	failureCount := 0

	// Process in batches
	for i := 0; i < totalAnimes; i += batchSize {
		end := i + batchSize
		if end > totalAnimes {
			end = totalAnimes
		}
		batch := animes[i:end]

		s.logger.Info("Processing Jikan batch",
			"progress", fmt.Sprintf("%d/%d", end, totalAnimes))

		for _, anime := range batch {
			releasedOn, err := s.jikanClient.GetAnimeAiredDate(anime.MalID)

			if err != nil {
				if errors.Is(err, ErrJikanNotFound) {
					s.logger.Debug("Anime not found in Jikan",
						"anime_id", anime.ID,
						"mal_id", anime.MalID,
						"title", anime.Title)
					notFoundCount++
				} else {
					s.logger.Debug("Failed to get aired date from Jikan",
						"anime_id", anime.ID,
						"mal_id", anime.MalID,
						"title", anime.Title,
						"error", err)
					failureCount++
				}
				continue
			}

			if releasedOn == "" {
				notFoundCount++
				continue
			}

			// Save to database
			result, err := s.db.Exec(`
				UPDATE animes 
				SET released_on = $1
				WHERE id = $2
			`, releasedOn, anime.ID)

			if err != nil {
				s.logger.Warn("Failed to save released_on",
					"anime_id", anime.ID,
					"error", err)
				failureCount++
				continue
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				successCount++
			}
		}

		// Wait before next batch (except for last batch)
		if end < totalAnimes {
			time.Sleep(batchInterval)
		}
	}

	s.logger.Info("Jikan released_on sync completed",
		"success", successCount,
		"not_found", notFoundCount,
		"failure", failureCount)

	return successCount, notFoundCount, failureCount, nil
}

// SyncReleasedOnForCurrentSeason syncs released_on from Jikan API for current and previous season animes
// that have released_on=NULL, then re-evaluates their status.
// This ensures that animes within the current season period are correctly labeled as "upcoming" vs "airing".
func (s *AnimeService) SyncReleasedOnForCurrentSeason(ctx context.Context, currentSeason, previousSeason string) (int, int, error) {
	s.logger.Info("Syncing released_on for current season animes without release date",
		"current_season", currentSeason,
		"previous_season", previousSeason)

	// released_onがNULLまたは空文字の今期・前期作品を取得
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, mal_anime_id
		FROM animes
		WHERE season IN ($1, $2)
		AND (released_on IS NULL OR released_on = '')
		AND mal_anime_id IS NOT NULL AND mal_anime_id != ''
	`, currentSeason, previousSeason)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query animes for released_on sync: %w", err)
	}
	defer rows.Close()

	type animeInfo struct {
		ID    int
		Title string
		MalID string
	}
	var animes []animeInfo

	for rows.Next() {
		var a animeInfo
		if err := rows.Scan(&a.ID, &a.Title, &a.MalID); err != nil {
			s.logger.Warn("Failed to scan anime for released_on sync", "error", err)
			continue
		}
		animes = append(animes, a)
	}

	totalAnimes := len(animes)
	if totalAnimes == 0 {
		s.logger.Info("All current/previous season animes already have released_on")
		return 0, 0, nil
	}

	s.logger.Info("Found current/previous season animes without released_on", "count", totalAnimes)

	successCount := 0
	failureCount := 0

	for _, anime := range animes {
		// コンテキストのキャンセル・タイムアウトを確認
		select {
		case <-ctx.Done():
			s.logger.Warn("SyncReleasedOnForCurrentSeason cancelled",
				"success", successCount, "failure", failureCount)
			return successCount, failureCount, ctx.Err()
		default:
		}

		releasedOn, err := s.jikanClient.GetAnimeAiredDate(anime.MalID)
		if err != nil {
			s.logger.Debug("Failed to get aired date from Jikan for current season anime",
				"anime_id", anime.ID,
				"mal_id", anime.MalID,
				"title", anime.Title,
				"error", err)
			failureCount++
			time.Sleep(constants.JikanRateLimitDelay)
			continue
		}

		if releasedOn == "" {
			time.Sleep(constants.JikanRateLimitDelay)
			continue
		}

		// released_onをDBに保存
		_, err = s.db.ExecContext(ctx, `
			UPDATE animes SET released_on = $1 WHERE id = $2
		`, releasedOn, anime.ID)
		if err != nil {
			s.logger.Warn("Failed to save released_on for current season anime",
				"anime_id", anime.ID, "error", err)
			failureCount++
			time.Sleep(constants.JikanRateLimitDelay)
			continue
		}

		s.logger.Debug("Saved released_on for current season anime",
			"anime_id", anime.ID,
			"title", anime.Title,
			"released_on", releasedOn)

		// released_on設定後にステータスを再判定
		if syncErr := s.SyncAnimeStatus(ctx, anime.ID, false); syncErr != nil {
			s.logger.Warn("Failed to re-evaluate status after released_on update",
				"anime_id", anime.ID, "error", syncErr)
		}

		successCount++
		time.Sleep(constants.JikanRateLimitDelay)
	}

	// キャッシュをクリア
	if successCount > 0 {
		s.ClearAnimeCache()
	}

	s.logger.Info("Current season released_on sync completed",
		"success", successCount,
		"failure", failureCount)

	return successCount, failureCount, nil
}

// SyncPriorityAnimeData syncs episodes, descriptions, and released_on for specific anime IDs.
// Used by priority sync mode to enrich only current season + popular animes.
func (s *AnimeService) SyncPriorityAnimeData(animeIDs []int) error {
	if len(animeIDs) == 0 {
		return nil
	}

	s.logger.Info("Starting priority anime data sync", "anime_count", len(animeIDs))

	// バッチクエリで全アニメ情報を一括取得（N+1問題の回避）
	type animeInfo struct {
		ID          int
		Title       string
		Description *string
		ReleasedOn  *string
		MalAnimeID  *string
		AnnictID    *int
	}
	animeMap := make(map[int]*animeInfo, len(animeIDs))

	// IN句用のプレースホルダーを構築
	placeholders := ""
	args := make([]interface{}, len(animeIDs))
	for i, id := range animeIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT id, title, description, released_on, mal_anime_id, annict_id
		FROM animes WHERE id IN (%s)
	`, placeholders), args...)
	if err != nil {
		return fmt.Errorf("failed to batch fetch anime info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var info animeInfo
		var desc, relOn, malID sql.NullString
		var annictID sql.NullInt64
		if err := rows.Scan(&info.ID, &info.Title, &desc, &relOn, &malID, &annictID); err != nil {
			s.logger.Warn("Failed to scan anime info", "error", err)
			continue
		}
		if desc.Valid && desc.String != "" {
			info.Description = &desc.String
		}
		if relOn.Valid && relOn.String != "" {
			info.ReleasedOn = &relOn.String
		}
		if malID.Valid && malID.String != "" {
			info.MalAnimeID = &malID.String
		}
		if annictID.Valid && annictID.Int64 > 0 {
			aid := int(annictID.Int64)
			info.AnnictID = &aid
		}
		animeMap[info.ID] = &info
	}

	s.logger.Info("Batch fetched anime info", "fetched", len(animeMap))

	// Step 1: Sync descriptions from TMDB
	s.logger.Info("Priority sync: descriptions from TMDB...")
	descSuccess, descSkip := 0, 0
	for _, id := range animeIDs {
		info, ok := animeMap[id]
		if !ok {
			continue
		}
		if info.Description != nil && *info.Description != "" {
			descSkip++
			continue
		}
		anime := models.Anime{
			ID:         info.ID,
			Title:      info.Title,
			MalAnimeID: info.MalAnimeID,
		}
		s.enrichAnimeWithDescription(&anime)
		if anime.Description != nil && *anime.Description != "" {
			descSuccess++
			// メモリ上のマップも更新
			info.Description = anime.Description
		}
	}
	s.logger.Info("Priority sync: descriptions completed", "success", descSuccess, "skipped", descSkip)

	// Step 2: Sync released_on from Jikan
	s.logger.Info("Priority sync: released_on from Jikan...")
	relSuccess, relSkip := 0, 0
	for _, id := range animeIDs {
		info, ok := animeMap[id]
		if !ok {
			continue
		}
		if info.ReleasedOn != nil && *info.ReleasedOn != "" {
			relSkip++
			continue
		}
		anime := models.Anime{
			ID:         info.ID,
			Title:      info.Title,
			MalAnimeID: info.MalAnimeID,
		}
		s.enrichAnimeWithReleasedOn(&anime)
		if anime.ReleasedOn != nil && *anime.ReleasedOn != "" {
			relSuccess++
		}
	}
	s.logger.Info("Priority sync: released_on completed", "success", relSuccess, "skipped", relSkip)

	s.logger.Info("Priority anime data sync completed")
	return nil
}
