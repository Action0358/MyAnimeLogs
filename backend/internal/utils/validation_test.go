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
