'use client'

import React from 'react'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import type { Route } from 'next'
import { buildSignInHref } from '@/lib/authReturnPath'

type SignInLinkProps = Omit<React.ComponentProps<typeof Link>, 'href'>

/**
 * ログイン後にいまのページへ戻ってこられる /signin へのリンク。
 *
 * 戻り先は /signin?next=<パス> で渡す。クエリ（作品詳細の ?from= や検索条件）と
 * ハッシュ（お知らせの #<id>）も引き継ぎたいが、useSearchParams は静的エクスポート時に Suspense 境界を要求し、
 * ヘッダーのように全ページに出る部品とは噛み合わない。そこで href はパスだけで
 * 組み、クリック時に window.location から最新のクエリ・ハッシュを含めて遷移し直す。
 * 新しいタブで開いた場合は href どおりパスだけが引き継がれる。
 */
export const SignInLink: React.FC<SignInLinkProps> = ({ onClick, ...props }) => {
  const pathname = usePathname()
  const router = useRouter()

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    onClick?.(e)
    if (e.defaultPrevented) return
    // 修飾キー付き（新しいタブ等）はブラウザの既定動作に任せる
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0) return
    e.preventDefault()
    const { pathname: current, search, hash } = window.location
    // 戻り先つきの動的URLのため typedRoutes の Route 型には収まらない
    router.push(buildSignInHref(`${current}${search}${hash}`) as Route)
  }

  return <Link {...props} href={buildSignInHref(pathname) as Route} onClick={handleClick} />
}
