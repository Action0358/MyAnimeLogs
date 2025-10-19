package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// 許可する画像ホスティングドメイン
var allowedImageDomains = []string{
	"i.imgur.com",
	"imgur.com",
	"avatars.githubusercontent.com",
	"github.com",
	"raw.githubusercontent.com",
	"gravatar.com",
	"secure.gravatar.com",
	"cloudinary.com",
	"res.cloudinary.com",
	"anilist.co",
	"s4.anilist.co",
	"cdn.myanimelist.net",
	"pbs.twimg.com",
	"twimg.com",
}

// ValidateImageURL は画像URLが信頼できるドメインからのものかを検証します
func ValidateImageURL(imageURL string) error {
	// 空文字列は許可（アイコンなし）
	if imageURL == "" {
		return nil
	}

	// URLをパース
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("無効なURL形式です")
	}

	// HTTPSのみ許可
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("HTTPSのURLのみ許可されています")
	}

	// ホスト名を取得
	hostname := strings.ToLower(parsedURL.Hostname())

	// 許可されたドメインかチェック
	for _, allowedDomain := range allowedImageDomains {
		// 完全一致またはサブドメイン一致
		if hostname == allowedDomain || strings.HasSuffix(hostname, "."+allowedDomain) {
			return nil
		}
	}

	return fmt.Errorf("許可されていないドメインです。信頼できる画像ホスティングサービス（Imgur、GitHub、Gravatarなど）のURLを使用してください")
}

// ValidateImageURLSize は画像URLの長さを検証します
func ValidateImageURLSize(imageURL string) error {
	const maxURLLength = 2048
	if len(imageURL) > maxURLLength {
		return fmt.Errorf("URLが長すぎます（最大%d文字）", maxURLLength)
	}
	return nil
}

// ValidateURL は一般的なURLの形式を検証します
func ValidateURL(urlStr string) error {
	// 空文字列は許可
	if urlStr == "" {
		return nil
	}

	// URLをパース
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("無効なURL形式です")
	}

	// スキームとホストが必要
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("無効なURL形式です")
	}

	// HTTPまたはHTTPSのみ許可
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("HTTPまたはHTTPSのURLのみ許可されています")
	}

	return nil
}
