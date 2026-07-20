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
	"myanimelogs/internal/utils"
)

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

// metadataFailureCatastrophicRatio - 個別メタデータ同期の失敗率がこれを超え、かつ
// 十分な試行数があるとき、全体障害（外部API障害/設定ミス等）とみなしバッチを
// success=false にする。少数の散発的な失敗（レート制限など回復可能）では落とさない。
const metadataFailureCatastrophicRatio = 0.5

// metadataFailureMinSample - 失敗率による致命判定を適用する最小試行数
// これ未満の試行ではサンプルが小さすぎるため、率では判定しない。
const metadataFailureMinSample = 10

// isCatastrophicMetadataFailure - 個別メタデータ同期の失敗が全体障害レベルかを判定する
func isCatastrophicMetadataFailure(failures, attempted int) bool {
	if attempted < metadataFailureMinSample {
		return false
	}
	return float64(failures)/float64(attempted) > metadataFailureCatastrophicRatio
}

// logMetadataRecoverableFailures - 個別メタデータ同期の失敗を可視化のためログ出力する。
// catastrophic（全体障害レベル）なら Error、散発的なら Warn とし、いずれも失敗件数・
// 試行数・サンプルを記録する。失敗が無ければ何もしない。
func (s *AnimeSyncScheduler) logMetadataRecoverableFailures(metadata metadataSyncResult, attempted int, catastrophic bool) {
	if metadata.recoverableFailures == 0 {
		return
	}
	logFn := s.logger.Warn
	if catastrophic {
		logFn = s.logger.Error
	}
	logFn("Metadata sync had recoverable per-item failures",
		"failures", metadata.recoverableFailures,
		"attempted", attempted,
		"synced", metadata.syncedCount,
		"catastrophic", catastrophic,
		"sample_errors", metadata.recoverableSamples)
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

// syncGenresAndCoverImages performs genre and cover image synchronization
// Returns error messages if any errors occur
func (s *AnimeSyncScheduler) syncGenresAndCoverImages() []string {
	var syncErrors []string

	// Sync genre information from AniList API
	syncErrors = s.executeSyncOperation(syncErrors, s.animeService.SyncGenresForAllAnimes, "genre information from AniList API")

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
	syncErrors := s.syncGenresAndCoverImages()
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
	// SyncTMDBMetadata resolves tmdb_id and Japanese descriptions from TMDB for up to
	// `limit` animes that need them (limit <= 0 means unlimited). Returns success,
	// not-found, failure counts.
	SyncTMDBMetadata(ctx context.Context, limit int) (int, int, int, error)
	// IsTMDBEnabled reports whether the TMDB client is configured/usable.
	IsTMDBEnabled() bool
	ClearAnimeCache() // キャッシュクリア
}

// AnimeSyncScheduler handles scheduled anime synchronization
type AnimeSyncScheduler struct {
	annictClient   *services.AnnictClient
	animeService   AnimeService
	anilistClient  *services.AniListClient // AnimeServiceと共有するAniListClientインスタンス（レートリミッター共有）
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

// NewAnimeSyncScheduler creates a new anime sync scheduler.
// anilistClient must be the same instance used by AnimeService to share the rate limiter.
// Returns an error if animeService is nil.
func NewAnimeSyncScheduler(annictClient *services.AnnictClient, animeService AnimeService, anilistClient *services.AniListClient, logger *common.Logger, enabled bool, runHour, runMinute int) (*AnimeSyncScheduler, error) {
	// Ensure logger is never nil - use a default logger if not provided
	if logger == nil {
		logger = common.NewLogger(common.LogLevelError, "text")
	}

	// AnimeService is a required dependency
	if animeService == nil {
		return nil, fmt.Errorf("animeService cannot be nil")
	}

	// Fall back to a new client only if none was provided (e.g., in tests)
	if anilistClient == nil {
		anilistClient = services.NewAniListClient()
	}

	return &AnimeSyncScheduler{
		annictClient:  annictClient,
		anilistClient: anilistClient,
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
	// 初回フル同期にタイムアウト制御を適用（12時間）。シャットダウン時も即中断する。
	ctx, cancel := s.newSyncContext(initialSyncTimeout)
	defer cancel()

	s.logger.Info("Starting initial full sync of all seasons (differential mode)")
	startTime := time.Now()

	// Use the provided next sync time (the first scheduled sync)
	s.statusTracker.StartSync(nextSyncTime)

	// Use startTime for all time-based calculations to ensure consistency
	currentYear := startTime.Year()
	currentSeason := utils.SeasonForTime(startTime)

	// Generate all seasons from syncStartYear-spring to current
	seasons := utils.GenerateAllSeasons(syncStartYear, currentYear, currentSeason)

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
	previousSeason := utils.PreviousSeasonForTime(startTime)
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
	metadata := s.syncMetadataIncrementally(ctx)
	syncErrors = append(syncErrors, metadata.opErrors...)
	metadataSyncedCount := metadata.syncedCount

	// TMDBメタデータ解決（tmdb_id・日本語あらすじ）
	syncErrors = s.runTMDBMetadataStage(ctx, syncErrors)

	duration := time.Since(startTime)
	// 成功判定: operation-level の致命エラーが無く、かつ個別失敗が全体障害レベルでない場合のみ成功。
	// レート制限等による散発的な個別ジャンル失敗（再試行可能）ではバッチを失敗扱いにしない。
	metadataAttempted := metadata.syncedCount + metadata.recoverableFailures
	catastrophic := isCatastrophicMetadataFailure(metadata.recoverableFailures, metadataAttempted)
	success := len(syncErrors) == 0 && !catastrophic
	s.logMetadataRecoverableFailures(metadata, metadataAttempted, catastrophic)

	// Update status tracker
	s.statusTracker.EndSync(success, totalAnimeCount, duration, syncErrors)

	// Record sync stats（個別失敗は failed_animes として計上）
	s.recordSyncStats("full", startTime, totalAnimeCount, 0, 0, metadata.recoverableFailures, metadataSyncedCount, success, len(syncErrors)+metadata.recoverableFailures)

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

// runTMDBMetadataStage resolves TMDB-derived metadata (tmdb_id / Japanese description)
// for a bounded number of animes per run. Skipped when TMDB is not configured or ctx is
// already cancelled. Only operation-level errors are appended to syncErrors (per-item
// not-found/failures are recoverable and retried in later runs).
func (s *AnimeSyncScheduler) runTMDBMetadataStage(ctx context.Context, syncErrors []string) []string {
	if ctx.Err() != nil || !s.animeService.IsTMDBEnabled() {
		return syncErrors
	}

	s.logger.Info("Starting TMDB metadata sync")
	success, notFound, failure, err := s.animeService.SyncTMDBMetadata(ctx, constants.TMDBMetadataBatchLimit)
	if err != nil {
		return append(syncErrors, fmt.Sprintf("TMDB metadata sync: %s", err))
	}
	s.logger.Info("TMDB metadata sync completed",
		"success", success,
		"not_found", notFound,
		"failure", failure)
	return syncErrors
}

// newSyncContext creates a timeout context that is ALSO cancelled when the scheduler
// receives a shutdown signal (stopChan). This ensures long-running phases (e.g. metadata
// sync, which makes many rate-limited AniList calls) stop promptly on graceful shutdown
// instead of running until the timeout. The watcher goroutine exits on either signal.
func (s *AnimeSyncScheduler) newSyncContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	go func() {
		select {
		case <-s.stopChan:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
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
	// 処理全体のタイムアウト: 6時間。シャットダウン時も即中断する。
	ctx, cancel := s.newSyncContext(syncTimeout)
	defer cancel()

	s.logger.Info("Starting scheduled anime sync (differential mode with timeout)")
	startTime := time.Now()

	// Calculate next sync time (24 hours from start time)
	nextSyncTime := startTime.Add(24 * time.Hour)
	s.statusTracker.StartSync(nextSyncTime)

	// Use startTime for all time-based calculations to ensure consistency
	currentSeason := utils.SeasonForTime(startTime)
	previousSeason := utils.PreviousSeasonForTime(startTime)

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
	metadata := s.syncMetadataIncrementally(ctx)
	syncErrors = append(syncErrors, metadata.opErrors...)
	metadataSyncedCount := metadata.syncedCount

	// Step 3: TMDBメタデータ解決（tmdb_id・日本語あらすじ）
	syncErrors = s.runTMDBMetadataStage(ctx, syncErrors)

	duration := time.Since(startTime)
	// 成功判定: operation-level の致命エラーが無く、かつ個別失敗が全体障害レベルでない場合のみ成功。
	metadataAttempted := metadata.syncedCount + metadata.recoverableFailures
	catastrophic := isCatastrophicMetadataFailure(metadata.recoverableFailures, metadataAttempted)
	success := len(syncErrors) == 0 && !catastrophic
	s.logMetadataRecoverableFailures(metadata, metadataAttempted, catastrophic)

	// Update status tracker
	s.statusTracker.EndSync(success, totalAnimeCount, duration, syncErrors)

	// Record sync stats（個別失敗は failed_animes として計上）
	s.recordSyncStats("incremental", startTime, totalAnimeCount, 0, 0, metadata.recoverableFailures, metadataSyncedCount, success, len(syncErrors)+metadata.recoverableFailures)

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

// getAnimesWithoutMetadata - メタデータ未取得作品を取得
func (s *AnimeSyncScheduler) getAnimesWithoutMetadata() ([]models.Anime, error) {
	return services.QueryAnimeSyncTargets(s.annictClient.DB, s.logger, "animes without metadata", `
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
	return services.QueryAnimeSyncTargets(s.annictClient.DB, s.logger, "airing animes", `
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
	return services.QueryAnimeSyncTargets(s.annictClient.DB, s.logger, "animes with stale metadata", `
		SELECT id, mal_anime_id, title, annict_id
		FROM animes
		WHERE metadata_synced_at < $1
		AND mal_anime_id IS NOT NULL
		AND sync_status != $2
		ORDER BY metadata_synced_at ASC
		LIMIT 50
	`, staleThreshold, string(constants.SyncStatusFailed))
}

// getFailedMetadataAnimes - 同期失敗(sync_status='failed')作品を再試行対象として取得
// 通常の取得クエリは failed を除外するため、レート制限で失敗した作品が永続的に滞留する。
// クールダウン(updated_at が cooldown より前)を満たす作品のみを対象とし、
// 失敗直後の連続再試行を避けつつ滞留を解消する。
func (s *AnimeSyncScheduler) getFailedMetadataAnimes(cooldown time.Duration) ([]models.Anime, error) {
	cooldownThreshold := time.Now().Add(-cooldown)
	return services.QueryAnimeSyncTargets(s.annictClient.DB, s.logger, "failed metadata animes", `
		SELECT id, mal_anime_id, title, annict_id
		FROM animes
		WHERE sync_status = $1
		AND mal_anime_id IS NOT NULL
		AND updated_at < $2
		ORDER BY updated_at ASC
		LIMIT $3
	`, string(constants.SyncStatusFailed), cooldownThreshold, constants.MetadataRetryBatchLimit)
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

// metadataSyncResult - メタデータ段階同期の集計結果
// opErrors（取得クエリ等の operation-level 致命エラー）と、個別ジャンル同期の
// recoverable な失敗（recoverableFailures）を分離して保持する。
type metadataSyncResult struct {
	opErrors            []string // operation-level（致命）エラー。バッチ成功判定を落とす
	syncedCount         int      // ジャンル同期に成功した件数
	recoverableFailures int      // 個別ジャンル同期の失敗件数（再試行可能）
	recoverableSamples  []string // 個別失敗のサンプル（ログ用、上限あり）
}

// syncMetadataIncrementally - 優先度ベースでメタデータを段階的に同期
// ctx はタイムアウト/シャットダウンの伝播に使用する。
func (s *AnimeSyncScheduler) syncMetadataIncrementally(ctx context.Context) metadataSyncResult {
	var result metadataSyncResult

	addItemResult := func(failCount, syncCount int, samples []string) {
		result.syncedCount += syncCount
		result.recoverableFailures += failCount
		for _, e := range samples {
			if len(result.recoverableSamples) >= metadataSampleErrorLimit {
				break
			}
			result.recoverableSamples = append(result.recoverableSamples, e)
		}
	}

	// 各優先度のDBクエリ前に ctx を確認し、キャンセル/タイムアウト時は早期returnする。
	// これがないと、syncMetadataForAnimes が即中断するにもかかわらず、後続優先度の
	// DBクエリと "Syncing ..." ログだけが無駄に実行されてしまう。

	// 優先度1: 新規作品（メタデータ未取得）
	if ctx.Err() != nil {
		return result
	}
	newAnimes, err := s.getAnimesWithoutMetadata()
	if err != nil {
		result.opErrors = append(result.opErrors, fmt.Sprintf("Failed to get new animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for new animes", "count", len(newAnimes))
		addItemResult(s.syncMetadataForAnimes(ctx, newAnimes, "new"))
	}

	// 優先度2: 放送中作品（週次更新）
	if ctx.Err() != nil {
		return result
	}
	airingAnimes, err := s.getAiringAnimes()
	if err != nil {
		result.opErrors = append(result.opErrors, fmt.Sprintf("Failed to get airing animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for airing animes", "count", len(airingAnimes))
		addItemResult(s.syncMetadataForAnimes(ctx, airingAnimes, "airing"))
	}

	// 優先度3: 古いメタデータ（月次更新）
	if ctx.Err() != nil {
		return result
	}
	staleAnimes, err := s.getAnimesWithStaleMetadata(constants.StaleDataThreshold)
	if err != nil {
		result.opErrors = append(result.opErrors, fmt.Sprintf("Failed to get stale animes: %s", err))
	} else {
		s.logger.Info("Syncing metadata for stale animes", "count", len(staleAnimes))
		addItemResult(s.syncMetadataForAnimes(ctx, staleAnimes, "stale"))
	}

	// 優先度4: 同期失敗作品の再試行（クールダウン経過後）
	// レート制限で failed になり通常クエリから除外され滞留した作品を解消する
	if ctx.Err() != nil {
		return result
	}
	retryAnimes, err := s.getFailedMetadataAnimes(constants.MetadataRetryCooldown)
	if err != nil {
		result.opErrors = append(result.opErrors, fmt.Sprintf("Failed to get retry-target animes: %s", err))
	} else if len(retryAnimes) > 0 {
		s.logger.Info("Retrying metadata for previously failed animes", "count", len(retryAnimes))
		addItemResult(s.syncMetadataForAnimes(ctx, retryAnimes, "retry-failed"))
	}

	return result
}

// metadataSampleErrorLimit - ログ用に保持する個別失敗エラーのサンプル上限
const metadataSampleErrorLimit = 10

// syncMetadataForAnimes - アニメリストのメタデータを同期
// 個別ジャンル同期の失敗は recoverable（次回バッチで再試行可能）として扱い、
// 失敗件数と少数のサンプルエラーを返す（バッチ全体の成功判定とは分離する）。
// ctx がキャンセル/タイムアウトされた場合は途中で中断する（graceful shutdown対応）。
// Returns: failure count, success count, sample error messages (capped)
func (s *AnimeSyncScheduler) syncMetadataForAnimes(ctx context.Context, animes []models.Anime, category string) (int, int, []string) {
	var sampleErrors []string
	successCount := 0
	failureCount := 0

	s.logger.Info(fmt.Sprintf("Starting metadata sync for %s animes", category), "count", len(animes))

	recordFailure := func(animeID int, cause error) {
		failureCount++
		if len(sampleErrors) < metadataSampleErrorLimit {
			sampleErrors = append(sampleErrors, fmt.Sprintf("Anime %d: %s", animeID, cause))
		}
		// 失敗時はsync_statusをfailedに更新
		if err := s.updateMetadataSyncFailure(animeID, cause); err != nil {
			s.logger.Error("Failed to update metadata sync failure", "anime_id", animeID, "error", err)
		}
	}

	for _, anime := range animes {
		// タイムアウト/シャットダウン時は新たな作品の処理を開始しない。
		// 未処理分は failed にせず、次回バッチで処理する。
		if ctx.Err() != nil {
			s.logger.Info("Metadata sync interrupted (context cancelled)",
				"category", category,
				"processed", successCount+failureCount)
			break
		}

		if anime.MalAnimeID == nil || *anime.MalAnimeID == "" {
			continue
		}

		malID := *anime.MalAnimeID

		// ジャンル同期（AniList API: 90req/min = 1.5req/sec）
		// 構造体フィールドのanilistClientを再利用（レートリミッター共有）
		genres, genreErr := s.anilistClient.GetGenresByMALId(ctx, malID)
		if genreErr != nil {
			// ctxキャンセル/タイムアウトが原因の失敗は failed 扱いにせず中断する
			// （回復可能。次回バッチで再試行される）
			if ctx.Err() != nil {
				s.logger.Info("Metadata sync interrupted during fetch (context cancelled)",
					"category", category,
					"processed", successCount+failureCount)
				break
			}
			s.logger.Warn("Failed to get genres",
				"anime_id", anime.ID,
				"mal_id", malID,
				"error", genreErr)
			recordFailure(anime.ID, genreErr)
			continue
		}

		// ジャンル保存（service/annict と同一の共通関数を使用。genres_synced_at も更新される）
		if saveErr := services.SaveGenresToDatabase(s.annictClient.DB, s.logger, anime.ID, genres); saveErr != nil {
			s.logger.Error("Failed to save genres", "anime_id", anime.ID, "error", saveErr)
			recordFailure(anime.ID, saveErr)
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

	return failureCount, successCount, sampleErrors
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
