/**
 * 認証付きAPIクライアント
 * 全てのAPI呼び出しで使用する共通クライアント
 */

import { reportMaintenanceIfNeeded } from '@/lib/maintenance/detection'

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

interface ApiError {
  message: string
  status: number
}

/**
 * APIエラークラス
 */
export class ApiClientError extends Error {
  status: number
  
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiClientError'
    this.status = status
  }
}

/**
 * 401エラー時のグローバルハンドラー
 */
let unauthorizedHandler: (() => void) | null = null

/**
 * トークンリフレッシュハンドラー
 */
let refreshTokenHandler: (() => Promise<string | null>) | null = null

/**
 * 進行中のトークンリフレッシュPromise（競合状態を防ぐ）
 */
let ongoingRefreshPromise: Promise<string | null> | null = null

/**
 * 401エラーハンドラーを登録
 */
export const setUnauthorizedHandler = (handler: () => void) => {
  unauthorizedHandler = handler
}

/**
 * トークンリフレッシュハンドラーを登録
 */
export const setRefreshTokenHandler = (handler: () => Promise<string | null>) => {
  refreshTokenHandler = handler
}

/**
 * 認証トークンを取得
 */
const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') return null
  try {
    const token = localStorage.getItem('auth_token')
    return token ? JSON.parse(token) : null
  } catch {
    return null
  }
}

/**
 * 認証付きfetchリクエスト（リトライなし）
 */
const apiClientInternal = async <T = any>(
  endpoint: string,
  options: RequestInit = {},
  token?: string
): Promise<T> => {
  const authToken = token || getAuthToken()
  
  const headers: Record<string, string> = {}
  
  // FormDataの場合はContent-Typeを設定しない（ブラウザが自動設定）
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  
  // 認証トークンがある場合は追加
  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`
  }
  
  // 既存のヘッダーをマージ
  if (options.headers) {
    const existingHeaders = new Headers(options.headers)
    existingHeaders.forEach((value, key) => {
      headers[key] = value
    })
  }
  
  const config: RequestInit = {
    ...options,
    headers,
  }
  
  const response = await fetch(`${API_BASE_URL}${endpoint}`, config)
  
  // レスポンスボディを取得
  const data = await response.json().catch(() => null)
  
  if (!response.ok) {
    // メンテナンス中（503 + MAINTENANCE）なら、全画面のメンテナンス表示へ切り替える
    reportMaintenanceIfNeeded(response.status, data)
    const errorMessage = data?.message || data?.error || `HTTPエラー: ${response.status}`
    const error = new ApiClientError(errorMessage, response.status)
    throw error
  }
  
  return data as T
}

/**
 * 認証付きfetchリクエスト（自動リフレッシュ対応）
 */
export const apiClient = async <T = any>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> => {
  try {
    return await apiClientInternal<T>(endpoint, options)
  } catch (error) {
    if (error instanceof ApiClientError && error.status === 401) {
      // 401エラーの場合、トークンをリフレッシュして再試行
      if (refreshTokenHandler) {
        console.log('🔄 401 error - refreshing token and retrying...')
        
        try {
          // 進行中のリフレッシュがあればそれを待つ（競合状態を防ぐ）
          if (!ongoingRefreshPromise) {
            ongoingRefreshPromise = refreshTokenHandler()
              .finally(() => {
                // リフレッシュ完了後にPromiseをクリア
                ongoingRefreshPromise = null
              })
          }
          
          const newToken = await ongoingRefreshPromise
          
          if (newToken) {
            // 新しいトークンで再試行
            return await apiClientInternal<T>(endpoint, options, newToken)
          }
        } catch (refreshError) {
          console.error('❌ Token refresh failed during API call:', refreshError)
          // リフレッシュ失敗時はunauthorizedHandlerを呼び出してログアウト
          if (unauthorizedHandler) {
            unauthorizedHandler()
          }
          // 元の401エラーを再スロー
          throw error
        }
      }
      
      // リフレッシュハンドラーがない場合はunauthorizedHandlerを呼び出す
      if (unauthorizedHandler) {
        unauthorizedHandler()
      }
    }
    
    // その他のエラー
    if (error instanceof ApiClientError) {
      throw error
    }
    
    // ネットワークエラーなど
    if (error instanceof TypeError) {
      throw new ApiClientError('ネットワークエラーが発生しました', 0)
    }
    
    throw new ApiClientError('予期しないエラーが発生しました', 500)
  }
}

/**
 * GETリクエスト
 */
export const get = <T = any>(endpoint: string, options?: RequestInit): Promise<T> => {
  return apiClient<T>(endpoint, { ...options, method: 'GET' })
}

/**
 * POSTリクエスト
 */
export const post = <T = any>(
  endpoint: string,
  body?: any,
  options?: RequestInit
): Promise<T> => {
  return apiClient<T>(endpoint, {
    ...options,
    method: 'POST',
    body: body ? JSON.stringify(body) : undefined,
  })
}

/**
 * PUTリクエスト
 */
export const put = <T = any>(
  endpoint: string,
  body?: any,
  options?: RequestInit
): Promise<T> => {
  return apiClient<T>(endpoint, {
    ...options,
    method: 'PUT',
    body: body ? JSON.stringify(body) : undefined,
  })
}

/**
 * DELETEリクエスト
 */
export const del = <T = any>(endpoint: string, options?: RequestInit): Promise<T> => {
  return apiClient<T>(endpoint, { ...options, method: 'DELETE' })
}

/**
 * ファイルアップロード用POSTリクエスト
 */
export const uploadFile = <T = any>(
  endpoint: string,
  formData: FormData
): Promise<T> => {
  return apiClient<T>(endpoint, {
    method: 'POST',
    body: formData,
  })
}

/**
 * 内部用POSTリクエスト（リトライなし）
 * トークンリフレッシュなど、無限ループを避けたい場合に使用
 */
export const postInternal = <T = any>(
  endpoint: string,
  body?: any,
  options?: RequestInit
): Promise<T> => {
  return apiClientInternal<T>(endpoint, {
    ...options,
    method: 'POST',
    body: body ? JSON.stringify(body) : undefined,
  })
}
