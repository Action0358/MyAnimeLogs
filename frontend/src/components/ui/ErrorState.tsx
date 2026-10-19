'use client'

import React from 'react'
import Link from 'next/link'
import { CircleAlert, RotateCw, WifiOff } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from './Button'

export type ErrorStateType = 'network' | 'unknown'

export interface ErrorStateProps {
  type: ErrorStateType
  onRetry: () => void
  /** page: ページ本文用 / compact: トップページの各セクション用（小さめ・outlineボタン・サブリンクなし） */
  variant?: 'page' | 'compact'
  /** variant=page で、検索のように上に操作部が残るページ向けに最小高さを抑える（PC 360px / SP 320px） */
  shortHeight?: boolean
  className?: string
}

const CONTENT = {
  network: {
    Icon: WifiOff,
    title: '接続できませんでした',
    description: '通信環境を確認して、もう一度お試しください。',
    subLabel: 'トップへ戻る',
    subHref: '/',
  },
  unknown: {
    Icon: CircleAlert,
    title: '読み込めませんでした',
    description: '時間をおいて、もう一度お試しください。',
    subLabel: '解決しない場合はお問い合わせ',
    subHref: '/contact',
  },
} as const

/**
 * エラーの種類を判定する。レスポンスなし（axios: error.response なし / fetch: TypeError・status 0）
 * またはオフラインの場合は network、それ以外は unknown。
 */
export const getErrorStateType = (error: unknown): ErrorStateType => {
  if (typeof navigator !== 'undefined' && navigator.onLine === false) return 'network'
  if (error instanceof TypeError) return 'network'
  if (typeof error === 'object' && error !== null) {
    const e = error as { isAxiosError?: boolean; response?: unknown; status?: unknown }
    if (e.isAxiosError && !e.response) return 'network'
    // apiClient は fetch の TypeError を status 0 の ApiClientError に変換する
    if (e.status === 0) return 'network'
  }
  return 'unknown'
}

/**
 * エラー状態の共通表示。空状態（EmptyState）と同じ寸法・配置で、
 * 円アイコン → 見出し → 説明 → 再読み込みボタン → 控えめなリンクの順に並べる。
 * 文言はページではなくエラーの種類で決め、APIの生メッセージは出さない。
 */
export const ErrorState: React.FC<ErrorStateProps> = ({
  type,
  onRetry,
  variant = 'page',
  shortHeight = false,
  className,
}) => {
  const { Icon, title, description, subLabel, subHref } = CONTENT[type]

  if (variant === 'compact') {
    return (
      <div
        role="alert"
        className={cn(
          'min-h-[240px] md:min-h-[280px] rounded-lg bg-slate-50 flex flex-col items-center justify-center px-4 py-5 md:p-6 text-center',
          className
        )}
      >
        <div className="w-12 h-12 rounded-full bg-slate-200/60 flex items-center justify-center text-slate-500">
          <Icon className="w-[22px] h-[22px]" strokeWidth={1.75} aria-hidden="true" />
        </div>
        <h3 className="mt-3.5 text-[15px] font-bold text-foreground">{title}</h3>
        <p className="mt-1 max-w-[260px] md:max-w-none text-[13px] leading-[1.7] text-muted-foreground [text-wrap:pretty] [word-break:auto-phrase]">
          {description}
        </p>
        <Button
          variant="outline-gray"
          size="sm"
          onClick={onRetry}
          className="mt-4 gap-1.5 h-11 md:h-9 px-4 md:px-3.5 text-[13.5px] md:text-[13px] font-semibold"
        >
          <RotateCw className="w-[15px] h-[15px]" aria-hidden="true" />
          再読み込み
        </Button>
      </div>
    )
  }

  return (
    <div
      role="alert"
      className={cn(
        shortHeight ? 'min-h-[320px] md:min-h-[360px]' : 'min-h-[400px] md:min-h-[420px]',
        'flex flex-col items-center justify-center py-8 text-center',
        className
      )}
    >
      <div className="w-16 h-16 md:w-[72px] md:h-[72px] rounded-full bg-slate-100 flex items-center justify-center mb-[18px] md:mb-5 text-slate-500">
        <Icon className="w-7 h-7 md:w-8 md:h-8" strokeWidth={1.75} aria-hidden="true" />
      </div>
      <h2 className="text-[17px] md:text-lg font-bold text-foreground mb-2">{title}</h2>
      <p className="text-[13.5px] md:text-sm leading-[1.7] text-muted-foreground mb-[22px] md:mb-6 max-w-[280px] md:max-w-[360px] [text-wrap:pretty] [word-break:auto-phrase]">
        {description}
      </p>
      <Button variant="primary" size="md" onClick={onRetry} className="gap-2 h-11 md:h-10">
        <RotateCw className="w-4 h-4" aria-hidden="true" />
        再読み込み
      </Button>
      <Link
        href={subHref}
        className="mt-2 md:mt-4 min-h-[44px] md:min-h-0 flex items-center text-[13px] font-medium text-muted-foreground underline underline-offset-[3px] hover:text-foreground"
      >
        {subLabel}
      </Link>
    </div>
  )
}

export default ErrorState
