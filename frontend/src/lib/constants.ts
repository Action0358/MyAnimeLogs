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
// 狭い画面で折り返して読みにくくならないよう、1行に収まる長さにする
export const LOGOUT_REASON_MESSAGES = {
  timeout: 'タイムアウトしました。再度ログインしてください',
  unauthorized: 'ログインが終了しました。再度ログインしてください',
} as const

/**
 * ログアウト理由の型
 */
export type LogoutReason = 'timeout' | 'unauthorized' | 'manual' | 'account-deleted'

/**
 * ログアウト理由に応じたリダイレクト先のマップ
 */
export const LOGOUT_REDIRECT_PATHS = {
  timeout: '/signin?reason=timeout',
  unauthorized: '/signin?reason=unauthorized',
  manual: '/',
  // 削除完了をトップページで伝えるのは難しいため専用ページへ送る。
  // 再ログインすれば新しいアカウントが作られる（データは戻らない）ことを
  // ここで明示しないと、削除できていないように見える。
  'account-deleted': '/account-deleted',
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

// ========================
// アニメ関連の定数
// ========================

/**
 * 人気アニメの取得上限
 * バックエンドのconstants.PopularAnimeLimitと同期
 */
export const POPULAR_ANIME_LIMIT = 100

/**
 * 今期アニメの取得上限
 * バックエンドのconstants.PopularAnimeLimitと同期
 */
export const CURRENT_SEASON_ANIME_LIMIT = 100

/**
 * トップページで表示する今期アニメの取得上限
 * 一覧ページ（CURRENT_SEASON_ANIME_LIMIT）とは別枠で、カルーセルとヒーロー背景が共有する
 */
export const HOME_CURRENT_SEASON_ANIME_LIMIT = 20

/**
 * アニメのステータス型
 */
export type AnimeStatus = 'airing' | 'finished' | 'upcoming'

/**
 * 有効なアニメステータスかどうかを判定する型ガード関数
 */
export const isValidAnimeStatus = (status: unknown): status is AnimeStatus => {
  return status === 'airing' || status === 'finished' || status === 'upcoming'
}

/**
 * ステータス文字列を安全にAnimeStatus型に変換
 * 無効な値の場合はundefinedを返す
 */
export const toAnimeStatus = (status: unknown): AnimeStatus | undefined => {
  return isValidAnimeStatus(status) ? status : undefined
}

// ========================
// 運営者情報
// ========================

/**
 * 対外表記に使う運営者名。
 * 個人開発の無償サービスのため、個人名ではなくサービス名義を用いる。
 */
export const OPERATOR_NAME = 'MyAnimeLogs 運営'

/**
 * 問い合わせ先メールアドレス。
 *
 * Google OAuth 同意画面のサポートメールと共通のため、問い合わせの主導線を
 * フォームへ移したあとも有効なまま維持する必要がある（Googleが開発者への
 * 連絡に使う）。
 *
 * 一般の問い合わせはフォーム（SUPPORT_FORM_URL）へ誘導し、こちらは
 * 個人情報の開示・削除などの請求窓口として残す。法的な請求の受付を
 * フォームだけにしないため。
 */
export const SUPPORT_EMAIL = 'myanimelogs-support@googlegroups.com'

/**
 * お問い合わせフォーム（Googleフォーム）。
 *
 * 送信者のメールアドレスは自動収集しない設定にしてあるため、匿名でも
 * 送信できる。回答はスプレッドシートに蓄積され、新着は運営者へメール通知される。
 */
export const SUPPORT_FORM_URL =
  'https://docs.google.com/forms/d/e/1FAIpQLSf7W1gZTxI7GhzpGniIpWACjDRXd-NEEmIdKEjZ9Kl4nKY6rg/viewform'
