/**
 * アプリケーション全体で使用する定数
 */

// ========================
// 認証関連の定数
// ========================

/**
 * 非アクティブタイムアウト（ミリ秒）
 * 環境変数で設定可能。デフォルトは30分（1800秒）
 */
export const INACTIVITY_TIMEOUT = process.env.NEXT_PUBLIC_INACTIVITY_TIMEOUT 
  ? parseInt(process.env.NEXT_PUBLIC_INACTIVITY_TIMEOUT, 10) * 1000 
  : 30 * 60 * 1000 // デフォルト30分

/**
 * トークンリフレッシュのマージン（ミリ秒）
 * トークン期限の何秒前にリフレッシュするか
 */
export const TOKEN_REFRESH_MARGIN = 60 * 1000 // 1分

/**
 * ログアウト理由のメッセージマップ
 * TypeScriptの型推論を活用するため、明示的な型注釈は省略
 */
export const LOGOUT_REASON_MESSAGES = {
  timeout: 'セッションがタイムアウトしました。再度ログインしてください。',
  unauthorized: 'ログインセッションが終了しました。再度ログインしてください。',
} as const

/**
 * ログアウト理由の型
 */
export type LogoutReason = 'timeout' | 'unauthorized' | 'manual'

/**
 * ログアウト理由に応じたリダイレクト先のマップ
 */
export const LOGOUT_REDIRECT_PATHS = {
  timeout: '/login?reason=timeout',
  unauthorized: '/login?reason=unauthorized',
  manual: '/',
} as const satisfies Record<LogoutReason, string>

// ========================
// ストレージキー
// ========================

export const STORAGE_KEYS = {
  TOKEN: 'auth_token',
  REFRESH_TOKEN: 'refresh_token',
  USER: 'user_data',
  TOKEN_EXPIRES_AT: 'token_expires_at',
  LAST_ACTIVITY_TIMESTAMP: 'last_activity_timestamp',
} as const

// ========================
// デバッグ関連の定数
// ========================

/**
 * トークン期限切れ警告の閾値（ミリ秒）
 * この時間以内になると警告を表示
 */
export const TOKEN_EXPIRING_SOON_THRESHOLD = 2 * 60 * 1000 // 2分

/**
 * 非アクティブ警告の閾値（ミリ秒）
 * この時間以内になると警告を表示
 */
export const INACTIVITY_WARNING_THRESHOLD = 5 * 60 * 1000 // 5分

// ========================
// アセットパス
// ========================

/**
 * ヒーローセクション用ロゴ画像のパス
 */
export const HERO_LOGO_PATH = '/logo.svg'

/**
 * ヘッダー・フッター用ロゴ画像のパス
 */
export const SUB_LOGO_PATH = '/logo2.svg'
