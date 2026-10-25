import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'メンテナンス中 | MyAnimeLogs',
  robots: { index: false, follow: false },
}

/**
 * メンテナンス画面の元になるページ（#1136）。
 *
 * 画面そのものは MaintenanceGate がこのパスで MaintenanceScreen に差し替えて描画する
 * （ヘッダー・フッターを含めない全画面にするため）。ここは中身を持たない。
 * 静的エクスポートした out/maintenance/index.html を、メンテナンス中にエッジ
 * （functions/_middleware.js）がどの URL に対しても返す。
 */
const MaintenancePage = () => null

export default MaintenancePage
