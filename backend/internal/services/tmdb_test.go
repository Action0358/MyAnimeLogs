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
