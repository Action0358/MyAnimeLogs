package models

import (
	"time"
)

type User struct {
	ID            int       `json:"id" db:"id"`
	AuthUserID    *string   `json:"auth_user_id,omitempty" db:"auth_user_id"`
	Name          string    `json:"name" db:"name"`
	Email         string    `json:"email" db:"email"`
	Description   *string   `json:"description" db:"description"`
	Icon          *string   `json:"icon" db:"icon"`
	FavoriteAnime *string   `json:"favorite_anime" db:"favorite_anime"`
	SnsLink       *string   `json:"sns_link" db:"sns_link"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// PresetAvatarIDs はユーザーが選択できるアバターの許可リスト。
//
// 画像アップロードは #763 で廃止した。users.icon にはURLではなくこのIDを
// 格納し、ここで許可リスト検証することで任意のURLを混入させない。
// varchar のまま任意の値を受け付けると、外部画像を持ち込ませないという
// 廃止の目的が達成できないため。
//
// 【重要】frontend/src/lib/avatars.ts の PRESET_AVATAR_IDS と同期させること。
var PresetAvatarIDs = []string{
	"preset-01",
	"preset-02",
	"preset-03",
	"preset-04",
	"preset-05",
	"preset-06",
	"preset-07",
	"preset-08",
}

// IsValidPresetAvatarID reports whether the given icon value is an allowed preset.
// 空文字はアイコン未設定を表すため許可する。
func IsValidPresetAvatarID(id string) bool {
	if id == "" {
		return true
	}
	for _, preset := range PresetAvatarIDs {
		if preset == id {
			return true
		}
	}
	return false
}

type UserUpdateRequest struct {
	Name          *string `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Description   *string `json:"description,omitempty" binding:"omitempty,max=500"`
	FavoriteAnime *string `json:"favorite_anime,omitempty" binding:"omitempty,max=200"`
	SnsLink       *string `json:"sns_link,omitempty" binding:"omitempty,url,max=500"`
	// Icon はプリセットのID。値の検証はハンドラーで IsValidPresetAvatarID により行う
	Icon *string `json:"icon,omitempty" binding:"omitempty,max=50"`
}

type UserResponse struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Description   *string   `json:"description"`
	Icon          *string   `json:"icon"`
	FavoriteAnime *string   `json:"favorite_anime"`
	SnsLink       *string   `json:"sns_link"`
	CreatedAt     time.Time `json:"created_at"`
}
