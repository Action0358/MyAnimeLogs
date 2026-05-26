package middleware

import "testing"

func TestIsOriginAllowed(t *testing.T) {
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://myanimelogs.com",
	}
	pagesDomains := []string{
		"myanimelogs.pages.dev",
		"myanimelogs-staging.pages.dev",
	}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		// 明示許可
		{"exact match: localhost dev", "http://localhost:3000", true},
		{"exact match: custom domain", "https://myanimelogs.com", true},

		// Cloudflare Pages apex
		{"apex: production pages", "https://myanimelogs.pages.dev", true},
		{"apex: staging pages", "https://myanimelogs-staging.pages.dev", true},

		// Cloudflare Pages preview / branch subdomains
		{"subdomain: production preview", "https://abc123.myanimelogs.pages.dev", true},
		{"subdomain: staging preview hash", "https://dbccfbdf.myanimelogs-staging.pages.dev", true},
		{"subdomain: staging branch alias", "https://staging.myanimelogs-staging.pages.dev", true},

		// 不一致系
		{"http for pages domain is rejected", "http://myanimelogs.pages.dev", false},
		{"unrelated domain", "https://evil.example.com", false},
		{"empty origin", "", false},
		{"prefix attack: pretend to end with allowed domain", "https://evilmyanimelogs.pages.dev", false},
		{"prefix attack: pretend to end with staging", "https://evilmyanimelogs-staging.pages.dev", false},
		{"close but not exact apex", "https://myanimelogs.pages.de", false},

		// サフィックス偽装: 許可ドメインを攻撃者ドメインのサブドメインとして配置
		{"suffix spoofing: allowed apex as subdomain of attacker", "https://myanimelogs.pages.dev.evil.com", false},
		{"suffix spoofing: allowed staging as subdomain of attacker", "https://myanimelogs-staging.pages.dev.evil.com", false},
		{"suffix spoofing: subdomain pattern under attacker", "https://abc.myanimelogs.pages.dev.evil.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOriginAllowed(tt.origin, allowedOrigins, pagesDomains)
			if got != tt.want {
				t.Errorf("isOriginAllowed(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

// BenchmarkIsOriginAllowed は CORS 判定がリクエストごとにアロケーションを発生させないことを担保する。
// go test -bench BenchmarkIsOriginAllowed -benchmem で allocs/op が 0 になることを確認する。
func BenchmarkIsOriginAllowed(b *testing.B) {
	allowedOrigins := []string{"http://localhost:3000"}
	pagesDomains := []string{"myanimelogs.pages.dev", "myanimelogs-staging.pages.dev"}

	cases := []string{
		"https://myanimelogs-staging.pages.dev",         // apex 一致
		"https://dbccfbdf.myanimelogs-staging.pages.dev", // サブドメイン一致
		"https://evil.example.com",                       // 不一致（ループ最後で false）
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isOriginAllowed(cases[i%len(cases)], allowedOrigins, pagesDomains)
	}
}
