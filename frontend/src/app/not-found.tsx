'use client'

import Link from 'next/link'
import { Button } from '../components/ui/button'
import { Card } from '../components/ui/card'
import { Home, ArrowLeft, Search } from 'lucide-react'

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4">
      <div className="text-center">
        <Card className="p-12 max-w-lg mx-auto shadow-xl border-0 bg-white/95 backdrop-blur">
          {/* 404 Animation */}
          <div className="mb-8">
            <div className="relative">
              <h1 className="text-9xl font-bold text-blue-100 select-none">404</h1>
              <div className="absolute inset-0 flex items-center justify-center">
                <Search className="h-16 w-16 text-blue-300 animate-pulse" />
              </div>
            </div>
          </div>

          {/* Content */}
          <div className="mb-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">
              Oops! Page not found
            </h2>
            <p className="text-gray-600 mb-6">
              The page you're looking for doesn't exist or has been moved. 
              Let's get you back on track.
            </p>
          </div>

          {/* Actions */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button asChild className="flex items-center">
              <Link href="/">
                <Home className="mr-2 h-4 w-4" />
                Go Home
              </Link>
            </Button>
            
            <Button variant="outline" onClick={() => window.history.back()} className="flex items-center">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Go Back
            </Button>
          </div>

          {/* Help Links */}
          <div className="mt-8 pt-6 border-t border-gray-200">
            <p className="text-sm text-gray-500 mb-4">Need help? Try these links:</p>
            <div className="flex flex-wrap justify-center gap-4 text-sm">
              <Link href="/auth/login" className="text-blue-600 hover:text-blue-500">
                Login
              </Link>
              <Link href="/user/dashboard" className="text-blue-600 hover:text-blue-500">
                Dashboard
              </Link>
              <Link href="/user/profile" className="text-blue-600 hover:text-blue-500">
                Profile
              </Link>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}