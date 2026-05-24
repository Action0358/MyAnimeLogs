package middleware

import (
	"net/http"
	"strings"

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
	// Fall back to email prefix
	if claims.Email != "" {
		if idx := strings.Index(claims.Email, "@"); idx > 0 {
			return claims.Email[:idx]
		}
	}
	return "ユーザー"
}

// validateSupabaseJWT parses and validates a Supabase-issued JWT.
func validateSupabaseJWT(tokenStr, jwtSecret string) (*SupabaseClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SupabaseClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SupabaseClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	return claims, nil
}

// SupabaseAuthMiddleware validates the Supabase JWT and injects user_id (internal int)
// into the Gin context. If the user does not yet have a row in public.users, one is
// created automatically on first access.
func SupabaseAuthMiddleware(db *database.DB, jwtSecret string, logger *common.Logger) gin.HandlerFunc {
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

		authUserID := claims.Subject
		email := claims.Email
		name := nameFromClaims(claims)

		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.GetOrCreateByAuthUserID(authUserID, email, name)
		if err != nil {
			logger.Error("failed to get or create user", "auth_user_id", authUserID, "error", err)
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{
				Error:   "server_error",
				Message: "ユーザー情報の取得に失敗しました",
				Code:    "SERVER_ERROR",
			})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user_name", user.Name)
		c.Next()
	}
}

// OptionalSupabaseAuthMiddleware validates the Supabase JWT when present but does
// not reject requests that have no Authorization header.
func OptionalSupabaseAuthMiddleware(db *database.DB, jwtSecret string, logger *common.Logger) gin.HandlerFunc {
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

		authUserID := claims.Subject
		email := claims.Email
		name := nameFromClaims(claims)

		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.GetOrCreateByAuthUserID(authUserID, email, name)
		if err != nil {
			logger.Error("failed to get or create user (optional auth)", "auth_user_id", authUserID, "error", err)
			c.Next()
			return
		}
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user_name", user.Name)
		c.Next()
	}
}
