import React from 'react'
import { useAuth } from '@/hooks/useAuth'
import { useAuthStore } from '@/stores/authStore'

interface AuthDebugInfoProps {
  className?: string
}

export const AuthDebugInfo: React.FC<AuthDebugInfoProps> = React.memo(({ className = '' }) => {
  const { isLoading, isAuthenticated, error, user } = useAuth()
  const { forceResetLoading } = useAuthStore()

  // Only show in development
  if (import.meta.env.PROD) {
    return null
  }

  return (
    <div className={`fixed bottom-4 left-4 bg-black bg-opacity-80 text-white p-3 rounded text-xs font-mono z-50 ${className}`}>
      <div className="text-yellow-300 font-bold mb-2">AUTH DEBUG</div>
      <div>isLoading: <span className={isLoading ? 'text-red-400' : 'text-green-400'}>{String(isLoading)}</span></div>
      <div>isAuthenticated: <span className={isAuthenticated ? 'text-green-400' : 'text-red-400'}>{String(isAuthenticated)}</span></div>
      <div>hasUser: <span className={user ? 'text-green-400' : 'text-red-400'}>{String(!!user)}</span></div>
      <div>error: <span className={error ? 'text-red-400' : 'text-green-400'}>{error || 'none'}</span></div>
      {user && (
        <div className="mt-1 text-blue-300">
          user: {user.email}
        </div>
      )}
      {isLoading && (
        <button 
          onClick={forceResetLoading}
          className="mt-2 px-2 py-1 bg-red-600 text-white text-xs rounded hover:bg-red-700"
        >
          🚨 Force Reset Loading
        </button>
      )}
    </div>
  )
})

AuthDebugInfo.displayName = 'AuthDebugInfo'

export default AuthDebugInfo