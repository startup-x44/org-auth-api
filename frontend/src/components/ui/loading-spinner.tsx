import React from 'react'
import { cn } from '@/lib/utils'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  className?: string
  fullScreen?: boolean
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ size = 'md', className, fullScreen = false }) => {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-8 w-8',
    lg: 'h-12 w-12'
  }

  const spinner = (
    <div
      className={cn(
        'animate-spin rounded-full border-2 border-gray-300 border-t-blue-600',
        sizeClasses[size],
        className
      )}
    />
  )

  if (fullScreen) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-indigo-100 flex items-center justify-center relative overflow-hidden">
        {/* Optimized background decoration */}
        <div className="absolute inset-0 opacity-10">
          <div className="absolute top-20 left-20 w-24 h-24 bg-blue-400 rounded-full animate-pulse"></div>
          <div className="absolute top-40 right-32 w-16 h-16 bg-indigo-400 rounded-full animate-pulse delay-75"></div>
          <div className="absolute bottom-32 left-1/3 w-32 h-32 bg-purple-400 rounded-full animate-pulse delay-150"></div>
        </div>
        
        <div className="text-center relative z-10 px-4">
          <div className="inline-flex items-center justify-center w-14 h-14 mb-4 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-xl shadow-lg">
            <div className="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
          </div>
          <div className="space-y-2">
            <h2 className="text-lg font-semibold text-gray-800">NILOAUTH</h2>
            <div className="flex items-center justify-center space-x-1">
              <div className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce"></div>
              <div className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce delay-75"></div>
              <div className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce delay-150"></div>
            </div>
            <p className="text-xs text-gray-600 mt-2">Initializing...</p>
          </div>
        </div>
      </div>
    )
  }

  return spinner
}

export default LoadingSpinner