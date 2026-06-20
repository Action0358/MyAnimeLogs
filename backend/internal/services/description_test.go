package services

import "testing"

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "HTMLタグと出典注記を除去",
			in:   "Hello <i>world</i>.<br><br>Second line. (Source: Crunchyroll)",
			want: "Hello world.\n\nSecond line.",
		},
		{
			name: "連続改行を段落区切りに圧縮",
			in:   "Line1\n\n\n\nLine2",
			want: "Line1\n\nLine2",
		},
		{
			name: "行頭行末の空白と連続スペースを圧縮",
			in:   "  foo   bar  \n  baz  ",
			want: "foo bar\nbaz",
		},
		{
			name: "HTMLエンティティをデコード",
			in:   "Tom &amp; Jerry",
			want: "Tom & Jerry",
		},
		{
			name: "日本語の出典注記も除去",
			in:   "あらすじ本文。（出典: AniList）",
			want: "あらすじ本文。",
		},
		{
			name: "空文字はそのまま",
			in:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeDescription(tt.in); got != tt.want {
				t.Errorf("sanitizeDescription(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsJapaneseText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"日本語の文章", "魔王を倒した勇者一行は、世界を巡る旅に出た。", true},
		{"ひらがなのみ", "ひらがなのぶんしょう", true},
		{"カタカナのみ", "カタカナノブンショウ", true},
		{"英語の文章", "During their decade-long quest to defeat the Demon King", false},
		// 英語あらすじにタイトル・人名で日本語が部分混入するケース（誤判定しないこと）
		{"英語＋日本語タイトル混入", "Frieren (フリーレン) is a mage who lives a long life.", false},
		{"英語＋日本語人名混入", "Based on the manga by Akasaka Aka (赤坂 アカ) and Mengo Yokoyari.", false},
		{"記号・数字のみ", "12345 !?@#", false},
		{"空文字", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isJapaneseText(tt.in); got != tt.want {
				t.Errorf("isJapaneseText(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
