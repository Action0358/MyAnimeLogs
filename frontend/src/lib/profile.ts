// プロフィール編集フォームで共有する定数・型・ヘルパー
// ProfileEditContent / MyPageContent / useProfileForm から参照する

// 文字数上限（バリデーション用）
export const MAX_NAME_LENGTH = 50
export const MAX_DESCRIPTION_LENGTH = 500
export const MAX_FAVORITE_ANIME_LENGTH = 200

// 編集可能なプロフィール項目の値
export interface ProfileFieldValues {
  name: string
  description: string
  favoriteAnime: string
}

// getProfileFieldValues が受け取るユーザー情報の最小形（API/AuthContext 双方に対応）
interface ProfileUserFields {
  name?: string
  description?: string
  favorite_anime?: string | null
}

// ユーザー情報から編集フォームの初期値を取り出す。
// ProfileEditContent（編集）と MyPageContent（表示）で重複していた導出処理を共通化する。
export const getProfileFieldValues = (
  user: ProfileUserFields | null | undefined
): ProfileFieldValues => ({
  name: user?.name || '',
  description: user?.description || '',
  favoriteAnime: user?.favorite_anime || '',
})
