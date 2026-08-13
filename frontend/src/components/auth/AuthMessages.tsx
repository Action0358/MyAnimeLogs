import React from 'react'

interface AuthMessagesProps {
  error?: string
  info?: string
}

export const AuthMessages: React.FC<AuthMessagesProps> = ({ error, info }) => {
  return (
    <>
      {info && (
        <div className="rounded-md border border-primary/20 bg-primary/5 p-4">
          <p className="text-sm text-primary">{info}</p>
        </div>
      )}
      {error && (
        <div className="rounded-md border border-destructive/20 bg-destructive/5 p-4">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}
    </>
  )
}
