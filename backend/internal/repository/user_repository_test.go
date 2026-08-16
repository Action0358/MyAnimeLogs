package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
	"myanimelogs/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "test@example.com"
	authUserID := "auth-uuid-1"
	expectedUser := models.User{
		ID:          1,
		Name:        "Test User",
		Email:       email,
		Description: testutil.StringPtr(""),
		AuthUserID:  &authUserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "auth_user_id", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email,
			expectedUser.Description, expectedUser.Icon, expectedUser.FavoriteAnime, expectedUser.SnsLink,
			expectedUser.AuthUserID, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, email, description, icon, favorite_anime, sns_link, auth_user_id, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	result, err := repo.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Name, result.Name)
	assert.Equal(t, expectedUser.Email, result.Email)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "nonexistent@example.com"

	mock.ExpectQuery(`SELECT id, name, email, description, icon, favorite_anime, sns_link, auth_user_id, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByEmail(email)

	assert.NoError(t, err)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "test@example.com"
	expectedError := errors.New("database connection error")

	mock.ExpectQuery(`SELECT id, name, email, description, icon, favorite_anime, sns_link, auth_user_id, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(expectedError)

	result, err := repo.GetByEmail(email)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user by email")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	userID := 1
	authUserID := "auth-uuid-1"
	expectedUser := models.User{
		ID:          userID,
		Name:        "Test User",
		Email:       "test@example.com",
		Description: testutil.StringPtr(""),
		AuthUserID:  &authUserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "auth_user_id", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email,
			expectedUser.Description, expectedUser.Icon, expectedUser.FavoriteAnime, expectedUser.SnsLink,
			expectedUser.AuthUserID, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, email, description, icon, favorite_anime, sns_link, auth_user_id, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.GetByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Name, result.Name)
	assert.Equal(t, expectedUser.Email, result.Email)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	userID := 999

	mock.ExpectQuery(`SELECT id, name, email, description, icon, favorite_anime, sns_link, auth_user_id, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(userID)

	assert.NoError(t, err)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	authUserID := "auth-uuid-new"
	user := &models.User{
		Name:       "New User",
		Email:      "new@example.com",
		AuthUserID: &authUserID,
	}

	expectedID := 1
	expectedTime := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(expectedID, expectedTime, expectedTime)

	mock.ExpectQuery(`INSERT INTO users \(name, email, auth_user_id, created_at, updated_at\) VALUES \(\$1, \$2, \$3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP\) RETURNING id, created_at, updated_at`).
		WithArgs(user.Name, user.Email, user.AuthUserID).
		WillReturnRows(rows)

	err = repo.Create(user)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, expectedTime, user.CreatedAt)
	assert.Equal(t, expectedTime, user.UpdatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	authUserID := "auth-uuid-existing"
	user := &models.User{
		Name:       "New User",
		Email:      "existing@example.com",
		AuthUserID: &authUserID,
	}

	expectedError := errors.New("duplicate key value violates unique constraint")

	mock.ExpectQuery(`INSERT INTO users \(name, email, auth_user_id, created_at, updated_at\) VALUES \(\$1, \$2, \$3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP\) RETURNING id, created_at, updated_at`).
		WithArgs(user.Name, user.Email, user.AuthUserID).
		WillReturnError(expectedError)

	err = repo.Create(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_SQLInjectionPrevention(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	maliciousEmail := "'; DROP TABLE users; --"

	mock.ExpectQuery(`SELECT id, name, email, description, icon, favorite_anime, sns_link, auth_user_id, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(maliciousEmail).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByEmail(maliciousEmail)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdatePartial_WithFields(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	userID := 1
	name := "Updated Name"
	favoriteAnime := "鬼滅の刃"
	snsLink := "https://twitter.com/updated"

	req := models.UserUpdateRequest{
		Name:          &name,
		FavoriteAnime: &favoriteAnime,
		SnsLink:       &snsLink,
	}

	authUserID := "auth-uuid-1"
	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "auth_user_id", "created_at", "updated_at"}).
		AddRow(userID, name, "test@example.com", "", nil, favoriteAnime, snsLink, authUserID, time.Now(), time.Now())

	mock.ExpectQuery(`UPDATE users SET .*name = \$\d+.*favorite_anime = \$\d+.*sns_link = \$\d+.*updated_at = CURRENT_TIMESTAMP WHERE id = \$\d+ RETURNING`).
		WithArgs(name, favoriteAnime, snsLink, userID).
		WillReturnRows(rows)

	result, err := repo.UpdatePartial(userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, name, result.Name)
	assert.NotNil(t, result.FavoriteAnime)
	assert.Equal(t, favoriteAnime, *result.FavoriteAnime)
	assert.NotNil(t, result.SnsLink)
	assert.Equal(t, snsLink, *result.SnsLink)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateIcon_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	userID := 1
	iconURL := "/uploads/icons/user_1_1234567890.jpg"
	authUserID := "auth-uuid-1"
	expectedUser := models.User{
		ID:          userID,
		Name:        "Test User",
		Email:       "test@example.com",
		Description: testutil.StringPtr("Test description"),
		Icon:        &iconURL,
		AuthUserID:  &authUserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "auth_user_id", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email, expectedUser.Description,
			expectedUser.Icon, expectedUser.FavoriteAnime, expectedUser.SnsLink, expectedUser.AuthUserID,
			expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`UPDATE users SET icon = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs(iconURL, userID).
		WillReturnRows(rows)

	result, err := repo.UpdateIcon(userID, iconURL)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, &iconURL, result.Icon)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateIcon_DeleteIcon(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	userID := 1
	authUserID := "auth-uuid-1"
	expectedUser := models.User{
		ID:          userID,
		Name:        "Test User",
		Email:       "test@example.com",
		Description: testutil.StringPtr("Test description"),
		Icon:        nil,
		AuthUserID:  &authUserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "auth_user_id", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email, expectedUser.Description,
			expectedUser.Icon, expectedUser.FavoriteAnime, expectedUser.SnsLink, expectedUser.AuthUserID,
			expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`UPDATE users SET icon = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs(nil, userID).
		WillReturnRows(rows)

	result, err := repo.UpdateIcon(userID, "")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Nil(t, result.Icon)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateIcon_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	userID := 1
	iconURL := "/uploads/icons/user_1_1234567890.jpg"

	mock.ExpectQuery(`UPDATE users SET icon = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs(iconURL, userID).
		WillReturnError(errors.New("database error"))

	result, err := repo.UpdateIcon(userID, iconURL)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update user icon")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- GetOrCreateByAuthUserID ---
//
// Google 認証への移行で最も壊れやすい経路のため、4つの分岐すべてを検証する。
// 特に relinked は、既存の視聴履歴・お気に入りを持つ行が新しい auth_user_id へ
// 正しく引き継がれるかを保証するもので、失敗すると既存ユーザーのデータが失われる。

const (
	selectByAuthUserIDQuery = `SELECT id, name, email, description, icon, favorite_anime, sns_link, auth_user_id, created_at, updated_at FROM users WHERE auth_user_id = \$1`
	linkByEmailQuery        = `WITH target AS`
	insertUserQuery         = `INSERT INTO users \(name, email, auth_user_id, created_at, updated_at\)`
)

func userRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "email", "description", "icon", "favorite_anime", "sns_link",
		"auth_user_id", "created_at", "updated_at",
	})
}

func userRowsWithPrevious() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "email", "description", "icon", "favorite_anime", "sns_link",
		"auth_user_id", "created_at", "updated_at", "previous_auth_user_id",
	})
}

func TestUserRepository_GetOrCreateByAuthUserID_Existing(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&database.DB{DB: db})
	authUserID := "auth-uuid-existing"
	now := time.Now()

	mock.ExpectQuery(selectByAuthUserIDQuery).
		WithArgs(authUserID).
		WillReturnRows(userRows().AddRow(
			7, "既存ユーザー", "user@example.com", nil, nil, nil, nil, authUserID, now, now,
		))

	user, result, err := repo.GetOrCreateByAuthUserID(authUserID, "user@example.com", "Google Name")

	assert.NoError(t, err)
	assert.Equal(t, 7, user.ID)
	assert.Equal(t, AuthLinkExisting, result)
	// 既存ユーザーは書き込みを一切行わない（毎リクエストの UPDATE を避けるため）
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetOrCreateByAuthUserID_ClaimsUnlinkedRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&database.DB{DB: db})
	authUserID := "auth-uuid-new"
	email := "legacy@example.com"
	now := time.Now()

	mock.ExpectQuery(selectByAuthUserIDQuery).
		WithArgs(authUserID).
		WillReturnError(sql.ErrNoRows)

	// auth_user_id が NULL の行（Supabase Auth 移行前のレコード）を引き当てる
	mock.ExpectQuery(linkByEmailQuery).
		WithArgs(authUserID, email, "Google Name").
		WillReturnRows(userRowsWithPrevious().AddRow(
			3, "移行前の表示名", email, nil, nil, nil, nil, authUserID, now, now, nil,
		))

	user, result, err := repo.GetOrCreateByAuthUserID(authUserID, email, "Google Name")

	assert.NoError(t, err)
	assert.Equal(t, 3, user.ID)
	assert.Equal(t, AuthLinkClaimed, result)
	// ユーザーが設定済みの表示名は Google プロフィール名で上書きしない
	assert.Equal(t, "移行前の表示名", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetOrCreateByAuthUserID_RelinksDifferentAuthUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&database.DB{DB: db})
	newAuthUserID := "auth-uuid-google"
	oldAuthUserID := "auth-uuid-password"
	email := "existing@example.com"
	now := time.Now()

	mock.ExpectQuery(selectByAuthUserIDQuery).
		WithArgs(newAuthUserID).
		WillReturnError(sql.ErrNoRows)

	// パスワード認証時代の auth_user_id が入った行を、Google の UUID へ張り替える
	mock.ExpectQuery(linkByEmailQuery).
		WithArgs(newAuthUserID, email, "Google Name").
		WillReturnRows(userRowsWithPrevious().AddRow(
			42, "ユーザーが変更した名前", email, nil, nil, nil, nil, newAuthUserID, now, now, oldAuthUserID,
		))

	user, result, err := repo.GetOrCreateByAuthUserID(newAuthUserID, email, "Google Name")

	assert.NoError(t, err)
	// 既存の user ID が維持されること = 視聴履歴・お気に入りが引き継がれること
	assert.Equal(t, 42, user.ID)
	assert.Equal(t, AuthLinkRelinked, result)
	assert.Equal(t, "ユーザーが変更した名前", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetOrCreateByAuthUserID_CreatesNewUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&database.DB{DB: db})
	authUserID := "auth-uuid-fresh"
	email := "new@example.com"
	now := time.Now()

	mock.ExpectQuery(selectByAuthUserIDQuery).
		WithArgs(authUserID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(linkByEmailQuery).
		WithArgs(authUserID, email, "Google Name").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(insertUserQuery).
		WithArgs("Google Name", email, authUserID).
		WillReturnRows(userRows().AddRow(
			99, "Google Name", email, nil, nil, nil, nil, authUserID, now, now,
		))

	user, result, err := repo.GetOrCreateByAuthUserID(authUserID, email, "Google Name")

	assert.NoError(t, err)
	assert.Equal(t, 99, user.ID)
	assert.Equal(t, AuthLinkCreated, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetOrCreateByAuthUserID_LinkQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&database.DB{DB: db})
	authUserID := "auth-uuid-err"

	mock.ExpectQuery(selectByAuthUserIDQuery).
		WithArgs(authUserID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(linkByEmailQuery).
		WithArgs(authUserID, "err@example.com", "Google Name").
		WillReturnError(errors.New("deadlock detected"))

	user, _, err := repo.GetOrCreateByAuthUserID(authUserID, "err@example.com", "Google Name")

	assert.Error(t, err)
	assert.Nil(t, user)
	// 曖昧な「作成に失敗」ではなく、リンク段階での失敗と分かること
	assert.Contains(t, err.Error(), "failed to link user by email")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 退会直後は JWT がまだ有効なため、削除済み auth_user_id で
// GetOrCreateByAuthUserID が呼ばれうる。このとき users 行を作り直して
// データを復活させてはならず、呼び出し側が 401 を返せるよう
// ErrAuthUserNotFound を返す必要がある。
func TestUserRepository_GetOrCreateByAuthUserID_DeletedAuthUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&database.DB{DB: db})
	authUserID := "auth-uuid-deleted"
	email := "deleted@example.com"

	mock.ExpectQuery(selectByAuthUserIDQuery).
		WithArgs(authUserID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(linkByEmailQuery).
		WithArgs(authUserID, email, "Google Name").
		WillReturnError(sql.ErrNoRows)
	// auth.users が存在しないため外部キー違反になる
	mock.ExpectQuery(insertUserQuery).
		WithArgs("Google Name", email, authUserID).
		WillReturnError(&pq.Error{Code: "23503", Constraint: "fk_users_auth_user_id"})

	user, _, err := repo.GetOrCreateByAuthUserID(authUserID, email, "Google Name")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrAuthUserNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 外部キー違反以外のINSERT失敗は、削除済み扱いにせず通常のエラーとして返す
func TestUserRepository_GetOrCreateByAuthUserID_InsertErrorIsNotTreatedAsDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(&database.DB{DB: db})
	authUserID := "auth-uuid-x"
	email := "x@example.com"

	mock.ExpectQuery(selectByAuthUserIDQuery).
		WithArgs(authUserID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(linkByEmailQuery).
		WithArgs(authUserID, email, "Name").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(insertUserQuery).
		WithArgs("Name", email, authUserID).
		WillReturnError(errors.New("connection reset"))

	user, _, err := repo.GetOrCreateByAuthUserID(authUserID, email, "Name")

	assert.Nil(t, user)
	assert.NotErrorIs(t, err, ErrAuthUserNotFound)
	assert.Contains(t, err.Error(), "failed to create user")
	assert.NoError(t, mock.ExpectationsWereMet())
}
