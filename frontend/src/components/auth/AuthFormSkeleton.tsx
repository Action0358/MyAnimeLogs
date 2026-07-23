import React from 'react'
import { Container } from '@/components/layout'

interface AuthFormSkeletonProps {
  title?: string
  subtitle?: string
  icon?: string
}

export const AuthFormSkeleton: React.FC<AuthFormSkeletonProps> = ({
  title = 'ログイン',
  subtitle = 'MyAnimeLogsへようこそ',
  icon = '📺',
}) => {
  return (
    <Container className="py-12">
      <div className="max-w-md mx-auto">
        <div className="text-center mb-8">
          {icon && (
            <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <span className="text-blue-600 text-3xl">{icon}</span>
            </div>
          )}
          <h1 className="text-3xl font-bold text-gray-900 mb-2">{title}</h1>
          {subtitle && <p className="text-gray-600">{subtitle}</p>}
        </div>

        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
          <div className="animate-pulse space-y-6">
            <div className="h-10 bg-gray-200 rounded"></div>
            <div className="h-10 bg-gray-200 rounded"></div>
            <div className="h-12 bg-gray-200 rounded"></div>
          </div>
        </div>
      </div>
    </Container>
  )
}
