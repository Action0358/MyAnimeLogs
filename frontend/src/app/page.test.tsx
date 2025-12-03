import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import '@testing-library/jest-dom'
import { AuthProvider } from '@/contexts/AuthContext'
import HomePage from './page'

// Next.jsのuseRouterをモック
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
  }),
}))

// テスト用のQueryClientを作成
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false, // テスト中は再試行しない
      },
    },
  })

// テスト用のラッパーコンポーネント
const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = createTestQueryClient()
  return (
    <AuthProvider>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </AuthProvider>
  )
}

describe('HomePage', () => {
  it('renders the main heading', () => {
    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    const heading = screen.getByRole('heading', {
      name: /あなたのアニメ視聴体験を記録・管理しよう/i,
    })

    expect(heading).toBeInTheDocument()
  })
})
