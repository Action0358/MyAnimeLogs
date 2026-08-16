package models

import "testing"

// users.icon は varchar のため、許可リスト検証がないと任意のURLを保存できてしまう。
// アップロード廃止（#763）の目的は「外部画像を持ち込ませない」ことなので、
// この検証が機能しないと廃止の意味がなくなる。
func TestIsValidPresetAvatarID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"許可リストの先頭", "preset-01", true},
		{"許可リストの末尾", "preset-08", true},
		{"空文字はアイコン未設定を表すため許可", "", true},
		{"存在しないプリセット", "preset-99", false},
		{"外部URL", "https://example.com/evil.png", false},
		{"移行前のアップロードパス", "/uploads/icons/user_14_1786463713.jpeg", false},
		{"大文字は別物として扱う", "PRESET-01", false},
		{"前後の空白を含む値", " preset-01 ", false},
		{"パストラバーサル", "../../etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPresetAvatarID(tt.id); got != tt.expected {
				t.Errorf("IsValidPresetAvatarID(%q) = %v, want %v", tt.id, got, tt.expected)
			}
		})
	}
}

// フロントエンド（frontend/src/lib/avatars.ts）と件数がずれると、
// 選べるのに保存できない（またはその逆）という不整合が起きる
func TestPresetAvatarIDsCount(t *testing.T) {
	if len(PresetAvatarIDs) != 8 {
		t.Errorf("PresetAvatarIDs の件数 = %d, want 8（frontend/src/lib/avatars.ts と同期させること）", len(PresetAvatarIDs))
	}
}
