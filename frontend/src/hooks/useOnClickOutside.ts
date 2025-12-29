'use client'

import { useEffect, RefObject } from 'react'

/**
 * 指定された要素の外側がクリックされたときにコールバックを実行するカスタムフック
 * mousedownとtouchstartの両方をリッスンしてモバイル端末にも対応
 */
export const useOnClickOutside = <T extends HTMLElement>(
  ref: RefObject<T | null>,
  handler: () => void
): void => {
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent | TouchEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        handler()
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    document.addEventListener('touchstart', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('touchstart', handleClickOutside)
    }
  }, [ref, handler])
}
