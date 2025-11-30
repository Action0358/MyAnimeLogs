'use client';

import React, { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { Container } from '@/components/layout';

function SearchContent() {
  const searchParams = useSearchParams();
  const keyword = searchParams.get('keyword');
  const season = searchParams.get('season');
  const genre = searchParams.get('genre');

  return (
    <Container className="py-12">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">検索結果</h1>

        {/* 検索条件の表示 */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-2">検索条件</h2>
          <div className="space-y-1 text-gray-700">
            {keyword && (
              <p>
                <span className="font-medium">キーワード:</span> {keyword}
              </p>
            )}
            {season && (
              <p>
                <span className="font-medium">放送時期:</span> {season}
              </p>
            )}
            {genre && (
              <p>
                <span className="font-medium">ジャンル:</span> {genre}
              </p>
            )}
          </div>
        </div>

        {/* 検索結果（今後実装） */}
        <div className="bg-white border border-gray-200 rounded-lg p-8 text-center">
          <p className="text-gray-600">
            検索機能は現在開発中です。
            <br />
            バックエンドAPIとの連携後に実装されます。
          </p>
        </div>
      </div>
    </Container>
  );
}

const SearchPage: React.FC = () => {
  return (
    <Suspense fallback={
      <Container className="py-12">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">読み込み中...</p>
        </div>
      </Container>
    }>
      <SearchContent />
    </Suspense>
  );
};

export default SearchPage;
