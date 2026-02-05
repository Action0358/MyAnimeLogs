package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKatakanaToHiragana(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "カタカナのみ",
			input:    "ブリーチ",
			expected: "ぶりーち",
		},
		{
			name:     "カタカナと漢字の混合",
			input:    "機動戦士ガンダム",
			expected: "機動戦士がんだむ",
		},
		{
			name:     "英語とカタカナの混合",
			input:    "NARUTOナルト",
			expected: "NARUTOなると",
		},
		{
			name:     "ひらがなのみ（変換なし）",
			input:    "ぶりーち",
			expected: "ぶりーち",
		},
		{
			name:     "英語のみ（変換なし）",
			input:    "BLEACH",
			expected: "BLEACH",
		},
		{
			name:     "空文字",
			input:    "",
			expected: "",
		},
		{
			name:     "長音記号を含む",
			input:    "ブリーチ",
			expected: "ぶりーち",
		},
		{
			name:     "半角カタカナ",
			input:    "ﾌﾞﾘｰﾁ",
			expected: "ぶりーち", // 長音記号はそのまま
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KatakanaToHiragana(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHiraganaToKatakana(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ひらがなのみ",
			input:    "ぶりーち",
			expected: "ブリーチ",
		},
		{
			name:     "ひらがなと漢字の混合",
			input:    "機動戦士がんだむ",
			expected: "機動戦士ガンダム",
		},
		{
			name:     "カタカナのみ（変換なし）",
			input:    "ブリーチ",
			expected: "ブリーチ",
		},
		{
			name:     "英語のみ（変換なし）",
			input:    "BLEACH",
			expected: "BLEACH",
		},
		{
			name:     "空文字",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HiraganaToKatakana(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsKatakana(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "カタカナを含む",
			input:    "ブリーチ",
			expected: true,
		},
		{
			name:     "ひらがなのみ",
			input:    "ぶりーち",
			expected: false,
		},
		{
			name:     "漢字のみ",
			input:    "進撃の巨人",
			expected: false,
		},
		{
			name:     "英語のみ",
			input:    "BLEACH",
			expected: false,
		},
		{
			name:     "混合（カタカナあり）",
			input:    "進撃のキョジン",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsKatakana(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsHiragana(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "ひらがなを含む",
			input:    "ぶりーち",
			expected: true,
		},
		{
			name:     "カタカナのみ",
			input:    "ブリーチ",
			expected: false,
		},
		{
			name:     "漢字のみ",
			input:    "進撃",
			expected: false,
		},
		{
			name:     "混合（ひらがなあり）",
			input:    "進撃の巨人",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsHiragana(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
