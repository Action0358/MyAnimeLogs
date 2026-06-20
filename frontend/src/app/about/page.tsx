'use client'

import React from 'react'
import { Container } from '@/components/layout'

const AboutPage: React.FC = () => {
  return (
    <Container className="py-12">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-gray-900 mb-8">
          MyAnimeLogsとは
        </h1>

        <div className="prose prose-lg max-w-none">
          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">
              サービス概要
            </h2>
            <p className="text-gray-700 leading-relaxed mb-4">
              MyAnimeLogsは、アニメ視聴記録を管理し、自分だけのアニメライブラリを作成できるサービスです。
              視聴したアニメを記録し、評価やレビューを残すことで、あなたのアニメライフをより充実させます。
            </p>
          </section>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">
              主な機能
            </h2>
            <ul className="space-y-3 text-gray-700 list-disc list-inside marker:text-blue-600">
              <li>アニメの視聴記録と評価管理</li>
              <li>今期放送中のアニメ情報</li>
              <li>人気アニメのランキング</li>
              <li>詳細な統計情報の表示</li>
              <li>アニメの検索機能</li>
            </ul>
          </section>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">
              使い方
            </h2>
            <ol className="space-y-3 text-gray-700 list-decimal list-inside">
              <li>アカウントを作成してログイン</li>
              <li>気になるアニメを検索</li>
              <li>視聴状況や評価を記録</li>
              <li>マイページで自分のアニメライブラリを確認</li>
              <li>統計ページで視聴傾向を分析</li>
            </ol>
          </section>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">
              お問い合わせ
            </h2>
            <p className="text-gray-700 leading-relaxed">
              ご質問やご要望がございましたら、お気軽にお問い合わせください。
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">
              データ提供元
            </h2>
            <p className="text-gray-700 leading-relaxed mb-3">
              本サービスのアニメ情報・配信情報は、以下の各サービスより提供されています。
            </p>
            <ul className="space-y-2 text-gray-700 list-disc list-inside marker:text-blue-600">
              <li>
                アニメ基本情報:{' '}
                <a
                  href="https://annict.com"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700"
                >
                  Annict
                </a>
              </li>
              <li>
                カバー画像・ジャンル:{' '}
                <a
                  href="https://anilist.co"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700"
                >
                  AniList
                </a>
              </li>
              <li>
                放送情報:{' '}
                <a
                  href="https://myanimelist.net"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700"
                >
                  MyAnimeList
                </a>{' '}
                （
                <a
                  href="https://jikan.moe"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700"
                >
                  Jikan API
                </a>
                経由）
              </li>
              <li>
                配信情報:{' '}
                <a
                  href="https://www.themoviedb.org/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700"
                >
                  TMDB
                </a>{' '}
                /{' '}
                <a
                  href="https://www.justwatch.com/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700"
                >
                  JustWatch
                </a>
              </li>
            </ul>
            <p className="text-sm text-gray-500 leading-relaxed mt-3">
              This product uses the TMDB API but is not endorsed or certified by TMDB.
            </p>
          </section>
        </div>
      </div>
    </Container>
  )
}

export default AboutPage
