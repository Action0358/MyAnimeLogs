package models

import (
	"strings"
	"testing"

	"myanimelogs/internal/constants"

	"github.com/stretchr/testify/assert"
)

func genreNames(n int) []string {
	genres := make([]string, n)
	for i := range genres {
		genres[i] = "g" + strings.Repeat("x", i)
	}
	return genres
}

func TestAnimeSearchParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  AnimeSearchParams
		wantErr bool
	}{
		{"条件なし", AnimeSearchParams{}, false},
		// バイト数で数えると、日本語は34文字程度で弾かれてしまう
		{"日本語キーワードは文字数で数える", AnimeSearchParams{Keyword: strings.Repeat("あ", constants.MaxSearchKeywordLength)}, false},
		{"キーワードが上限超え", AnimeSearchParams{Keyword: strings.Repeat("あ", constants.MaxSearchKeywordLength+1)}, true},
		{"ジャンル数がちょうど上限", AnimeSearchParams{Genre: strings.Join(genreNames(constants.MaxSearchGenres), ",")}, false},
		{"ジャンル数が上限超え", AnimeSearchParams{Genre: strings.Join(genreNames(constants.MaxSearchGenres+1), ",")}, true},
		{"ジャンル名が長すぎる", AnimeSearchParams{Genre: strings.Repeat("x", constants.MaxSearchFilterLength+1)}, true},
		// 重複除去すると1個になるが、ログには元の文字列が残るので弾く
		{"重複だけの巨大なジャンル指定", AnimeSearchParams{Genre: strings.Repeat("a,", 6000)}, true},
		{"シーズンが長すぎる", AnimeSearchParams{Season: strings.Repeat("x", constants.MaxSearchFilterLength+1)}, true},
		{"並び順が長すぎる", AnimeSearchParams{Sort: strings.Repeat("x", constants.MaxSearchFilterLength+1)}, true},
		{"メディアが長すぎる", AnimeSearchParams{Media: strings.Repeat("x", constants.MaxSearchFilterLength+1)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdvancedSearchParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  AdvancedSearchParams
		wantErr bool
	}{
		{"条件なし", AdvancedSearchParams{}, false},
		{"キーワードがちょうど上限", AdvancedSearchParams{Query: strings.Repeat("a", constants.MaxSearchKeywordLength)}, false},
		{"キーワードが上限超え", AdvancedSearchParams{Query: strings.Repeat("a", constants.MaxSearchKeywordLength+1)}, true},
		{"ジャンル数がちょうど上限", AdvancedSearchParams{Genres: genreNames(constants.MaxSearchGenres)}, false},
		{"ジャンル数が上限超え", AdvancedSearchParams{Genres: genreNames(constants.MaxSearchGenres + 1)}, true},
		{"ジャンル名が長すぎる", AdvancedSearchParams{Genres: []string{strings.Repeat("x", constants.MaxSearchFilterLength+1)}}, true},
		{"視聴ステータスが長すぎる", AdvancedSearchParams{UserStatus: strings.Repeat("x", constants.MaxSearchFilterLength+1)}, true},
		{"並び方向が長すぎる", AdvancedSearchParams{Order: strings.Repeat("x", constants.MaxSearchFilterLength+1)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
