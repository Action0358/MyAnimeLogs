/**
 * ログイン後の戻り先の受け渡し。
 *
 * /signin?next=<パス> で受け取った戻り先を、Googleへ遷移する直前に
 * sessionStorage へ退避し、/auth/callback で取り出して遷移する。
 * OAuth の redirectTo にクエリを載せると Supabase の許可URLとの照合に
 * 影響するため、同じタブで生き残る sessionStorage を経由する。
 */

const STORAGE_KEY = 'auth_return_path'

/** 戻り先にしても意味がない、またはループするパス */
const EXCLUDED_PATHS = ['/signin', '/login', '/register', '/auth', '/account-deleted']

const isExcluded = (path: string): boolean => {
  const pathname = path.split(/[?#]/)[0]
  return EXCLUDED_PATHS.some((p) => pathname === p || pathname.startsWith(`${p}/`))
}

/**
 * 戻り先として安全なサイト内パスだけを通す。
 * 外部サイトへ飛ばすオープンリダイレクトを防ぐため、"/" 始まりに限り、
 * "//"（プロトコル相対URL）を拒否する。
 */
export const sanitizeReturnPath = (raw: string | null | undefined): string | null => {
  if (!raw) return null
  if (!raw.startsWith('/') || raw.startsWith('//')) return null
  // バックスラッシュや制御文字はブラウザが除去・変換し、"//" と同じ扱いになりうる
  if (/[\\\u0000-\u001f\u007f]/.test(raw)) return null
  if (raw === '/' || isExcluded(raw)) return null
  return raw
}

/** 戻り先つきの /signin のURL。戻り先が無効ならクエリを付けない */
export const buildSignInHref = (returnPath: string | null | undefined): string => {
  const safe = sanitizeReturnPath(returnPath)
  return safe ? `/signin?next=${encodeURIComponent(safe)}` : '/signin'
}

export const saveReturnPath = (returnPath: string | null | undefined): void => {
  try {
    const safe = sanitizeReturnPath(returnPath)
    if (safe) {
      sessionStorage.setItem(STORAGE_KEY, safe)
    } else {
      // 前回の中断したログインの戻り先を持ち越さない
      sessionStorage.removeItem(STORAGE_KEY)
    }
  } catch {
    // ストレージが使えない環境ではトップへ戻すだけなので握りつぶす
  }
}

/** 退避した戻り先を取り出して消す。無ければ "/" */
export const consumeReturnPath = (): string => {
  try {
    const stored = sessionStorage.getItem(STORAGE_KEY)
    sessionStorage.removeItem(STORAGE_KEY)
    return sanitizeReturnPath(stored) ?? '/'
  } catch {
    return '/'
  }
}
