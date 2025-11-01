/**
 * 認証付きAPIクライアント
 * 全てのAPI呼び出しで使用する共通クライアント
 */

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

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
 * 401エラーハンドラーを登録
 */
export const setUnauthorizedHandler = (handler: () => void) => {
  unauthorizedHandler = handler
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
 * 認証付きfetchリクエスト
 */
export const apiClient = async <T = any>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> => {
  const token = getAuthToken()
  
  const headers: Record<string, string> = {}
  
  // FormDataの場合はContent-Typeを設定しない（ブラウザが自動設定）
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  
  // 認証トークンがある場合は追加
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
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
  
  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, config)
    
    // レスポンスボディを取得
    const data = await response.json().catch(() => null)
    
    if (!response.ok) {
      const errorMessage = data?.message || data?.error || `HTTPエラー: ${response.status}`
      const error = new ApiClientError(errorMessage, response.status)
      
      // 401エラーの場合はグローバルハンドラーを呼び出す
      if (response.status === 401 && unauthorizedHandler) {
        unauthorizedHandler()
      }
      
      throw error
    }
    
    return data as T
  } catch (error) {
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
