import { useState } from 'react'

interface UseAuthFormState {
  error: string
  isLoading: boolean
}

interface UseAuthFormResult {
  error: string
  isLoading: boolean
  setError: (error: string) => void
  handleSubmit: (fn: () => Promise<void>) => (e: React.FormEvent) => Promise<void>
}

export const useAuthForm = (): UseAuthFormResult => {
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit =
    (fn: () => Promise<void>) =>
    async (e: React.FormEvent) => {
      e.preventDefault()
      setError('')
      setIsLoading(true)

      try {
        await fn()
      } catch (err) {
        setError(err instanceof Error ? err.message : 'エラーが発生しました')
      } finally {
        setIsLoading(false)
      }
    }

  return {
    error,
    isLoading,
    setError,
    handleSubmit,
  }
}
