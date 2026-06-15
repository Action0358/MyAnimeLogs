package services

import (
	"testing"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestAnimeService_NewAnimeService tests the constructor
func TestAnimeService_NewAnimeService(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger)

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.anilistClient)
}

// TestAnimeService_SetSearchDefaults tests default value setting
func TestAnimeService_SetSearchDefaults(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger)

	tests := []struct {
		name     string
		input    models.AdvancedSearchParams
		expected models.AdvancedSearchParams
	}{
		{
			name:  "空のパラメータ - デフォルト値が設定される",
			input: models.AdvancedSearchParams{},
			expected: models.AdvancedSearchParams{
				Page:    1,
				PerPage: 20,
				Sort:    "title",
				Order:   "asc",
			},
		},
		{
			name: "一部のパラメータのみ指定 - 残りはデフォルト",
			input: models.AdvancedSearchParams{
				Page: 2,
				Sort: "rating",
			},
			expected: models.AdvancedSearchParams{
				Page:    2,
				PerPage: 20,
				Sort:    "rating",
				Order:   "asc",
			},
		},
		{
			name: "すべてのパラメータ指定 - 変更なし",
			input: models.AdvancedSearchParams{
				Page:    3,
				PerPage: 50,
				Sort:    "year",
				Order:   "desc",
			},
			expected: models.AdvancedSearchParams{
				Page:    3,
				PerPage: 50,
				Sort:    "year",
				Order:   "desc",
			},
		},
		{
			name: "無効なページ番号 - 1に修正",
			input: models.AdvancedSearchParams{
				Page: 0,
			},
			expected: models.AdvancedSearchParams{
				Page:    1,
				PerPage: 20,
				Sort:    "title",
				Order:   "asc",
			},
		},
		{
			name: "無効なPerPage - デフォルトに修正",
			input: models.AdvancedSearchParams{
				PerPage: 0,
			},
			expected: models.AdvancedSearchParams{
				Page:    1,
				PerPage: 20,
				Sort:    "title",
				Order:   "asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.input
			service.setSearchDefaults(&params)

			assert.Equal(t, tt.expected, params)
		})
	}
}

// TestAnimeService_BuildOrderByClause tests ORDER BY clause generation
func TestAnimeService_BuildOrderByClause(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger)

	tests := []struct {
		name     string
		params   models.AdvancedSearchParams
		expected string
	}{
		{
			name: "タイトル昇順",
			params: models.AdvancedSearchParams{
				Sort:  "title",
				Order: "asc",
			},
			expected: "ORDER BY a.title ASC",
		},
		{
			name: "タイトル降順",
			params: models.AdvancedSearchParams{
				Sort:  "title",
				Order: "desc",
			},
			expected: "ORDER BY a.title DESC",
		},
		{
			name: "評価降順",
			params: models.AdvancedSearchParams{
				Sort:  "rating",
				Order: "desc",
			},
			expected: "ORDER BY avg_ratings.avg_rating DESC NULLS LAST",
		},
		{
			name: "年降順",
			params: models.AdvancedSearchParams{
				Sort:  "year",
				Order: "desc",
			},
			expected: "ORDER BY SUBSTRING(a.season FROM 1 FOR 4) DESC, a.season DESC",
		},
		{
			name: "人気度降順",
			params: models.AdvancedSearchParams{
				Sort:  "popularity",
				Order: "desc",
			},
			expected: "ORDER BY a.watchers_count DESC",
		},
		{
			name: "更新日降順",
			params: models.AdvancedSearchParams{
				Sort:  "updated",
				Order: "desc",
			},
			expected: "ORDER BY a.updated_at DESC",
		},
		{
			name: "無効なソート - デフォルト（タイトル昇順）",
			params: models.AdvancedSearchParams{
				Sort:  "invalid",
				Order: "asc",
			},
			expected: "ORDER BY a.title ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.buildOrderByClause(tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAnimeService_BuildBaseQuery tests base query construction
func TestAnimeService_BuildBaseQuery(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger)

	tests := []struct {
		name             string
		params           models.AdvancedSearchParams
		userID           *int
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:   "認証なし - 基本クエリのみ",
			params: models.AdvancedSearchParams{},
			userID: nil,
			shouldContain: []string{
				"FROM animes_with_computed_status a",
			},
			shouldNotContain: []string{
				"LEFT JOIN anime_genres",
				"LEFT JOIN (", // avg_ratings subquery
				"INNER JOIN user_watch_status",
			},
		},
		{
			name: "ジャンルフィルター - ジャンルJOIN",
			params: models.AdvancedSearchParams{
				Genres: []string{"アクション"},
			},
			userID: nil,
			shouldContain: []string{
				"FROM animes_with_computed_status a",
				"LEFT JOIN anime_genres ag",
			},
		},
		{
			name: "評価フィルター - 評価サブクエリJOIN",
			params: models.AdvancedSearchParams{
				MinRating: 4.0,
			},
			userID: nil,
			shouldContain: []string{
				"FROM animes_with_computed_status a",
				"LEFT JOIN (",
				"SELECT anime_id, AVG(rating) as avg_rating",
				"FROM user_ratings",
				"AS avg_ratings",
			},
		},
		{
			name: "評価ソート - 評価サブクエリJOIN",
			params: models.AdvancedSearchParams{
				Sort: "rating",
			},
			userID: nil,
			shouldContain: []string{
				"LEFT JOIN (",
				"avg_ratings",
			},
		},
		{
			name: "ユーザーステータスフィルター - 視聴ステータスINNER JOIN",
			params: models.AdvancedSearchParams{
				UserStatus: "watching",
			},
			userID: func() *int { id := 123; return &id }(),
			shouldContain: []string{
				"INNER JOIN user_watch_status ws",
			},
		},
		{
			name: "複合フィルター - 複数JOIN",
			params: models.AdvancedSearchParams{
				Genres:     []string{"アクション"},
				MinRating:  4.0,
				UserStatus: "watching",
			},
			userID: func() *int { id := 123; return &id }(),
			shouldContain: []string{
				"LEFT JOIN anime_genres ag",
				"LEFT JOIN (",
				"avg_ratings",
				"INNER JOIN user_watch_status ws",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.buildBaseQuery(tt.params, tt.userID)

			for _, expected := range tt.shouldContain {
				assert.Contains(t, result, expected, "Query should contain: "+expected)
			}

			for _, notExpected := range tt.shouldNotContain {
				assert.NotContains(t, result, notExpected, "Query should not contain: "+notExpected)
			}
		})
	}
}

// TestAnimeService_BuildWhereClause tests WHERE clause construction
func TestAnimeService_BuildWhereClause(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger)

	tests := []struct {
		name          string
		params        models.AdvancedSearchParams
		userID        *int
		expectedArgs  int
		shouldContain []string
	}{
		{
			name:          "パラメータなし - WHERE句なし",
			params:        models.AdvancedSearchParams{},
			userID:        nil,
			expectedArgs:  0,
			shouldContain: []string{},
		},
		{
			name: "クエリのみ",
			params: models.AdvancedSearchParams{
				Query: "進撃",
			},
			userID:       nil,
			expectedArgs: 1, // カタカナなしの場合は1つの引数のみ
			shouldContain: []string{
				"ILIKE",
			},
		},
		{
			name: "年フィルター",
			params: models.AdvancedSearchParams{
				Year: 2024,
			},
			userID:       nil,
			expectedArgs: 1,
			shouldContain: []string{
				"a.season LIKE",
			},
		},
		{
			name: "シーズンフィルター",
			params: models.AdvancedSearchParams{
				Season: "2024-spring",
			},
			userID:       nil,
			expectedArgs: 1,
			shouldContain: []string{
				"a.season",
			},
		},
		{
			name: "シーズンフィルター（日本語形式）",
			params: models.AdvancedSearchParams{
				Season: "2024年春",
			},
			userID:       nil,
			expectedArgs: 1,
			shouldContain: []string{
				"a.season",
			},
		},
		{
			name: "ジャンルフィルター",
			params: models.AdvancedSearchParams{
				Genres: []string{"アクション", "冒険"},
			},
			userID:       nil,
			expectedArgs: 2,
			shouldContain: []string{
				"ag.genre IN",
			},
		},
		{
			name: "評価範囲フィルター（両方指定）",
			params: models.AdvancedSearchParams{
				MinRating: 3.0,
				MaxRating: 5.0,
			},
			userID:       nil,
			expectedArgs: 2,
			shouldContain: []string{
				"avg_ratings.avg_rating BETWEEN",
			},
		},
		{
			name: "評価範囲フィルター（最小のみ）",
			params: models.AdvancedSearchParams{
				MinRating: 4.0,
			},
			userID:       nil,
			expectedArgs: 1,
			shouldContain: []string{
				"avg_ratings.avg_rating >=",
			},
		},
		{
			name: "ユーザーステータスフィルター",
			params: models.AdvancedSearchParams{
				UserStatus: "watching",
			},
			userID:       func() *int { id := 123; return &id }(),
			expectedArgs: 2, // user_id + status
			shouldContain: []string{
				"ws.user_id",
				"ws.status",
			},
		},
		{
			name: "複合フィルター",
			params: models.AdvancedSearchParams{
				Query:      "アニメ",
				Year:       2024,
				Season:     "2024-spring",
				Genres:     []string{"アクション"},
				MinRating:  4.0,
				UserStatus: "watching",
			},
			userID:       func() *int { id := 123; return &id }(),
			expectedArgs: 9, // query(3: original, katakana, hiragana) + year + season + genre + minRating + user_id + userStatus
			shouldContain: []string{
				"ILIKE",
				"a.season LIKE",
				"ag.genre IN",
				"avg_ratings.avg_rating >=",
				"ws.user_id",
				"ws.status",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whereClause, args, _ := service.buildWhereClause(tt.params, tt.userID)

			assert.Equal(t, tt.expectedArgs, len(args), "Number of arguments should match")

			for _, expected := range tt.shouldContain {
				assert.Contains(t, whereClause, expected, "WHERE clause should contain: "+expected)
			}
		})
	}
}

// TestAnimeService_ErrorConstants tests error constant definitions
func TestAnimeService_ErrorConstants(t *testing.T) {
	assert.NotNil(t, ErrAnimeNotFound)
	assert.Equal(t, "anime not found", ErrAnimeNotFound.Error())
}

// TestAnimeService_ImplementsInterface verifies that AnimeService implements AnimeServiceInterface
func TestAnimeService_ImplementsInterface(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger)

	// Verify that service implements AnimeServiceInterface (compile-time check)
	var _ AnimeServiceInterface = service

	// If this compiles, the interface is correctly implemented
	assert.NotNil(t, service)
}

// TestAnimeService_JikanClientIntegration tests Jikan client integration
func TestAnimeService_JikanClientIntegration(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger)

	// Verify that JikanClient is initialized
	assert.NotNil(t, service.jikanClient)
	assert.NotNil(t, service.jikanClient.httpClient)
	assert.NotNil(t, service.jikanClient.rateLimiter)
}
