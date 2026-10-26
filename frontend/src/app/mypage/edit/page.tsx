'use client'

import React from 'react'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { SettingsContent } from '@/components/mypage'

const SettingsPage: React.FC = () => {
  return (
    <ProtectedRoute>
      <SettingsContent />
    </ProtectedRoute>
  )
}

export default SettingsPage
