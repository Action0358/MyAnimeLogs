package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newAdminTestRouter builds a minimal router whose /admin route is guarded by
// RequireAdmin. isAdmin controls whether the upstream auth stub marks the
// request as an admin; a nil value simulates an unauthenticated request where
// "is_admin" was never set in the context.
func newAdminTestRouter(isAdmin *bool) *gin.Engine {
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		if isAdmin != nil {
			c.Set("is_admin", *isAdmin)
		}
	}, RequireAdmin(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestRequireAdmin(t *testing.T) {
	adminTrue := true
	adminFalse := false

	tests := []struct {
		name       string
		isAdmin    *bool
		wantStatus int
	}{
		{"admin user passes", &adminTrue, http.StatusOK},
		{"non-admin user is forbidden", &adminFalse, http.StatusForbidden},
		{"missing is_admin is forbidden", nil, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newAdminTestRouter(tt.isAdmin)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestIsAdminFromClaims(t *testing.T) {
	tests := []struct {
		name   string
		claims *SupabaseClaims
		want   bool
	}{
		{
			name:   "app_metadata.role == admin",
			claims: &SupabaseClaims{AppMetadata: map[string]interface{}{"role": "admin"}},
			want:   true,
		},
		{
			name:   "app_metadata.role != admin",
			claims: &SupabaseClaims{AppMetadata: map[string]interface{}{"role": "user"}},
			want:   false,
		},
		{
			name:   "top-level authenticated role does not grant admin",
			claims: &SupabaseClaims{Role: "authenticated"},
			want:   false,
		},
		{
			name:   "nil app_metadata",
			claims: &SupabaseClaims{},
			want:   false,
		},
		{
			name:   "role present but not a string",
			claims: &SupabaseClaims{AppMetadata: map[string]interface{}{"role": 1}},
			want:   false,
		},
		{
			name:   "nil claims",
			claims: nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAdminFromClaims(tt.claims))
		})
	}
}
