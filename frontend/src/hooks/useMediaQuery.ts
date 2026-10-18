import { useEffect, useState } from 'react'

/**
 * メディアクエリの一致状態を返す。SSR時は`window`が無いため常にfalseを返し、
 * マウント後に`matchMedia`で実際の値へ補正する（初回だけ再レンダリングが発生する）。
 */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(false)

  useEffect(() => {
    const mediaQueryList = window.matchMedia(query)
    setMatches(mediaQueryList.matches)

    const handleChange = (event: MediaQueryListEvent) => setMatches(event.matches)
    mediaQueryList.addEventListener('change', handleChange)
    return () => mediaQueryList.removeEventListener('change', handleChange)
  }, [query])

  return matches
}
