'use client'

import { useEffect } from 'react'
import { fetchLatestBuildId, getClientBuildId } from '@/lib/buildSkew'

/** 前面にいる間の確認間隔 */
const CHECK_INTERVAL_MS = 5 * 60 * 1000
/** タブへの復帰・フォーカスのたびに確認しないよう、この間隔をあける */
const MIN_CHECK_GAP_MS = 30 * 1000
/** 読み込み直しの連続を防ぐ（読み込み後も食い違うなら、繰り返さない） */
const RELOAD_GUARD_MS = 60 * 1000
const RELOAD_KEY = 'build_skew_reloaded_at'

/** 認可コードを一度だけ使う画面。読み込み直すと、使用済みの code で失敗する */
const NO_RELOAD_PATHS = ['/auth/callback']

const canReloadNow = (): boolean => {
  if (NO_RELOAD_PATHS.some((p) => window.location.pathname.startsWith(p))) return false
  try {
    const last = Number(sessionStorage.getItem(RELOAD_KEY) ?? 0)
    return Date.now() - last > RELOAD_GUARD_MS
  } catch {
    return true
  }
}

const reload = () => {
  try {
    sessionStorage.setItem(RELOAD_KEY, String(Date.now()))
  } catch {
    // ストレージが使えなくても、読み込み直し自体はできる
  }
  window.location.reload()
}

/** サイト内の通常のリンククリックか（新しいタブ・修飾キー・ダウンロード等は対象外） */
const findInternalLink = (event: MouseEvent): HTMLAnchorElement | null => {
  if (event.defaultPrevented || event.button !== 0) return null
  if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return null
  const anchor = (event.target as Element | null)?.closest?.('a')
  if (!anchor || !anchor.href) return null
  if (anchor.target && anchor.target !== '_self') return null
  if (anchor.hasAttribute('download')) return null
  const url = new URL(anchor.href, window.location.href)
  if (url.origin !== window.location.origin) return null
  // 同じページ内のハッシュ移動は、読み込み直さなくてよい
  if (
    url.pathname === window.location.pathname &&
    url.search === window.location.search &&
    url.hash
  ) {
    return null
  }
  return anchor
}

/**
 * デプロイされて古くなったタブを、遷移が失敗する前に見つけて直す（#1311）。
 *
 * - タブが前面に戻ったとき: 古ければそのまま読み込み直す（戻った直後は操作の途中ではない）
 * - 前面で使用中に古くなったとき: 以降のサイト内リンクを、ブラウザごとの移動に切り替える
 *   （クライアント遷移だと、生テキストの .txt へ移ってしまうため）
 */
export const BuildSkewGuard: React.FC = () => {
  useEffect(() => {
    const clientId = getClientBuildId()
    // 読めない環境（開発サーバーなど）では、何もしない
    if (!clientId) return

    let isStale = false
    let lastCheckAt = 0
    let isChecking = false

    const check = async (): Promise<boolean> => {
      if (isStale) return true
      if (isChecking) return false
      isChecking = true
      lastCheckAt = Date.now()
      try {
        const latest = await fetchLatestBuildId()
        if (latest && latest !== clientId) isStale = true
      } finally {
        isChecking = false
      }
      return isStale
    }

    const onReturn = async () => {
      if (document.visibilityState !== 'visible') return
      if (!isStale && Date.now() - lastCheckAt < MIN_CHECK_GAP_MS) return
      if ((await check()) && canReloadNow()) reload()
    }

    const onClick = (event: MouseEvent) => {
      if (!isStale) return
      const anchor = findInternalLink(event)
      if (!anchor) return
      // Next.js のクライアント遷移を止め、ブラウザごと移動する
      event.preventDefault()
      event.stopImmediatePropagation()
      window.location.assign(anchor.href)
    }

    const timer = window.setInterval(() => {
      if (document.visibilityState === 'visible') void check()
    }, CHECK_INTERVAL_MS)

    document.addEventListener('visibilitychange', onReturn)
    window.addEventListener('focus', onReturn)
    // Next.js の Link より先に処理するため、キャプチャ段階で受ける
    document.addEventListener('click', onClick, true)

    return () => {
      window.clearInterval(timer)
      document.removeEventListener('visibilitychange', onReturn)
      window.removeEventListener('focus', onReturn)
      document.removeEventListener('click', onClick, true)
    }
  }, [])

  return null
}

export default BuildSkewGuard
