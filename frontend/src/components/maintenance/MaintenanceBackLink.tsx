'use client'

import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { ChevronLeft } from 'lucide-react'
import { useMaintenanceActive } from '@/lib/maintenance/useMaintenanceActive'

/**
 * メンテナンス中に /news を開いたときだけ出す「戻る」リンク。
 * お知らせ詳細の「‹ お知らせ一覧」と同じ位置・同じ型にして、タイトルの上に置く。
 * 行き先の "/" はゲートがメンテナンス画面に差し替える。
 *
 * 通常時も静的 HTML には入れておき、CSS で隠す。エッジが <html data-maintenance> を
 * 付けたときは JS が動く前から表示され、リロード時に後から押し出されて出ない。
 */
export const MaintenanceBackLink: React.FC = () => {
  const active = useMaintenanceActive()

  return (
    <Link
      href={'/' as Route}
      data-show-when={active ? undefined : 'maintenance'}
      className="-my-2.5 mb-2.5 inline-flex min-h-[44px] items-center gap-1 text-[12.5px] font-semibold text-muted-foreground transition-colors hover:text-primary md:mb-4 md:min-h-0 md:text-[13px]"
    >
      <ChevronLeft aria-hidden="true" className="h-3.5 w-3.5 md:h-4 md:w-4" />
      戻る
    </Link>
  )
}

export default MaintenanceBackLink
