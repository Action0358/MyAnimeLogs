/**
 * シンプルなトースト通知フック
 * react-hot-toastを導入するまでの暫定実装
 */

import { useState, useCallback } from 'react'

export type ToastType = 'success' | 'error' | 'info' | 'warning'

export interface Toast {
  id: string
  message: string
  type: ToastType
}

export const useToast = () => {
  const [toasts, setToasts] = useState<Toast[]>([])

  const showToast = useCallback((message: string, type: ToastType = 'info') => {
    const id = `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    const newToast: Toast = { id, message, type }

    setToasts((prev) => [...prev, newToast])

    // 3秒後に自動削除
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id))
    }, 3000)
  }, [])

  const success = useCallback(
    (message: string) => {
      showToast(message, 'success')
    },
    [showToast]
  )

  const error = useCallback(
    (message: string) => {
      showToast(message, 'error')
    },
    [showToast]
  )

  const info = useCallback(
    (message: string) => {
      showToast(message, 'info')
    },
    [showToast]
  )

  const warning = useCallback(
    (message: string) => {
      showToast(message, 'warning')
    },
    [showToast]
  )

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  return {
    toasts,
    showToast,
    success,
    error,
    info,
    warning,
    removeToast,
  }
}
