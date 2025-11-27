package constants

import "time"

// AniList API関連の定数
const (
	// AniListAPIRateLimit - AniList APIのレート制限（リクエスト/分）
	// 公式ドキュメント: https://anilist.gitbook.io/anilist-apiv2-docs/overview/rate-limiting
	// 90 requests per minute
	AniListAPIRateLimit = 90

	// AniListAPIRateLimitDelay - AniList APIのレート制限を回避するための待機時間
	// 注意: この定数は後方互換性のために残されていますが、
	// 実際のレート制限は AniListClient の内部レートリミッター（golang.org/x/time/rate）で管理されています。
	// レートリミッターはトークンバケット方式を使用し、複数のgoroutineからの同時リクエストを適切に制御します。
	//
	// 計算: 60秒 / 90リクエスト = 666ms
	// 安全マージンを考慮して700msに設定（約85リクエスト/分）
	AniListAPIRateLimitDelay = 700 * time.Millisecond
)
