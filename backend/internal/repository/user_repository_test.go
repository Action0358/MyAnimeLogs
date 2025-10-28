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
	expectedUser := models.User{
		ID:            1,
		Name:          "Test User",
		Email:         email,
		Password:      "hashed_password",
		Description:   testutil.StringPtr(""),
		Icon:          nil,
		FavoriteAnime: nil,
		SnsLink:       nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "description", "icon", "favorite_anime", "sns_link", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email, expectedUser.Password,
			expectedUser.Description, expectedUser.Icon, expectedUser.FavoriteAnime, expectedUser.SnsLink, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, email, password, description, icon, favorite_anime, sns_link, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	result, err := repo.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Name, result.Name)
	assert.Equal(t, expectedUser.Email, result.Email)
	assert.Equal(t, expectedUser.Password, result.Password)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "nonexistent@example.com"

	mock.ExpectQuery(`SELECT id, name, email, password, description, icon, favorite_anime, sns_link, created_at, updated_at FROM users WHERE email = \$1`).
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

	mock.ExpectQuery(`SELECT id, name, email, password, description, icon, favorite_anime, sns_link, created_at, updated_at FROM users WHERE email = \$1`).
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
	expectedUser := models.User{
		ID:            userID,
		Name:          "Test User",
		Email:         "test@example.com",
		Description:   testutil.StringPtr(""),
		Icon:          nil,
		FavoriteAnime: nil,
		SnsLink:       nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email,
			expectedUser.Description, expectedUser.Icon, expectedUser.FavoriteAnime, expectedUser.SnsLink, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, email, description, icon, favorite_anime, sns_link, created_at, updated_at FROM users WHERE id = \$1`).
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

	mock.ExpectQuery(`SELECT id, name, email, description, icon, favorite_anime, sns_link, created_at, updated_at FROM users WHERE id = \$1`).
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

	user := &models.User{
		Name:     "New User",
		Email:    "new@example.com",
		Password: "hashed_password",
	}

	expectedID := 1
	expectedTime := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(expectedID, expectedTime, expectedTime)

	mock.ExpectQuery(`INSERT INTO users \(name, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP\) RETURNING id, created_at, updated_at`).
		WithArgs(user.Name, user.Email, user.Password).
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

	user := &models.User{
		Name:     "New User",
		Email:    "existing@example.com",
		Password: "hashed_password",
	}

	expectedError := errors.New("duplicate key value violates unique constraint")

	mock.ExpectQuery(`INSERT INTO users \(name, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP\) RETURNING id, created_at, updated_at`).
		WithArgs(user.Name, user.Email, user.Password).
		WillReturnError(expectedError)

	err = repo.Create(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_EmailExists_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "existing@example.com"

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
		WithArgs(email).
		WillReturnRows(rows)

	result, err := repo.EmailExists(email)

	assert.NoError(t, err)
	assert.True(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_EmailExists_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "nonexistent@example.com"

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
		WithArgs(email).
		WillReturnRows(rows)

	result, err := repo.EmailExists(email)

	assert.NoError(t, err)
	assert.False(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_EmailExists_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "test@example.com"
	expectedError := errors.New("database connection error")

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
		WithArgs(email).
		WillReturnError(expectedError)

	result, err := repo.EmailExists(email)

	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "failed to check email existence")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_SQLInjectionPrevention(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	// SQLインジェクション攻撃を試みる入力
	maliciousEmail := "'; DROP TABLE users; --"

	// パラメータ化クエリが使用されているため、この値はそのまま検索される
	mock.ExpectQuery(`SELECT id, name, email, password, description, icon, favorite_anime, sns_link, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(maliciousEmail).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByEmail(maliciousEmail)

	assert.NoError(t, err)
	assert.Nil(t, result)

	// SQLインジェクションが防がれていることを確認
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_NullHandling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	user := &models.User{
		Name:        "User With Nulls",
		Email:       "null@example.com",
		Password:    "hashed_password",
		Description: testutil.StringPtr(""), // 空文字列
		Icon:        nil,                    // NULL値
	}

	expectedID := 1
	expectedTime := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(expectedID, expectedTime, expectedTime)

	mock.ExpectQuery(`INSERT INTO users \(name, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP\) RETURNING id, created_at, updated_at`).
		WithArgs(user.Name, user.Email, user.Password).
		WillReturnRows(rows)

	err = repo.Create(user)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, user.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_WithNullFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "user@example.com"

	// 空文字列とNULL値を含む行をモック
	rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "description", "icon", "favorite_anime", "sns_link", "created_at", "updated_at"}).
		AddRow(1, "Test User", email, "hashed_password", "", nil, nil, nil, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT id, name, email, password, description, icon, favorite_anime, sns_link, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	result, err := repo.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "Test User", result.Name)
	assert.Equal(t, email, result.Email)
	assert.NotNil(t, result.Description)
	assert.Equal(t, "", *result.Description) // 空文字列
	assert.Nil(t, result.Icon)               // NULL値

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_WithNewFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := &database.DB{DB: db}
	repo := NewUserRepository(dbWrapper)

	email := "user@example.com"
	favoriteAnime := "進撃の巨人"
	snsLink := "https://twitter.com/user"

	rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "description", "icon", "favorite_anime", "sns_link", "created_at", "updated_at"}).
		AddRow(1, "Test User", email, "hashed_password", "Test description", nil, favoriteAnime, snsLink, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT id, name, email, password, description, icon, favorite_anime, sns_link, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	result, err := repo.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "Test User", result.Name)
	assert.Equal(t, email, result.Email)
	assert.NotNil(t, result.FavoriteAnime)
	assert.Equal(t, favoriteAnime, *result.FavoriteAnime)
	assert.NotNil(t, result.SnsLink)
	assert.Equal(t, snsLink, *result.SnsLink)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdatePartial_WithNewFields(t *testing.T) {
	// Use regex matcher for more flexible query matching
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

	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "created_at", "updated_at"}).
		AddRow(userID, name, "test@example.com", "", nil, favoriteAnime, snsLink, time.Now(), time.Now())

	// Use regex pattern that matches the UPDATE query structure
	// This pattern validates that all expected fields are updated
	// while being flexible about their order
	mock.ExpectQuery(`UPDATE users SET .*name = \$\d+.*favorite_anime = \$\d+.*sns_link = \$\d+.*updated_at = CURRENT_TIMESTAMP WHERE id = \$\d+ RETURNING id, name, email, description, icon, favorite_anime, sns_link, created_at, updated_at`).
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
	expectedUser := models.User{
		ID:            userID,
		Name:          "Test User",
		Email:         "test@example.com",
		Description:   testutil.StringPtr("Test description"),
		Icon:          &iconURL,
		FavoriteAnime: nil,
		SnsLink:       nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email, expectedUser.Description,
			expectedUser.Icon, expectedUser.FavoriteAnime, expectedUser.SnsLink, expectedUser.CreatedAt, expectedUser.UpdatedAt)

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
	expectedUser := models.User{
		ID:            userID,
		Name:          "Test User",
		Email:         "test@example.com",
		Description:   testutil.StringPtr("Test description"),
		Icon:          nil,
		FavoriteAnime: nil,
		SnsLink:       nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "description", "icon", "favorite_anime", "sns_link", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email, expectedUser.Description,
			expectedUser.Icon, expectedUser.FavoriteAnime, expectedUser.SnsLink, expectedUser.CreatedAt, expectedUser.UpdatedAt)

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
