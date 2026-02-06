package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

// KatakanaToHiragana converts katakana characters to hiragana
// カタカナをひらがなに変換する（半角カタカナも対応）
// golang.org/x/text/widthを使用して半角カタカナ→全角カタカナ変換後、カタカナ→ひらがな変換
func KatakanaToHiragana(s string) string {
	// 半角カタカナを全角カタカナに正規化（濁点・半濁点も結合）
	normalized := norm.NFC.String(width.Fold.String(s))

	return strings.Map(func(r rune) rune {
		// カタカナ範囲 (ァ-ヶ: U+30A1-U+30F6) をひらがなに変換
		if r >= 'ァ' && r <= 'ヶ' {
			return r - 0x60
		}
		return r
	}, normalized)
}

// HiraganaToKatakana converts hiragana characters to katakana
// ひらがなをカタカナに変換する
func HiraganaToKatakana(s string) string {
	return strings.Map(func(r rune) rune {
		// ひらがな範囲 (ぁ-ゖ: U+3041-U+3096) をカタカナに変換
		if r >= 'ぁ' && r <= 'ゖ' {
			return r + 0x60
		}
		return r
	}, s)
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
