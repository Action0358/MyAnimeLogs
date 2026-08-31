package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SupabaseClaims represents the JWT claims issued by Supabase Auth.
type SupabaseClaims struct {
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone"`
	Role         string                 `json:"role"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
	AppMetadata  map[string]interface{} `json:"app_metadata"`
	jwt.RegisteredClaims
}

// userIDCache stores the authUserID → internal userID mapping to avoid DB lookups on every request.
var userIDCache sync.Map

// InvalidateUserIDCache drops the cached mapping for the given Supabase auth UUID.
//
// アカウント削除時に必ず呼ぶこと。呼ばないと、削除済みユーザーの内部IDを
// キャッシュから返し続けてしまう。
func InvalidateUserIDCache(authUserID string) {
	userIDCache.Delete(authUserID)
}

// jwksCache caches EC public keys fetched from the Supabase JWKS endpoint.
type jwksCache struct {
	mu        sync.RWMutex
	keys      map[string]*ecdsa.PublicKey
	fetchedAt time.Time
	ttl       time.Duration
}

var globalJWKSCache = &jwksCache{
	keys: make(map[string]*ecdsa.PublicKey),
	ttl:  15 * time.Minute,
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

// jwksHTTPClient is a dedicated HTTP client with a timeout for JWKS fetches.
// Using the default http.Get (no timeout) risks blocking the middleware indefinitely
// if the Supabase auth endpoint is slow or unresponsive.
var jwksHTTPClient = &http.Client{Timeout: 5 * time.Second}

// fetchJWKS fetches EC public keys from the given JWKS URL.
func fetchJWKS(url string) (map[string]*ecdsa.PublicKey, error) {
	resp, err := jwksHTTPClient.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks jwksResponse
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	keys := make(map[string]*ecdsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "EC" || key.Alg != "ES256" {
			continue
		}
		xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
		if err != nil {
			continue
		}
		keys[key.Kid] = &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
	}

	return keys, nil
}

// getJWKSPublicKey returns the cached EC public key for the given kid,
// refreshing the cache from the JWKS endpoint when stale.
// Uses double-checked locking to prevent Thundering Herd: multiple goroutines
// waiting on the write lock would otherwise each trigger a redundant fetchJWKS.
func getJWKSPublicKey(kid, jwksURL string) (*ecdsa.PublicKey, error) {
	if kid == "" {
		return nil, fmt.Errorf("token header missing kid")
	}

	isFresh := func() bool { return time.Since(globalJWKSCache.fetchedAt) < globalJWKSCache.ttl }

	globalJWKSCache.mu.RLock()
	if isFresh() {
		key, ok := globalJWKSCache.keys[kid]
		globalJWKSCache.mu.RUnlock()
		if ok {
			return key, nil
		}
		// Cache is fresh but kid is unknown — do not fetch; reject immediately.
		return nil, fmt.Errorf("key with kid %q not found in JWKS", kid)
	}
	globalJWKSCache.mu.RUnlock()

	globalJWKSCache.mu.Lock()
	defer globalJWKSCache.mu.Unlock()

	// Re-check after acquiring write lock: another goroutine may have already refreshed.
	if isFresh() {
		if key, ok := globalJWKSCache.keys[kid]; ok {
			return key, nil
		}
		return nil, fmt.Errorf("key with kid %q not found in JWKS", kid)
	}

	keys, err := fetchJWKS(jwksURL)
	if err != nil {
		return nil, err
	}

	globalJWKSCache.keys = keys
	globalJWKSCache.fetchedAt = time.Now()

	key, ok := keys[kid]
	if !ok {
		return nil, fmt.Errorf("key with kid %q not found in JWKS", kid)
	}
	return key, nil
}

// isAdminFromClaims reports whether the Supabase JWT grants the admin role.
// The role is read from app_metadata, which can only be set server-side
// (via the Supabase Admin API), never by the user's self-service signup —
// so it is safe to trust as an authorization signal.
func isAdminFromClaims(claims *SupabaseClaims) bool {
	if claims == nil || claims.AppMetadata == nil {
		return false
	}
	if role, ok := claims.AppMetadata["role"].(string); ok && strings.ToLower(role) == "admin" {
		return true
	}
	return false
}

// avatarURLFromClaims extracts the user's avatar image URL from Supabase JWT metadata.
// Google 認証では avatar_url または picture にプロフィール画像URLが入る（プロバイダ依存のため両対応）。
func avatarURLFromClaims(claims *SupabaseClaims) string {
	if claims.UserMetadata == nil {
		return ""
	}
	if u, ok := claims.UserMetadata["avatar_url"].(string); ok && u != "" {
		return u
	}
	if u, ok := claims.UserMetadata["picture"].(string); ok && u != "" {
		return u
	}
	return ""
}

// nameFromClaims extracts a display name from Supabase JWT metadata.
func nameFromClaims(claims *SupabaseClaims) string {
	if claims.UserMetadata != nil {
		if n, ok := claims.UserMetadata["name"].(string); ok && n != "" {
			return n
		}
		if n, ok := claims.UserMetadata["full_name"].(string); ok && n != "" {
			return n
		}
	}
	if claims.Email != "" {
		if idx := strings.Index(claims.Email, "@"); idx > 0 {
			return claims.Email[:idx]
		}
	}
	return "ユーザー"
}

// validateSupabaseJWT parses and validates a Supabase-issued JWT.
// Supports both HS256 (cloud Supabase) and ES256 (new local CLI) signing methods.
func validateSupabaseJWT(tokenStr, jwtSecret string) (*SupabaseClaims, error) {
	// Peek at the header without full validation to determine the signing algorithm.
	unverified, _, err := jwt.NewParser().ParseUnverified(tokenStr, &SupabaseClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	switch unverified.Method.Alg() {
	case "ES256":
		return validateES256JWT(tokenStr)
	default:
		return validateHS256JWT(tokenStr, jwtSecret)
	}
}

// validateHS256JWT validates a Supabase JWT signed with HS256 (cloud Supabase).
func validateHS256JWT(tokenStr, jwtSecret string) (*SupabaseClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&SupabaseClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		},
		jwt.WithAudience("authenticated"),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*SupabaseClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	return claims, nil
}

// validateES256JWT validates a Supabase JWT signed with ES256 (new local Supabase CLI).
// Public keys are fetched from the Supabase JWKS endpoint and cached for 15 minutes.
func validateES256JWT(tokenStr string) (*SupabaseClaims, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = "http://localhost:54321"
	}
	jwksURL := supabaseURL + "/auth/v1/.well-known/jwks.json"

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&SupabaseClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			kid, _ := t.Header["kid"].(string)
			return getJWKSPublicKey(kid, jwksURL)
		},
		jwt.WithAudience("authenticated"),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*SupabaseClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	return claims, nil
}

// resolveUserID returns the internal user ID for the given Supabase auth UUID.
// Results are cached in userIDCache to avoid a DB round-trip on every request.
func resolveUserID(userRepo *repository.UserRepository, logger *common.Logger, authUserID, email, name string) (int, string, string, error) {
	if cached, ok := userIDCache.Load(authUserID); ok {
		id := cached.(int)
		return id, email, name, nil
	}
	user, linkResult, err := userRepo.GetOrCreateByAuthUserID(authUserID, email, name)
	if err != nil {
		return 0, "", "", err
	}
	// Google 移行に伴う紐づけ替えは、既存データの引き継ぎが正しく行われたかを
	// 後から追跡できるよう記録する。通常ログイン（existing）は件数が多いため出力しない。
	if linkResult != repository.AuthLinkExisting {
		logger.Info("linked users row to auth user",
			"result", linkResult.String(),
			"user_id", user.ID,
			"auth_user_id", authUserID,
		)
	}
	userIDCache.Store(authUserID, user.ID)
	return user.ID, user.Email, user.Name, nil
}

// SupabaseAuthMiddleware validates the Supabase JWT and injects user_id (internal int)
// into the Gin context. If the user does not yet have a row in public.users, one is
// created automatically on first access.
func SupabaseAuthMiddleware(db *database.DB, jwtSecret string, logger *common.Logger) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(db)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse{
				Error:   common.ErrMsgUnauthorized,
				Message: "認証トークンが必要です",
				Code:    common.ErrCodeUnauthorized,
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse{
				Error:   common.ErrMsgUnauthorized,
				Message: "無効な認証トークン形式です",
				Code:    common.ErrCodeUnauthorized,
			})
			c.Abort()
			return
		}

		claims, err := validateSupabaseJWT(parts[1], jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse{
				Error:   common.ErrMsgUnauthorized,
				Message: "無効または期限切れのトークンです",
				Code:    common.ErrCodeUnauthorized,
			})
			c.Abort()
			return
		}

		userID, email, name, err := resolveUserID(userRepo, logger, claims.Subject, claims.Email, nameFromClaims(claims))
		if err != nil {
			// 退会直後は JWT がまだ有効期限内のため、削除済みアカウントで
			// アクセスされうる。サーバー障害ではないので 401 を返す。
			if errors.Is(err, repository.ErrAuthUserNotFound) {
				logger.Info("request from deleted account", "auth_user_id", claims.Subject)
				c.JSON(http.StatusUnauthorized, common.ErrorResponse{
					Error:   common.ErrMsgUnauthorized,
					Message: "このアカウントは削除されています",
					Code:    common.ErrCodeUnauthorized,
				})
				c.Abort()
				return
			}
			logger.Error("failed to get or create user", "auth_user_id", claims.Subject, "error", err)
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{
				Error:   common.ErrMsgInternalServer,
				Message: "ユーザー情報の取得に失敗しました",
				Code:    common.ErrCodeInternalServer,
			})
			c.Abort()
			return
		}

		c.Set("auth_user_id", claims.Subject)
		c.Set("user_id", userID)
		c.Set("user_email", email)
		c.Set("user_name", name)
		c.Set("user_avatar_url", avatarURLFromClaims(claims))
		c.Set("is_admin", isAdminFromClaims(claims))
		c.Next()
	}
}

// OptionalSupabaseAuthMiddleware validates the Supabase JWT when present but does
// not reject requests that have no Authorization header.
func OptionalSupabaseAuthMiddleware(db *database.DB, jwtSecret string, logger *common.Logger) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(db)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := validateSupabaseJWT(parts[1], jwtSecret)
		if err != nil {
			c.Next()
			return
		}

		userID, email, name, err := resolveUserID(userRepo, logger, claims.Subject, claims.Email, nameFromClaims(claims))
		if err != nil {
			// 削除済みアカウントは障害ではないため ERROR で鳴らさない。
			// オプショナル認証なので、未ログイン扱いで処理を続行する。
			if errors.Is(err, repository.ErrAuthUserNotFound) {
				logger.Info("request from deleted account (optional auth)", "auth_user_id", claims.Subject)
			} else {
				logger.Error("failed to get or create user (optional auth)", "auth_user_id", claims.Subject, "error", err)
			}
			c.Next()
			return
		}
		c.Set("auth_user_id", claims.Subject)
		c.Set("user_id", userID)
		c.Set("user_email", email)
		c.Set("user_name", name)
		c.Set("user_avatar_url", avatarURLFromClaims(claims))
		c.Set("is_admin", isAdminFromClaims(claims))
		c.Next()
	}
}
