'use client'

import React, { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { cn } from '@/lib/utils'

interface CopyEmailButtonProps {
  email: string
  className?: string
}

// クリップボード操作はサーバーコンポーネントにできないため、このボタンだけ切り出す
export const CopyEmailButton: React.FC<CopyEmailButtonProps> = ({ email, className }) => {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(email)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      type="button"
      onClick={handleCopy}
      className={cn(
        'flex h-9 flex-shrink-0 items-center gap-1.5 whitespace-nowrap rounded-md border border-border bg-card px-3.5 text-xs font-semibold text-foreground/80 transition-colors hover:bg-muted focus:outline-none focus-visible:ring-2 focus-visible:ring-ring',
        className
      )}
    >
      {copied ? (
        <>
          <Check className="h-3.5 w-3.5 text-primary" aria-hidden="true" />
          コピーしました
        </>
      ) : (
        <>
          <Copy className="h-3.5 w-3.5" aria-hidden="true" />
          コピー
        </>
      )}
    </button>
  )
}

export default CopyEmailButton
