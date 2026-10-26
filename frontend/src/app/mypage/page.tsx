'use client'

import React from 'react'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { MyPageContent } from '@/components/mypage'

const MyPage: React.FC = () => {
  return (
    <ProtectedRoute>
      <MyPageContent />
    </ProtectedRoute>
  )
}

export default MyPage
