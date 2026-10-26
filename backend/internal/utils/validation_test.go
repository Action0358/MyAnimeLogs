package utils

import (
	"testing"
)

func TestValidateImageURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "空文字列は許可",
			url:     "",
			wantErr: false,
		},
		{
			name:    "Imgur URL",
			url:     "https://i.imgur.com/abc123.png",
			wantErr: false,
		},
		{
			name:    "GitHub Avatars URL",
			url:     "https://avatars.githubusercontent.com/u/123456",
			wantErr: false,
		},
		{
			name:    "Gravatar URL",
			url:     "https://secure.gravatar.com/avatar/abc123",
			wantErr: false,
		},
		{
			name:    "Cloudinary URL",
			url:     "https://res.cloudinary.com/demo/image/upload/sample.jpg",
			wantErr: false,
		},
		{
			name:    "HTTPは拒否",
			url:     "http://i.imgur.com/abc123.png",
			wantErr: true,
		},
		{
			name:    "許可されていないドメイン",
			url:     "https://example.com/image.png",
			wantErr: true,
		},
		{
			name:    "無効なURL形式",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "悪意のあるURL",
			url:     "https://malicious-site.com/phishing.png",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateImageURLSize(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "通常のURL",
			url:     "https://i.imgur.com/abc123.png",
			wantErr: false,
		},
		{
			name:    "長いURL（2048文字以下）",
			url:     "https://example.com/" + string(make([]byte, 2000)),
			wantErr: false,
		},
		{
			name:    "長すぎるURL（2048文字超）",
			url:     "https://example.com/" + string(make([]byte, 2100)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageURLSize(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageURLSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "空文字列は許可",
			url:     "",
			wantErr: false,
		},
		{
			name:    "有効なHTTPS URL",
			url:     "https://twitter.com/username",
			wantErr: false,
		},
		{
			name:    "有効なHTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "GitHub URL",
			url:     "https://github.com/username",
			wantErr: false,
		},
		{
			name:    "Twitter URL",
			url:     "https://twitter.com/username",
			wantErr: false,
		},
		{
			name:    "無効なURL形式",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "スキームなし",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "無効なスキーム",
			url:     "ftp://example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeExternalURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "https はそのまま", raw: "https://example.com/anime", want: "https://example.com/anime"},
		{name: "http はそのまま", raw: "http://example.com", want: "http://example.com"},
		{name: "前後の空白は除く", raw: "  https://example.com  ", want: "https://example.com"},
		{name: "大文字のスキームも許可", raw: "HTTPS://example.com", want: "HTTPS://example.com"},
		{name: "空文字列", raw: "", want: ""},
		{name: "javascript: は捨てる", raw: "javascript:alert(document.cookie)", want: ""},
		{name: "大文字混じりの javascript: も捨てる", raw: "JavaScript:alert(1)", want: ""},
		{name: "前置空白付きの javascript: も捨てる", raw: "  javascript:alert(1)", want: ""},
		{name: "data: は捨てる", raw: "data:text/html,<script>alert(1)</script>", want: ""},
		{name: "vbscript: は捨てる", raw: "vbscript:msgbox(1)", want: ""},
		{name: "ホスト無しは捨てる", raw: "https://", want: ""},
		{name: "プロトコル相対URLは捨てる", raw: "//evil.example", want: ""},
		{name: "相対パスは捨てる", raw: "/anime/1", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeExternalURL(tt.raw); got != tt.want {
				t.Errorf("SanitizeExternalURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
