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
            <ul className="space-y-3 text-gray-700">
              <li className="flex items-start">
                <span className="text-blue-600 mr-2">•</span>
                <span>アニメの視聴記録と評価管理</span>
              </li>
              <li className="flex items-start">
                <span className="text-blue-600 mr-2">•</span>
                <span>今期放送中のアニメ情報</span>
              </li>
              <li className="flex items-start">
                <span className="text-blue-600 mr-2">•</span>
                <span>人気アニメのランキング</span>
              </li>
              <li className="flex items-start">
                <span className="text-blue-600 mr-2">•</span>
                <span>詳細な統計情報の表示</span>
              </li>
              <li className="flex items-start">
                <span className="text-blue-600 mr-2">•</span>
                <span>アニメの検索機能</span>
              </li>
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

          <section>
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">
              お問い合わせ
            </h2>
            <p className="text-gray-700 leading-relaxed">
              ご質問やご要望がございましたら、お気軽にお問い合わせください。
            </p>
          </section>
        </div>
      </div>
    </Container>
  )
}

export default AboutPage
