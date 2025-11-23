'use client'

import { LoadingSpinner } from '../components/ui/loading-spinner'

export default function Loading() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="text-center">
        <LoadingSpinner size="xl" variant="primary" />
        <p className="mt-6 text-lg text-gray-600">Loading NILOAUTH...</p>
        <p className="mt-2 text-sm text-gray-500">Please wait while we prepare your experience</p>
      </div>
    </div>
  )
}