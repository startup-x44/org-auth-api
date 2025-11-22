/**
 * Authentication Layout - For login, MFA, and other auth pages
 * Clean, minimal design with branding and security messaging
 */

import React from 'react'
import { Link } from 'react-router-dom'
import { Shield, Lock, CheckCircle } from 'lucide-react'

interface AuthLayoutProps {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="flex">
        {/* Left side - Branding and features */}
        <div className="hidden lg:flex lg:flex-1 lg:flex-col lg:justify-center lg:py-12 lg:px-8 xl:px-12">
          <div className="max-w-md mx-auto">
            {/* Logo/Branding */}
            <div className="mb-8">
              <Link to="/" className="flex items-center">
                <div className="flex items-center justify-center w-12 h-12 bg-blue-600 rounded-xl">
                  <Shield className="h-7 w-7 text-white" />
                </div>
                <div className="ml-3">
                  <h1 className="text-2xl font-bold text-gray-900">NILOAUTH</h1>
                  <p className="text-sm text-gray-600">Authentication Service</p>
                </div>
              </Link>
            </div>

            {/* Features */}
            <div className="space-y-8">
              <div>
                <h2 className="text-2xl font-bold text-gray-900 mb-4">
                  Enterprise-Grade Security
                </h2>
                <p className="text-gray-600">
                  Secure authentication and authorization for modern applications with 
                  enterprise-level features and compliance.
                </p>
              </div>

              <div className="space-y-4">
                <div className="flex items-start">
                  <CheckCircle className="h-5 w-5 text-green-500 mt-1 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-gray-900">Multi-Factor Authentication</h3>
                    <p className="text-sm text-gray-600">Additional security layer with TOTP support</p>
                  </div>
                </div>

                <div className="flex items-start">
                  <CheckCircle className="h-5 w-5 text-green-500 mt-1 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-gray-900">Role-Based Access Control</h3>
                    <p className="text-sm text-gray-600">Granular permissions and organization management</p>
                  </div>
                </div>

                <div className="flex items-start">
                  <CheckCircle className="h-5 w-5 text-green-500 mt-1 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-gray-900">Enterprise SSO</h3>
                    <p className="text-sm text-gray-600">OAuth 2.0 and SAML integration support</p>
                  </div>
                </div>

                <div className="flex items-start">
                  <CheckCircle className="h-5 w-5 text-green-500 mt-1 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-gray-900">Audit & Compliance</h3>
                    <p className="text-sm text-gray-600">Comprehensive logging and compliance reporting</p>
                  </div>
                </div>
              </div>

              {/* Security badges */}
              <div className="border-t border-gray-200 pt-8">
                <p className="text-xs text-gray-500 mb-3">Security & Compliance</p>
                <div className="flex space-x-4">
                  <div className="flex items-center space-x-1">
                    <Lock className="h-3 w-3 text-gray-400" />
                    <span className="text-xs text-gray-600">SOC 2 Compliant</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Lock className="h-3 w-3 text-gray-400" />
                    <span className="text-xs text-gray-600">GDPR Ready</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Lock className="h-3 w-3 text-gray-400" />
                    <span className="text-xs text-gray-600">ISO 27001</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Right side - Auth form */}
        <div className="flex flex-1 flex-col justify-center py-12 px-4 sm:px-6 lg:flex-none lg:px-20 xl:px-24">
          <div className="mx-auto w-full max-w-sm lg:w-96">
            {/* Mobile logo */}
            <div className="lg:hidden mb-8">
              <Link to="/" className="flex items-center justify-center">
                <div className="flex items-center justify-center w-12 h-12 bg-blue-600 rounded-xl">
                  <Shield className="h-7 w-7 text-white" />
                </div>
                <div className="ml-3">
                  <h1 className="text-xl font-bold text-gray-900">NILOAUTH</h1>
                  <p className="text-sm text-gray-600">Authentication</p>
                </div>
              </Link>
            </div>

            {/* Auth form content */}
            <div className="bg-white py-8 px-6 shadow-xl rounded-xl border border-gray-100">
              {children}
            </div>

            {/* Footer links */}
            <div className="mt-8">
              <div className="flex justify-center space-x-6 text-sm text-gray-600">
                <Link to="/privacy" className="hover:text-gray-900 transition-colors">
                  Privacy Policy
                </Link>
                <Link to="/terms" className="hover:text-gray-900 transition-colors">
                  Terms of Service
                </Link>
                <Link to="/support" className="hover:text-gray-900 transition-colors">
                  Support
                </Link>
              </div>
              
              <div className="mt-4 text-center">
                <p className="text-xs text-gray-500">
                  © {new Date().getFullYear()} NILOAUTH. All rights reserved.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Background decoration */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden">
        <div className="absolute -top-40 -right-32 w-80 h-80 bg-gradient-to-br from-blue-400 to-purple-400 rounded-full opacity-10 blur-3xl" />
        <div className="absolute -bottom-40 -left-32 w-80 h-80 bg-gradient-to-tr from-green-400 to-blue-400 rounded-full opacity-10 blur-3xl" />
      </div>
    </div>
  )
}