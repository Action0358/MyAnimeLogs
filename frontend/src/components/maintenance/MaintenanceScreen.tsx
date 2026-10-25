'use client'

import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Clock, Wrench } from 'lucide-react'
import { MaintenanceFooter, MaintenanceHeader } from './MaintenanceChrome'
import { ENDS_AT_ROW_ATTR, ENDS_AT_TEXT_ATTR, endsAtText } from '@/lib/maintenance/shared'

export { formatEndsAt } from '@/lib/maintenance/shared'

export interface MaintenanceScreenProps {
  /** 終了予定（ISO 8601）。未指定・不正な値のときは終了予定の行を出さない */
  endsAt?: string
  /** お知らせの掲載先。APIに依存しない場所（静的ページなど）を指定する */
  newsHref?: string
}

/**
 * メンテナンス画面。ヘッダーはロゴのみ・フッターはコピーライトのみの独立した画面で、
 * ナビ・ログイン・お問い合わせは出さない（メンテナンス中はどれも動かないため）。
 * 予定された作業の案内なので role="status" とし、障害のような緊急性は伝えない。
 * ErrorState と同じ型（円アイコン → 見出し → 説明）で揃える。
 * 再読み込みボタンは置かない（他サービスにもなく、押すと一瞬通常画面が見えてちらつくため）。
 * 終了の確認は、ブラウザの再読み込みか、お知らせで行う。
 */
export const MaintenanceScreen: React.FC<MaintenanceScreenProps> = ({
  endsAt,
  newsHref = '/news',
}) => {
  const endsAtLine = endsAtText(endsAt)

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <MaintenanceHeader />

      <main
        role="status"
        className="flex-1 min-h-[400px] md:min-h-[420px] flex flex-col items-center justify-center px-6 py-12 md:p-12 text-center"
      >
        <div className="w-16 h-16 md:w-[72px] md:h-[72px] rounded-full bg-slate-100 flex items-center justify-center mb-[18px] md:mb-5 text-slate-500">
          <Wrench className="w-7 h-7 md:w-8 md:h-8" strokeWidth={1.75} aria-hidden="true" />
        </div>
        <h1 className="text-[17px] md:text-lg font-bold text-foreground mb-2">メンテナンス中です</h1>
        <p className="text-[13.5px] md:text-sm leading-[1.7] text-slate-600 max-w-[300px] md:max-w-[360px] [text-wrap:pretty] [word-break:auto-phrase]">
          サービス向上のため、現在作業を行っています。
          <br className="hidden md:block" />
          終了までしばらくお待ちください。
        </p>
        {/* 終了予定の行は常に描画し、ないときは hidden で隠す。エッジ（functions/_middleware.js）が
            静的に書き出したこの画面を返すとき、終了予定を差し込んで hidden を外すため。
            hidden は外側の要素に付ける（flex を付けた要素では display が上書きされて効かない） */}
        <div {...{ [ENDS_AT_ROW_ATTR]: '' }} hidden={!endsAtLine} className="mt-4">
          <div className="min-h-8 py-1.5 px-3 rounded-md bg-slate-50 text-slate-700 text-[12.5px] md:text-[13px] font-semibold flex items-center gap-1.5">
            <Clock className="w-4 h-4 shrink-0" aria-hidden="true" />
            <span {...{ [ENDS_AT_TEXT_ATTR]: '' }}>{endsAtLine}</span>
          </div>
        </div>
        <Link
          href={newsHref as Route}
          className="mt-4 md:mt-6 min-h-[44px] md:min-h-0 flex items-center text-[13px] font-medium text-slate-600 underline underline-offset-[3px] hover:text-foreground"
        >
          お知らせを見る
        </Link>
      </main>

      <MaintenanceFooter />
    </div>
  )
}

export default MaintenanceScreen
