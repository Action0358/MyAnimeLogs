package services

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupabaseAdminClient_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		key      string
		expected bool
	}{
		{"両方設定済み", "http://localhost:54321", "service-role-key", true},
		{"URLのみ", "http://localhost:54321", "", false},
		{"キーのみ", "", "service-role-key", false},
		{"どちらも未設定", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSupabaseAdminClient(tt.baseURL, tt.key)
			assert.Equal(t, tt.expected, c.IsConfigured())
		})
	}
}

func TestSupabaseAdminClient_DeleteUser_Success(t *testing.T) {
	var gotPath, gotMethod, gotAPIKey, gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotAPIKey = r.Header.Get("apikey")
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewSupabaseAdminClient(server.URL, "test-service-role-key")
	err := c.DeleteUser("auth-uuid-1")

	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/auth/v1/admin/users/auth-uuid-1", gotPath)
	// Admin API は apikey と Authorization の両方を要求する
	assert.Equal(t, "test-service-role-key", gotAPIKey)
	assert.Equal(t, "Bearer test-service-role-key", gotAuth)
}

// 既に存在しない場合は、呼び出し側から見て削除済みと同義
func TestSupabaseAdminClient_DeleteUser_NotFoundIsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewSupabaseAdminClient(server.URL, "key")
	assert.NoError(t, c.DeleteUser("missing-user"))
}

func TestSupabaseAdminClient_DeleteUser_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer server.Close()

	c := NewSupabaseAdminClient(server.URL, "key")
	err := c.DeleteUser("auth-uuid-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// service_role キーがエラーメッセージ経由で漏れないことを保証する
func TestSupabaseAdminClient_DeleteUser_ErrorDoesNotLeakKey(t *testing.T) {
	const secret = "super-secret-service-role-key"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden"}`))
	}))
	defer server.Close()

	c := NewSupabaseAdminClient(server.URL, secret)
	err := c.DeleteUser("auth-uuid-1")

	require.Error(t, err)
	assert.NotContains(t, err.Error(), secret)
}

func TestSupabaseAdminClient_DeleteUser_NotConfigured(t *testing.T) {
	c := NewSupabaseAdminClient("", "")
	err := c.DeleteUser("auth-uuid-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestSupabaseAdminClient_DeleteUser_EmptyAuthUserID(t *testing.T) {
	c := NewSupabaseAdminClient("http://localhost:54321", "key")
	err := c.DeleteUser("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

// baseURL の末尾スラッシュがあってもパスが二重にならないこと
func TestSupabaseAdminClient_TrimsTrailingSlash(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewSupabaseAdminClient(strings.TrimSuffix(server.URL, "/")+"/", "key")
	require.NoError(t, c.DeleteUser("auth-uuid-1"))
	assert.Equal(t, "/auth/v1/admin/users/auth-uuid-1", gotPath)
}
