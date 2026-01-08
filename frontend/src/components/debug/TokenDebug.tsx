'use client'

import { useAuth } from '@/contexts/AuthContext'
import { useEffect, useState } from 'react'
import { storage } from '@/lib/utils'
import { 
  STORAGE_KEYS, 
  TOKEN_EXPIRING_SOON_THRESHOLD, 
  INACTIVITY_TIMEOUT,
  INACTIVITY_WARNING_THRESHOLD 
} from '@/lib/constants'

export function TokenDebug() {
  const { user, lastActivityTime } = useAuth()
  const [timeLeft, setTimeLeft] = useState<number>(0)
  const [tokenExpiresAt, setTokenExpiresAt] = useState<number | null>(null)
  const [inactivityTimeLeft, setInactivityTimeLeft] = useState<number>(INACTIVITY_TIMEOUT)

  useEffect(() => {
    // localStorageから有効期限を取得
    const updateExpiresAt = () => {
      const expiresAt = storage.getNumber(STORAGE_KEYS.TOKEN_EXPIRES_AT)
      setTokenExpiresAt(expiresAt)
    }

    // 初回取得
    updateExpiresAt()

    // storageイベントリスナーで他のタブでの変更を検知
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === STORAGE_KEYS.TOKEN_EXPIRES_AT) {
        updateExpiresAt()
      }
    }

    window.addEventListener('storage', handleStorageChange)
    return () => window.removeEventListener('storage', handleStorageChange)
  }, [])

  useEffect(() => {
    if (!tokenExpiresAt) {
      setTimeLeft(0)
      return
    }

    const interval = setInterval(() => {
      const left = Math.max(0, tokenExpiresAt - Date.now())
      setTimeLeft(left)
    }, 1000)

    return () => clearInterval(interval)
  }, [tokenExpiresAt])

  // 非アクティブ時間のカウントダウン
  // AuthContextから最終アクティビティ時刻を取得して計算
  useEffect(() => {
    if (!user) return

    const interval = setInterval(() => {
      const elapsed = Date.now() - lastActivityTime
      const left = Math.max(0, INACTIVITY_TIMEOUT - elapsed)
      setInactivityTimeLeft(left)
    }, 1000)

    return () => clearInterval(interval)
  }, [user, lastActivityTime])

  if (!user || !tokenExpiresAt) return null

  const minutes = Math.floor(timeLeft / 60000)
  const seconds = Math.floor((timeLeft % 60000) / 1000)
  const isExpiringSoon = timeLeft < TOKEN_EXPIRING_SOON_THRESHOLD

  const inactivityMinutes = Math.floor(inactivityTimeLeft / 60000)
  const inactivitySeconds = Math.floor((inactivityTimeLeft % 60000) / 1000)
  const isInactivityWarning = inactivityTimeLeft < INACTIVITY_WARNING_THRESHOLD

  return (
    <div className="fixed bottom-4 right-4 bg-gray-800 text-white p-4 rounded-lg shadow-lg z-50 min-w-[240px]">
      <div className="text-sm font-mono space-y-3">
        {/* トークン有効期限 */}
        <div>
          <div className="text-gray-400 mb-1">Token expires in:</div>
          <div className={`text-2xl font-bold ${isExpiringSoon ? 'text-yellow-400' : 'text-green-400'}`}>
            {minutes}:{seconds.toString().padStart(2, '0')}
          </div>
          {isExpiringSoon && timeLeft > 0 && (
            <div className="text-yellow-400 text-xs mt-2 flex items-center gap-1">
              <span>⚠️</span>
              <span>Will refresh soon</span>
            </div>
          )}
          {timeLeft === 0 && (
            <div className="text-red-400 text-xs mt-2 flex items-center gap-1">
              <span>❌</span>
              <span>Token expired</span>
            </div>
          )}
        </div>

        {/* 非アクティブタイムアウト */}
        <div className="border-t border-gray-700 pt-3">
          <div className="text-gray-400 mb-1">Inactivity timeout:</div>
          <div className={`text-xl font-bold ${isInactivityWarning ? 'text-red-400' : 'text-blue-400'}`}>
            {inactivityMinutes}:{inactivitySeconds.toString().padStart(2, '0')}
          </div>
          {isInactivityWarning && inactivityTimeLeft > 0 && (
            <div className="text-red-400 text-xs mt-2 flex items-center gap-1">
              <span>⚠️</span>
              <span>Move mouse to reset</span>
            </div>
          )}
          {inactivityTimeLeft === 0 && (
            <div className="text-red-400 text-xs mt-2 flex items-center gap-1">
              <span>❌</span>
              <span>Logging out...</span>
            </div>
          )}
        </div>

        {/* ユーザー情報 */}
        <div className="text-gray-500 text-xs border-t border-gray-700 pt-2">
          User: {user.name}
        </div>
      </div>
    </div>
  )
}
