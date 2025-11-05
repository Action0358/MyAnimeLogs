package scheduler

import (
	"fmt"
	"sync"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
)

// AnimeSyncScheduler handles scheduled anime synchronization
type AnimeSyncScheduler struct {
	annictClient  *services.AnnictClient
	logger        *common.Logger
	mu            sync.Mutex // Protects ticker field
	ticker        *time.Ticker
	stopChan      chan struct{}
	enabled       bool
	runHour       int
	runMinute     int
	statusTracker *SyncStatusTracker
}

// NewAnimeSyncScheduler creates a new anime sync scheduler
func NewAnimeSyncScheduler(annictClient *services.AnnictClient, logger *common.Logger, enabled bool, runHour, runMinute int) *AnimeSyncScheduler {
	// Ensure logger is never nil - use a default logger if not provided
	if logger == nil {
		logger = common.NewLogger(common.LogLevelError, "text")
	}

	return &AnimeSyncScheduler{
		annictClient:  annictClient,
		logger:        logger,
		stopChan:      make(chan struct{}),
		enabled:       enabled,
		runHour:       runHour,
		runMinute:     runMinute,
		statusTracker: NewSyncStatusTracker(enabled),
	}
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
		// Use interruptible sleep to avoid resource leak
		timer := time.NewTimer(time.Until(next))
		defer timer.Stop()

		select {
		case <-timer.C:
			// First execution
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
	if s.logger != nil {
		s.logger.Info("Stopping anime sync scheduler")
	}
}

// syncCurrentSeasonAnimes syncs animes for the current and previous season
func (s *AnimeSyncScheduler) syncCurrentSeasonAnimes() {
	s.logger.Info("Starting scheduled anime sync")
	startTime := time.Now()

	// Calculate next sync time (24 hours from now)
	nextSyncTime := time.Now().Add(24 * time.Hour)
	s.statusTracker.StartSync(nextSyncTime)

	// Get current time once for consistent season calculation
	now := time.Now()
	currentSeason := getSeasonForTime(now)
	previousSeason := getPreviousSeasonForTime(now)

	// Sync current season
	currentCount, currentErr := s.syncSeason(currentSeason)

	// Sync previous season (for ongoing shows)
	previousCount, previousErr := s.syncSeason(previousSeason)

	duration := time.Since(startTime)
	totalCount := currentCount + previousCount
	success := currentErr == nil && previousErr == nil

	// Collect errors
	var errors []string
	if currentErr != nil {
		errors = append(errors, fmt.Sprintf("Current season (%s): %s", currentSeason, currentErr.Error()))
	}
	if previousErr != nil {
		errors = append(errors, fmt.Sprintf("Previous season (%s): %s", previousSeason, previousErr.Error()))
	}

	// Update status tracker
	s.statusTracker.EndSync(success, totalCount, duration, errors)

	// Log results
	if !success {
		s.logger.Error("Scheduled anime sync completed with errors",
			"duration", duration.String(),
			"current_season", currentSeason,
			"current_count", currentCount,
			"current_error", currentErr,
			"previous_season", previousSeason,
			"previous_count", previousCount,
			"previous_error", previousErr)
	} else {
		s.logger.Info("Scheduled anime sync completed successfully",
			"duration", duration.String(),
			"current_season", currentSeason,
			"current_count", currentCount,
			"previous_season", previousSeason,
			"previous_count", previousCount,
			"total_count", totalCount)
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
