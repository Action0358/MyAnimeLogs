package services

import (
	"context"
	"testing"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"

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

// Redisは任意構成のため、未設定の環境でもenrichがpanicせず
// 「キャッシュ無し」として素通りすることを確認する。
func TestEnrichCacheHelpers_NoRedis(t *testing.T) {
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
