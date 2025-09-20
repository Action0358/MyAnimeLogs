import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import '@testing-library/jest-dom'
import HomePage from './page'

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
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
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
      name: /MyAnimeLogs へようこそ/i,
    })

    expect(heading).toBeInTheDocument()
  })
})
