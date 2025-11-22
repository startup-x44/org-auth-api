/**
 * Route Guards - Authentication and authorization protection
 * Handles protected routes, admin routes, and public routes
 */

import React from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import LoadingSpinner from '../ui/loading-spinner'

interface RouteGuardProps {
  children: React.ReactNode
}

/**
 * Protected Route - Requires authentication
 */
export function ProtectedRoute({ children }: RouteGuardProps) {
  const { isAuthenticated, isLoading } = useAuthStore()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" state={{ from: location.pathname }} replace />
  }

  return <>{children}</>
}

/**
 * Admin Route - Requires admin permissions
 */
export function AdminRoute({ children }: RouteGuardProps) {
  const { hasPermission, isLoading } = useAuthStore()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (!hasPermission('users:read')) {
    return (
      <Navigate 
        to="/unauthorized" 
        state={{ 
          from: location.pathname,
          reason: 'insufficient_permissions',
          required: ['users:read']
        }} 
        replace 
      />
    )
  }

  return <>{children}</>
}

/**
 * Public Route - Redirects authenticated users to dashboard
 */
export function PublicRoute({ children }: RouteGuardProps) {
  const { isAuthenticated, isLoading } = useAuthStore()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }

  return <>{children}</>
}

/**
 * Role-based Route - Requires specific role
 */
interface RoleRouteProps extends RouteGuardProps {
  requiredRole: string
}

export function RoleRoute({ children, requiredRole }: RoleRouteProps) {
  const { user, isLoading } = useAuthStore()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (!user || user.role !== requiredRole) {
    return (
      <Navigate 
        to="/unauthorized" 
        state={{ 
          from: location.pathname,
          reason: 'insufficient_roles',
          required: [requiredRole]
        }} 
        replace 
      />
    )
  }

  return <>{children}</>
}

/**
 * Permission-based Route - Requires specific permissions
 */
interface PermissionRouteProps extends RouteGuardProps {
  requiredPermissions: string[]
  requireAll?: boolean // If true, requires ALL permissions. If false, requires ANY permission
}

export function PermissionRoute({ 
  children, 
  requiredPermissions, 
  requireAll = false 
}: PermissionRouteProps) {
  const { hasPermission, isLoading } = useAuthStore()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  const hasRequiredPermissions = requireAll
    ? requiredPermissions.every(permission => hasPermission(permission))
    : requiredPermissions.some(permission => hasPermission(permission))

  if (!hasRequiredPermissions) {
    return (
      <Navigate 
        to="/unauthorized" 
        state={{ 
          from: location.pathname,
          reason: 'insufficient_permissions',
          required: requiredPermissions
        }} 
        replace 
      />
    )
  }

  return <>{children}</>
}

/**
 * Conditional Route - Custom condition function
 */
interface ConditionalRouteProps extends RouteGuardProps {
  condition: () => boolean
  fallbackPath?: string
  reason?: string
}

export function ConditionalRoute({ 
  children, 
  condition, 
  fallbackPath = '/unauthorized',
  reason = 'custom_condition_failed'
}: ConditionalRouteProps) {
  const location = useLocation()

  if (!condition()) {
    return (
      <Navigate 
        to={fallbackPath} 
        state={{ 
          from: location.pathname,
          reason
        }} 
        replace 
      />
    )
  }

  return <>{children}</>
}