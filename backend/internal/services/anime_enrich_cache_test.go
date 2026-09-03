package services

import (
	"context"
	"testing"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"

	"github.com/stretchr/testify/assert"
)

// キャッシュキーは ClearAnimeCache がプレフィックス一致で消す対象でもあるため、
// constants 側のプレフィックスから外れていないことを確認する。
func TestEnrichCacheKeys(t *testing.T) {
	assert.Equal(t, "watch_providers:9908", watchProvidersCacheKey(9908))
	assert.Equal(t, "movie_duration_missing:9907", enrichMissingCacheKey(constants.CacheKeyMovieDurationMissing, 9907))
	assert.Equal(t, "broadcast_info_missing:9908", enrichMissingCacheKey(constants.CacheKeyBroadcastInfoMissing, 9908))
	assert.Equal(t, "released_on_missing:9909", enrichMissingCacheKey(constants.CacheKeyReleasedOnMissing, 9909))
}

// 取得できなかった項目のマーカーは項目ごとに別のキーでなければならない。
// 同じキーを共有すると、片方の欠損でもう片方の取得まで止まる。
func TestEnrichMissingCacheKeys_AreDistinctPerField(t *testing.T) {
	keys := map[string]bool{}
	for _, prefix := range []string{
		constants.CacheKeyMovieDurationMissing,
		constants.CacheKeyBroadcastInfoMissing,
		constants.CacheKeyReleasedOnMissing,
	} {
		key := enrichMissingCacheKey(prefix, 1)
		assert.False(t, keys[key], "キーが重複している: %s", key)
		keys[key] = true
	}
	assert.Len(t, keys, 3)
}

// 一時的な失敗はデータ不在より短く切り上げ、再挑戦の余地を残す。
func TestJikanEnrichCacheTTLs(t *testing.T) {
	assert.Less(t, constants.JikanEnrichErrorCacheTTL, constants.JikanEnrichMissingCacheTTL)
}

// キャッシュを一切持たない構成でも panic せず「キャッシュ無し」として
// 素通りすることを確認する。
func TestEnrichCacheHelpers_NoCache(t *testing.T) {
	s := &AnimeService{logger: common.NewLogger(common.LogLevelDebug, "text")}
	ctx := context.Background()

	assert.False(t, s.isEnrichKnownMissing(ctx, constants.CacheKeyBroadcastInfoMissing, 9907))

	providers, found := s.getCachedWatchProviders(ctx, 9908)
	assert.False(t, found)
	assert.Nil(t, providers)

	// 書き込み側も同様にno-opであること
	assert.NotPanics(t, func() {
		s.markEnrichMissing(ctx, constants.CacheKeyBroadcastInfoMissing, 9907, 0)
		s.cacheWatchProviders(ctx, 9908, nil)
	})
}

// newMemoryOnlyService は Redis が無い（STG・本番と同じ）構成を作る。
func newMemoryOnlyService() *AnimeService {
	return &AnimeService{
		logger: common.NewLogger(common.LogLevelDebug, "text"),
		cache: NewMemoryCache(CacheConfig{
			DefaultTTL:      time.Minute,
			CleanupInterval: time.Minute,
			MaxItems:        constants.CacheMaxItems,
		}),
	}
}

// Redisが無い環境でもメモリキャッシュにフォールバックすることを確認する（#803）。
// ここが効かないと、STG・本番で詳細取得のキャッシュがすべてno-opになる。
func TestEnrichCacheHelpers_FallsBackToMemory(t *testing.T) {
	s := newMemoryOnlyService()
	defer s.cache.Close()
	ctx := context.Background()

	// 取得できなかった記録が読み戻せること
	assert.False(t, s.isEnrichKnownMissing(ctx, constants.CacheKeyBroadcastInfoMissing, 9907))
	s.markEnrichMissing(ctx, constants.CacheKeyBroadcastInfoMissing, 9907, time.Minute)
	assert.True(t, s.isEnrichKnownMissing(ctx, constants.CacheKeyBroadcastInfoMissing, 9907))

	// 項目ごとにキーが分かれており、片方の欠損が他方に波及しないこと
	assert.False(t, s.isEnrichKnownMissing(ctx, constants.CacheKeyReleasedOnMissing, 9907))
	// 作品ごとにも分かれていること
	assert.False(t, s.isEnrichKnownMissing(ctx, constants.CacheKeyBroadcastInfoMissing, 9908))
}

// 配信サービスは「0件」も有効な値としてキャッシュする。
// 配信が無い作品ほどフォールバックまで走って遅くなるため、
// 0件をヒットとして扱えないと毎回TMDBを叩き直すことになる。
func TestCacheWatchProviders_EmptyResultIsCached(t *testing.T) {
	s := newMemoryOnlyService()
	defer s.cache.Close()
	ctx := context.Background()

	s.cacheWatchProviders(ctx, 9908, []models.WatchProvider{})

	providers, found := s.getCachedWatchProviders(ctx, 9908)
	assert.True(t, found, "0件でもキャッシュヒットとして扱う")
	assert.Empty(t, providers)
}

func TestCacheWatchProviders_RoundTrip(t *testing.T) {
	s := newMemoryOnlyService()
	defer s.cache.Close()
	ctx := context.Background()

	s.cacheWatchProviders(ctx, 9908, []models.WatchProvider{
		{ProviderID: 8, ProviderName: "Netflix", LogoURL: "https://example.com/n.png"},
	})

	providers, found := s.getCachedWatchProviders(ctx, 9908)
	assert.True(t, found)
	assert.Len(t, providers, 1)
	assert.Equal(t, "Netflix", providers[0].ProviderName)
}
