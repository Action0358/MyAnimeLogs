package services

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

// あらすじ正規化用の正規表現（パッケージ初期化時に1回だけコンパイル）。
var (
	descBrTagRe      = regexp.MustCompile(`(?i)<br\s*/?>`)         // <br> → 改行
	descHTMLTagRe    = regexp.MustCompile(`<[^>]+>`)              // 残存HTMLタグ除去
	descSourceNoteRe = regexp.MustCompile(`(?im)\s*[\(（]\s*(source|出典)[:：][^)）]*[\)）]\s*$`) // 末尾の出典注記
	descMultiSpaceRe = regexp.MustCompile(`[ \t]{2,}`)           // 連続スペース
	descMultiBlankRe = regexp.MustCompile(`\n{3,}`)              // 3つ以上の連続改行
)

// sanitizeDescription はソース（TMDB/AniList等）でバラバラなあらすじの書式を統一する。
// HTMLタグ・出典注記を除去し、段落は維持したまま余分な空行・行頭行末の空白を圧縮する。
func sanitizeDescription(s string) string {
	if s == "" {
		return ""
	}

	// <br> を改行へ、その他のHTMLタグを除去
	s = descBrTagRe.ReplaceAllString(s, "\n")
	s = descHTMLTagRe.ReplaceAllString(s, "")
	// HTMLエンティティをデコード（&amp; など）
	s = html.UnescapeString(s)
	// 末尾の "(Source: ...)" / "（出典: ...）" を除去
	s = descSourceNoteRe.ReplaceAllString(s, "")
	// 改行コードを統一
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// 行ごとに連続スペースを圧縮し、行頭・行末の空白を除去
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		line = descMultiSpaceRe.ReplaceAllString(line, " ")
		lines[i] = strings.TrimSpace(line)
	}
	s = strings.Join(lines, "\n")

	// 3つ以上の連続改行は段落区切り（空行1つ）に圧縮
	s = descMultiBlankRe.ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}

// containsJapanese は文字列に日本語（ひらがな・カタカナ・漢字）が含まれるかを判定する。
// あらすじが既に日本語かどうかの判定に用いる。
func containsJapanese(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return true
		}
	}
	return false
}
