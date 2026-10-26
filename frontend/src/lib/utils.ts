import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// フォームバリデーション用ユーティリティ
export const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

export const validatePassword = (password: string): boolean => {
  return password.length >= 8
}

// エラーメッセージ取得
export const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error) return error.message
  if (typeof error === 'string') return error
  return 'An unexpected error occurred'
}

// ローカルストレージ操作
export const storage = {
  /**
   * localStorageから値を取得（エラーを捕捉してnullを返す）
   * 注意: 型安全性が必要な場合はgetString, getNumber, getObjectを使用してください
   */
  get: (key: string): unknown => {
    if (typeof window === 'undefined') return null
    try {
      const item = localStorage.getItem(key)
      if (!item) return null
      return JSON.parse(item)
    } catch (error) {
      console.error(`Failed to get item from localStorage with key "${key}":`, error)
      return null
    }
  },

  /**
   * localStorageから文字列を取得（型検証付き）
   */
  getString: (key: string): string | null => {
    const value = storage.get(key)
    return typeof value === 'string' ? value : null
  },

  /**
   * localStorageから数値を取得（型検証付き）
   */
  getNumber: (key: string): number | null => {
    const value = storage.get(key)
    return typeof value === 'number' && !isNaN(value) ? value : null
  },

  /**
   * localStorageからオブジェクトを取得（型検証付き）
   * @param validator オブジェクトの型を検証する関数
   */
  getObject: <T>(key: string, validator: (value: unknown) => value is T): T | null => {
    const value = storage.get(key)
    return validator(value) ? value : null
  },

  set: (key: string, value: unknown) => {
    if (typeof window === 'undefined') return
    try {
      localStorage.setItem(key, JSON.stringify(value))
    } catch (error) {
      console.error(`Failed to set item in localStorage with key "${key}":`, error)
    }
  },

  remove: (key: string) => {
    if (typeof window === 'undefined') return
    try {
      localStorage.removeItem(key)
    } catch (error) {
      console.error(`Failed to remove item from localStorage with key "${key}":`, error)
    }
  },
}
