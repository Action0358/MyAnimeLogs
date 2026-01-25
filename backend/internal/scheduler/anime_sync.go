package scheduler

import (
	"fmt"
	"sync"
	"time"

	"myanimelogs/internal/common"
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
	CreateEpisodesForAllAnimes() (int, int, error)
	EnrichEpisodesWithTitles(animeID *int) (int, int, []string, error)
	SyncGenresForAllAnimes() (int, int, []string, error)
	SyncEpisodeDetailsForAllAnimes() (int, int, []string, error)
	SyncCoverImagesForAllAnimes() (int, int, []string, error)
}

// AnimeSyncScheduler handles scheduled anime synchronization
type AnimeSyncScheduler struct {
	annictClient   *services.AnnictClient
	animeService   AnimeService
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
	s.logger.Info("Starting initial full sync of all seasons")
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

	// Sync all seasons
	for _, season := range seasons {
		count, err := s.syncSeason(season)
		totalAnimeCount += count
		if err != nil {
			errors = append(errors, fmt.Sprintf("Season %s: %s", season, err.Error()))
		}
	}

	// Sync episode data (creation and title enrichment only during initial sync)
	// Skip heavy operations (genres and episode details) to reduce startup time
	errors = append(errors, s.syncEpisodeData(true)...)

	duration := time.Since(startTime)
	success := len(errors) == 0

	// Update status tracker
	s.statusTracker.EndSync(success, totalAnimeCount, duration, errors)

	// Log results
	if !success {
		s.logger.Error("Initial full sync completed with errors",
			"duration", duration.String(),
			"total_seasons", len(seasons),
			"total_animes", totalAnimeCount,
			"error_count", len(errors),
			"errors", errors)
	} else {
		s.logger.Info("Initial full sync completed successfully",
			"duration", duration.String(),
			"total_seasons", len(seasons),
			"total_animes", totalAnimeCount,
			"note", "Genre and episode details sync will run in the first scheduled sync")
	}
}

// syncCurrentSeasonAnimes syncs animes for the current and previous season
func (s *AnimeSyncScheduler) syncCurrentSeasonAnimes() {
	s.logger.Info("Starting scheduled anime sync")
	startTime := time.Now()

	// Calculate next sync time (24 hours from start time)
	nextSyncTime := startTime.Add(24 * time.Hour)
	s.statusTracker.StartSync(nextSyncTime)

	// Use startTime for all time-based calculations to ensure consistency
	currentSeason := getSeasonForTime(startTime)
	previousSeason := getPreviousSeasonForTime(startTime)

	// Step 1: Sync anime works for current and previous seasons
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

	// Sync episode data (full sync including genres and episode details)
	errors = append(errors, s.syncEpisodeData(false)...)

	duration := time.Since(startTime)
	success := len(errors) == 0

	// Update status tracker
	s.statusTracker.EndSync(success, totalAnimeCount, duration, errors)

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
		s.logger.Info("Scheduled anime sync completed successfully",
			"duration", duration.String(),
			"current_season", currentSeason,
			"current_count", currentCount,
			"previous_season", previousSeason,
			"previous_count", previousCount,
			"total_count", totalAnimeCount)
	}
}

// syncEpisodeData creates episodes and enriches them with titles from Annict API
// Returns a slice of error messages if any errors occur
// If isInitialSync is true, skips heavy operations (genres and episode details)
func (s *AnimeSyncScheduler) syncEpisodeData(isInitialSync bool) []string {
	var errors []string

	// Step 1: Create episodes for animes that don't have episodes yet
	s.logger.Info("Creating episodes for animes without episodes")
	episodeSuccessCount, episodeFailureCount, episodeErr := s.animeService.CreateEpisodesForAllAnimes()

	if episodeErr != nil {
		errors = append(errors, fmt.Sprintf("Episode creation: %s", episodeErr.Error()))
	} else {
		s.logger.Info("Episode creation completed",
			"success_count", episodeSuccessCount,
			"failure_count", episodeFailureCount)
	}

	// Step 2: Enrich episode titles from Annict API
	s.logger.Info("Enriching episode titles from Annict API")
	titleSuccessCount, titleFailureCount, titleErrors, titleErr := s.animeService.EnrichEpisodesWithTitles(nil)

	if titleErr != nil {
		errors = append(errors, fmt.Sprintf("Episode title enrichment: %s", titleErr.Error()))
	} else {
		s.logger.Info("Episode title enrichment completed",
			"success_count", titleSuccessCount,
			"failure_count", titleFailureCount)

		// Add individual title errors to the error list (limit to avoid log spam)
		errors = appendErrorsWithLimit(errors, titleErrors, 10, "title")
	}

	// Skip heavy operations during initial sync to reduce startup time
	if isInitialSync {
		s.logger.Info("Skipping genre and episode details sync during initial sync (will run in background)")
		return errors
	}

	// Try to acquire sync lock
	if !s.tryAcquireSyncLock() {
		s.logger.Warn("Skipping genre and episode details sync: another sync is already in progress")
		return errors
	}
	defer s.releaseSyncLock()

	// Perform genre and episode details synchronization
	heavySyncErrors := s.syncGenresAndEpisodeDetails()
	errors = append(errors, heavySyncErrors...)

	return errors
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
