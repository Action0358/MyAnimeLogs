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

// initialSyncTimeout is the maximum duration for the initial full sync operation
const initialSyncTimeout = 12 * time.Hour

// appendErrorsWithLimit appends errors to the error list with a maximum limit
// to prevent log spam. Returns the updated error list.
func appendErrorsWithLimit(syncErrors []string, newErrors []string, maxErrors int, errorType string) []string {
	if len(newErrors) == 0 {
		return syncErrors
	}

	if len(newErrors) <= maxErrors {
		return append(syncErrors, newErrors...)
	}

	// Add first maxErrors items
	syncErrors = append(syncErrors, newErrors[:maxErrors]...)
	// Add summary message
	syncErrors = append(syncErrors, fmt.Sprintf("... and %d more %s errors", len(newErrors)-maxErrors, errorType))
	return syncErrors
}

// syncOperation represents a synchronization operation that returns success count, failure count, errors, and error
type syncOperation func() (int, int, []string, error)

// executeSyncOperation executes a sync operation with consistent logging and error handling
// Returns updated error list
func (s *AnimeSyncScheduler) executeSyncOperation(syncErrors []string, operation syncOperation, operationName string) []string {
	s.logger.Info(fmt.Sprintf("Syncing %s", operationName))
	successCount, failureCount, opErrors, err := operation()

	if err != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("%s sync: %s", operationName, err.Error()))
	} else {
		s.logger.Info(fmt.Sprintf("%s sync completed", operationName),
			"success_count", successCount,
			"failure_count", failureCount)

		// Add individual errors to the error list (limit to avoid log spam)
		syncErrors = appendErrorsWithLimit(syncErrors, opErrors, 10, operationName)
	}

	return syncErrors
}

// syncGenresAndEpisodeDetails performs genre and cover image synchronization
// Returns error messages if any errors occur
func (s *AnimeSyncScheduler) syncGenresAndEpisodeDetails() []string {
	var syncErrors []string

	// Sync genre information from Jikan API
	syncErrors = s.executeSyncOperation(syncErrors, s.animeService.SyncGenresForAllAnimes, "genre information from Jikan API")

	// Sync cover images from AniList API
	syncErrors = s.executeSyncOperation(syncErrors, s.animeService.SyncCoverImagesForAllAnimes, "cover images from AniList API")

	return syncErrors
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
	syncErrors := s.syncGenresAndEpisodeDetails()
	duration := time.Since(startTime)

	// Clear cache after background sync to reflect new data
	s.animeService.ClearAnimeCache()

	// Log completion with status
	if len(syncErrors) > 0 {
		// Calculate sample size inline to avoid dependency on min function
		sampleSize := 5
		if len(syncErrors) < sampleSize {
			sampleSize = len(syncErrors)
		}
		s.logger.Error("Background sync completed with errors",
			"duration", duration.String(),
			"error_count", len(syncErrors),
			"sample_errors", syncErrors[:sampleSize])
	} else {
		s.logger.Info("Background sync completed successfully",
			"duration", duration.String())
	}
}

// AnimeService defines the interface for anime-related operations needed by the scheduler
type AnimeService interface {
	SyncGenresForAllAnimes() (int, int, []string, error)
	SyncCoverImagesForAllAnimes() (int, int, []string, error)
	SyncReleasedOnForCurrentSeason(ctx context.Context, currentSeason, previousSeason string) (int, int, error)
	SyncAllAnimeStatuses(ctx context.Context) (int, int, error)
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
	var syncErrors []string

	// 全シーズンを同期（差分判定により既存データはスキップ）
	for _, season := range seasons {
		count, err := s.syncSeason(season)
		totalAnimeCount += count
		if err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("Season %s: %s", season, err.Error()))
		}
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "season sync", startTime, totalAnimeCount, syncErrors) {
		return
	}

	// 今期・前期作品のreleased_on同期（Jikan API）
	previousSeason := getPreviousSeasonForTime(startTime)
	s.logger.Info("Syncing released_on for current/previous season animes (initial sync)")
	releasedOnSuccess, releasedOnFailure, releasedOnErr := s.animeService.SyncReleasedOnForCurrentSeason(ctx, currentSeason, previousSeason)
	if releasedOnErr != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Released_on sync: %s", releasedOnErr.Error()))
	} else {
		s.logger.Info("Released_on sync completed (initial sync)",
			"success_count", releasedOnSuccess,
			"failure_count", releasedOnFailure)
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "released_on sync", startTime, totalAnimeCount, syncErrors) {
		return
	}

	// アニメステータス同期（season + released_on ベース）
	s.logger.Info("Syncing anime statuses (initial sync)")
	statusSuccess, statusFailure, statusErr := s.animeService.SyncAllAnimeStatuses(ctx)
	if statusErr != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Status sync: %s", statusErr.Error()))
	} else {
		s.logger.Info("Anime status sync completed (initial sync)",
			"success_count", statusSuccess,
			"failure_count", statusFailure)
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "status sync", startTime, totalAnimeCount, syncErrors) {
		return
	}

	// 初回同期時はメタデータも同期（優先度ベース）
	s.logger.Info("Starting initial metadata sync")
	metadataErrors, metadataSyncedCount := s.syncMetadataIncrementally()
	syncErrors = append(syncErrors, metadataErrors...)

	duration := time.Since(startTime)
	success := len(syncErrors) == 0

	// Update status tracker
	s.statusTracker.EndSync(success, totalAnimeCount, duration, syncErrors)

	// Record sync stats
	s.recordSyncStats("full", startTime, totalAnimeCount, 0, 0, 0, metadataSyncedCount, success, len(syncErrors))

	// Clear cache after sync to reflect new data
	s.animeService.ClearAnimeCache()

	// Log results
	if !success {
		s.logger.Error("Initial full sync completed with errors",
			"duration", duration.String(),
			"total_seasons", len(seasons),
			"total_animes", totalAnimeCount,
			"error_count", len(syncErrors),
			"errors", syncErrors)
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
func (s *AnimeSyncScheduler) checkSyncTimeout(ctx context.Context, stepName string, startTime time.Time, totalAnimeCount int, syncErrors []string) bool {
	select {
	case <-ctx.Done():
		s.logger.Warn("Sync timed out",
			"step", stepName,
			"elapsed", time.Since(startTime).String())
		s.statusTracker.EndSync(false, totalAnimeCount, time.Since(startTime), append(syncErrors, "sync timed out after "+stepName))
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
	var syncErrors []string

	if currentErr != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Current season (%s): %s", currentSeason, currentErr.Error()))
	}
	if previousErr != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Previous season (%s): %s", previousSeason, previousErr.Error()))
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "Step 1", startTime, totalAnimeCount, syncErrors) {
		return
	}

	// Step 1.5: 今期・前期作品のreleased_on同期（Jikan API）
	// released_onがNULLの作品に対してJikanから放送開始日を取得し、ステータスを再判定
	s.logger.Info("Syncing released_on for current/previous season animes")
	releasedOnSuccess, releasedOnFailure, releasedOnErr := s.animeService.SyncReleasedOnForCurrentSeason(ctx, currentSeason, previousSeason)
	if releasedOnErr != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Released_on sync: %s", releasedOnErr.Error()))
	} else {
		s.logger.Info("Released_on sync completed",
			"success_count", releasedOnSuccess,
			"failure_count", releasedOnFailure)
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "Step 1.5", startTime, totalAnimeCount, syncErrors) {
		return
	}

	// Step 1.6: アニメステータス同期（season + released_on ベース）
	// 過去シーズンの作品を finished に、released_on が未来の作品を upcoming に更新
	s.logger.Info("Syncing anime statuses")
	statusSuccess, statusFailure, statusErr := s.animeService.SyncAllAnimeStatuses(ctx)
	if statusErr != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Status sync: %s", statusErr.Error()))
	} else {
		s.logger.Info("Anime status sync completed",
			"success_count", statusSuccess,
			"failure_count", statusFailure)
	}

	// タイムアウト確認
	if s.checkSyncTimeout(ctx, "Step 1.6", startTime, totalAnimeCount, syncErrors) {
		return
	}

	// Step 2: メタデータの段階的同期（優先度ベース）
	// 全作品ではなく、優先度の高いもののみ同期
	s.logger.Info("Starting incremental metadata sync")
	metadataErrors, metadataSyncedCount := s.syncMetadataIncrementally()
	syncErrors = append(syncErrors, metadataErrors...)

	duration := time.Since(startTime)
	success := len(syncErrors) == 0

	// Update status tracker
	s.statusTracker.EndSync(success, totalAnimeCount, duration, syncErrors)

	// Record sync stats
	s.recordSyncStats("incremental", startTime, totalAnimeCount, 0, 0, 0, metadataSyncedCount, success, len(syncErrors))

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
			"error_count", len(syncErrors))
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
		WHERE status = $1
		AND (metadata_synced_at IS NULL OR metadata_synced_at < NOW() - INTERVAL '7 days')
		AND mal_anime_id IS NOT NULL
		ORDER BY metadata_synced_at ASC NULLS FIRST
		LIMIT 50
	`, constants.AnimeStatusAiring.String())
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
	var syncErrors []string
	metadataSyncedCount := 0

	// 優先度1: 新規作品（メタデータ未取得）
	newAnimes, err := s.getAnimesWithoutMetadata()
	if err != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Failed to get new animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for new animes", "count", len(newAnimes))
		newErrs, syncCount := s.syncMetadataForAnimes(newAnimes, "new")
		syncErrors = append(syncErrors, newErrs...)
		metadataSyncedCount += syncCount
	}

	// 優先度2: 放送中作品（週次更新）
	airingAnimes, err := s.getAiringAnimes()
	if err != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Failed to get airing animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for airing animes", "count", len(airingAnimes))
		airingErrs, syncCount := s.syncMetadataForAnimes(airingAnimes, "airing")
		syncErrors = append(syncErrors, airingErrs...)
		metadataSyncedCount += syncCount
	}

	// 優先度3: 古いメタデータ（月次更新）
	staleAnimes, err := s.getAnimesWithStaleMetadata(constants.StaleDataThreshold)
	if err != nil {
		syncErrors = append(syncErrors, fmt.Sprintf("Failed to get stale animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for stale animes", "count", len(staleAnimes))
		staleErrs, syncCount := s.syncMetadataForAnimes(staleAnimes, "stale")
		syncErrors = append(syncErrors, staleErrs...)
		metadataSyncedCount += syncCount
	}

	return syncErrors, metadataSyncedCount
}

// syncMetadataForAnimes - アニメリストのメタデータを同期
// Returns: error messages, success count
func (s *AnimeSyncScheduler) syncMetadataForAnimes(animes []models.Anime, category string) ([]string, int) {
	var syncErrors []string
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
			syncErrors = append(syncErrors, fmt.Sprintf("Anime %d: genre sync failed - %s", anime.ID, genreErr))
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

	return syncErrors, successCount
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
