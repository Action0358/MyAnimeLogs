import axios from 'axios'
import { setMaintenance } from './maintenanceStore'
import { MAINTENANCE_META_NAME } from './shared'

/** バックエンドの MAINTENANCE エラーコード（backend/internal/common/errors.go と合わせる） */
export const MAINTENANCE_CODE = 'MAINTENANCE'

interface MaintenanceBody {
  code?: unknown
  ends_at?: unknown
}

/** 503 + code=MAINTENANCE か。外部API障害などの SERVICE_UNAVAILABLE は含めない */
export const isMaintenanceResponse = (status: number | undefined, body: unknown): boolean => {
  if (status !== 503 || typeof body !== 'object' || body === null) return false
  return (body as MaintenanceBody).code === MAINTENANCE_CODE
}

/** レスポンスがメンテナンスなら状態へ反映する。反映したら true */
export const reportMaintenanceIfNeeded = (status: number | undefined, body: unknown): boolean => {
  if (!isMaintenanceResponse(status, body)) return false
  const endsAt = (body as MaintenanceBody).ends_at
  setMaintenance({ endsAt: typeof endsAt === 'string' ? endsAt : undefined })
  return true
}

let axiosInterceptorInstalled = false

/**
 * 生成クライアント（axios）のレスポンスを監視し、メンテナンス応答を検出する。
 * fetch ベースの apiClient は client.ts 側で検出する。
 * 二重登録を防ぐため一度だけ登録する。
 */
export const installMaintenanceAxiosInterceptor = (): void => {
  if (axiosInterceptorInstalled) return
  axiosInterceptorInstalled = true
  axios.interceptors.response.use(
    (response) => response,
    (error: unknown) => {
      const response = (error as { response?: { status?: number; data?: unknown } })?.response
      reportMaintenanceIfNeeded(response?.status, response?.data)
      return Promise.reject(error)
    }
  )
}

/**
 * 静的ファイル /maintenance.json を確認する。バックエンドが止まっていても
 * フロント単独でメンテナンスを宣言できるようにするための経路。
 * 取得できない・形式が違う場合は何もしない（通常運用を妨げない）。
 */
export const checkStaticMaintenance = async (): Promise<void> => {
  if (typeof fetch === 'undefined') return
  try {
    const res = await fetch('/maintenance.json', { cache: 'no-store' })
    if (!res.ok) return
    const data: unknown = await res.json()
    if (typeof data !== 'object' || data === null) return
    const { enabled, endsAt } = data as { enabled?: unknown; endsAt?: unknown }
    if (enabled === true) {
      setMaintenance({ endsAt: typeof endsAt === 'string' ? endsAt : undefined })
    }
  } catch {
    // 取得失敗は無視する
  }
}

/**
 * エッジがメンテナンス中の開放ページ（/news）に差し込んだ meta を読む。
 * あればメンテナンス状態にする（content は終了予定。空なら終了予定なし）。
 * API を呼ばなくても /news のヘッダー・フッターをメンテナンス用にできる。
 */
export const readEdgeMaintenanceMarker = (): void => {
  if (typeof document === 'undefined') return
  const meta = document.querySelector<HTMLMetaElement>(`meta[name="${MAINTENANCE_META_NAME}"]`)
  if (!meta) return
  setMaintenance({ endsAt: meta.content || undefined })
}
