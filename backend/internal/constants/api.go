package constants

import "time"

// AniList API関連の定数
const (
	// AniListAPIRateLimit - AniList APIのレート制限（リクエスト/分）
	// 公式ドキュメント: https://anilist.gitbook.io/anilist-apiv2-docs/overview/rate-limiting
	// 90 requests per minute
	AniListAPIRateLimit = 90

	// AniListAPIRateLimitSafetyMargin - レート制限の安全マージン（0.94 = 94%）
	// APIの制限（90リクエスト/分）の94%（約85リクエスト/分）に制限することで、
	// バーストやネットワーク遅延による制限超過を防ぐ
	AniListAPIRateLimitSafetyMargin = 0.94

	// AniListAPIRateLimitDelay - AniList APIのレート制限を回避するための待機時間
	// 注意: この定数は後方互換性のために残されていますが、
	// 実際のレート制限は AniListClient の内部レートリミッター（golang.org/x/time/rate）で管理されています。
	// レートリミッターはトークンバケット方式を使用し、複数のgoroutineからの同時リクエストを適切に制御します。
	//
	// 計算: 60秒 / 90リクエスト = 666ms
	// 安全マージンを考慮して700msに設定（約85リクエスト/分）
	AniListAPIRateLimitDelay = 700 * time.Millisecond
)

// キャッシュ関連の定数
const (
	// CoverImageCacheTTL - カバー画像キャッシュの有効期限
	// 7日間キャッシュを保持し、期限切れ後はAPIから再取得
	CoverImageCacheTTL = 7 * 24 * time.Hour

	// ImageUpdateWorkerPoolSize - 画像更新用ワーカープールのサイズ
	// 高負荷時のgoroutine爆発を防ぐため、同時実行数を制限
	ImageUpdateWorkerPoolSize = 10

	// ImageUpdateQueueSize - 画像更新キューのバッファサイズ
	// キューが満杯の場合、新しいタスクはドロップされる
	ImageUpdateQueueSize = 100
)
