package services

import (
	"database/sql"
	"testing"

	"myanimelogs/internal/database"
	"myanimelogs/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 複数作品のジャンルを1クエリでまとめて取得し、各Animeへ振り分けること
func TestAttachGenresBulk_AssignsGenresPerAnime(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	db := &database.DB{DB: sqlDB}

	svc := NewAnimeService(db, newTestServiceLogger())
	defer svc.Close()

	animes := []models.Anime{{ID: 1}, {ID: 2}}

	mock.ExpectQuery(`SELECT anime_id, genre FROM anime_genres`).
		WithArgs(pq.Array([]int64{1, 2})).
		WillReturnRows(sqlmock.NewRows([]string{"anime_id", "genre"}).
			AddRow(1, "アクション").
			AddRow(1, "コメディ").
			AddRow(2, "ドラマ"))

	svc.attachGenresBulk(animes)

	assert.Equal(t, []string{"アクション", "コメディ"}, animes[0].Genres)
	assert.Equal(t, []string{"ドラマ"}, animes[1].Genres)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 作品が無い場合はクエリを発行しないこと
func TestAttachGenresBulk_NoAnimes_SkipsQuery(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	db := &database.DB{DB: sqlDB}

	svc := NewAnimeService(db, newTestServiceLogger())
	defer svc.Close()

	svc.attachGenresBulk([]models.Anime{})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// 外部URLは http/https のときだけ返し、javascript: 等や NULL は nil にすること
func TestSafeExternalURLPtr(t *testing.T) {
	tests := []struct {
		name string
		raw  sql.NullString
		want *string
	}{
		{name: "https はそのまま返す", raw: sql.NullString{String: "https://example.com", Valid: true}, want: strPtr("https://example.com")},
		{name: "javascript: は nil", raw: sql.NullString{String: "javascript:alert(1)", Valid: true}, want: nil},
		{name: "空文字列は nil", raw: sql.NullString{String: "", Valid: true}, want: nil},
		{name: "NULL は nil", raw: sql.NullString{}, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, safeExternalURLPtr(tt.raw))
		})
	}
}
