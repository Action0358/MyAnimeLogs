'use client';

import type React from 'react';
import { Container } from '@/components/layout';
import { CurrentSeasonAnime, PopularAnime } from '@/components/anime';
import { useAuth } from '@/contexts/AuthContext';
import Link from 'next/link';

const HomePage: React.FC = () => {
  const { isAuthenticated } = useAuth();

  return (
    <Container className="py-12">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          MyAnimeLogs へようこそ
        </h1>
        <p className="text-xl text-gray-600 mb-8">
          あなたのアニメ視聴体験を記録・管理しましょう
        </p>
        
        {!isAuthenticated && (
          <div className="flex justify-center space-x-4">
            <Link 
              href="/register"
              className="inline-flex items-center justify-center h-11 px-6 text-base font-medium rounded-md bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800 transition-colors"
            >
              今すぐ始める
            </Link>
          </div>
        )}
        
        {/* 機能紹介 */}
        <div className="mt-16">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="text-center p-6 bg-white rounded-lg border border-gray-200">
              <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mx-auto mb-4">
                <span className="text-blue-600 text-xl">📺</span>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">アニメ管理</h3>
              <p className="text-gray-600">視聴中・完了・予定のアニメを簡単に管理</p>
            </div>
            
            <div className="text-center p-6 bg-white rounded-lg border border-gray-200">
              <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mx-auto mb-4">
                <span className="text-green-600 text-xl">📊</span>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">視聴統計</h3>
              <p className="text-gray-600">視聴時間やジャンル傾向を可視化</p>
            </div>
            
            <div className="text-center p-6 bg-white rounded-lg border border-gray-200">
              <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mx-auto mb-4">
                <span className="text-purple-600 text-xl">⭐</span>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">評価・レビュー</h3>
              <p className="text-gray-600">アニメに評価とコメントを投稿</p>
            </div>
            
            <div className="text-center p-6 bg-white rounded-lg border border-gray-200">
              <div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center mx-auto mb-4">
                <span className="text-orange-600 text-xl">🔍</span>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">アニメ検索</h3>
              <p className="text-gray-600">豊富なデータベースから作品を検索</p>
            </div>
          </div>
        </div>

        {/* 今期アニメセクション */}
        <div className="mt-16">
          <CurrentSeasonAnime />
        </div>

        {/* 人気アニメセクション */}
        <div className="mt-16">
          <PopularAnime />
        </div>
      </div>
    </Container>
  );
};

export default HomePage;