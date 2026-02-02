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
	service := NewAnimeService(nil, logger, nil)

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.anilistClient)
}

// TestAnimeService_SetSearchDefaults tests default value setting
func TestAnimeService_SetSearchDefaults(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger, nil)

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
	service := NewAnimeService(nil, logger, nil)

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
	service := NewAnimeService(nil, logger, nil)

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
				"FROM animes a",
			},
			shouldNotContain: []string{
				"LEFT JOIN anime_genres",
				"LEFT JOIN (", // avg_ratings subquery
				"INNER JOIN user_favorites",
			},
		},
		{
			name: "ジャンルフィルター - ジャンルJOIN",
			params: models.AdvancedSearchParams{
				Genres: []string{"アクション"},
			},
			userID: nil,
			shouldContain: []string{
				"FROM animes a",
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
				"FROM animes a",
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
			name: "ユーザーステータスフィルター - お気に入りINNER JOIN",
			params: models.AdvancedSearchParams{
				UserStatus: "watching",
			},
			userID: func() *int { id := 123; return &id }(),
			shouldContain: []string{
				"INNER JOIN user_favorites uf",
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
				"INNER JOIN user_favorites uf",
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
	service := NewAnimeService(nil, logger, nil)

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
			expectedArgs: 2, // query (FTS用), likePattern (ILIKE用、3カラムで共有)
			shouldContain: []string{
				"to_tsvector",
				"plainto_tsquery",
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
				Season: "spring",
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
				"uf.user_id",
				"uf.status",
			},
		},
		{
			name: "複合フィルター",
			params: models.AdvancedSearchParams{
				Query:      "アニメ",
				Year:       2024,
				Season:     "spring",
				Genres:     []string{"アクション"},
				MinRating:  4.0,
				UserStatus: "watching",
			},
			userID:       func() *int { id := 123; return &id }(),
			expectedArgs: 8, // query(2) + year + season + genre + minRating + user_id + userStatus
			shouldContain: []string{
				"to_tsvector",
				"a.season LIKE",
				"ag.genre IN",
				"avg_ratings.avg_rating >=",
				"uf.user_id",
				"uf.status",
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
	assert.NotNil(t, ErrEpisodeNotFound)
	assert.Equal(t, "anime not found", ErrAnimeNotFound.Error())
	assert.Equal(t, "episode not found", ErrEpisodeNotFound.Error())
}

// TestAnimeService_ImplementsInterface verifies that AnimeService implements AnimeServiceInterface
func TestAnimeService_ImplementsInterface(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger, nil)

	// Verify that service implements AnimeServiceInterface (compile-time check)
	var _ AnimeServiceInterface = service

	// If this compiles, the interface is correctly implemented
	assert.NotNil(t, service)
}

// TestAnimeService_JikanClientIntegration tests Jikan client integration
func TestAnimeService_JikanClientIntegration(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := NewAnimeService(nil, logger, nil)

	// Verify that JikanClient is initialized
	assert.NotNil(t, service.jikanClient)
	assert.NotNil(t, service.jikanClient.httpClient)
	assert.NotNil(t, service.jikanClient.rateLimiter)
}

func TestNormalizeEpisodes(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := &AnimeService{logger: logger}

	tests := []struct {
		name        string
		input       []models.JikanEpisode
		wantRegular int
		wantSpecial int
	}{
		{
			name: "通常エピソードのみ",
			input: []models.JikanEpisode{
				{MalID: 1, Title: "Episode 1", Filler: false, Recap: false},
				{MalID: 2, Title: "Episode 2", Filler: false, Recap: false},
			},
			wantRegular: 2,
			wantSpecial: 0,
		},
		{
			name: "特別編を含む（recap）",
			input: []models.JikanEpisode{
				{MalID: 1, Title: "Episode 1", Filler: false, Recap: false},
				{MalID: 2, Title: "Special Recap", Filler: false, Recap: true},
				{MalID: 3, Title: "Episode 2", Filler: false, Recap: false},
			},
			wantRegular: 2,
			wantSpecial: 1,
		},
		{
			name: "特別編を含む（filler）",
			input: []models.JikanEpisode{
				{MalID: 1, Title: "Episode 1", Filler: false, Recap: false},
				{MalID: 2, Title: "Filler Episode", Filler: true, Recap: false},
				{MalID: 3, Title: "Episode 2", Filler: false, Recap: false},
			},
			wantRegular: 2,
			wantSpecial: 1,
		},
		{
			name: "複数の特別編",
			input: []models.JikanEpisode{
				{MalID: 1, Title: "Episode 1", Filler: false, Recap: false},
				{MalID: 2, Title: "Recap 1", Filler: false, Recap: true},
				{MalID: 3, Title: "Episode 2", Filler: false, Recap: false},
				{MalID: 4, Title: "Filler 1", Filler: true, Recap: false},
				{MalID: 5, Title: "Episode 3", Filler: false, Recap: false},
			},
			wantRegular: 3,
			wantSpecial: 2,
		},
		{
			name:        "空のエピソードリスト",
			input:       []models.JikanEpisode{},
			wantRegular: 0,
			wantSpecial: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regular, special := service.normalizeEpisodes(tt.input)

			assert.Equal(t, tt.wantRegular, len(regular), "通常エピソード数が一致しない")
			assert.Equal(t, tt.wantSpecial, len(special), "特別編数が一致しない")

			// 番号が正しく振り直されているか確認
			for i, ep := range regular {
				assert.Equal(t, i+1, ep.Number, "通常エピソードの番号が正しくない")
				assert.False(t, ep.IsSpecial, "通常エピソードのIsSpecialがfalseでない")
			}

			for i, ep := range special {
				assert.Equal(t, i+1, ep.Number, "特別編の番号が正しくない")
				assert.True(t, ep.IsSpecial, "特別編のIsSpecialがtrueでない")
			}
		})
	}
}

func TestNormalizeEpisodes_PreservesFlags(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	service := &AnimeService{logger: logger}

	input := []models.JikanEpisode{
		{MalID: 1, Title: "Episode 1", Filler: false, Recap: false},
		{MalID: 2, Title: "Recap Episode", Filler: false, Recap: true},
		{MalID: 3, Title: "Filler Episode", Filler: true, Recap: false},
	}

	regular, special := service.normalizeEpisodes(input)

	// 通常エピソードのフラグ確認
	assert.Len(t, regular, 1)
	assert.False(t, regular[0].Filler)
	assert.False(t, regular[0].Recap)

	// 特別編のフラグ確認
	assert.Len(t, special, 2)
	// Recap episode
	recapFound := false
	fillerFound := false
	for _, ep := range special {
		if ep.Recap {
			recapFound = true
			assert.False(t, ep.Filler)
		}
		if ep.Filler {
			fillerFound = true
			assert.False(t, ep.Recap)
		}
	}
	assert.True(t, recapFound, "Recapエピソードが見つからない")
	assert.True(t, fillerFound, "Fillerエピソードが見つからない")
}
