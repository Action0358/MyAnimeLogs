package services

import (
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
