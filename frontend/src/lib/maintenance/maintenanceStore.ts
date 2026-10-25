/**
 * メンテナンス状態の保持。React の外（APIクライアント・axios interceptor）から更新し、
 * MaintenanceGate が useSyncExternalStore で購読する。
 *
 * 一度メンテナンスと判定したら、ページを再読み込みするまで解除しない。
 * 解除の確認はブラウザの再読み込みで行い、状態がちらつくのを避ける。
 */
export interface MaintenanceState {
  active: boolean
  /** 終了予定（ISO 8601）。不明なら undefined */
  endsAt?: string
}

const INACTIVE: MaintenanceState = { active: false }

let state: MaintenanceState = INACTIVE
const listeners = new Set<() => void>()

export const getMaintenanceState = (): MaintenanceState => state

// SSG / SSR では常に非メンテナンスとして描画する（ハイドレーション不整合を避ける）
export const getServerMaintenanceState = (): MaintenanceState => INACTIVE

export const subscribeMaintenance = (listener: () => void): (() => void) => {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export const setMaintenance = (next: { endsAt?: string | null }): void => {
  const endsAt = next.endsAt || undefined
  // 同じ内容での更新では再描画させない
  if (state.active && state.endsAt === endsAt) return
  state = { active: true, endsAt }
  listeners.forEach((listener) => listener())
}

/** テスト用: 状態を初期化する */
export const resetMaintenance = (): void => {
  state = INACTIVE
  listeners.forEach((listener) => listener())
}
