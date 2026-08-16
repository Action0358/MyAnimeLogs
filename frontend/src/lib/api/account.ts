import { del, ApiClientError } from './client'

/**
 * アカウントを完全に削除する。
 *
 * 視聴記録・お気に入り・プロフィールがすべて削除され、復旧はできない。
 * 呼び出し側は成功後、必ず signOut してトークンを破棄すること。
 * JWTは最大1時間有効なまま残るが、バックエンドは削除済みアカウントからの
 * リクエストに401を返す。
 */
export const deleteAccount = async (): Promise<void> => {
  try {
    await del('/api/users/me')
  } catch (error) {
    if (error instanceof ApiClientError) {
      // 503 は service_role キー未設定など、サーバー側の設定不備
      if (error.status === 503) {
        throw new Error(
          'アカウント削除機能が一時的に利用できません。時間をおいて再度お試しください'
        )
      }
      throw new Error('アカウントの削除に失敗しました。時間をおいて再度お試しください')
    }
    throw error
  }
}
