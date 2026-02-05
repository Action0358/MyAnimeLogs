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
	// width.Foldは半角カタカナを全角に変換するが、英数字はそのまま
	// norm.NFCで濁点・半濁点を結合文字から合成済み文字に変換
	normalized := norm.NFC.String(width.Fold.String(s))

	var result strings.Builder
	result.Grow(len(normalized))

	for _, r := range normalized {
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
