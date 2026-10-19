package services

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SupabaseAdminClient calls the Supabase Auth Admin API with the service_role key.
//
// service_role キーは RLS を無視する強い権限を持つため、以下を厳守する。
//   - 値をログ・エラーメッセージ・APIレスポンスに含めない
//   - フロントエンドへ渡さない（Admin API の呼び出しは必ずバックエンド経由）
type SupabaseAdminClient struct {
	baseURL        string
	serviceRoleKey string
	httpClient     *http.Client
}

func NewSupabaseAdminClient(baseURL, serviceRoleKey string) *SupabaseAdminClient {
	return &SupabaseAdminClient{
		baseURL:        strings.TrimRight(baseURL, "/"),
		serviceRoleKey: serviceRoleKey,
		// 認証基盤への同期呼び出しのため、無期限に待たないようタイムアウトを設ける
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// IsConfigured reports whether the Admin API can be called.
// 未設定の環境では削除エンドポイントを 503 で返し、原因を明示できるようにする。
func (c *SupabaseAdminClient) IsConfigured() bool {
	return c.baseURL != "" && c.serviceRoleKey != ""
}

// DeleteUser removes a user from auth.users.
//
// public.users は fk_users_auth_user_id の ON DELETE CASCADE により連鎖削除され、
// さらに user_favorites / user_watch_status も連鎖する。
// search_logs も ON DELETE CASCADE により連鎖削除される（migration 054）。
func (c *SupabaseAdminClient) DeleteUser(authUserID string) error {
	if !c.IsConfigured() {
		return fmt.Errorf("supabase admin client is not configured")
	}
	if authUserID == "" {
		return fmt.Errorf("authUserID is required")
	}

	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", c.baseURL, authUserID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build delete user request: %w", err)
	}
	req.Header.Set("apikey", c.serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// err にリクエストURLは含まれるが、ヘッダ（=キー）は含まれない
		return fmt.Errorf("failed to call supabase admin api: %w", err)
	}
	defer resp.Body.Close()

	// 既に存在しない場合は、呼び出し側から見て削除済みと同義なので成功扱いにする
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("supabase admin api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
