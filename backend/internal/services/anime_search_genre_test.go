package services

import (
	"strings"
	"testing"

	"myanimelogs/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ジャンル未指定なら句を足さない。空文字を返して呼び出し側に判定させる。
func TestBuildGenreFilterClause_NoGenres(t *testing.T) {
	for _, genres := range [][]string{nil, {}} {
		clause, args, nextIndex := buildGenreFilterClause(genres, 3)

		assert.Empty(t, clause)
		assert.Empty(t, args)
		assert.Equal(t, 3, nextIndex, "句を足さないならプレースホルダ番号も進めない")
	}
}

// 単一ジャンルは従来どおり動く（後方互換）。
func TestBuildGenreFilterClause_SingleGenre(t *testing.T) {
	clause, args, nextIndex := buildGenreFilterClause([]string{"アクション"}, 1)

	assert.Equal(t, "id IN (SELECT anime_id FROM anime_genres WHERE genre IN ($1))", clause)
	assert.Equal(t, []interface{}{"アクション"}, args)
	assert.Equal(t, 2, nextIndex)
}

// 複数ジャンルはIN句になり、いずれかを持つ作品にヒットする（OR）。
// 以前は genre = $1 の完全一致だったため "アクション,ドラマ" が必ず0件だった（#811）。
func TestBuildGenreFilterClause_MultipleGenres(t *testing.T) {
	clause, args, nextIndex := buildGenreFilterClause([]string{"アクション", "ドラマ"}, 1)

	assert.Equal(t, "id IN (SELECT anime_id FROM anime_genres WHERE genre IN ($1, $2))", clause)
	assert.Equal(t, []interface{}{"アクション", "ドラマ"}, args)
	assert.Equal(t, 3, nextIndex)

	// ANDになっていないこと。ANDだと該当が激減する
	assert.NotContains(t, clause, " AND ")
}

// 他の条件の後ろに置かれてもプレースホルダ番号が衝突しない。
// ここがずれると引数の対応が崩れ、無関係な値で絞り込まれる。
func TestBuildGenreFilterClause_ContinuesArgIndex(t *testing.T) {
	clause, args, nextIndex := buildGenreFilterClause([]string{"SF", "ホラー", "ロマンス"}, 5)

	assert.Equal(t, "id IN (SELECT anime_id FROM anime_genres WHERE genre IN ($5, $6, $7))", clause)
	assert.Len(t, args, 3)
	assert.Equal(t, 8, nextIndex)
}

// プレースホルダの数と引数の数は必ず一致する。
// ずれるとPostgres側で "bind message supplies N parameters" エラーになる。
func TestBuildGenreFilterClause_PlaceholderCountMatchesArgs(t *testing.T) {
	genres := []string{"アクション", "コメディ", "ドラマ", "ファンタジー"}
	clause, args, _ := buildGenreFilterClause(genres, 1)

	assert.Equal(t, len(args), strings.Count(clause, "$"))
}

// カンマ区切りの genre パラメータが個々のジャンルに分解されること。
func TestAnimeSearchParams_GenreList(t *testing.T) {
	tests := []struct {
		name  string
		genre string
		want  []string
	}{
		{"未指定", "", nil},
		{"単一", "アクション", []string{"アクション"}},
		{"複数", "ドラマ,アクション", []string{"アクション", "ドラマ"}},
		// 手書きURLやコピペで混入しやすい形。捨てずに拾えること
		{"前後の空白は落とす", " アクション , ドラマ ", []string{"アクション", "ドラマ"}},
		{"空要素は無視する", "アクション,,ドラマ,", []string{"アクション", "ドラマ"}},
		{"重複は取り除く", "アクション,ドラマ,アクション", []string{"アクション", "ドラマ"}},
		{"カンマのみは未指定と同じ", ",,,", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := models.AnimeSearchParams{Genre: tt.genre}
			assert.Equal(t, tt.want, params.GenreList())
		})
	}
}

// 並び順だけが違う同じ検索が、別のキャッシュエントリに分かれないこと。
// 分かれるとキャッシュヒット率が落ちるだけでなく、片方だけ古い結果を返しうる。
func TestGenerateSearchCacheKey_GenreOrderIndependent(t *testing.T) {
	s := &AnimeService{}

	base := models.AnimeSearchParams{Page: 1, PerPage: 24}

	forward := base
	forward.Genre = "アクション,ドラマ"
	reverse := base
	reverse.Genre = "ドラマ,アクション"

	assert.Equal(t, s.generateSearchCacheKey(forward), s.generateSearchCacheKey(reverse))
}

// ジャンルが違えば別のキーになること（上の正規化で潰しすぎていないこと）。
func TestGenerateSearchCacheKey_DistinctPerGenreSet(t *testing.T) {
	s := &AnimeService{}
	base := models.AnimeSearchParams{Page: 1, PerPage: 24}

	single := base
	single.Genre = "アクション"
	multi := base
	multi.Genre = "アクション,ドラマ"
	none := base

	keys := []string{
		s.generateSearchCacheKey(none),
		s.generateSearchCacheKey(single),
		s.generateSearchCacheKey(multi),
	}

	seen := map[string]bool{}
	for _, key := range keys {
		require.False(t, seen[key], "キーが重複している: %s", key)
		seen[key] = true
	}
}
