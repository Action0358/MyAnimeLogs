import React from 'react'

interface AuthMessagesProps {
  error?: string
  info?: string
}

export const AuthMessages: React.FC<AuthMessagesProps> = ({ error, info }) => {
  return (
    <>
      {info && (
        <div className="p-4 bg-blue-50 border border-blue-200 rounded-md">
          <p className="text-sm text-blue-700">{info}</p>
        </div>
      )}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}
    </>
  )
}
