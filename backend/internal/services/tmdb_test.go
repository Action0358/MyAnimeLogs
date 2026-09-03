package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeAnimeTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"第N期（アラビア数字）", "ワンパンマン 第3期", "ワンパンマン"},
		{"第N期（スペースなし）", "【推しの子】第3期", "【推しの子】"},
		{"第N期（漢数字）", "ワンパンマン 第二期", "ワンパンマン"},
		{"第Nシーズン", "葬送のフリーレン 第2シーズン", "葬送のフリーレン"},
		{"第Nクール", "アニメ 第2クール", "アニメ"},
		{"Season N", "My Hero Academia Season 5", "My Hero Academia"},
		{"The Final Season", "進撃の巨人 The Final Season Part 2", "進撃の巨人"},
		{"Final Season", "進撃の巨人 Final Season", "進撃の巨人"},
		{"2nd Season", "モブサイコ100 2nd Season", "モブサイコ100"},
		{"3rd Season", "アニメ 3rd Season", "アニメ"},
		{"4th Season", "アニメ 4th Season", "アニメ"},
		{"小文字season+副題", "Re:ゼロから始める異世界生活 4th season 喪失編", "Re:ゼロから始める異世界生活"},
		{"小文字season（英タイトル）", "My Hero Academia season 7", "My Hero Academia"},
		{"小文字final season", "進撃の巨人 final season", "進撃の巨人"},
		{"末尾空白付きPart", "アニメ Part 2  ", "アニメ"},
		{"末尾空白付きローマ数字", "アニメ II ", "アニメ"},
		{"全角ローマ数字（副題の前に挿入）", "無職転生Ⅱ ～異世界行ったら本気だす～", "無職転生 ～異世界行ったら本気だす～"},
		{"全角ローマ数字（末尾）", "アニメⅢ", "アニメ"},
		{"変換不要", "鬼滅の刃", "鬼滅の刃"},
		{"空文字列", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAnimeTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProviderLogoURL(t *testing.T) {
	tests := []struct {
		name     string
		logoPath string
		size     string
		expected string
	}{
		{"通常のパス", "/abc123.png", "w92", "https://image.tmdb.org/t/p/w92/abc123.png"},
		{"サイズ指定なし → デフォルトw92", "/abc123.png", "", "https://image.tmdb.org/t/p/w92/abc123.png"},
		{"空パス → 空文字列", "", "w92", ""},
		{"大きいサイズ", "/abc123.png", "w500", "https://image.tmdb.org/t/p/w500/abc123.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetProviderLogoURL(tt.logoPath, tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindBestMatch(t *testing.T) {
	c := &TMDBClient{}

	reZero := TMDBTVShow{ID: 65942, Name: "Re:ゼロから始める異世界生活", OriginCountry: []string{"JP"}}
	unrelated := TMDBTVShow{ID: 999, Name: "全く別のアニメ", OriginCountry: []string{"JP"}}
	usShow := TMDBTVShow{ID: 111, Name: "Some US Drama", OriginCountry: []string{"US"}}

	tests := []struct {
		name     string
		title    string
		results  []TMDBTVShow
		expectID int // 0 means expect nil
	}{
		{
			name:     "正規化後タイトルでJP本編にマッチ",
			title:    "Re:ゼロから始める異世界生活",
			results:  []TMDBTVShow{reZero},
			expectID: 65942,
		},
		{
			name:     "類似しない結果は採用しない（誤マッチ防止）",
			title:    "Re:ゼロから始める異世界生活",
			results:  []TMDBTVShow{unrelated},
			expectID: 0,
		},
		{
			name:     "類似する結果を先頭以外からも選ぶ",
			title:    "Re:ゼロから始める異世界生活",
			results:  []TMDBTVShow{unrelated, reZero},
			expectID: 65942,
		},
		{
			name:     "非JPでも類似すれば採用",
			title:    "Some US Drama",
			results:  []TMDBTVShow{usShow},
			expectID: 111,
		},
		{
			name:     "空の結果はnil",
			title:    "なにか",
			results:  []TMDBTVShow{},
			expectID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.findBestMatch(tt.results, tt.title)
			if tt.expectID == 0 {
				assert.Nil(t, got)
				return
			}
			if assert.NotNil(t, got) {
				assert.Equal(t, tt.expectID, got.ID)
			}
		})
	}
}

func TestTitlesSimilar(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		candidate    string
		originalName string
		expected     bool
	}{
		{"完全一致", "Re:ゼロから始める異世界生活", "Re:ゼロから始める異世界生活", "", true},
		{"包含（クエリが候補に含まれる）", "リゼロ", "リゼロ 新編集版", "", true},
		{"original_name側でマッチ", "Re:Zero", "リゼロ", "Re:ZERO -Starting Life in Another World-", true},
		{"全く別物", "鬼滅の刃", "呪術廻戦", "", false},
		{"空クエリ", "", "なにか", "", false},
		// 短すぎるクエリの包含による誤マッチを防止（最低長 minContainmentRunes）
		{"短いクエリ no は naruto に誤マッチしない", "no", "Naruto", "", false},
		{"短いクエリ in は長文に誤マッチしない", "in", "Shingeki no Kyojin", "", false},
		{"3文字以上の包含はマッチ", "ナルト", "ナルト 疾風伝", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := titlesSimilar(normalizeForComparison(tt.query), tt.candidate, tt.originalName)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"同一", "kitten", "kitten", 0},
		{"古典例", "kitten", "sitting", 3},
		{"片方空", "", "abc", 3},
		{"両方空", "", "", 0},
		{"長さ非対称（引数順入替）", "abc", "abcdef", 3},
		{"日本語", "リゼロ", "リセロ", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, levenshtein([]rune(tt.a), []rune(tt.b)))
			// 引数の順序に依存しないこと（対称性）
			assert.Equal(t, tt.expected, levenshtein([]rune(tt.b), []rune(tt.a)))
		})
	}
}

func TestFindBestMovieMatch(t *testing.T) {
	c := &TMDBClient{}

	kimiNoNaWa := TMDBMovie{ID: 372058, Title: "君の名は。", OriginalLanguage: "ja"}
	unrelatedJa := TMDBMovie{ID: 999, Title: "全く別の劇場版", OriginalLanguage: "ja"}
	enMovie := TMDBMovie{ID: 111, Title: "Some English Film", OriginalLanguage: "en"}

	tests := []struct {
		name     string
		title    string
		results  []TMDBMovie
		expectID int // 0 means expect nil
	}{
		{"jaで類似する作品にマッチ", "君の名は。", []TMDBMovie{kimiNoNaWa}, 372058},
		{"類似しないja作品は採用しない", "君の名は。", []TMDBMovie{unrelatedJa}, 0},
		{"先頭以外の類似作品を選ぶ", "君の名は。", []TMDBMovie{unrelatedJa, kimiNoNaWa}, 372058},
		{"非jaでも類似すれば採用", "Some English Film", []TMDBMovie{enMovie}, 111},
		{"空結果はnil", "なにか", []TMDBMovie{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.findBestMovieMatch(tt.results, tt.title)
			if tt.expectID == 0 {
				assert.Nil(t, got)
				return
			}
			if assert.NotNil(t, got) {
				assert.Equal(t, tt.expectID, got.ID)
			}
		})
	}
}
