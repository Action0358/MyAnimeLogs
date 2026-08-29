package services

import (
	"context"
	"testing"

	"myanimelogs/internal/common"

	"github.com/stretchr/testify/assert"
)

// キャッシュキーは ClearAnimeCache がプレフィックス一致で消す対象でもあるため、
// constants 側のプレフィックスから外れていないことを確認する。
func TestEnrichCacheKeys(t *testing.T) {
	assert.Equal(t, "watch_providers:9908", watchProvidersCacheKey(9908))
	assert.Equal(t, "movie_duration_missing:9907", movieDurationMissingCacheKey(9907))
}

// Redisは任意構成のため、未設定の環境でもenrichがpanicせず
// 「キャッシュ無し」として素通りすることを確認する。
func TestEnrichCacheHelpers_NoRedis(t *testing.T) {
	s := &AnimeService{logger: common.NewLogger(common.LogLevelDebug, "text")}
	ctx := context.Background()

	assert.False(t, s.isMovieDurationKnownMissing(ctx, 9907))

	providers, found := s.getCachedWatchProviders(ctx, 9908)
	assert.False(t, found)
	assert.Nil(t, providers)

	// 書き込み側も同様にno-opであること
	assert.NotPanics(t, func() {
		s.markMovieDurationMissing(ctx, 9907, 0)
		s.cacheWatchProviders(ctx, 9908, nil)
	})
}
