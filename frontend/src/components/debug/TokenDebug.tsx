'use client'

import { useAuth } from '@/contexts/AuthContext'
import { useEffect, useState } from 'react'
import { storage } from '@/lib/utils'

const TOKEN_EXPIRES_AT_KEY = 'token_expires_at'

export function TokenDebug() {
  const { user } = useAuth()
  const [timeLeft, setTimeLeft] = useState<number>(0)
  const [tokenExpiresAt, setTokenExpiresAt] = useState<number | null>(null)

  useEffect(() => {
    // localStorageから有効期限を取得
    const updateExpiresAt = () => {
      try {
        const expiresAt = storage.getUnsafe(TOKEN_EXPIRES_AT_KEY)
        setTokenExpiresAt(expiresAt)
      } catch {
        setTokenExpiresAt(null)
      }
    }

    updateExpiresAt()

    // 1秒ごとに更新
    const interval = setInterval(updateExpiresAt, 1000)
    return () => clearInterval(interval)
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

  if (!user || !tokenExpiresAt) return null

  const minutes = Math.floor(timeLeft / 60000)
  const seconds = Math.floor((timeLeft % 60000) / 1000)
  const isExpiringSoon = timeLeft < 120000 // 2分以内

  return (
    <div className="fixed bottom-4 right-4 bg-gray-800 text-white p-4 rounded-lg shadow-lg z-50 min-w-[200px]">
      <div className="text-sm font-mono">
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
        <div className="text-gray-500 text-xs mt-2 border-t border-gray-700 pt-2">
          User: {user.name}
        </div>
      </div>
    </div>
  )
}
