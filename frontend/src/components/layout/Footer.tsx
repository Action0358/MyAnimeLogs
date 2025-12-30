import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import Container from './Container'

const Footer: React.FC = () => {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-white border-t border-gray-200 mt-16">
      <Container>
        <div className="py-12">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-8">
            {/* サイト情報 */}
            <div>
              <div className="h-7 flex items-center mb-4">
                <Link href={"/" as Route}>
                  <img
                    src="/logo2.svg"
                    alt="MyAnimeLogs"
                    className="h-10 w-auto"
                  />
                </Link>
              </div>
              <p className="text-gray-600 mb-3">あなたの視聴記録がここに</p>
              <p className="text-sm text-gray-500">
                アニメデータは{' '}
                <Link
                  href="https://annict.com"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700"
                >
                  Annict
                </Link>{' '}
                より提供
              </p>
            </div>

            {/* サイトマップ */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                サイトマップ
              </h3>
              <ul className="space-y-2">
                <li>
                  <Link
                    href="/"
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    ホーム
                  </Link>
                </li>
                <li>
                  <Link
                    href="/search"
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    検索
                  </Link>
                </li>
                <li>
                  <Link
                    href="/mypage"
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    マイページ
                  </Link>
                </li>
              </ul>
            </div>

            {/* サポート */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                サポート
              </h3>
              <ul className="space-y-2">
                <li>
                  <Link
                    href={'/help' as Route}
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    ヘルプ
                  </Link>
                </li>
                <li>
                  <Link
                    href={'/contact' as Route}
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    お問い合わせ
                  </Link>
                </li>
                <li>
                  <Link
                    href={'/privacy' as Route}
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    プライバシーポリシー
                  </Link>
                </li>
                <li>
                  <Link
                    href={'/terms' as Route}
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    利用規約
                  </Link>
                </li>
              </ul>
            </div>
          </div>

          {/* コピーライト */}
          <div className="border-t border-gray-200 mt-8 pt-8">
            <div className="flex flex-col md:flex-row justify-between items-center">
              <p className="text-gray-500 text-sm">
                &copy; {currentYear} MyAnimeLogs. All rights reserved.
              </p>
              <div className="flex items-center space-x-4 mt-4 md:mt-0">
                <span className="text-gray-500 text-sm">
                  Made with ❤️ for anime lovers
                </span>
              </div>
            </div>
          </div>
        </div>
      </Container>
    </footer>
  )
}

export default Footer
