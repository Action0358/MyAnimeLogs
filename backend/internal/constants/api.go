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

	// BackgroundTaskWorkerPoolSize - バックグラウンドタスク用ワーカープールのサイズ
	// 高負荷時のgoroutine爆発を防ぐため、同時実行数を制限
	BackgroundTaskWorkerPoolSize = 10

	// BackgroundTaskQueueSize - バックグラウンドタスクキューのバッファサイズ
	// キューが満杯の場合、新しいタスクはドロップされる
	BackgroundTaskQueueSize = 100

	// BackgroundTaskDBTimeout - バックグラウンドタスクのDB操作タイムアウト
	// DBが応答しない場合にワーカーがハングするのを防ぐ
	BackgroundTaskDBTimeout = 10 * time.Second
)

// APIレスポンスキャッシュ関連の定数
const (
	// CurrentSeasonAnimeCacheTTL - 今期アニメキャッシュの有効期限
	// 頻繁に更新されないため、30分キャッシュ
	CurrentSeasonAnimeCacheTTL = 30 * time.Minute

	// PopularAnimeCacheTTL - 人気アニメキャッシュの有効期限
	// 頻繁に更新されないため、30分キャッシュ
	PopularAnimeCacheTTL = 30 * time.Minute

	// FilterOptionsCacheTTL - フィルターオプションキャッシュの有効期限
	// ジャンル・年・シーズン一覧は1時間キャッシュ
	FilterOptionsCacheTTL = 1 * time.Hour

	// SearchResultCacheTTL - 検索結果キャッシュの有効期限
	// 検索結果は10分キャッシュ
	SearchResultCacheTTL = 10 * time.Minute

	// CacheCleanupInterval - キャッシュクリーンアップ間隔
	CacheCleanupInterval = 5 * time.Minute

	// CacheMaxItems - キャッシュの最大アイテム数
	CacheMaxItems = 1000
)

// キャッシュキープレフィックス
const (
	CacheKeyCurrentSeason   = "current_season:"
	CacheKeyPopular         = "popular:"
	CacheKeyFilterOptions   = "filter_options"
	CacheKeySearch          = "search:"
	CacheKeyAdvSearch       = "adv_search:"
	CacheKeyPopularSearches = "popular_searches:"
)

// 人気検索キャッシュ関連の定数
const (
	// PopularSearchesCacheTTL - 人気検索キャッシュの有効期限
	// 10分キャッシュ（頻繁に変わらないが、新しい検索トレンドを反映）
	PopularSearchesCacheTTL = 10 * time.Minute
)

// Annict API関連の定数
const (
	// PopularAnimeLimit - 人気アニメの取得上限
	// パフォーマンスとユーザー体験のバランスを考慮してTop100に制限
	PopularAnimeLimit = 100

	// AnnictAPIPerPage - Annict APIの1ページあたりの取得件数
	AnnictAPIPerPage = 50
)

// バッチ処理関連の定数
const (
	// BatchSyncMaxErrors - バッチ同期処理で収集するエラーメッセージの上限
	// ログの肥大化を防ぎつつ、デバッグに必要な情報を保持
	BatchSyncMaxErrors = 50

	// CoverImageSyncWorkerPoolSize - カバー画像同期用ワーカープールのサイズ
	// AniListのレートリミッターと連携して並列処理を行う
	CoverImageSyncWorkerPoolSize = 5
)

// 同期処理タイムアウト関連の定数
const (
	// EpisodeSyncTimeout - エピソード同期処理の最大実行時間
	// Annict + Jikan両方のエピソード同期に適用
	EpisodeSyncTimeout = 3 * time.Hour
)

// 検索履歴関連の定数
const (
	// SearchHistoryMaxItems - ユーザーごとの検索履歴の最大保持件数
	SearchHistoryMaxItems = 50

	// SearchHistoryDefaultLimit - 検索履歴取得時のデフォルト件数
	SearchHistoryDefaultLimit = 20
)
