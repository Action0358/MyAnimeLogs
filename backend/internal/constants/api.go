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

// AniList APIリトライ関連の定数
const (
	// AniListMaxRetries - executeGraphQLQuery のリトライ上限
	// レート制限(429)はRetry-Afterを尊重して待機するため、分単位のリセット窓を
	// またげるよう従来の3回から増やす
	AniListMaxRetries = 5

	// AniListBaseRetryDelay - Retry-Afterが無い場合の指数バックオフ基準時間
	AniListBaseRetryDelay = 2 * time.Second

	// AniListMaxRetryBackoff - リトライ時の最大待機時間（Retry-After含む上限）
	// AniListのレート制限窓は1分のため、最大1分まで待機を許容する
	AniListMaxRetryBackoff = 60 * time.Second
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

	// AnimeDetailCacheTTL - アニメ詳細（ユーザー固有部分を除く）のキャッシュ有効期限
	// 初回アクセス時はAniList/Jikan/TMDBへの同期問い合わせで数秒かかることがあり、
	// 同じ作品への再訪・複数ユーザーからの同時アクセスでこれを繰り返さないために
	// レスポンス単位でキャッシュする（#851）。検索結果と同じ10分に揃える
	AnimeDetailCacheTTL = 10 * time.Minute

	// JikanEnrichMissingCacheTTL - Jikan側にデータが無かった項目を覚えておく期間
	// 毎回問い合わせても結果が変わらないため、詳細取得のたびに同期的な
	// リトライを繰り返さないよう記憶する
	JikanEnrichMissingCacheTTL = 24 * time.Hour

	// JikanEnrichErrorCacheTTL - Jikanの呼び出しが失敗した項目を覚えておく期間
	// レート制限や一時的な障害の可能性があるため、データ不在より短くする
	JikanEnrichErrorCacheTTL = 1 * time.Hour

	// WatchProvidersCacheTTL - 配信サービスキャッシュの有効期限
	// 詳細取得のたびにTMDBを同期的に叩くとレスポンスが1〜3秒伸びるため、
	// 作品単位でキャッシュする。配信状況の変化は日単位のため24時間で十分。
	WatchProvidersCacheTTL = 24 * time.Hour

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
	CacheKeyWatchProviders  = "watch_providers:"
	CacheKeyAnimeDetail     = "anime_detail:"
	// 以下はJikanから取得できなかった項目のマーカー。
	// キーの有無だけが意味を持ち、値は参照しない
	CacheKeyMovieDurationMissing = "movie_duration_missing:"
	CacheKeyBroadcastInfoMissing = "broadcast_info_missing:"
	CacheKeyReleasedOnMissing    = "released_on_missing:"
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

	// MetadataRetryCooldown - メタデータ同期失敗(sync_status='failed')作品を
	// 再試行対象に含めるまでのクールダウン期間。失敗直後の連続再試行を避けつつ、
	// レート制限で滞留した作品を次回以降のバッチで解消する
	MetadataRetryCooldown = 6 * time.Hour

	// MetadataRetryBatchLimit - 1回のバッチで再試行する失敗作品の上限
	MetadataRetryBatchLimit = 50

	// TMDBMetadataBatchLimit - 1回の定期同期で解決するTMDBメタデータ作品の上限
	// TMDBレート制限(40req/10秒)下でバッチ処理(40件/約11秒)するため、200件で約55秒。
	// 滞留分は次回以降の定期同期で段階的に解消する。
	TMDBMetadataBatchLimit = 200
)

// 同期処理タイムアウト関連の定数
const (
	// JikanRateLimitDelay - Jikan APIのレート制限対策の待機時間
	// Jikan APIは3リクエスト/秒の制限があるため、約2.8req/secに抑える
	JikanRateLimitDelay = 350 * time.Millisecond
)

// 検索履歴関連の定数
const (
	// SearchHistoryMaxItems - ユーザーごとの検索履歴の最大保持件数
	SearchHistoryMaxItems = 50

	// SearchHistoryDefaultLimit - 検索履歴取得時のデフォルト件数
	SearchHistoryDefaultLimit = 20
)
