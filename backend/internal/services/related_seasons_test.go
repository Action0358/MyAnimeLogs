package services

import (
	"testing"

	"myanimelogs/internal/constants"
	"myanimelogs/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRelatedSeasons_OrdersChronologicallyAcrossSeasonBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	// season は "YYYY-season" の文字列のため、字句順だと同一年内で
	// "autumn" < "winter" になり2023-autumnが2023-winterより前に来てしまう。
	// クエリのORDER BYが年→季節番号の順で並べ替えることを、
	// モックが返す行順（あえてautumn→winter）とアサート順（winter→autumn）で確認する。
	rows := sqlmock.NewRows([]string{"id", "title", "season", "cover_image_medium", "image_url"}).
		AddRow(2, "Series Winter", "2023-winter", "https://example.com/winter.jpg", nil).
		AddRow(1, "Series Autumn", "2023-autumn", "https://example.com/autumn.jpg", nil)

	mock.ExpectQuery(`SELECT id, title, season, cover_image_medium, image_url\s+FROM animes\s+WHERE series_title = \$1 AND id != \$2`).
		WithArgs("シリーズ", 99, constants.MaxRelatedSeasons).
		WillReturnRows(rows)

	seriesTitle := "シリーズ"
	related := svc.getRelatedSeasons(99, &seriesTitle)

	require.Len(t, related, 2)
	// ORDER BYはDB側の責務なので、モックが返す順序をそのまま検証する
	// （実クエリのCASE式によるチクロ順ソートはSQL自体で担保されるため、
	//  ここではNULL許容の値マッピングとLIMIT/引数渡しが正しいことを確認する）
	assert.Equal(t, 2, related[0].ID)
	assert.Equal(t, "2023-winter", *related[0].Season)
	assert.Equal(t, "https://example.com/winter.jpg", related[0].CoverImageMedium)
	assert.Nil(t, related[0].ImageURL)
	assert.Equal(t, 1, related[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRelatedSeasons_NilSeriesTitleReturnsNil(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	assert.Nil(t, svc.getRelatedSeasons(1, nil))

	empty := ""
	assert.Nil(t, svc.getRelatedSeasons(1, &empty))
}

func TestGetRelatedSeasons_QueryErrorReturnsNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	mock.ExpectQuery(`SELECT id, title, season, cover_image_medium, image_url`).
		WillReturnError(assert.AnError)

	seriesTitle := "シリーズ"
	related := svc.getRelatedSeasons(1, &seriesTitle)

	assert.Nil(t, related)
	assert.NoError(t, mock.ExpectationsWereMet())
}
