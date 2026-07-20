package services

import (
	"fmt"
	"testing"

	"myanimelogs/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryAnimeSyncTargets_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	logger := newTestServiceLogger()

	malID1 := "100"
	malID2 := "200"
	rows := sqlmock.NewRows([]string{"id", "mal_anime_id", "title", "annict_id"}).
		AddRow(1, malID1, "Anime One", 11).
		AddRow(2, malID2, "Anime Two", 22)

	mock.ExpectQuery(`SELECT id, mal_anime_id, title, annict_id FROM animes WHERE status = \$1`).
		WithArgs("airing").
		WillReturnRows(rows)

	animes, err := QueryAnimeSyncTargets(dbWrapper, logger, "airing animes",
		`SELECT id, mal_anime_id, title, annict_id FROM animes WHERE status = $1`, "airing")

	assert.NoError(t, err)
	require.Len(t, animes, 2)
	assert.Equal(t, 1, animes[0].ID)
	require.NotNil(t, animes[0].MalAnimeID)
	assert.Equal(t, malID1, *animes[0].MalAnimeID)
	assert.Equal(t, "Anime One", animes[0].Title)
	assert.Equal(t, 2, animes[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryAnimeSyncTargets_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	logger := newTestServiceLogger()

	mock.ExpectQuery(`SELECT id, mal_anime_id, title, annict_id FROM animes`).
		WillReturnError(fmt.Errorf("connection refused"))

	animes, err := QueryAnimeSyncTargets(dbWrapper, logger, "sync targets",
		`SELECT id, mal_anime_id, title, annict_id FROM animes`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync targets")
	assert.Nil(t, animes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryAnimeSyncTargets_NilDB(t *testing.T) {
	// db が nil でもパニックせずエラーを返す（Exported 関数の防御的ガード）
	animes, err := QueryAnimeSyncTargets(nil, newTestServiceLogger(), "sync targets",
		`SELECT id, mal_anime_id, title, annict_id FROM animes`)

	assert.Error(t, err)
	assert.Nil(t, animes)
	assert.Contains(t, err.Error(), "database connection not available")
}

func TestQueryAnimeSyncTargets_NilLoggerDoesNotPanic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}

	// スキャン失敗する行を含めて logger.Warn 経路を通す。logger=nil でもパニックしないこと。
	rows := sqlmock.NewRows([]string{"id", "mal_anime_id", "title", "annict_id"}).
		AddRow("not-a-number", "100", "Broken", 11)

	mock.ExpectQuery(`SELECT id, mal_anime_id, title, annict_id FROM animes`).
		WillReturnRows(rows)

	animes, err := QueryAnimeSyncTargets(dbWrapper, nil, "sync targets",
		`SELECT id, mal_anime_id, title, annict_id FROM animes`)

	assert.NoError(t, err)
	assert.Empty(t, animes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryAnimeSyncTargets_SkipsInvalidRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	logger := newTestServiceLogger()

	// 2行目はid列に不正な値を入れてスキャン失敗させ、スキップされることを確認する
	rows := sqlmock.NewRows([]string{"id", "mal_anime_id", "title", "annict_id"}).
		AddRow(1, "100", "Valid Anime", 11).
		AddRow("not-a-number", "200", "Broken Anime", 22)

	mock.ExpectQuery(`SELECT id, mal_anime_id, title, annict_id FROM animes`).
		WillReturnRows(rows)

	animes, err := QueryAnimeSyncTargets(dbWrapper, logger, "sync targets",
		`SELECT id, mal_anime_id, title, annict_id FROM animes`)

	assert.NoError(t, err)
	require.Len(t, animes, 1)
	assert.Equal(t, "Valid Anime", animes[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}
