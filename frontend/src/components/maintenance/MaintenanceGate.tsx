'use client'

import React, { useEffect, useSyncExternalStore } from 'react'
import { usePathname } from 'next/navigation'
import {
  checkStaticMaintenance,
  installMaintenanceAxiosInterceptor,
} from '@/lib/maintenance/detection'
import {
  getMaintenanceState,
  getServerMaintenanceState,
  subscribeMaintenance,
} from '@/lib/maintenance/maintenanceStore'
import { MaintenanceScreen } from './MaintenanceScreen'

/**
 * メンテナンス中も開けるパス。お知らせ（/news）は静的生成でAPIに依存しないため、
 * メンテナンス画面の「お知らせを見る」の行き先として生かしておく。
 */
export const isMaintenanceExemptPath = (pathname: string | null): boolean =>
  pathname === '/news' || !!pathname?.startsWith('/news/')

/**
 * アプリ全体を包み、メンテナンス中は子要素の代わりにメンテナンス画面を出す。
 * 静的エクスポート（Cloudflare Pages）ではサーバー側の middleware が使えないため、
 * クライアントで次の2経路から検出する。
 *   1. APIの 503 + MAINTENANCE（バックエンドが宣言している場合）
 *   2. /maintenance.json（バックエンドが止まっていてもフロント単独で宣言する場合）
 * 判定が付くまでは通常どおり描画するので、メンテナンス開始直後は一瞬通常画面が見える。
 */
export const MaintenanceGate: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const pathname = usePathname()
  const state = useSyncExternalStore(
    subscribeMaintenance,
    getMaintenanceState,
    getServerMaintenanceState
  )

  useEffect(() => {
    installMaintenanceAxiosInterceptor()
    void checkStaticMaintenance()
  }, [])

  if (state.active && !isMaintenanceExemptPath(pathname)) {
    return <MaintenanceScreen endsAt={state.endsAt} />
  }
  return <>{children}</>
}

export default MaintenanceGate
