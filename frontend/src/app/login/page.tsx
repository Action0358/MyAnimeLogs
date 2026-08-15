'use client'

import React, { Suspense } from 'react'
import { RedirectToSignIn } from '@/components/auth/RedirectToSignIn'

/**
 * 旧パス。ログインと新規登録を /signin に統合したため転送のみを行う。
 * 外部リンクやブックマークを壊さないために残している。
 */
const LegacyAuthPage: React.FC = () => (
  <Suspense fallback={null}>
    <RedirectToSignIn />
  </Suspense>
)

export default LegacyAuthPage
