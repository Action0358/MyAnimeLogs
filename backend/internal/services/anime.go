package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
	"myanimelogs/internal/database"
)

var (
	ErrAnimeNotFound         = errors.New("anime not found")
	ErrSearchHistoryNotFound = errors.New("search history item not found")
)

// animeSelectColumns - アニメ取得用の共通SELECT句
// animes_with_computed_status VIEWを使用してステータスを動的に計算
// VIEWでcover_image_*カラムに対してCOALESCEが適用済み
const animeSelectColumns = `
	id, title, image_url, season, computed_status as status,
	mal_anime_id, annict_id, created_at, updated_at, watchers_count,
	cover_image_extra_large, cover_image_large, cover_image_medium,
	released_on, media
`

// aliasedAnimeSelectColumns - テーブルエイリアス付きアニメ取得用SELECT句
// AdvancedSearchAnimesなどのJOINを含むクエリで使用
// VIEWでcover_image_*カラムに対してCOALESCEが適用済み
const aliasedAnimeSelectColumns = `
	a.id, a.title, a.image_url, a.season, a.computed_status as status,
	a.mal_anime_id, a.annict_id, a.created_at, a.updated_at, a.watchers_count,
	a.cover_image_extra_large, a.cover_image_large, a.cover_image_medium,
	a.released_on, a.media
`

// aliasedAnimeGroupByColumns - テーブルエイリアス付きGROUP BY句
// AdvancedSearchAnimesなどのJOINを含むクエリで使用
const aliasedAnimeGroupByColumns = `
	a.id, a.title, a.image_url, a.season, a.computed_status,
	a.mal_anime_id, a.annict_id, a.created_at, a.updated_at, a.watchers_count,
	a.cover_image_extra_large, a.cover_image_large, a.cover_image_medium, a.released_on, a.media
`

type AnimeService struct {
	db                  *database.DB
	logger              *common.Logger
	anilistClient       *AniListClient
	jikanClient         *JikanClient
	tmdbClient          *TMDBClient   // TMDB APIクライアント
	claudeClient        *ClaudeClient // Claude APIクライアント（あらすじの日本語翻訳・整形）
	annictClient        *AnnictClient // Annict APIクライアント（循環参照回避のため後から設定）
	backgroundTaskQueue chan func()   // バックグラウンドタスク用ワーカープールのキュー
	cache               *MemoryCache  // レスポンスキャッシュ
	redisCache          *RedisCache   // Redisキャッシュ（オプション）
}

// SetAnnictClient sets the AnnictClient on AnimeService (called after both are initialized to avoid circular dependency)
func (s *AnimeService) SetAnnictClient(ac *AnnictClient) {
	s.annictClient = ac
}

// GetAniListClient returns the shared AniListClient instance.
// Used by the scheduler to share the same rate-limited client instead of creating a separate instance.
func (s *AnimeService) GetAniListClient() *AniListClient {
	return s.anilistClient
}

func NewAnimeService(db *database.DB, logger *common.Logger) *AnimeService {
	// キャッシュを初期化
	cacheConfig := CacheConfig{
		DefaultTTL:      constants.SearchResultCacheTTL,
		CleanupInterval: constants.CacheCleanupInterval,
		MaxItems:        constants.CacheMaxItems,
	}

	// Redisキャッシュを初期化（オプション - 接続失敗時はnilのまま）
	var redisCache *RedisCache
	if rc, err := NewRedisCache(logger); err != nil {
		logger.Warn("Redis cache not available, using memory cache only", "error", err)
	} else {
		redisCache = rc
	}

	s := &AnimeService{
		db:                  db,
		logger:              logger,
		anilistClient:       NewAniListClient(),
		jikanClient:         NewJikanClient(logger),
		tmdbClient:          NewTMDBClient(logger),
		claudeClient:        NewClaudeClient(logger),
		backgroundTaskQueue: make(chan func(), constants.BackgroundTaskQueueSize),
		cache:               NewMemoryCache(cacheConfig),
		redisCache:          redisCache,
	}

	// ワーカープールを起動
	s.startBackgroundTaskWorkers()

	return s
}

// startBackgroundTaskWorkers - バックグラウンドタスク用ワーカープールを起動
// 高負荷時のgoroutine爆発を防ぐため、固定数のワーカーでタスクを処理
func (s *AnimeService) startBackgroundTaskWorkers() {
	for i := 0; i < constants.BackgroundTaskWorkerPoolSize; i++ {
		go func() {
			for task := range s.backgroundTaskQueue {
				// パニックからの復帰（アプリケーション全体のクラッシュを防止）
				func() {
					defer func() {
						if r := recover(); r != nil {
							s.logger.Error("Panic recovered in image update worker",
								"panic", r)
						}
					}()
					task()
				}()
			}
		}()
	}
}

// enqueueBackgroundTask - バックグラウンドタスクをキューに追加
// キューが満杯の場合はタスクをドロップ（ノンブロッキング）
func (s *AnimeService) enqueueBackgroundTask(task func()) {
	select {
	case s.backgroundTaskQueue <- task:
		// タスクをキューに追加成功
	default:
		// キューが満杯の場合はドロップ（次回リクエスト時に再試行される）
		s.logger.Debug("Background task queue is full, task dropped")
	}
}

// Close - AnimeServiceのリソースを解放
// ワーカープールを graceful にシャットダウンし、goroutineリークを防止
func (s *AnimeService) Close() {
	if s.backgroundTaskQueue != nil {
		close(s.backgroundTaskQueue)
	}
	if s.cache != nil {
		s.cache.Close()
	}
	if s.redisCache != nil {
		s.redisCache.Close()
	}
}

// ClearAnimeCache - アニメ関連のキャッシュをクリア
// データ同期後に呼び出して、最新データを反映させる
// メモリキャッシュとRedisキャッシュを並行してクリアする
func (s *AnimeService) ClearAnimeCache() {
	s.logger.Info("Clearing anime cache after data sync")

	// クリア対象のキー（完全一致）
	exactKeys := []string{
		constants.CacheKeyFilterOptions,
	}

	// クリア対象のプレフィックス
	prefixKeys := []string{
		constants.CacheKeyPopularSearches,
		constants.CacheKeyCurrentSeason,
		constants.CacheKeyPopular,
		constants.CacheKeySearch,
		constants.CacheKeyAdvSearch,
		// 同期でtmdb_id / series_tmdb_id が更新されると配信サービスの取得結果も
		// 変わりうるため、あわせて捨てる
		constants.CacheKeyWatchProviders,
		// 同期でMAL IDや作品情報が更新されると再取得で成功しうるため、
		// 「取得できなかった」記録も捨てて再挑戦させる
		constants.CacheKeyMovieDurationMissing,
	}

	var wg sync.WaitGroup

	// メモリキャッシュをクリア（goroutine）
	if s.cache != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, key := range exactKeys {
				s.cache.Delete(key)
			}
			for _, prefix := range prefixKeys {
				s.cache.DeleteByPrefix(prefix)
			}
			s.logger.Debug("Memory cache cleared")
		}()
	}

	// Redisキャッシュをクリア（goroutine）
	if s.redisCache != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			for _, key := range exactKeys {
				if err := s.redisCache.Delete(ctx, key); err != nil {
					s.logger.Warn("Failed to clear key from Redis", "key", key, "error", err)
				}
			}
			for _, prefix := range prefixKeys {
				if err := s.redisCache.DeleteByPrefix(ctx, prefix); err != nil {
					s.logger.Warn("Failed to clear prefix from Redis", "prefix", prefix, "error", err)
				}
			}
			s.logger.Debug("Redis cache cleared")
		}()
	}

	// 両方の完了を待つ
	wg.Wait()

	s.logger.Info("Anime cache cleared successfully")
}
