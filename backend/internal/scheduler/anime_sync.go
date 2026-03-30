package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
)

// seasonOrder defines the order of seasons in a year
var seasonOrder = []string{"winter", "spring", "summer", "autumn"}

// backgroundSyncDelay is the delay before starting background sync after initial sync
const backgroundSyncDelay = 30 * time.Second

// syncStartYear is the year from which anime synchronization begins
// 1979 is chosen as it marks the start of Mobile Suit Gundam, a landmark in anime history
const syncStartYear = 1979

// syncTimeout is the maximum duration for the entire sync operation
const syncTimeout = 6 * time.Hour

// episodeSyncTimeout is the maximum duration for episode sync operations
const episodeSyncTimeout = 3 * time.Hour

// initialSyncTimeout is the maximum duration for the initial full sync operation
const initialSyncTimeout = 12 * time.Hour

// appendErrorsWithLimit appends errors to the error list with a maximum limit
// to prevent log spam. Returns the updated error list.
func appendErrorsWithLimit(errors []string, newErrors []string, maxErrors int, errorType string) []string {
	if len(newErrors) == 0 {
		return errors
	}

	if len(newErrors) <= maxErrors {
		return append(errors, newErrors...)
	}

	// Add first maxErrors items
	errors = append(errors, newErrors[:maxErrors]...)
	// Add summary message
	errors = append(errors, fmt.Sprintf("... and %d more %s errors", len(newErrors)-maxErrors, errorType))
	return errors
}

// syncOperation represents a synchronization operation that returns success count, failure count, errors, and error
type syncOperation func() (int, int, []string, error)

// executeSyncOperation executes a sync operation with consistent logging and error handling
// Returns updated error list
func (s *AnimeSyncScheduler) executeSyncOperation(errors []string, operation syncOperation, operationName string) []string {
	s.logger.Info(fmt.Sprintf("Syncing %s", operationName))
	successCount, failureCount, opErrors, err := operation()

	if err != nil {
		errors = append(errors, fmt.Sprintf("%s sync: %s", operationName, err.Error()))
	} else {
		s.logger.Info(fmt.Sprintf("%s sync completed", operationName),
			"success_count", successCount,
			"failure_count", failureCount)

		// Add individual errors to the error list (limit to avoid log spam)
		errors = appendErrorsWithLimit(errors, opErrors, 10, operationName)
	}

	return errors
}

// syncGenresAndEpisodeDetails performs genre, episode details, and cover image synchronization
// Returns error messages if any errors occur
func (s *AnimeSyncScheduler) syncGenresAndEpisodeDetails() []string {
	var errors []string

	// Sync genre information from Jikan API
	errors = s.executeSyncOperation(errors, s.animeService.SyncGenresForAllAnimes, "genre information from Jikan API")

	// Sync episode details (air_date and duration) from Jikan API
	errors = s.executeSyncOperation(errors, s.animeService.SyncEpisodeDetailsForAllAnimes, "episode details (air_date and duration) from Jikan API")

	// Sync cover images from AniList API
	errors = s.executeSyncOperation(errors, s.animeService.SyncCoverImagesForAllAnimes, "cover images from AniList API")

	return errors
}

// tryAcquireSyncLock attempts to acquire the sync lock
// Returns true if lock was acquired, false if sync is already in progress
func (s *AnimeSyncScheduler) tryAcquireSyncLock() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.syncInProgress {
		return false
	}
	s.syncInProgress = true
	return true
}

// releaseSyncLock releases the sync lock
func (s *AnimeSyncScheduler) releaseSyncLock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncInProgress = false
}

// runBackgroundSync runs heavy sync operations in background after initial sync
// with proper shutdown handling and mutual exclusion
func (s *AnimeSyncScheduler) runBackgroundSync() {
	s.logger.Info("Starting background sync of genres and episode details")

	// Wait for initial sync to complete, with shutdown handling
	select {
	case <-time.After(backgroundSyncDelay):
		// Continue with background sync
	case <-s.stopChan:
		s.logger.Info("Background sync cancelled due to shutdown")
		return
	}

	// Try to acquire sync lock
	if !s.tryAcquireSyncLock() {
		s.logger.Warn("Background sync skipped: another sync is already in progress")
		return
	}
	defer s.releaseSyncLock()

	startTime := time.Now()
	errors := s.syncGenresAndEpisodeDetails()
	duration := time.Since(startTime)

	// Clear cache after background sync to reflect new data
	s.animeService.ClearAnimeCache()

	// Log completion with status
	if len(errors) > 0 {
		// Calculate sample size inline to avoid dependency on min function
		sampleSize := 5
		if len(errors) < sampleSize {
			sampleSize = len(errors)
		}
		s.logger.Error("Background sync completed with errors",
			"duration", duration.String(),
			"error_count", len(errors),
			"sample_errors", errors[:sampleSize])
	} else {
		s.logger.Info("Background sync completed successfully",
			"duration", duration.String())
	}
}

// AnimeService defines the interface for anime-related operations needed by the scheduler
type AnimeService interface {
	SyncEpisodesFromAnnict(animeID *int) (int, int, []string, error)
	SyncEpisodesFromJikan(animeID *int) (int, int, []string, error)
	SyncEpisodesFromJikanWithContext(ctx context.Context, animeID *int) (int, int, []string, error)
	SyncGenresForAllAnimes() (int, int, []string, error)
	SyncEpisodeDetailsForAllAnimes() (int, int, []string, error)
	SyncCoverImagesForAllAnimes() (int, int, []string, error)
	ClearAnimeCache() // キャッシュクリア
}

// AnimeSyncScheduler handles scheduled anime synchronization
type AnimeSyncScheduler struct {
	annictClient   *services.AnnictClient
	animeService   AnimeService
	anilistClient  *services.AniListClient // AniListClientを再利用（レートリミッター共有）
	logger         *common.Logger
	mu             sync.Mutex // Protects ticker field and syncInProgress
	ticker         *time.Ticker
	stopChan       chan struct{}
	enabled        bool
	runHour        int
	runMinute      int
	statusTracker  *SyncStatusTracker
	initialSynced  bool // Tracks if initial full sync has been performed
	syncInProgress bool // Tracks if heavy sync operations are in progress
}

// NewAnimeSyncScheduler creates a new anime sync scheduler
// Returns an error if animeService is nil
func NewAnimeSyncScheduler(annictClient *services.AnnictClient, animeService AnimeService, logger *common.Logger, enabled bool, runHour, runMinute int) (*AnimeSyncScheduler, error) {
	// Ensure logger is never nil - use a default logger if not provided
	if logger == nil {
		logger = common.NewLogger(common.LogLevelError, "text")
	}

	// AnimeService is a required dependency
	if animeService == nil {
		return nil, fmt.Errorf("animeService cannot be nil")
	}

	return &AnimeSyncScheduler{
		annictClient:  annictClient,
		animeService:  animeService,
		logger:        logger,
		stopChan:      make(chan struct{}),
		enabled:       enabled,
		runHour:       runHour,
		runMinute:     runMinute,
		statusTracker: NewSyncStatusTracker(enabled),
		initialSynced: false,
	}, nil
}

// GetStatus returns the current sync status
func (s *AnimeSyncScheduler) GetStatus() models.SyncStatus {
	return s.statusTracker.GetStatus(time.Now())
}

// GetStatusInterface returns the status as interface{} for handler compatibility
func (s *AnimeSyncScheduler) GetStatusInterface() interface{} {
	return s.statusTracker.GetStatus(time.Now())
}

// Start starts the scheduler
func (s *AnimeSyncScheduler) Start() {
	if !s.enabled {
		if s.logger != nil {
			s.logger.Info("Anime sync scheduler is disabled")
		}
		return
	}

	if s.logger != nil {
		s.logger.Info("Starting anime sync scheduler")
	}

	// Calculate time until next scheduled run time
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), s.runHour, s.runMinute, 0, 0, time.Local)
	if now.After(next) {
		next = next.Add(24 * time.Hour)
	}

	if s.logger != nil {
		s.logger.Info("Next anime sync scheduled",
			"next_run", next.Format("2006-01-02 15:04:05"),
			"duration_until", time.Until(next).String(),
			"run_hour", s.runHour,
			"run_minute", s.runMinute)
	}

	// Set initial next sync time
	s.statusTracker.SetNextSyncTime(next)

	// Wait until first execution
	go func() {
		// Perform initial full sync immediately on first startup
		if !s.initialSynced {
			s.logger.Info("Performing initial full sync of all seasons")
			s.syncAllSeasons(next)
			s.initialSynced = true

			// Run heavy operations in background after initial sync
			go s.runBackgroundSync()
		}

		// Use interruptible sleep to avoid resource leak
		timer := time.NewTimer(time.Until(next))
		defer timer.Stop()

		select {
		case <-timer.C:
			// First scheduled execution (incremental sync)
			s.syncCurrentSeasonAnimes()
		case <-s.stopChan:
			if s.logger != nil {
				s.logger.Info("Anime sync scheduler stopped before first execution")
			}
			return
		}

		// Set up ticker for daily execution (use local variable)
		ticker := time.NewTicker(24 * time.Hour)

		// Store ticker with mutex protection
		s.mu.Lock()
		s.ticker = ticker
		s.mu.Unlock()

		for {
			select {
			case <-ticker.C:
				s.syncCurrentSeasonAnimes()
			case <-s.stopChan:
				if s.logger != nil {
					s.logger.Info("Anime sync scheduler stopped")
				}
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *AnimeSyncScheduler) Stop() {
	// Access ticker with mutex protection
	s.mu.Lock()
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.mu.Unlock()

	close(s.stopChan)
	s.logger.Info("Stopping anime sync scheduler")
}

// syncAllSeasons syncs all seasons from syncStartYear-spring to current season
// nextSyncTime is the time of the next scheduled sync (not 24 hours from now)
func (s *AnimeSyncScheduler) syncAllSeasons(nextSyncTime time.Time) {
	// 初回フル同期にタイムアウト制御を適用（12時間）
	ctx, cancel := context.WithTimeout(context.Background(), initialSyncTimeout)
	defer cancel()

	s.logger.Info("Starting initial full sync of all seasons (differential mode)")
	startTime := time.Now()

	// Use the provided next sync time (the first scheduled sync)
	s.statusTracker.StartSync(nextSyncTime)

	// Use startTime for all time-based calculations to ensure consistency
	currentYear := startTime.Year()
	currentSeason := getSeasonForTime(startTime)

	// Generate all seasons from syncStartYear-spring to current
	seasons := generateAllSeasons(syncStartYear, currentYear, currentSeason)

	totalAnimeCount := 0
	var errors []string

	// 全シーズンを同期（差分判定により既存データはスキップ）
	for _, season := range seasons {
		count, err := s.syncSeason(season)
		totalAnimeCount += count
		if err != nil {
			errors = append(errors, fmt.Sprintf("Season %s: %s", season, err.Error()))
		}
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "season sync", startTime, totalAnimeCount, errors) {
		return
	}

	// エピソード同期: syncEpisodeDataを使用（Annict→Jikan順序 + タイムアウト）
	episodeFailureCount, episodeErrors := s.syncEpisodeData(true)
	errors = append(errors, episodeErrors...)

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "episode sync", startTime, totalAnimeCount, errors) {
		return
	}

	// 初回同期時はメタデータも同期（優先度ベース）
	s.logger.Info("Starting initial metadata sync")
	metadataErrors, metadataSyncedCount := s.syncMetadataIncrementally()
	errors = append(errors, metadataErrors...)

	duration := time.Since(startTime)
	success := len(errors) == 0

	// Update status tracker
	s.statusTracker.EndSync(success, totalAnimeCount, duration, errors)

	// Record sync stats
	s.recordSyncStats("full", startTime, totalAnimeCount, 0, 0, episodeFailureCount, metadataSyncedCount, success, len(errors))

	// Clear cache after sync to reflect new data
	s.animeService.ClearAnimeCache()

	// Log results
	if !success {
		s.logger.Error("Initial full sync completed with errors",
			"duration", duration.String(),
			"total_seasons", len(seasons),
			"total_animes", totalAnimeCount,
			"error_count", len(errors),
			"errors", errors)
	} else {
		s.logger.Info("Initial full sync completed successfully (differential mode)",
			"duration", duration.String(),
			"total_seasons", len(seasons),
			"total_animes", totalAnimeCount,
			"note", "Metadata sync completed with priority-based approach")
	}
}

// checkSyncTimeout checks if the context has been cancelled and handles timeout logging/status update.
// Returns true if timed out (caller should return), false if OK to continue.
func (s *AnimeSyncScheduler) checkSyncTimeout(ctx context.Context, stepName string, startTime time.Time, totalAnimeCount int, errors []string) bool {
	select {
	case <-ctx.Done():
		s.logger.Warn("Sync timed out",
			"step", stepName,
			"elapsed", time.Since(startTime).String())
		s.statusTracker.EndSync(false, totalAnimeCount, time.Since(startTime), append(errors, "sync timed out after "+stepName))
		return true
	default:
		return false
	}
}

// syncCurrentSeasonAnimes syncs animes for the current and previous season
func (s *AnimeSyncScheduler) syncCurrentSeasonAnimes() {
	// 処理全体のタイムアウト: 6時間
	ctx, cancel := context.WithTimeout(context.Background(), syncTimeout)
	defer cancel()

	s.logger.Info("Starting scheduled anime sync (differential mode with timeout)")
	startTime := time.Now()

	// Calculate next sync time (24 hours from start time)
	nextSyncTime := startTime.Add(24 * time.Hour)
	s.statusTracker.StartSync(nextSyncTime)

	// Use startTime for all time-based calculations to ensure consistency
	currentSeason := getSeasonForTime(startTime)
	previousSeason := getPreviousSeasonForTime(startTime)

	// Step 1: 基本情報の差分同期（Annict）
	// needsSyncメソッドにより変更があるもののみUPDATE
	currentCount, currentErr := s.syncSeason(currentSeason)
	previousCount, previousErr := s.syncSeason(previousSeason)

	totalAnimeCount := currentCount + previousCount
	var errors []string

	if currentErr != nil {
		errors = append(errors, fmt.Sprintf("Current season (%s): %s", currentSeason, currentErr.Error()))
	}
	if previousErr != nil {
		errors = append(errors, fmt.Sprintf("Previous season (%s): %s", previousSeason, previousErr.Error()))
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "Step 1", startTime, totalAnimeCount, errors) {
		return
	}

	// Step 2: エピソード作成 - Annict（主軸）→ Jikan（補完）
	// Step 2a: Annictからエピソード作成（主軸・エピソードレコード作成）
	s.logger.Info("Syncing episodes from Annict API (primary)")
	annictEpSuccess, annictEpFailure, annictEpErrors, annictEpErr := s.animeService.SyncEpisodesFromAnnict(nil)
	if annictEpErr != nil {
		errors = append(errors, fmt.Sprintf("Episode sync from Annict: %s", annictEpErr.Error()))
	} else {
		s.logger.Info("Episode sync from Annict completed",
			"success_count", annictEpSuccess,
			"failure_count", annictEpFailure)
		errors = appendErrorsWithLimit(errors, annictEpErrors, 10, "annict-episode")
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "Step 2a", startTime, totalAnimeCount, errors) {
		return
	}

	// Step 2b: Jikanからエピソード補完（duration/air_date/filler/recap追加）
	s.logger.Info("Syncing episodes from Jikan API (supplemental, with context)")
	episodeSuccessCount, episodeFailureCount, episodeErrors, episodeErr := s.animeService.SyncEpisodesFromJikanWithContext(ctx, nil)

	if episodeErr != nil {
		errors = append(errors, fmt.Sprintf("Episode sync from Jikan: %s", episodeErr.Error()))
	} else {
		s.logger.Info("Episode sync from Jikan completed",
			"success_count", episodeSuccessCount,
			"failure_count", episodeFailureCount)

		// Add individual episode errors to the error list (limit to avoid log spam)
		errors = appendErrorsWithLimit(errors, episodeErrors, 10, "episode")
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "Step 2", startTime, totalAnimeCount, errors) {
		return
	}

	// Step 3: メタデータの段階的同期（優先度ベース）
	// 全作品ではなく、優先度の高いもののみ同期
	s.logger.Info("Starting incremental metadata sync")
	metadataErrors, metadataSyncedCount := s.syncMetadataIncrementally()
	errors = append(errors, metadataErrors...)

	duration := time.Since(startTime)
	success := len(errors) == 0

	// Update status tracker
	s.statusTracker.EndSync(success, totalAnimeCount, duration, errors)

	// Record sync stats
	s.recordSyncStats("incremental", startTime, totalAnimeCount, 0, 0, episodeFailureCount, metadataSyncedCount, success, len(errors))

	// Clear cache after sync to reflect new data
	s.animeService.ClearAnimeCache()

	// Log results
	if !success {
		s.logger.Error("Scheduled anime sync completed with errors",
			"duration", duration.String(),
			"current_season", currentSeason,
			"current_count", currentCount,
			"current_error", currentErr,
			"previous_season", previousSeason,
			"previous_count", previousCount,
			"previous_error", previousErr,
			"error_count", len(errors))
	} else {
		s.logger.Info("Scheduled anime sync completed successfully (differential mode)",
			"duration", duration.String(),
			"current_season", currentSeason,
			"current_count", currentCount,
			"previous_season", previousSeason,
			"previous_count", previousCount,
			"total_count", totalAnimeCount)
	}
}

// syncEpisodeData syncs episodes from Jikan API
// Returns a slice of error messages if any errors occur
// If isInitialSync is true, skips heavy operations (genres and episode details)
func (s *AnimeSyncScheduler) syncEpisodeData(isInitialSync bool) (int, []string) {
	var errors []string
	totalFailureCount := 0

	// エピソード同期にタイムアウト付きコンテキストを使用（3時間）
	ctx, cancel := context.WithTimeout(context.Background(), episodeSyncTimeout)
	defer cancel()

	// Step 1: Annictからエピソード作成（主軸）
	s.logger.Info("Syncing episodes from Annict API (primary)")
	annictSuccess, annictFailure, annictErrors, annictErr := s.animeService.SyncEpisodesFromAnnict(nil)
	if annictErr != nil {
		errors = append(errors, fmt.Sprintf("Episode sync from Annict: %s", annictErr.Error()))
	} else {
		s.logger.Info("Episode sync from Annict completed",
			"success_count", annictSuccess,
			"failure_count", annictFailure)
		totalFailureCount += annictFailure
		errors = appendErrorsWithLimit(errors, annictErrors, 10, "annict-episode")
	}

	// Step 2: Jikanからエピソード補完（duration/air_date/filler/recap）
	s.logger.Info("Syncing episodes from Jikan API (supplemental, with context)")
	episodeSuccessCount, episodeFailureCount, episodeErrors, episodeErr := s.animeService.SyncEpisodesFromJikanWithContext(ctx, nil)

	if episodeErr != nil {
		errors = append(errors, fmt.Sprintf("Episode sync from Jikan: %s", episodeErr.Error()))
	} else {
		s.logger.Info("Episode sync from Jikan completed",
			"success_count", episodeSuccessCount,
			"failure_count", episodeFailureCount)
		totalFailureCount += episodeFailureCount

		// Add individual episode errors to the error list (limit to avoid log spam)
		errors = appendErrorsWithLimit(errors, episodeErrors, 10, "episode")
	}

	// Skip heavy operations during initial sync to reduce startup time
	if isInitialSync {
		s.logger.Info("Skipping genre and episode details sync during initial sync (will run in background)")
		return totalFailureCount, errors
	}

	// Try to acquire sync lock
	if !s.tryAcquireSyncLock() {
		s.logger.Warn("Skipping genre and episode details sync: another sync is already in progress")
		return totalFailureCount, errors
	}
	defer s.releaseSyncLock()

	// Perform genre and episode details synchronization
	heavySyncErrors := s.syncGenresAndEpisodeDetails()
	errors = append(errors, heavySyncErrors...)

	return totalFailureCount, errors
}

// syncSeason syncs animes for a specific season
func (s *AnimeSyncScheduler) syncSeason(season string) (int, error) {
	s.logger.Info("Syncing season", "season", season)

	animes, err := s.annictClient.FetchAndSyncWorksBySeason(season)
	if err != nil {
		s.logger.Error("Failed to sync season",
			"season", season,
			"error", err)
		return 0, err
	}

	s.logger.Info("Season sync completed",
		"season", season,
		"count", len(animes))

	return len(animes), nil
}

// getSeasonForTime returns the season string for a given time (e.g., "2025-autumn")
// This is a pure function that takes time as input for easier testing
func getSeasonForTime(t time.Time) string {
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

// getPreviousSeasonForTime returns the previous season string for a given time
// This is a pure function that takes time as input for easier testing
func getPreviousSeasonForTime(t time.Time) string {
	year := t.Year()
	month := t.Month()

	var season string
	switch {
	case month >= 1 && month <= 3:
		season = "autumn"
		year--
	case month >= 4 && month <= 6:
		season = "winter"
	case month >= 7 && month <= 9:
		season = "spring"
	default:
		season = "summer"
	}

	return fmt.Sprintf("%d-%s", year, season)
}

// generateAllSeasons generates all season strings from startYear-spring to endSeason
// For startYear=1979, it starts from 1979-spring (skipping 1979-winter)
// Returns empty slice if endSeason is before the effective start season (startYear-spring)
func generateAllSeasons(startYear, endYear int, endSeason string) []string {
	var result []string

	// Effective start season is spring of startYear (winter is skipped)
	effectiveStartSeason := fmt.Sprintf("%d-spring", startYear)

	// Check if endSeason is before effective start season
	// This handles cases like generateAllSeasons(2024, 2024, "2024-winter")
	if isSeasonBefore(endSeason, effectiveStartSeason) {
		return result
	}

	for year := startYear; year <= endYear; year++ {
		for _, season := range seasonOrder {
			// Skip winter of the start year (e.g., 1979-winter)
			if year == startYear && season == "winter" {
				continue
			}

			seasonStr := fmt.Sprintf("%d-%s", year, season)
			result = append(result, seasonStr)

			// Stop if we've reached the end season
			if seasonStr == endSeason {
				return result
			}
		}
	}

	return result
}

// isSeasonBefore returns true if season1 is chronologically before season2
func isSeasonBefore(season1, season2 string) bool {
	var year1, year2 int
	var seasonName1, seasonName2 string

	fmt.Sscanf(season1, "%d-%s", &year1, &seasonName1)
	fmt.Sscanf(season2, "%d-%s", &year2, &seasonName2)

	if year1 != year2 {
		return year1 < year2
	}

	// Compare season order within the same year
	seasonIndex := map[string]int{"winter": 0, "spring": 1, "summer": 2, "autumn": 3}
	return seasonIndex[seasonName1] < seasonIndex[seasonName2]
}

// getAnimesByQuery - 共通のアニメ取得ヘルパー関数
func (s *AnimeSyncScheduler) getAnimesByQuery(query string, args ...interface{}) ([]models.Anime, error) {
	var animes []models.Anime

	rows, err := s.annictClient.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var anime models.Anime
		if err := rows.Scan(&anime.ID, &anime.MalAnimeID, &anime.Title, &anime.AnnictID); err != nil {
			s.logger.Warn("Failed to scan anime", "error", err)
			continue
		}
		animes = append(animes, anime)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating anime rows: %w", err)
	}

	return animes, nil
}

// getAnimesWithoutMetadata - メタデータ未取得作品を取得
func (s *AnimeSyncScheduler) getAnimesWithoutMetadata() ([]models.Anime, error) {
	return s.getAnimesByQuery(`
		SELECT id, mal_anime_id, title, annict_id
		FROM animes
		WHERE metadata_synced_at IS NULL
		AND mal_anime_id IS NOT NULL
		AND sync_status != $1
		ORDER BY created_at DESC
		LIMIT 100
	`, string(constants.SyncStatusFailed))
}

// getAiringAnimes - 放送中作品を取得
func (s *AnimeSyncScheduler) getAiringAnimes() ([]models.Anime, error) {
	return s.getAnimesByQuery(`
		SELECT id, mal_anime_id, title, annict_id
		FROM animes
		WHERE status = 'currently_airing'
		AND (metadata_synced_at IS NULL OR metadata_synced_at < NOW() - INTERVAL '7 days')
		AND mal_anime_id IS NOT NULL
		ORDER BY metadata_synced_at ASC NULLS FIRST
		LIMIT 50
	`)
}

// getAnimesWithStaleMetadata - 古いメタデータを持つ作品を取得
func (s *AnimeSyncScheduler) getAnimesWithStaleMetadata(staleDuration time.Duration) ([]models.Anime, error) {
	staleThreshold := time.Now().Add(-staleDuration)
	return s.getAnimesByQuery(`
		SELECT id, mal_anime_id, title, annict_id
		FROM animes
		WHERE metadata_synced_at < $1
		AND mal_anime_id IS NOT NULL
		AND sync_status != $2
		ORDER BY metadata_synced_at ASC
		LIMIT 50
	`, staleThreshold, string(constants.SyncStatusFailed))
}

// updateMetadataSyncTime - メタデータ同期日時を更新
func (s *AnimeSyncScheduler) updateMetadataSyncTime(animeID int) error {
	_, err := s.annictClient.DB.Exec(`
		UPDATE animes
		SET metadata_synced_at = CURRENT_TIMESTAMP,
			sync_status = $1,
			sync_error = NULL
		WHERE id = $2
	`, string(constants.SyncStatusSynced), animeID)
	return err
}

// updateMetadataSyncFailure - メタデータ同期失敗時のステータス更新
func (s *AnimeSyncScheduler) updateMetadataSyncFailure(animeID int, syncErr error) error {
	_, err := s.annictClient.DB.Exec(`
		UPDATE animes
		SET sync_status = $1,
			sync_error = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, string(constants.SyncStatusFailed), syncErr.Error(), animeID)
	return err
}

// syncMetadataIncrementally - 優先度ベースでメタデータを段階的に同期
// Returns: error messages, metadata synced count
func (s *AnimeSyncScheduler) syncMetadataIncrementally() ([]string, int) {
	var errors []string
	metadataSyncedCount := 0

	// 優先度1: 新規作品（メタデータ未取得）
	newAnimes, err := s.getAnimesWithoutMetadata()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to get new animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for new animes", "count", len(newAnimes))
		syncErrors, syncCount := s.syncMetadataForAnimes(newAnimes, "new")
		errors = append(errors, syncErrors...)
		metadataSyncedCount += syncCount
	}

	// 優先度2: 放送中作品（週次更新）
	airingAnimes, err := s.getAiringAnimes()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to get airing animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for airing animes", "count", len(airingAnimes))
		syncErrors, syncCount := s.syncMetadataForAnimes(airingAnimes, "airing")
		errors = append(errors, syncErrors...)
		metadataSyncedCount += syncCount
	}

	// 優先度3: 古いメタデータ（月次更新）
	staleAnimes, err := s.getAnimesWithStaleMetadata(constants.StaleDataThreshold)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to get stale animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for stale animes", "count", len(staleAnimes))
		syncErrors, syncCount := s.syncMetadataForAnimes(staleAnimes, "stale")
		errors = append(errors, syncErrors...)
		metadataSyncedCount += syncCount
	}

	return errors, metadataSyncedCount
}

// syncMetadataForAnimes - アニメリストのメタデータを同期
// Returns: error messages, success count
func (s *AnimeSyncScheduler) syncMetadataForAnimes(animes []models.Anime, category string) ([]string, int) {
	var errors []string
	successCount := 0
	failureCount := 0

	s.logger.Info(fmt.Sprintf("Starting metadata sync for %s animes", category), "count", len(animes))

	for _, anime := range animes {
		if anime.MalAnimeID == nil || *anime.MalAnimeID == "" {
			continue
		}

		malID := *anime.MalAnimeID

		// ジャンル同期（AniList API: 90req/min = 1.5req/sec）
		// 構造体フィールドのanilistClientを再利用（レートリミッター共有）
		ctx := context.Background()
		genres, genreErr := s.anilistClient.GetGenresByMALId(ctx, malID)
		if genreErr != nil {
			s.logger.Warn("Failed to get genres",
				"anime_id", anime.ID,
				"mal_id", malID,
				"error", genreErr)
			failureCount++
			errors = append(errors, fmt.Sprintf("Anime %d: genre sync failed - %s", anime.ID, genreErr))
			// 失敗時はsync_statusをfailedに更新
			if err := s.updateMetadataSyncFailure(anime.ID, genreErr); err != nil {
				s.logger.Error("Failed to update metadata sync failure", "anime_id", anime.ID, "error", err)
			}
			continue
		}

		// ジャンル保存
		if saveErr := s.saveGenres(anime.ID, genres); saveErr != nil {
			s.logger.Error("Failed to save genres", "anime_id", anime.ID, "error", saveErr)
			failureCount++
			// 失敗時はsync_statusをfailedに更新
			if err := s.updateMetadataSyncFailure(anime.ID, saveErr); err != nil {
				s.logger.Error("Failed to update metadata sync failure", "anime_id", anime.ID, "error", err)
			}
			continue
		}

		// ジャンル保存成功時のみメタデータ同期日時を更新
		successCount++
		if err := s.updateMetadataSyncTime(anime.ID); err != nil {
			s.logger.Error("Failed to update metadata sync time", "anime_id", anime.ID, "error", err)
		}
	}

	s.logger.Info(fmt.Sprintf("Metadata sync completed for %s animes", category),
		"success", successCount,
		"failure", failureCount)

	return errors, successCount
}

// saveGenres - ジャンルをDBに保存（トランザクション対応）
func (s *AnimeSyncScheduler) saveGenres(animeID int, genres []string) error {
	// トランザクション開始
	tx, err := s.annictClient.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 既存のジャンルを削除
	_, err = tx.Exec("DELETE FROM anime_genres WHERE anime_id = $1", animeID)
	if err != nil {
		return fmt.Errorf("failed to delete existing genres: %w", err)
	}

	// 新しいジャンルを挿入
	for _, genre := range genres {
		_, err := tx.Exec(`
			INSERT INTO anime_genres (anime_id, genre)
			VALUES ($1, $2)
		`, animeID, genre)
		if err != nil {
			return fmt.Errorf("failed to insert genre %s: %w", genre, err)
		}
	}

	// コミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// recordSyncStats - 同期統計をDBに記録
func (s *AnimeSyncScheduler) recordSyncStats(syncType string, startTime time.Time, totalAnimes, updatedAnimes, skippedAnimes, failedAnimes, metadataSynced int, success bool, errorCount int) {
	duration := int(time.Since(startTime).Seconds())

	_, err := s.annictClient.DB.Exec(`
		INSERT INTO sync_stats (
			sync_type, started_at, completed_at, duration,
			total_animes, updated_animes, skipped_animes, failed_animes,
			metadata_synced, success, error_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, syncType, startTime, time.Now(), duration,
		totalAnimes, updatedAnimes, skippedAnimes, failedAnimes,
		metadataSynced, success, errorCount)

	if err != nil {
		s.logger.Error("Failed to record sync stats", "error", err)
	} else {
		s.logger.Debug("Sync stats recorded",
			"sync_type", syncType,
			"duration", duration,
			"total_animes", totalAnimes,
			"success", success)
	}
}
