package scheduler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"

	"github.com/stretchr/testify/assert"
)

// newTestLogger creates a logger that discards output for testing
func newTestLogger() *common.Logger {
	handler := slog.NewTextHandler(io.Discard, nil)
	return &common.Logger{
		Logger: slog.New(handler),
	}
}

// newTestAnnictClient creates a minimal AnnictClient for testing
func newTestAnnictClient() *services.AnnictClient {
	return services.NewAnnictClient("")
}

// mockAnimeService is a mock implementation of AnimeService for testing
type mockAnimeService struct{}

func (m *mockAnimeService) SyncGenresForAllAnimes() (int, int, []string, error) {
	return 0, 0, nil, nil
}

func (m *mockAnimeService) SyncCoverImagesForAllAnimes() (int, int, []string, error) {
	return 0, 0, nil, nil
}

func (m *mockAnimeService) SyncReleasedOnForCurrentSeason(ctx context.Context, currentSeason, previousSeason string) (int, int, error) {
	return 0, 0, nil
}

func (m *mockAnimeService) SyncAllAnimeStatuses(ctx context.Context) (int, int, error) {
	return 0, 0, nil
}

func (m *mockAnimeService) SyncTMDBMetadata(ctx context.Context, limit int) (int, int, int, error) {
	return 0, 0, 0, nil
}

func (m *mockAnimeService) IsTMDBEnabled() bool {
	return false
}

func (m *mockAnimeService) ClearAnimeCache() {
	// No-op for testing
}

func newTestAnimeService() AnimeService {
	return &mockAnimeService{}
}

func TestGetSeasonForTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"January - Winter", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"March - Winter", time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"April - Spring", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"June - Spring", time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"July - Summer", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"September - Summer", time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"October - Autumn", time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), "2025-autumn"},
		{"December - Autumn", time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "2025-autumn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSeasonForTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPreviousSeasonForTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"January - Previous Autumn", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2024-autumn"},
		{"April - Previous Winter", time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"July - Previous Spring", time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"October - Previous Summer", time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"Year boundary", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "2024-autumn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPreviousSeasonForTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewAnimeSyncScheduler(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	animeService := newTestAnimeService()
	scheduler, err := NewAnimeSyncScheduler(client, animeService, nil, logger, true, 3, 0)

	assert.NoError(t, err)
	assert.NotNil(t, scheduler)
	assert.True(t, scheduler.enabled)
	assert.NotNil(t, scheduler.stopChan)
	assert.NotNil(t, scheduler.logger)
	assert.NotNil(t, scheduler.annictClient)
	assert.NotNil(t, scheduler.animeService)
	assert.NotNil(t, scheduler.statusTracker)
	assert.Equal(t, 3, scheduler.runHour)
	assert.Equal(t, 0, scheduler.runMinute)
}

func TestSchedulerDisabled(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	animeService := newTestAnimeService()
	scheduler, err := NewAnimeSyncScheduler(client, animeService, nil, logger, false, 3, 0)

	assert.NoError(t, err)
	assert.NotNil(t, scheduler)
	assert.False(t, scheduler.enabled)
	assert.NotNil(t, scheduler.annictClient)

	// Start should not panic and should log that it's disabled
	scheduler.Start()

	// Stop should not panic
	scheduler.Stop()
}

func TestGetStatus(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	animeService := newTestAnimeService()
	scheduler, err := NewAnimeSyncScheduler(client, animeService, nil, logger, true, 3, 0)

	assert.NoError(t, err)
	now := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
	status := scheduler.GetStatus()

	assert.True(t, status.Enabled)
	assert.Equal(t, "2025-autumn", getSeasonForTime(now))
	assert.Equal(t, "2025-summer", getPreviousSeasonForTime(now))
}

func TestSyncStatusTracker_DayReset(t *testing.T) {
	tracker := NewSyncStatusTracker(true)

	// Set last reset to yesterday
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tracker.mu.Lock()
	tracker.lastResetDate = yesterday
	tracker.totalSyncsToday = 5
	tracker.mu.Unlock()

	// Start sync today - should reset because it's a new day
	nextSync := now.Add(24 * time.Hour)
	tracker.StartSync(nextSync)

	tracker.mu.RLock()
	count := tracker.totalSyncsToday
	tracker.mu.RUnlock()

	assert.Equal(t, 0, count, "Counter should reset on new day")

	// End sync - should be 1
	tracker.EndSync(true, 10, time.Second, nil)

	tracker.mu.RLock()
	count = tracker.totalSyncsToday
	tracker.mu.RUnlock()

	assert.Equal(t, 1, count, "Counter should be 1 after first sync of new day")
}

func TestSyncStatusTracker_SameDayNoReset(t *testing.T) {
	tracker := NewSyncStatusTracker(true)

	// Set last reset to today
	now := time.Now()
	tracker.mu.Lock()
	tracker.lastResetDate = now
	tracker.totalSyncsToday = 5
	tracker.mu.Unlock()

	// Start sync today - should NOT reset
	nextSync := now.Add(2 * time.Hour)
	tracker.StartSync(nextSync)

	tracker.mu.RLock()
	count := tracker.totalSyncsToday
	tracker.mu.RUnlock()

	assert.Equal(t, 5, count, "Counter should not reset on same day")

	// End sync - should increment to 6
	tracker.EndSync(true, 10, time.Second, nil)

	tracker.mu.RLock()
	count = tracker.totalSyncsToday
	tracker.mu.RUnlock()

	assert.Equal(t, 6, count, "Counter should increment")
}

func TestSyncStatusTracker_StatusFields(t *testing.T) {
	tracker := NewSyncStatusTracker(true)

	// Start a sync
	nextSync := time.Now().Add(24 * time.Hour)
	tracker.StartSync(nextSync)

	tracker.mu.RLock()
	isRunning := tracker.isRunning
	tracker.mu.RUnlock()

	assert.True(t, isRunning, "Should be running after StartSync")

	// End the sync
	errors := []string{"test error"}
	tracker.EndSync(false, 42, 5*time.Second, errors)

	tracker.mu.RLock()
	isRunning = tracker.isRunning
	success := tracker.lastSyncSuccess
	count := tracker.lastSyncCount
	duration := tracker.lastSyncDuration
	syncErrors := tracker.lastSyncErrors
	tracker.mu.RUnlock()

	assert.False(t, isRunning, "Should not be running after EndSync")
	assert.False(t, success, "Should record failure")
	assert.Equal(t, 42, count, "Should record count")
	assert.Equal(t, 5*time.Second, duration, "Should record duration")
	assert.Equal(t, errors, syncErrors, "Should record errors")
}

func TestSchedulerWithNilLogger(t *testing.T) {
	client := newTestAnnictClient()
	animeService := newTestAnimeService()
	// enabled=false でテストして、実際の同期が開始されないようにする
	scheduler, err := NewAnimeSyncScheduler(client, animeService, nil, nil, false, 3, 0)

	assert.NoError(t, err)
	// Should not panic - logger is initialized in constructor
	assert.NotNil(t, scheduler.logger)

	// Start and Stop should not panic (enabled=false なので同期は開始されない)
	scheduler.Start()
	scheduler.Stop()
}

func TestSchedulerCustomRunTime(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	animeService := newTestAnimeService()
	scheduler, err := NewAnimeSyncScheduler(client, animeService, nil, logger, true, 14, 30)

	assert.NoError(t, err)
	assert.Equal(t, 14, scheduler.runHour)
	assert.Equal(t, 30, scheduler.runMinute)
}

func TestNewAnimeSyncSchedulerWithNilAnimeService(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	scheduler, err := NewAnimeSyncScheduler(client, nil, nil, logger, true, 3, 0)

	assert.Error(t, err)
	assert.Nil(t, scheduler)
	assert.Contains(t, err.Error(), "animeService cannot be nil")
}

// Mutex and concurrency tests

// mockAnimeServiceWithDelay is a mock implementation with configurable delay for mutex testing
type mockAnimeServiceWithDelay struct {
	syncDelay time.Duration
}

func (m *mockAnimeServiceWithDelay) SyncGenresForAllAnimes() (int, int, []string, error) {
	if m.syncDelay > 0 {
		time.Sleep(m.syncDelay)
	}
	return 10, 0, nil, nil
}

func (m *mockAnimeServiceWithDelay) SyncCoverImagesForAllAnimes() (int, int, []string, error) {
	if m.syncDelay > 0 {
		time.Sleep(m.syncDelay)
	}
	return 10, 0, nil, nil
}

func (m *mockAnimeServiceWithDelay) SyncReleasedOnForCurrentSeason(ctx context.Context, currentSeason, previousSeason string) (int, int, error) {
	return 0, 0, nil
}

func (m *mockAnimeServiceWithDelay) SyncAllAnimeStatuses(ctx context.Context) (int, int, error) {
	return 0, 0, nil
}

func (m *mockAnimeServiceWithDelay) SyncTMDBMetadata(ctx context.Context, limit int) (int, int, int, error) {
	return 0, 0, 0, nil
}

func (m *mockAnimeServiceWithDelay) IsTMDBEnabled() bool {
	return false
}

func (m *mockAnimeServiceWithDelay) ClearAnimeCache() {
	// No-op for testing
}

// TestTryAcquireSyncLock tests the sync lock acquisition mechanism
func TestTryAcquireSyncLock(t *testing.T) {
	logger := newTestLogger()
	mockService := &mockAnimeServiceWithDelay{}

	scheduler, err := NewAnimeSyncScheduler(nil, mockService, nil, logger, false, 0, 0)
	assert.NoError(t, err)

	// First acquisition should succeed
	assert.True(t, scheduler.tryAcquireSyncLock(), "First tryAcquireSyncLock should succeed")

	// Second acquisition should fail (lock already held)
	assert.False(t, scheduler.tryAcquireSyncLock(), "Second tryAcquireSyncLock should fail when lock is already held")

	// Release lock
	scheduler.releaseSyncLock()

	// Third acquisition should succeed after release
	assert.True(t, scheduler.tryAcquireSyncLock(), "tryAcquireSyncLock should succeed after lock is released")

	scheduler.releaseSyncLock()
}

// TestConcurrentSyncPrevention tests that concurrent syncs are prevented
func TestConcurrentSyncPrevention(t *testing.T) {
	logger := newTestLogger()
	mockService := &mockAnimeServiceWithDelay{
		syncDelay: 10 * time.Millisecond,
	}

	scheduler, err := NewAnimeSyncScheduler(nil, mockService, nil, logger, false, 0, 0)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	syncAttempts := 5
	successfulAcquisitions := 0
	failedAcquisitions := 0
	var mu sync.Mutex

	// Try to run multiple syncs concurrently
	for i := 0; i < syncAttempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if scheduler.tryAcquireSyncLock() {
				mu.Lock()
				successfulAcquisitions++
				mu.Unlock()

				// Simulate sync work
				time.Sleep(20 * time.Millisecond)

				scheduler.releaseSyncLock()
			} else {
				mu.Lock()
				failedAcquisitions++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// At least one sync should have succeeded
	assert.Greater(t, successfulAcquisitions, 0, "Expected at least one successful sync acquisition")

	// Some syncs should have been prevented (failed to acquire lock)
	assert.Greater(t, failedAcquisitions, 0, "Expected some sync attempts to be prevented by the lock")

	// Total should equal attempts
	assert.Equal(t, syncAttempts, successfulAcquisitions+failedAcquisitions,
		"Total attempts should equal successful + failed")

	t.Logf("Sync attempts: %d successful, %d prevented", successfulAcquisitions, failedAcquisitions)
}

// TestReleaseSyncLockIdempotency tests that releasing lock multiple times is safe
func TestReleaseSyncLockIdempotency(t *testing.T) {
	logger := newTestLogger()
	mockService := &mockAnimeServiceWithDelay{}

	scheduler, err := NewAnimeSyncScheduler(nil, mockService, nil, logger, false, 0, 0)
	assert.NoError(t, err)

	// Acquire lock
	assert.True(t, scheduler.tryAcquireSyncLock(), "Failed to acquire lock")

	// Release lock multiple times (should not panic)
	scheduler.releaseSyncLock()
	scheduler.releaseSyncLock()
	scheduler.releaseSyncLock()

	// Should be able to acquire lock again
	assert.True(t, scheduler.tryAcquireSyncLock(), "Should be able to acquire lock after multiple releases")

	scheduler.releaseSyncLock()
}

// TestSyncLockWithDefer tests that defer properly releases the lock
func TestSyncLockWithDefer(t *testing.T) {
	logger := newTestLogger()
	mockService := &mockAnimeServiceWithDelay{}

	scheduler, err := NewAnimeSyncScheduler(nil, mockService, nil, logger, false, 0, 0)
	assert.NoError(t, err)

	// Simulate a function that uses defer to release lock
	func() {
		assert.True(t, scheduler.tryAcquireSyncLock(), "Failed to acquire lock")
		defer scheduler.releaseSyncLock()

		// Simulate some work
		time.Sleep(10 * time.Millisecond)
	}()

	// Lock should be released after function returns
	assert.True(t, scheduler.tryAcquireSyncLock(), "Lock should be released after defer")

	scheduler.releaseSyncLock()
}

// TestGenerateAllSeasons tests the generateAllSeasons function
func TestGenerateAllSeasons(t *testing.T) {
	tests := []struct {
		name      string
		startYear int
		endYear   int
		endSeason string
		expected  []string
	}{
		{
			name:      "Single year - spring to autumn",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2024-autumn",
			expected:  []string{"2024-spring", "2024-summer", "2024-autumn"},
		},
		{
			name:      "Single year - spring to summer",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2024-summer",
			expected:  []string{"2024-spring", "2024-summer"},
		},
		{
			name:      "Single year - spring only",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2024-spring",
			expected:  []string{"2024-spring"},
		},
		{
			name:      "Edge case - endSeason is winter of startYear (before effective start)",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2024-winter",
			expected:  nil,
		},
		{
			name:      "Cross year - 2024 spring to 2025 winter",
			startYear: 2024,
			endYear:   2025,
			endSeason: "2025-winter",
			expected:  []string{"2024-spring", "2024-summer", "2024-autumn", "2025-winter"},
		},
		{
			name:      "Multiple years - 1979 to 1980 autumn",
			startYear: 1979,
			endYear:   1980,
			endSeason: "1980-autumn",
			expected: []string{
				"1979-spring", "1979-summer", "1979-autumn",
				"1980-winter", "1980-spring", "1980-summer", "1980-autumn",
			},
		},
		{
			name:      "Edge case - endSeason before startYear",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2023-autumn",
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateAllSeasons(tt.startYear, tt.endYear, tt.endSeason)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsSeasonBefore tests the isSeasonBefore helper function
func TestIsSeasonBefore(t *testing.T) {
	tests := []struct {
		name     string
		season1  string
		season2  string
		expected bool
	}{
		{"Same season", "2024-spring", "2024-spring", false},
		{"Earlier year", "2023-autumn", "2024-spring", true},
		{"Later year", "2025-winter", "2024-autumn", false},
		{"Same year - winter before spring", "2024-winter", "2024-spring", true},
		{"Same year - spring before summer", "2024-spring", "2024-summer", true},
		{"Same year - summer before autumn", "2024-summer", "2024-autumn", true},
		{"Same year - autumn after summer", "2024-autumn", "2024-summer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSeasonBefore(tt.season1, tt.season2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCatastrophicMetadataFailure(t *testing.T) {
	tests := []struct {
		name      string
		failures  int
		attempted int
		expected  bool
	}{
		{"試行数ゼロ", 0, 0, false},
		{"最小サンプル未満は率判定しない", 5, 5, false},
		{"最小サンプル未満(全失敗でも)", 9, 9, false},
		{"十分なサンプルで閾値超過(全失敗)", 50, 50, true},
		{"十分なサンプルで閾値超過(過半数)", 6, 10, true},
		{"十分なサンプルだが閾値ちょうど(50%)は致命ではない", 5, 10, false},
		{"十分なサンプルで散発的(20%)", 4, 20, false},
		{"大規模で少数失敗", 3, 8000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCatastrophicMetadataFailure(tt.failures, tt.attempted)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncMetadataForAnimes_ContextCancelled(t *testing.T) {
	// ctx が既にキャンセル済みなら、外部API/DBに触れずに即中断する
	// （nil クライアントでも panic しないことで確認）。
	s := &AnimeSyncScheduler{logger: newTestLogger()}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 事前にキャンセル

	malID := "53065"
	animes := []models.Anime{{ID: 1, MalAnimeID: &malID}}

	failCount, successCount, samples := s.syncMetadataForAnimes(ctx, animes, "test")

	assert.Equal(t, 0, failCount, "中断時は失敗扱いにしない")
	assert.Equal(t, 0, successCount)
	assert.Empty(t, samples)
}

func TestSyncMetadataIncrementally_ContextCancelled(t *testing.T) {
	// ctx が既にキャンセル済みなら、最初のDBクエリ前に早期returnし、
	// nil クライアントに触れない（panic しない）ことで確認する。
	s := &AnimeSyncScheduler{logger: newTestLogger()}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := s.syncMetadataIncrementally(ctx)

	assert.Equal(t, 0, result.syncedCount)
	assert.Equal(t, 0, result.recoverableFailures)
	assert.Empty(t, result.opErrors)
}

// tmdbStageMock は runTMDBMetadataStage 検証用のモック。
// mockAnimeService を埋め込み、TMDB関連メソッドのみ上書きする。
type tmdbStageMock struct {
	mockAnimeService
	enabled   bool
	err       error
	calls     int
	lastLimit int
}

func (m *tmdbStageMock) IsTMDBEnabled() bool { return m.enabled }

func (m *tmdbStageMock) SyncTMDBMetadata(ctx context.Context, limit int) (int, int, int, error) {
	m.calls++
	m.lastLimit = limit
	if m.err != nil {
		return 0, 0, 0, m.err
	}
	return 1, 0, 0, nil
}

func TestRunTMDBMetadataStage(t *testing.T) {
	t.Run("TMDB有効なら制限付きで呼び出される", func(t *testing.T) {
		mock := &tmdbStageMock{enabled: true}
		s := &AnimeSyncScheduler{logger: newTestLogger(), animeService: mock}
		errs := s.runTMDBMetadataStage(context.Background(), nil)
		assert.Equal(t, 1, mock.calls)
		assert.Equal(t, constants.TMDBMetadataBatchLimit, mock.lastLimit)
		assert.Empty(t, errs)
	})

	t.Run("TMDB無効ならスキップ", func(t *testing.T) {
		mock := &tmdbStageMock{enabled: false}
		s := &AnimeSyncScheduler{logger: newTestLogger(), animeService: mock}
		errs := s.runTMDBMetadataStage(context.Background(), nil)
		assert.Equal(t, 0, mock.calls)
		assert.Empty(t, errs)
	})

	t.Run("ctxキャンセル済みならスキップ", func(t *testing.T) {
		mock := &tmdbStageMock{enabled: true}
		s := &AnimeSyncScheduler{logger: newTestLogger(), animeService: mock}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		errs := s.runTMDBMetadataStage(ctx, nil)
		assert.Equal(t, 0, mock.calls)
		assert.Empty(t, errs)
	})

	t.Run("operation-levelエラーはsyncErrorsに追加", func(t *testing.T) {
		mock := &tmdbStageMock{enabled: true, err: fmt.Errorf("db down")}
		s := &AnimeSyncScheduler{logger: newTestLogger(), animeService: mock}
		errs := s.runTMDBMetadataStage(context.Background(), nil)
		assert.Equal(t, 1, mock.calls)
		assert.Len(t, errs, 1)
	})
}
