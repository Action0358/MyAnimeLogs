import axios from 'axios'
import {
  getMaintenanceState,
  resetMaintenance,
  setMaintenance,
  subscribeMaintenance,
} from '../maintenanceStore'
import {
  checkStaticMaintenance,
  installMaintenanceAxiosInterceptor,
  isMaintenanceResponse,
  reportMaintenanceIfNeeded,
} from '../detection'

beforeEach(() => {
  resetMaintenance()
})

describe('isMaintenanceResponse', () => {
  it('503 + code=MAINTENANCE だけをメンテナンスとみなす', () => {
    expect(isMaintenanceResponse(503, { code: 'MAINTENANCE' })).toBe(true)
  })

  it('外部API障害などの SERVICE_UNAVAILABLE は含めない', () => {
    expect(isMaintenanceResponse(503, { code: 'SERVICE_UNAVAILABLE' })).toBe(false)
    expect(isMaintenanceResponse(500, { code: 'MAINTENANCE' })).toBe(false)
    expect(isMaintenanceResponse(503, null)).toBe(false)
    expect(isMaintenanceResponse(undefined, undefined)).toBe(false)
  })
})

describe('reportMaintenanceIfNeeded', () => {
  it('メンテナンス応答なら状態を更新し、ends_at を保持する', () => {
    const reported = reportMaintenanceIfNeeded(503, {
      code: 'MAINTENANCE',
      ends_at: '2026-10-25T06:00:00Z',
    })
    expect(reported).toBe(true)
    expect(getMaintenanceState()).toEqual({ active: true, endsAt: '2026-10-25T06:00:00Z' })
  })

  it('それ以外では状態を変えない', () => {
    expect(reportMaintenanceIfNeeded(503, { code: 'SERVICE_UNAVAILABLE' })).toBe(false)
    expect(getMaintenanceState().active).toBe(false)
  })
})

describe('maintenanceStore', () => {
  it('同じ内容の更新では購読者を呼ばない', () => {
    const listener = jest.fn()
    const unsubscribe = subscribeMaintenance(listener)
    setMaintenance({ endsAt: 'x' })
    setMaintenance({ endsAt: 'x' })
    expect(listener).toHaveBeenCalledTimes(1)
    unsubscribe()
  })
})

describe('checkStaticMaintenance', () => {
  const mockFetch = (impl: () => Promise<unknown>) => {
    global.fetch = jest.fn(impl) as unknown as typeof fetch
  }

  it('enabled: true ならメンテナンス状態にする', async () => {
    mockFetch(async () => ({
      ok: true,
      json: async () => ({ enabled: true, endsAt: '2026-10-25T06:00:00Z' }),
    }))
    await checkStaticMaintenance()
    expect(getMaintenanceState()).toEqual({ active: true, endsAt: '2026-10-25T06:00:00Z' })
    expect(global.fetch).toHaveBeenCalledWith('/maintenance.json', { cache: 'no-store' })
  })

  it('enabled: false・取得失敗・不正な形式では何もしない', async () => {
    mockFetch(async () => ({ ok: true, json: async () => ({ enabled: false }) }))
    await checkStaticMaintenance()
    mockFetch(async () => ({ ok: false }))
    await checkStaticMaintenance()
    mockFetch(async () => {
      throw new Error('offline')
    })
    await checkStaticMaintenance()
    mockFetch(async () => ({ ok: true, json: async () => 'broken' }))
    await checkStaticMaintenance()
    expect(getMaintenanceState().active).toBe(false)
  })
})

describe('installMaintenanceAxiosInterceptor', () => {
  it('生成クライアント（axios）の503 MAINTENANCEを検出し、エラー自体は呼び出し元へ返す', async () => {
    installMaintenanceAxiosInterceptor()
    installMaintenanceAxiosInterceptor() // 二重登録しても1回分だけ動く

    const adapter = (config: never) =>
      Promise.reject(
        Object.assign(new Error('Request failed'), {
          config,
          response: {
            status: 503,
            data: { code: 'MAINTENANCE', ends_at: '2026-10-25T06:00:00Z' },
          },
        })
      )

    await expect(axios.get('/api/animes', { adapter } as never)).rejects.toThrow('Request failed')
    expect(getMaintenanceState()).toEqual({ active: true, endsAt: '2026-10-25T06:00:00Z' })
  })
})
