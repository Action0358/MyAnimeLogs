import React from 'react'
import { cn } from '@/lib/utils'

interface ContainerProps {
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  padding?: boolean
  center?: boolean
  children: React.ReactNode
  className?: string
}

const Container: React.FC<ContainerProps> = ({
  size = 'lg',
  padding = true,
  center = true,
  children,
  className,
}) => {
  const sizes = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-4xl',
    // サイト共通のコンテナ幅（本文・フッター）。ヘッダーは列に収めず全幅に置く
    xl: 'max-w-[1180px]',
    full: 'max-w-full',
  }

  return (
    <div
      className={cn(
        'w-full',
        sizes[size],
        center && 'mx-auto',
        padding && 'px-4 sm:px-6 lg:px-8',
        className
      )}
    >
      {children}
    </div>
  )
}

export default Container
