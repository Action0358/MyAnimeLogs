'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'

const LoginPage: React.FC = () => {
  const { login } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setIsLoading(true)

    try {
      await login(email, password)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'ログインに失敗しました')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Container className="py-12">
      <div className="max-w-md mx-auto">
        {/* ヘッダー */}
        <div className="text-center mb-8">
          <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-blue-600 text-3xl">📺</span>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">ログイン</h1>
          <p className="text-gray-600">MyAnimeLogsへようこそ</p>
        </div>

        {/* ログインフォーム */}
        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="p-4 bg-red-50 border border-red-200 rounded-md">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <Input
              label="メールアドレス"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="example@email.com"
              required
              autoComplete="email"
            />

            <Input
              label="パスワード"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="8文字以上"
              required
              autoComplete="current-password"
            />

            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isLoading}
              className="w-full"
            >
              ログイン
            </Button>
          </form>

          {/* 新規登録リンク */}
          <div className="mt-6 pt-6 border-t border-gray-200 text-center">
            <p className="text-sm text-gray-600">
              アカウントをお持ちでない方は{' '}
              <Link
                href="/register"
                className="text-blue-600 hover:text-blue-700 font-medium transition-colors"
              >
                新規登録
              </Link>
            </p>
          </div>
        </div>

        {/* 機能紹介 */}
        <div className="mt-8">
          <p className="text-center text-sm text-gray-600 mb-4">
            ログインすると以下の機能をご利用いただけます
          </p>
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <span className="text-blue-600 text-lg">📺</span>
              </div>
              <p className="text-sm text-gray-700 font-medium">視聴管理</p>
            </div>

            <div className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <span className="text-green-600 text-lg">📊</span>
              </div>
              <p className="text-sm text-gray-700 font-medium">視聴統計</p>
            </div>

            <div className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <span className="text-purple-600 text-lg">⭐</span>
              </div>
              <p className="text-sm text-gray-700 font-medium">評価投稿</p>
            </div>

            <div className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <span className="text-orange-600 text-lg">🔍</span>
              </div>
              <p className="text-sm text-gray-700 font-medium">アニメ検索</p>
            </div>
          </div>
        </div>
      </div>
    </Container>
  )
}

export default LoginPage
