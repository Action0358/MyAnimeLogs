package services

import (
	"testing"

	"myanimelogs/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildKeywordSearchClause_EscapesLikeWildcards(t *testing.T) {
	tests := []struct {
		name     string
		keyword  string
		wantArgs []interface{}
	}{
		{"%はそのままの文字として扱う", "100%", []interface{}{`%100\%%`}},
		{"_はそのままの文字として扱う", "a_b", []interface{}{`%a\_b%`}},
		{"バックスラッシュもエスケープする", `a\b`, []interface{}{`%a\\b%`}},
		{"ワイルドカードが無ければ変わらない", "abc", []interface{}{"%abc%"}},
		{"かな変換後のパターンもエスケープする", "いちご100%", []interface{}{`%いちご100\%%`, `%イチゴ100\%%`, `%いちご100\%%`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, args, _ := buildKeywordSearchClause(tt.keyword, 1, "")
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestGetSearchSuggestions_EscapesLikeWildcards(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())

	mock.ExpectQuery(`SELECT title`).
		WithArgs(`100\%%`, 10).
		WillReturnRows(sqlmock.NewRows([]string{"title"}))

	_, err = service.GetSearchSuggestions("100%", 10)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogSearch_SkipsAnonymousSearch(t *testing.T) {
	// DBを渡さないので、INSERTまで進むとpanicする
	service := NewAnimeService(nil, newTestServiceLogger())

	assert.NotPanics(t, func() {
		service.logSearch("進撃", map[string]interface{}{}, nil, 1)
	})
}

func TestLogSearch_RecordsLoggedInSearch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())

	userID := 7
	mock.ExpectExec(`INSERT INTO search_logs`).
		WithArgs(userID, "進撃", "{}", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	service.logSearch("進撃", map[string]interface{}{}, &userID, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}
