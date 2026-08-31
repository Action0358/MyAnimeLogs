package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"myanimelogs/internal/database"
	"myanimelogs/internal/models"

	"github.com/lib/pq"
)

// ErrAuthUserNotFound は、JWT は有効だが対応する auth.users が存在しない場合に返る。
// アカウント削除直後、まだ失効していない JWT でアクセスされたときに発生する。
var ErrAuthUserNotFound = errors.New("auth user no longer exists")

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, name, email, description, favorite_anime, sns_link,
		       auth_user_id, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Description, &user.FavoriteAnime, &user.SnsLink,
		&user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, name, email, description, favorite_anime, sns_link,
		       auth_user_id, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Description, &user.FavoriteAnime, &user.SnsLink,
		&user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// AuthLinkResult describes how a users row was resolved for an incoming Supabase auth UUID.
// The caller logs this so that the Google 移行期のリンク発生状況を追跡できる。
type AuthLinkResult int

const (
	// AuthLinkExisting: auth_user_id で直接見つかった（移行後の通常ケース）
	AuthLinkExisting AuthLinkResult = iota
	// AuthLinkClaimed: auth_user_id が未設定の行をメール一致で引き当てた（Supabase Auth 移行前の行）
	AuthLinkClaimed
	// AuthLinkRelinked: 別の auth_user_id が設定済みの行を、同一メールで新しい auth_user_id へ張り替えた
	AuthLinkRelinked
	// AuthLinkCreated: 新規作成
	AuthLinkCreated
)

func (r AuthLinkResult) String() string {
	switch r {
	case AuthLinkExisting:
		return "existing"
	case AuthLinkClaimed:
		return "claimed"
	case AuthLinkRelinked:
		return "relinked"
	case AuthLinkCreated:
		return "created"
	default:
		return "unknown"
	}
}

// GetOrCreateByAuthUserID returns an existing user or creates a new one for the given Supabase auth UUID.
// Handles three cases in order:
//  1. Returning user: SELECT by auth_user_id (avoids write on every request)
//  2. Link by email: メールが一致する既存行に auth_user_id を張り替える
//  3. New user: INSERT
//
// 2 が必要な理由: users.email には UNIQUE 制約があるため、同一メールの行が既にあると
// 3 の INSERT は unique violation で失敗し、ログイン自体が 500 になる。
// メール／パスワード認証から Google 認証へ移行したユーザーは auth.users の UUID が
// 変わりうるため、auth_user_id が NULL のケースだけでなく「別の UUID が入っている」
// ケースも張り替える必要がある。メールアドレスは Google が検証済みのものが JWT に
// 入るため、メール一致を本人性の根拠として扱ってよい。
func (r *UserRepository) GetOrCreateByAuthUserID(authUserID, email, name string) (*models.User, AuthLinkResult, error) {
	var user models.User
	scanUser := func(row *sql.Row) error {
		return row.Scan(
			&user.ID, &user.Name, &user.Email,
			&user.Description, &user.FavoriteAnime, &user.SnsLink,
			&user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
		)
	}

	// 1. Existing Supabase user (most requests after migration)
	err := scanUser(r.db.QueryRow(`
		SELECT id, name, email, description, favorite_anime, sns_link,
		       auth_user_id, created_at, updated_at
		FROM users WHERE auth_user_id = $1
	`, authUserID))
	if err == nil {
		return &user, AuthLinkExisting, nil
	}
	if err != sql.ErrNoRows {
		return nil, AuthLinkExisting, fmt.Errorf("failed to get user by auth_user_id: %w", err)
	}

	// 2. Link the existing row for this email to the new auth_user_id.
	//    行ロック付きの CTE で読み取りと更新を1文にまとめ、同時ログイン時の競合を防ぐ。
	//    表示名はユーザーが変更している可能性があるため上書きしない（空のときだけ補完）。
	var previousAuthUserID *string
	err = r.db.QueryRow(`
		WITH target AS (
			SELECT id, auth_user_id AS previous_auth_user_id
			FROM users WHERE email = $2
			FOR UPDATE
		)
		UPDATE users u
		SET auth_user_id = $1,
		    name = CASE WHEN COALESCE(u.name, '') = '' THEN $3 ELSE u.name END,
		    updated_at = CURRENT_TIMESTAMP
		FROM target
		WHERE u.id = target.id
		RETURNING u.id, u.name, u.email, u.description, u.favorite_anime, u.sns_link,
		          u.auth_user_id, u.created_at, u.updated_at, target.previous_auth_user_id
	`, authUserID, email, name).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Description, &user.FavoriteAnime, &user.SnsLink,
		&user.AuthUserID, &user.CreatedAt, &user.UpdatedAt, &previousAuthUserID,
	)
	if err == nil {
		if previousAuthUserID == nil {
			return &user, AuthLinkClaimed, nil
		}
		return &user, AuthLinkRelinked, nil
	}
	if err != sql.ErrNoRows {
		return nil, AuthLinkExisting, fmt.Errorf("failed to link user by email: %w", err)
	}

	// 3. New user
	err = scanUser(r.db.QueryRow(`
		INSERT INTO users (name, email, auth_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (auth_user_id) WHERE auth_user_id IS NOT NULL
		DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, updated_at = CURRENT_TIMESTAMP
		RETURNING id, name, email, description, favorite_anime, sns_link,
		          auth_user_id, created_at, updated_at
	`, name, email, authUserID))
	if err != nil {
		// auth.users への外部キー違反 = そのアカウントは削除済み。
		// 退会直後の JWT はまだ有効期限内のため、このパスに入りうる。
		// 500 ではなく 401 を返せるよう、呼び出し側が判別できる形にする。
		if isForeignKeyViolation(err) {
			return nil, AuthLinkExisting, ErrAuthUserNotFound
		}
		return nil, AuthLinkExisting, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, AuthLinkCreated, nil
}

// isForeignKeyViolation reports whether err is a PostgreSQL foreign key violation (SQLSTATE 23503).
func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23503"
	}
	return false
}

func (r *UserRepository) Create(user *models.User) error {
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, auth_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`, user.Name, user.Email, user.AuthUserID).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(user *models.User) error {
	err := r.db.QueryRow(`
		UPDATE users SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, name, email, description, created_at, updated_at
	`, user.Name, user.Description, user.ID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func addFieldToUpdate(setParts *[]string, args *[]interface{}, argIndex *int, fieldName string, value *string) {
	if value != nil {
		*setParts = append(*setParts, fmt.Sprintf("%s = $%d", fieldName, *argIndex))
		*args = append(*args, *value)
		*argIndex++
	}
}

func (r *UserRepository) UpdatePartial(userID int, req models.UserUpdateRequest) (*models.User, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	addFieldToUpdate(&setParts, &args, &argIndex, "name", req.Name)
	addFieldToUpdate(&setParts, &args, &argIndex, "description", req.Description)
	addFieldToUpdate(&setParts, &args, &argIndex, "favorite_anime", req.FavoriteAnime)
	addFieldToUpdate(&setParts, &args, &argIndex, "sns_link", req.SnsLink)

	if len(setParts) == 0 {
		return r.GetByID(userID)
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, name, email, description, favorite_anime, sns_link,
		          auth_user_id, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var user models.User
	err := r.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description,
		&user.FavoriteAnime, &user.SnsLink, &user.AuthUserID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}
