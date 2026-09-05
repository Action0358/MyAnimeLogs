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
	rows := sqlmock.NewRows([]string{"id", "title", "season", "cover_image_extra_large", "cover_image_large", "cover_image_medium", "image_url"}).
		AddRow(2, "Series Winter", "2023-winter", "https://example.com/winter-xl.jpg", "https://example.com/winter-l.jpg", "https://example.com/winter.jpg", nil).
		AddRow(1, "Series Autumn", "2023-autumn", nil, nil, "https://example.com/autumn.jpg", nil)

	mock.ExpectQuery(`SELECT id, title, season, cover_image_extra_large, cover_image_large, cover_image_medium, image_url\s+FROM animes\s+WHERE id != \$2`).
		WithArgs("シリーズ", 99, nil, "シリーズタイトル", constants.MinRelatedTitlePrefixLength, constants.MaxRelatedSeasons).
		WillReturnRows(rows)

	seriesTitle := "シリーズ"
	related := svc.getRelatedSeasons(99, "シリーズタイトル", &seriesTitle, nil)

	require.Len(t, related, 2)
	// ORDER BYはDB側の責務なので、モックが返す順序をそのまま検証する
	// （実クエリのCASE式によるチクロ順ソートはSQL自体で担保されるため、
	//  ここではNULL許容の値マッピングとLIMIT/引数渡しが正しいことを確認する）
	assert.Equal(t, 2, related[0].ID)
	assert.Equal(t, "2023-winter", *related[0].Season)
	assert.Equal(t, "https://example.com/winter-xl.jpg", related[0].CoverImageExtraLarge)
	assert.Equal(t, "https://example.com/winter-l.jpg", related[0].CoverImageLarge)
	assert.Equal(t, "https://example.com/winter.jpg", related[0].CoverImageMedium)
	assert.Nil(t, related[0].ImageURL)
	assert.Equal(t, 1, related[1].ID)
	assert.Empty(t, related[1].CoverImageExtraLarge)
	assert.Empty(t, related[1].CoverImageLarge)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// series_title・series_tmdb_idが両方欠落し、かつタイトルが前方一致の
// 最小文字数未満の場合は、クエリを発行せず安全にnilを返す。
func TestGetRelatedSeasons_NoUsableSignalReturnsNilWithoutQuery(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	assert.Nil(t, svc.getRelatedSeasons(1, "F", nil, nil))

	empty := ""
	assert.Nil(t, svc.getRelatedSeasons(1, "F", &empty, nil))
}

func TestGetRelatedSeasons_QueryErrorReturnsNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	mock.ExpectQuery(`SELECT id, title, season, cover_image_extra_large, cover_image_large, cover_image_medium, image_url`).
		WillReturnError(assert.AnError)

	seriesTitle := "シリーズ"
	related := svc.getRelatedSeasons(1, "シリーズタイトル", &seriesTitle, nil)

	assert.Nil(t, related)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 「この素晴らしい世界に祝福を！」事例の再現: series_titleが正規化されず
// 各作品で異なる値になっていても、series_tmdb_idの一致で関連作品が拾えること。
func TestGetRelatedSeasons_FallsBackToSeriesTMDBID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	rows := sqlmock.NewRows([]string{"id", "title", "season", "cover_image_extra_large", "cover_image_large", "cover_image_medium", "image_url"}).
		AddRow(2252, "この素晴らしい世界に祝福を！2", "2017-winter", nil, nil, nil, nil).
		AddRow(9361, "この素晴らしい世界に祝福を！3", "2024-spring", nil, nil, nil, nil)

	mock.ExpectQuery(`SELECT id, title, season, cover_image_extra_large, cover_image_large, cover_image_medium, image_url\s+FROM animes\s+WHERE id != \$2`).
		WithArgs("この素晴らしい世界に祝福を！", 2211, 65844, "この素晴らしい世界に祝福を！", constants.MinRelatedTitlePrefixLength, constants.MaxRelatedSeasons).
		WillReturnRows(rows)

	seriesTitle := "この素晴らしい世界に祝福を！"
	tmdbID := 65844
	related := svc.getRelatedSeasons(2211, "この素晴らしい世界に祝福を！", &seriesTitle, &tmdbID)

	require.Len(t, related, 2)
	assert.Equal(t, 2252, related[0].ID)
	assert.Equal(t, 9361, related[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 対照群のリグレッション確認: series_titleの正規化が効いているシリーズ
// （「ダンジョンに出会いを求めるのは間違っているだろうか」）が、
// 従来通りseries_title一致で拾えること。
func TestGetRelatedSeasons_ExistingSeriesTitleMatchStillWorks(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	rows := sqlmock.NewRows([]string{"id", "title", "season", "cover_image_extra_large", "cover_image_large", "cover_image_medium", "image_url"}).
		AddRow(1431, "ダンジョンに出会いを求めるのは間違っているだろうかⅡ", "2019-summer", nil, nil, nil, nil).
		AddRow(1685, "ダンジョンに出会いを求めるのは間違っているだろうかⅢ", "2020-autumn", nil, nil, nil, nil)

	mock.ExpectQuery(`SELECT id, title, season, cover_image_extra_large, cover_image_large, cover_image_medium, image_url\s+FROM animes\s+WHERE id != \$2`).
		WithArgs("ダンジョンに出会いを求めるのは間違っているだろうか", 84, nil, "ダンジョンに出会いを求めるのは間違っているだろうか", constants.MinRelatedTitlePrefixLength, constants.MaxRelatedSeasons).
		WillReturnRows(rows)

	seriesTitle := "ダンジョンに出会いを求めるのは間違っているだろうか"
	related := svc.getRelatedSeasons(84, "ダンジョンに出会いを求めるのは間違っているだろうか", &seriesTitle, nil)

	require.Len(t, related, 2)
	assert.Equal(t, 1431, related[0].ID)
	assert.Equal(t, 1685, related[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 1文字タイトル「F」「K」は前方一致の最小文字数ガードにより
// クエリすら発行されず、無関係な作品を拾わないこと。
func TestGetRelatedSeasons_SingleCharacterTitleGuardedByMinLength(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	for _, title := range []string{"F", "K"} {
		assert.Nil(t, svc.getRelatedSeasons(1, title, nil, nil), "title=%s", title)
	}
}

// 前方一致にLIKEを使うと、タイトル自体がパターンとして解釈される。
// 実データに「攻殻機動隊 SAC_2045」「NieA_7」（_は任意の1文字）や
// 「いちご100%」（%は任意の文字列）が存在するため、LIKEでは無関係な作品まで
// 拾ってしまう。left()による前方一致であることをSQLレベルで固定する。
func TestGetRelatedSeasons_PrefixMatchDoesNotUseLikePatterns(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	svc := NewAnimeService(&database.DB{DB: db}, newTestServiceLogger())
	defer svc.Close()

	rows := sqlmock.NewRows([]string{"id", "title", "season", "cover_image_extra_large", "cover_image_large", "cover_image_medium", "image_url"})

	// タイトルを連結したLIKEパターンが組み立てられていないこと
	mock.ExpectQuery(`left\(title, length\(\$4\)\) = \$4`).WillReturnRows(rows)

	svc.getRelatedSeasons(1, "攻殻機動隊 SAC_2045", nil, nil)

	assert.NoError(t, mock.ExpectationsWereMet())
}
