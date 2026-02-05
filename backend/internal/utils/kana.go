package utils

import (
	"strings"
	"unicode"
)

// KatakanaToHiragana converts katakana characters to hiragana
// カタカナをひらがなに変換する
// Unicode範囲: カタカナ U+30A0-U+30FF, ひらがな U+3040-U+309F
// カタカナとひらがなの差は0x60 (96)
func KatakanaToHiragana(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	for _, r := range s {
		// カタカナ範囲 (ァ-ヶ: U+30A1-U+30F6) をひらがなに変換
		// 長音記号(ー U+30FC)やカタカナ中点(・ U+30FB)は変換しない
		if r >= 'ァ' && r <= 'ヶ' {
			// カタカナからひらがなへの変換 (0x60を引く)
			result.WriteRune(r - 0x60)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// HiraganaToKatakana converts hiragana characters to katakana
// ひらがなをカタカナに変換する
func HiraganaToKatakana(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	for _, r := range s {
		// ひらがな範囲 (ぁ-ゖ: U+3041-U+3096) をカタカナに変換
		if r >= 'ぁ' && r <= 'ゖ' {
			// ひらがなからカタカナへの変換 (0x60を足す)
			result.WriteRune(r + 0x60)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ContainsKatakana checks if the string contains any katakana characters
func ContainsKatakana(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Katakana) {
			return true
		}
	}
	return false
}

// ContainsHiragana checks if the string contains any hiragana characters
func ContainsHiragana(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Hiragana) {
			return true
		}
	}
	return false
}
