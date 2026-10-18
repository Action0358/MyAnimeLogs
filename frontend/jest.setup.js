import '@testing-library/jest-dom'

// jsdomはmatchMediaを実装していないため、レスポンシブ判定を使うコンポーネントのテスト用にポリフィルする
if (typeof window !== 'undefined' && !window.matchMedia) {
  window.matchMedia = (query) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    dispatchEvent: () => false,
  })
}