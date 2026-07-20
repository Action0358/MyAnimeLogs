package services

import (
	"io"
	"log/slog"
	"testing"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServiceLogger creates a logger that discards output for testing
func newTestServiceLogger() *common.Logger {
	handler := slog.NewTextHandler(io.Discard, nil)
	return &common.Logger{
		Logger: slog.New(handler),
	}
}

func TestSaveGenresToDatabase_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	logger := newTestServiceLogger()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM anime_genres WHERE anime_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))
	prep := mock.ExpectPrepare(`INSERT INTO anime_genres`)
	prep.ExpectExec().WithArgs(1, "アクション").WillReturnResult(sqlmock.NewResult(1, 1))
	prep.ExpectExec().WithArgs(1, "ファンタジー").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(`UPDATE animes SET genres_synced_at = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = SaveGenresToDatabase(dbWrapper, logger, 1, []string{"アクション", "ファンタジー"})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveGenresToDatabase_EmptyGenresClearsExisting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	logger := newTestServiceLogger()

	// ジャンルが空でも既存データの削除とgenres_synced_at更新は行われる
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM anime_genres WHERE anime_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`UPDATE animes SET genres_synced_at = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = SaveGenresToDatabase(dbWrapper, logger, 1, nil)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveGenresToDatabase_NilDB(t *testing.T) {
	err := SaveGenresToDatabase(nil, newTestServiceLogger(), 1, []string{"アクション"})
	assert.Error(t, err)
}
