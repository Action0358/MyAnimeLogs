package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testJWTSecret   = "test-jwt-secret"
	testSupabaseURL = "https://project.supabase.co"
	testIssuer      = testSupabaseURL + "/auth/v1"
)

func testClaims(issuer string) *SupabaseClaims {
	return &SupabaseClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   "auth-user",
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{"authenticated"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}
}

func signHS256(t *testing.T, issuer string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims(issuer)).SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return token
}

func TestValidateSupabaseJWT_Issuer(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		supabaseURL string
		issuer      string
		wantErr     error
	}{
		{"production: own issuer", "production", testSupabaseURL, testIssuer, nil},
		{"production: URL with trailing slash", "production", testSupabaseURL + "/", testIssuer, nil},
		{"production: other project", "production", testSupabaseURL, "https://other.supabase.co/auth/v1", jwt.ErrTokenInvalidIssuer},
		{"production: base URL only", "production", testSupabaseURL, testSupabaseURL, jwt.ErrTokenInvalidIssuer},
		{"production: missing issuer", "production", testSupabaseURL, "", jwt.ErrTokenRequiredClaimMissing},
		{"staging: own issuer", "staging", testSupabaseURL, testIssuer, nil},
		{"staging: other project", "staging", testSupabaseURL, "https://other.supabase.co/auth/v1", jwt.ErrTokenInvalidIssuer},
		{"staging: missing issuer", "staging", testSupabaseURL, "", jwt.ErrTokenRequiredClaimMissing},

		// ローカルは SUPABASE_URL と iss のホスト名が異なる
		{"development: docker host differs from issuer", "development", "http://host.docker.internal:54321", "http://127.0.0.1:54321/auth/v1", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENVIRONMENT", tt.environment)
			t.Setenv("SUPABASE_URL", tt.supabaseURL)

			claims, err := validateSupabaseJWT(signHS256(t, tt.issuer), testJWTSecret)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "auth-user", claims.Subject)
		})
	}
}

func TestValidateSupabaseJWT_IssuerDoesNotReplaceOtherChecks(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SUPABASE_URL", testSupabaseURL)

	t.Run("wrong audience", func(t *testing.T) {
		claims := testClaims(testIssuer)
		claims.Audience = jwt.ClaimStrings{"anon"}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testJWTSecret))
		require.NoError(t, err)

		_, err = validateSupabaseJWT(token, testJWTSecret)
		assert.ErrorIs(t, err, jwt.ErrTokenInvalidAudience)
	})

	t.Run("wrong signature", func(t *testing.T) {
		_, err := validateSupabaseJWT(signHS256(t, testIssuer), "other-secret")
		assert.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
	})
}

func TestValidateSupabaseJWT_IssuerWithES256(t *testing.T) {
	const kid = "test-kid"
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwksResponse{Keys: []jwkKey{{
			Kty: "EC",
			Kid: kid,
			Alg: "ES256",
			Crv: "P-256",
			X:   base64.RawURLEncoding.EncodeToString(key.X.FillBytes(make([]byte, 32))),
			Y:   base64.RawURLEncoding.EncodeToString(key.Y.FillBytes(make([]byte, 32))),
		}}})
	}))
	defer server.Close()

	// JWKS のキャッシュはパッケージ全体で共有するので、前後で空にする
	resetJWKSCache := func() {
		globalJWKSCache.mu.Lock()
		defer globalJWKSCache.mu.Unlock()
		globalJWKSCache.keys = make(map[string]*ecdsa.PublicKey)
		globalJWKSCache.fetchedAt = time.Time{}
	}
	resetJWKSCache()
	t.Cleanup(resetJWKSCache)

	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SUPABASE_URL", server.URL)

	sign := func(issuer string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, testClaims(issuer))
		token.Header["kid"] = kid
		signed, err := token.SignedString(key)
		require.NoError(t, err)
		return signed
	}

	_, err = validateSupabaseJWT(sign(server.URL+"/auth/v1"), testJWTSecret)
	assert.NoError(t, err)

	_, err = validateSupabaseJWT(sign("https://other.supabase.co/auth/v1"), testJWTSecret)
	assert.ErrorIs(t, err, jwt.ErrTokenInvalidIssuer)
}
