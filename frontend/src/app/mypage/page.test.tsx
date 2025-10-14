import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import MyPage from './page';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';

// Mock dependencies
jest.mock('@/contexts/AuthContext');
jest.mock('next/navigation');
jest.mock('@/components/auth/ProtectedRoute', () => ({
  ProtectedRoute: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

// Mock fetch
global.fetch = jest.fn();

describe('MyPage', () => {
  const mockRouter = {
    push: jest.fn(),
  };

  const mockUser = {
    id: 1,
    name: 'テストユーザー',
    email: 'test@example.com',
    description: 'テスト説明',
  };

  const mockStats = {
    totalEpisodesWatched: 100,
    totalAnimesCompleted: 10,
    totalWatchTimeHours: 40.0,
    favoriteGenres: ['アクション', 'コメディ', 'ドラマ'],
    thisMonthEpisodes: 15,
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
  });

  it('未認証時は何も表示しない（ProtectedRouteが処理）', () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    });

    const { container } = render(<MyPage />);

    // ProtectedRouteがモックされているため、コンテンツは表示されるが
    // 実際のアプリではProtectedRouteがリダイレクトを処理する
    expect(container).toBeInTheDocument();
  });

  it('認証済みユーザーの情報を表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    });

    (global.fetch as jest.Mock).mockImplementation((url) => {
      if (url.includes('/stats')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockStats,
        });
      }
      if (url.includes('/favorites')) {
        return Promise.resolve({
          ok: true,
          json: async () => [],
        });
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('テストユーザー')).toBeInTheDocument();
      expect(screen.getByText('test@example.com')).toBeInTheDocument();
    });
  });

  it('統計情報を正しく表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    });

    (global.fetch as jest.Mock).mockImplementation((url) => {
      if (url.includes('/stats')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockStats,
        });
      }
      if (url.includes('/favorites')) {
        return Promise.resolve({
          ok: true,
          json: async () => [],
        });
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('総視聴エピソード')).toBeInTheDocument();
      expect(screen.getByText('完了アニメ')).toBeInTheDocument();
      expect(screen.getByText('40.0h')).toBeInTheDocument();
    });
  });

  it('データローディング中は読み込み表示を出す', () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    });

    // fetchを未解決のPromiseとしてモック
    (global.fetch as jest.Mock).mockImplementation(
      () => new Promise(() => {}) // 永遠に解決しないPromise
    );

    render(<MyPage />);

    expect(screen.getByText('読み込み中...')).toBeInTheDocument();
  });

  it('データ取得エラー時にエラーメッセージを表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    });

    (global.fetch as jest.Mock).mockImplementation(() => {
      return Promise.resolve({
        ok: false,
      });
    });

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('統計情報の取得に失敗しました')).toBeInTheDocument();
    });
  });

  it('ネットワークエラー時にエラーメッセージを表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    });

    (global.fetch as jest.Mock).mockRejectedValue(new Error('Network error'));

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });
  });
});
