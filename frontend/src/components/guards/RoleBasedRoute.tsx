import React from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { logger } from '../../utils/logger'

export interface RoleBasedRouteProps {
  children: React.ReactNode
  allowedRoles: string[]
  strictMode?: boolean // If true, blocks role inheritance
  fallback?: React.ReactNode
  redirectPath?: string
}

/**
 * RoleBasedRoute component for strict role-based access control
 * 
 * This component implements the policy-based RBAC system:
 * - strictMode=true: Only exact role matches are allowed (no inheritance)
 * - strictMode=false: Role hierarchy applies (superadmin > admin > member > user)
 */
export const RoleBasedRoute: React.FC<RoleBasedRouteProps> = ({
  children,
  allowedRoles,
  strictMode = false,
  fallback,
  redirectPath = '/unauthorized'
}) => {
  const { user, isAuthenticated, isLoading, hasRole, canAccess } = useAuthStore()
  const location = useLocation()

  // Show loading state
  if (isLoading) {
    return fallback || (
      <div className="flex justify-center items-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  // Check authentication
  if (!isAuthenticated || !user) {
    logger.warn('RoleBasedRoute: User not authenticated', {
      path: location.pathname,
      allowedRoles,
      strictMode
    })
    
    return <Navigate to="/auth/login" state={{ from: location.pathname }} replace />
  }

  // Get user's roles for logging
  const userRoles = [
    ...(user.roles?.map(r => r.name) || []),
    ...(user.global_role ? [user.global_role] : []),
    ...(user.is_superadmin ? ['superadmin'] : [])
  ]

  // Strict mode: No inheritance, exact role match required
  if (strictMode) {
    const hasExactRole = allowedRoles.some(role => hasRole(role))
    
    if (!hasExactRole) {
      logger.warn('RoleBasedRoute: Access denied in strict mode', {
        path: location.pathname,
        userRoles,
        allowedRoles,
        userId: user.id,
        reason: 'exact_role_match_required'
      })

      return (
        <Navigate 
          to={redirectPath}
          state={{ 
            from: location.pathname,
            reason: 'role_mismatch_strict',
            required: allowedRoles,
            actual: userRoles,
            message: `This area is restricted to ${allowedRoles.join(', ')} only. Role inheritance is disabled.`
          }} 
          replace 
        />
      )
    }
  } else {
    // Non-strict mode: Role hierarchy applies
    let hasAccess = false

    // Check each allowed role
    for (const role of allowedRoles) {
      if (hasRole(role)) {
        hasAccess = true
        break
      }
      
      // Check role hierarchy for user-facing routes
      if (role === 'user' || role === 'member') {
        // For user/member routes, check if user has higher privileges
        if (hasRole('admin') || hasRole('superadmin') || user.is_superadmin) {
          // Allow higher roles to access lower-level routes
          // But only for user-facing resources, not role-specific dashboards
          if (!location.pathname.includes('/user/dashboard') && !location.pathname.includes('/member/dashboard')) {
            hasAccess = true
            break
          }
        }
      }
    }

    if (!hasAccess) {
      logger.warn('RoleBasedRoute: Access denied with role hierarchy', {
        path: location.pathname,
        userRoles,
        allowedRoles,
        userId: user.id,
        reason: 'insufficient_role_hierarchy'
      })

      return (
        <Navigate 
          to={redirectPath}
          state={{ 
            from: location.pathname,
            reason: 'insufficient_role',
            required: allowedRoles,
            actual: userRoles,
            message: `Access denied. Required roles: ${allowedRoles.join(' or ')}`
          }} 
          replace 
        />
      )
    }
  }

  // Access granted
  logger.debug('RoleBasedRoute: Access granted', {
    path: location.pathname,
    userRoles,
    allowedRoles,
    strictMode,
    userId: user.id
  })

  return <>{children}</>
}

/**
 * UserOnlyRoute - Strict route for user/member roles only
 */
export const UserOnlyRoute: React.FC<{
  children: React.ReactNode
  fallback?: React.ReactNode
}> = ({ children, fallback }) => {
  return (
    <RoleBasedRoute
      allowedRoles={['user', 'member']}
      strictMode={true}
      fallback={fallback}
      redirectPath="/unauthorized"
    >
      {children}
    </RoleBasedRoute>
  )
}

/**
 * AdminOnlyRoute - Route for admin/superadmin roles
 */
export const AdminOnlyRoute: React.FC<{
  children: React.ReactNode
  fallback?: React.ReactNode
}> = ({ children, fallback }) => {
  return (
    <RoleBasedRoute
      allowedRoles={['admin', 'superadmin']}
      strictMode={false} // Allow hierarchy for admin routes
      fallback={fallback}
      redirectPath="/unauthorized"
    >
      {children}
    </RoleBasedRoute>
  )
}

/**
 * SuperAdminOnlyRoute - Strict route for superadmin only
 */
export const SuperAdminOnlyRoute: React.FC<{
  children: React.ReactNode
  fallback?: React.ReactNode
}> = ({ children, fallback }) => {
  return (
    <RoleBasedRoute
      allowedRoles={['superadmin']}
      strictMode={true}
      fallback={fallback}
      redirectPath="/unauthorized"
    >
      {children}
    </RoleBasedRoute>
  )
}

export default RoleBasedRoute