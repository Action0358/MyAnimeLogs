'use client'

import React, { useEffect, useSyncExternalStore } from 'react'
import type { Route } from 'next'
import { usePathname, useRouter } from 'next/navigation'
import {
  checkStaticMaintenance,
  installMaintenanceAxiosInterceptor,
  readEdgeMaintenanceMarker,
} from '@/lib/maintenance/detection'
import {
  getMaintenanceState,
  getServerMaintenanceState,
  subscribeMaintenance,
} from '@/lib/maintenance/maintenanceStore'
import { MAINTENANCE_PAGE_PATH, isMaintenanceExemptPath } from '@/lib/maintenance/shared'
import { MaintenanceScreen } from './MaintenanceScreen'

export { isMaintenanceExemptPath } from '@/lib/maintenance/shared'

/** /maintenance と /maintenance/ のどちらでも true */
const isMaintenancePage = (pathname: string | null): boolean =>
  pathname?.replace(/\/$/, '') === MAINTENANCE_PAGE_PATH

/**
 * アプリ全体を包み、メンテナンス中は子要素の代わりにメンテナンス画面を出す。
 *
 * ページを開いた時点のメンテナンスは、エッジ（functions/_middleware.js）が
 * /maintenance の静的 HTML をスクリプト抜きで返すため、このコンポーネントは動かない。
 * ここで扱うのは、エッジを通らない次の場合。
 *   1. 閲覧中にメンテナンスが始まった（APIの 503 + MAINTENANCE を検出）
 *   2. /maintenance.json での宣言（エッジの判定をすり抜けたときの保険）
 *   3. メンテナンス中に開放している /news（エッジが差し込む meta を読む）
 *
 * /maintenance はエッジが返す画面の元になるページで、ここで画面に差し替えて描画する
 * （静的 HTML にヘッダー・フッターを含めないため）。メンテナンス中でないときに
 * 直接開かれたら、トップへ戻す。
 */
export const MaintenanceGate: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const pathname = usePathname()
  const router = useRouter()
  const state = useSyncExternalStore(
    subscribeMaintenance,
    getMaintenanceState,
    getServerMaintenanceState
  )
  const onMaintenancePage = isMaintenancePage(pathname)

  useEffect(() => {
    installMaintenanceAxiosInterceptor()
    readEdgeMaintenanceMarker()
    void checkStaticMaintenance()
  }, [])

  useEffect(() => {
    if (onMaintenancePage && !state.active) router.replace('/' as Route)
  }, [onMaintenancePage, state.active, router])

  if (onMaintenancePage || (state.active && !isMaintenanceExemptPath(pathname))) {
    return <MaintenanceScreen endsAt={state.endsAt} />
  }
  return <>{children}</>
}

export default MaintenanceGate
