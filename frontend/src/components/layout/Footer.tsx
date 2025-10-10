import React from 'react'
import Link from 'next/link'
import Container from './Container'

const Footer: React.FC = () => {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-white border-t border-gray-200 mt-16">
      <Container>
        <div className="py-12">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            {/* サイト情報 */}
            <div className="md:col-span-2">
              <div className="flex items-center space-x-2 mb-4">
                <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                  <span className="text-white font-bold text-sm">MA</span>
                </div>
                <span className="text-xl font-bold text-gray-900">
                  MyAnimeLogs
                </span>
              </div>
              <p className="text-gray-600 mb-4">
                あなたのアニメ視聴体験を記録・管理プラットフォーム
              </p>
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

            {/* ナビゲーション */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                サイトマップ
              </h3>
              <ul className="space-y-2">
                <li>
                  <Link
                    href="/animes"
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    アニメ一覧
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
                    href="/about"
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    サイトについて
                  </Link>
                </li>
              </ul>
            </div>

            {/* サポート・法的情報 */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                サポート
              </h3>
              <ul className="space-y-2">
                <li>
                  <Link
                    href="/help"
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    ヘルプ
                  </Link>
                </li>
                <li>
                  <Link
                    href="/contact"
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    お問い合わせ
                  </Link>
                </li>
                <li>
                  <Link
                    href="/privacy"
                    className="text-gray-600 hover:text-gray-900 transition-colors"
                  >
                    プライバシーポリシー
                  </Link>
                </li>
                <li>
                  <Link
                    href="/terms"
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
