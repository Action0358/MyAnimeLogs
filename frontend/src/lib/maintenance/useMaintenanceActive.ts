import { useSyncExternalStore } from 'react'
import {
  getMaintenanceState,
  getServerMaintenanceState,
  subscribeMaintenance,
} from './maintenanceStore'

/** メンテナンス中か。SSG/SSR とハイドレーション時は常に false */
export const useMaintenanceActive = (): boolean =>
  useSyncExternalStore(
    subscribeMaintenance,
    () => getMaintenanceState().active,
    () => getServerMaintenanceState().active
  )
