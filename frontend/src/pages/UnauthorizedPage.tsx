import React from 'react'
import { useLocation, Link, useNavigate } from 'react-router-dom'
import { AlertTriangle, ArrowLeft, Home, RefreshCw, Shield } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { useAuthStore } from '@/stores/authStore'

interface UnauthorizedPageProps {}

interface UnauthorizedState {
  from?: string
  reason?: string
  required?: string | string[]
  actual?: string[]
  message?: string
}

export const UnauthorizedPage: React.FC<UnauthorizedPageProps> = () => {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, hasRole } = useAuthStore()
  
  const state = location.state as UnauthorizedState | null

  const getCorrectDashboard = () => {
    if (!user) return '/auth/login'
    
    if (user.is_superadmin || hasRole('superadmin')) {
      return '/admin/dashboard'
    }
    if (hasRole('admin')) {
      return '/admin/dashboard'
    }
    if (hasRole('user') || hasRole('member')) {
      return '/user/dashboard'  
    }
    return '/dashboard'
  }

  const handleGoBack = () => {
    if (window.history.length > 1) {
      navigate(-1)
    } else {
      navigate(getCorrectDashboard())
    }
  }

  const getErrorMessage = () => {
    if (state?.message) {
      return state.message
    }
    
    switch (state?.reason) {
      case 'role_mismatch_strict':
        return `This area is restricted to ${Array.isArray(state.required) ? state.required.join(' or ') : state.required} users only.`
      case 'insufficient_permissions':
        return 'You don\'t have the required permissions to access this page.'  
      case 'insufficient_roles':
      case 'insufficient_role':
        return 'Your current role doesn\'t allow access to this page.'
      case 'custom_condition_failed':
        return 'You don\'t meet the requirements to access this page.'
      default:
        return 'You are not authorized to access this page.'
    }
  }

  const getRoleStatusColor = (role: string) => {
    if (user?.roles?.some(r => r.name === role) || user?.global_role === role || (role === 'superadmin' && user?.is_superadmin)) {
      return 'bg-green-100 text-green-800 border-green-300'
    }
    return 'bg-gray-100 text-gray-600 border-gray-300'
  }

  const getRequiredText = () => {
    if (!state?.required) return null
    
    const required = Array.isArray(state.required) ? state.required : [state.required]
    const type = state.reason === 'insufficient_roles' ? 'roles' : 'permissions'
    
    return (
      <p className="text-sm text-gray-600 mt-2">
        Required {type}: <code className="bg-gray-100 px-2 py-1 rounded text-xs">{required.join(', ')}</code>
      </p>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          <div className="text-center">
            {/* Icon */}
            <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-red-100 mb-6">
              <AlertTriangle className="h-8 w-8 text-red-600" aria-hidden="true" />
            </div>

            {/* Heading */}
            <h1 className="text-2xl font-bold text-gray-900 mb-2">
              Access Denied
            </h1>

            {/* Error message */}
            <p className="text-gray-600 mb-4">
              {getErrorMessage()}
            </p>

            {/* Required permissions/roles */}
            {getRequiredText()}

            {/* Role Information */}
            {state?.required && state?.actual && (
              <div className="bg-gray-50 rounded-lg p-4 mt-4">
                <h3 className="font-semibold text-gray-900 mb-3 flex items-center text-sm">
                  <Shield className="w-4 h-4 mr-2" />
                  Role Information
                </h3>
                
                <div className="grid grid-cols-1 gap-3">
                  {/* Required Roles */}
                  <div>
                    <p className="text-xs font-medium text-gray-700 mb-2">Required Roles:</p>
                    <div className="flex flex-wrap gap-1">
                      {(Array.isArray(state.required) ? state.required : [state.required]).map((role) => (
                        <Badge 
                          key={role} 
                          variant="outline"
                          className="bg-red-50 text-red-700 border-red-300 text-xs"
                        >
                          {role}
                        </Badge>
                      ))}
                    </div>
                  </div>

                  {/* Your Roles */}
                  <div>
                    <p className="text-xs font-medium text-gray-700 mb-2">Your Roles:</p>
                    <div className="flex flex-wrap gap-1">
                      {state.actual.length > 0 ? (
                        state.actual.map((role) => (
                          <Badge 
                            key={role} 
                            variant="outline"
                            className={`${getRoleStatusColor(role)} text-xs`}
                          >
                            {role}
                          </Badge>
                        ))
                      ) : (
                        <Badge variant="outline" className="bg-gray-100 text-gray-500 text-xs">
                          No roles assigned
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Attempted path */}
            {state?.from && (
              <p className="text-xs text-gray-500 mt-4">
                Attempted to access: <code className="bg-gray-100 px-1 py-0.5 rounded">{state.from}</code>
              </p>
            )}
          </div>

          {/* Actions */}
          <div className="mt-8 space-y-3">
            <button
              onClick={handleGoBack}
              className="w-full flex justify-center items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Go Back
            </button>

            <Link
              to={getCorrectDashboard()}
              className="w-full flex justify-center items-center px-4 py-2 border border-transparent rounded-md shadow-sm bg-blue-600 text-sm font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <Home className="h-4 w-4 mr-2" />
              Go to Dashboard
            </Link>

            <button
              onClick={() => window.location.reload()}
              className="w-full flex justify-center items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Refresh Page
            </button>
          </div>

          {/* Contact support */}
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-xs text-gray-500 text-center">
              If you believe this is an error, please contact your administrator or{' '}
              <a 
                href="/support" 
                className="text-blue-600 hover:text-blue-500 underline"
              >
                support
              </a>
              .
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default UnauthorizedPage