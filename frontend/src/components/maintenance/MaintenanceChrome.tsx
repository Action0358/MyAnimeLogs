'use client'

import React from 'react'
import Image from 'next/image'
import { SUB_LOGO_PATH } from '@/lib/constants'

/**
 * 静的 HTML に通常版と並べて出すときの印。'maintenance' なら、エッジが
 * <html data-maintenance> を付けたときだけ CSS で表示される（globals.css）。
 * 'normal' はその逆。クライアントがメンテナンスを検出した後は付けない。
 */
export type ShowWhen = 'maintenance' | 'normal'

interface ChromeProps {
  showWhen?: ShowWhen
}

/**
 * メンテナンス中に使う、ロゴのみのヘッダー（リンクなし）。ナビ・ログインは出さない
 * （メンテナンス中はどれも動かないため）。メンテナンス画面と、メンテナンス中に
 * 開ける /news の両方で同じ見た目にする。
 */
export const MaintenanceHeader: React.FC<ChromeProps> = ({ showWhen }) => (
  <header
    data-show-when={showWhen}
    className="h-14 md:h-16 flex items-center px-4 md:px-7 border-b border-border"
  >
    <Image
      src={SUB_LOGO_PATH}
      alt="MyAnimeLogs"
      width={1792}
      height={576}
      priority
      className="h-9 md:h-12 w-auto"
    />
  </header>
)

/** メンテナンス中に使う、コピーライトのみのフッター */
export const MaintenanceFooter: React.FC<ChromeProps> = ({ showWhen }) => (
  <footer
    data-show-when={showWhen}
    className="border-t border-border px-4 md:px-7 py-[22px] text-[11.5px] md:text-xs text-muted-foreground"
  >
    &copy; {new Date().getFullYear()} MyAnimeLogs. All rights reserved.
  </footer>
)
