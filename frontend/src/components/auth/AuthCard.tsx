import React from 'react'
import { Container } from '@/components/layout'

interface AuthCardProps {
  title: string
  subtitle?: string
  icon?: string
  children: React.ReactNode
}

export const AuthCard: React.FC<AuthCardProps> = ({ title, subtitle, icon, children }) => {
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
          {children}
        </div>
      </div>
    </Container>
  )
}
