import React from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { useTenantStore } from '../../stores/tenantStore'
import { logger } from '../../utils/logger'

// Types
export interface ProtectedRouteProps {
  children: React.ReactNode
  requiredPermissions?: string | string[]
  requiredRoles?: string | string[]
  requireAuth?: boolean
  fallback?: React.ReactNode
  redirectTo?: string
}

export interface AdminRouteProps {
  children: React.ReactNode
  fallback?: React.ReactNode
  redirectTo?: string
}

// Helper function to check if user has required permissions
const hasRequiredPermissions = (
  userPermissions: string[],
  requiredPermissions: string | string[] | undefined
): boolean => {
  if (!requiredPermissions) return true
  
  const permissions = Array.isArray(requiredPermissions) ? requiredPermissions : [requiredPermissions]
  return permissions.some(permission => userPermissions.includes(permission))
}

// Helper function to check if user has required roles
const hasRequiredRoles = (
  userRoles: string[],
  requiredRoles: string | string[] | undefined
): boolean => {
  if (!requiredRoles) return true
  
  const roles = Array.isArray(requiredRoles) ? requiredRoles : [requiredRoles]
  return roles.some(role => userRoles.includes(role))
}

/**
 * ProtectedRoute component
 * Protects routes based on authentication status, permissions, and roles
 */
export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requiredPermissions,
  requiredRoles,
  requireAuth = true,
  fallback,
  redirectTo = '/auth/login',
}) => {
  const { isAuthenticated, isLoading, user } = useAuth()
  const permissions = useTenantStore((state) => state.currentUserPermissions) || []
  const roles = useTenantStore((state) => state.currentUserRoles) || []
  const location = useLocation()

  // Show loading state while checking auth
  if (isLoading) {
    return fallback || <div className="flex justify-center items-center min-h-screen">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
  }

  // Check authentication requirement
  if (requireAuth && !isAuthenticated) {
    logger.warn('Access denied: User not authenticated', {
      requestedPath: location.pathname,
      requiredAuth: requireAuth,
    })

    return <Navigate 
      to={redirectTo} 
      state={{ from: location.pathname }} 
      replace 
    />
  }

  // Check permissions requirement
  if (isAuthenticated && !hasRequiredPermissions(permissions, requiredPermissions)) {
    logger.warn('Access denied: Insufficient permissions', {
      requestedPath: location.pathname,
      userPermissions: permissions,
      requiredPermissions,
      userId: user?.id,
    })

    return <Navigate 
      to="/unauthorized" 
      state={{ 
        from: location.pathname,
        reason: 'insufficient_permissions',
        required: requiredPermissions 
      }} 
      replace 
    />
  }

  // Check roles requirement
  if (isAuthenticated && !hasRequiredRoles(roles, requiredRoles)) {
    logger.warn('Access denied: Insufficient roles', {
      requestedPath: location.pathname,
      userRoles: roles,
      requiredRoles,
      userId: user?.id,
    })

    return <Navigate 
      to="/unauthorized" 
      state={{ 
        from: location.pathname,
        reason: 'insufficient_roles',
        required: requiredRoles 
      }} 
      replace 
    />
  }

  // All checks passed, render children
  return <>{children}</>
}

/**
 * AdminRoute component
 * Shorthand for routes that require admin role
 */
export const AdminRoute: React.FC<AdminRouteProps> = ({
  children,
  fallback,
  redirectTo = '/unauthorized',
}) => {
  return (
    <ProtectedRoute
      requiredRoles={['admin', 'super_admin']}
      fallback={fallback}
      redirectTo={redirectTo}
    >
      {children}
    </ProtectedRoute>
  )
}

/**
 * SuperAdminRoute component
 * For routes that require super admin access
 */
export const SuperAdminRoute: React.FC<AdminRouteProps> = ({
  children,
  fallback,
  redirectTo = '/unauthorized',
}) => {
  return (
    <ProtectedRoute
      requiredRoles="super_admin"
      fallback={fallback}
      redirectTo={redirectTo}
    >
      {children}
    </ProtectedRoute>
  )
}

/**
 * PublicRoute component
 * For routes that should only be accessible when NOT authenticated (e.g., login page)
 */
export const PublicRoute: React.FC<{
  children: React.ReactNode
  redirectTo?: string
}> = ({
  children,
  redirectTo = '/dashboard',
}) => {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation()

  // Show loading state
  if (isLoading) {
    return <div className="flex justify-center items-center min-h-screen">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
  }

  // If authenticated, redirect to dashboard
  if (isAuthenticated) {
    const from = (location.state as any)?.from || redirectTo
    return <Navigate to={from} replace />
  }

  return <>{children}</>
}

/**
 * ConditionalRoute component
 * For advanced permission checking with custom logic
 */
export const ConditionalRoute: React.FC<{
  children: React.ReactNode
  condition: (user: any, permissions: string[], roles: string[]) => boolean
  fallback?: React.ReactNode
  redirectTo?: string
}> = ({
  children,
  condition,
  fallback,
  redirectTo = '/unauthorized',
}) => {
  const { isAuthenticated, isLoading, user } = useAuth()
  const permissions = useTenantStore((state) => state.currentUserPermissions) || []
  const roles = useTenantStore((state) => state.currentUserRoles) || []
  const location = useLocation()

  // Show loading state
  if (isLoading) {
    return fallback || <div className="flex justify-center items-center min-h-screen">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
  }

  // Check authentication
  if (!isAuthenticated) {
    return <Navigate to="/auth/login" state={{ from: location.pathname }} replace />
  }

  // Check custom condition
  if (!condition(user, permissions, roles)) {
    logger.warn('Access denied: Custom condition failed', {
      requestedPath: location.pathname,
      userId: user?.id,
    })

    return <Navigate 
      to={redirectTo} 
      state={{ 
        from: location.pathname,
        reason: 'custom_condition_failed' 
      }} 
      replace 
    />
  }

  return <>{children}</>
}

/**
 * PermissionGate component
 * For conditional rendering within components based on permissions
 */
export const PermissionGate: React.FC<{
  children: React.ReactNode
  requiredPermissions?: string | string[]
  requiredRoles?: string | string[]
  fallback?: React.ReactNode
  requireAll?: boolean // Whether all permissions/roles are required (AND) or just one (OR)
}> = ({
  children,
  requiredPermissions,
  requiredRoles,
  fallback = null,
  requireAll = false,
}) => {
  const { isAuthenticated } = useAuth()
  const permissions = useTenantStore((state) => state.currentUserPermissions) || []
  const roles = useTenantStore((state) => state.currentUserRoles) || []

  if (!isAuthenticated) {
    return <>{fallback}</>
  }

  // Check permissions
  if (requiredPermissions) {
    const permissionsArray = Array.isArray(requiredPermissions) ? requiredPermissions : [requiredPermissions]
    const hasPermissions = requireAll 
      ? permissionsArray.every(permission => permissions.includes(permission))
      : permissionsArray.some(permission => permissions.includes(permission))
    
    if (!hasPermissions) {
      return <>{fallback}</>
    }
  }

  // Check roles
  if (requiredRoles) {
    const rolesArray = Array.isArray(requiredRoles) ? requiredRoles : [requiredRoles]
    const hasRoles = requireAll
      ? rolesArray.every(role => roles.includes(role))
      : rolesArray.some(role => roles.includes(role))
    
    if (!hasRoles) {
      return <>{fallback}</>
    }
  }

  return <>{children}</>
}

/**
 * RoleGate component
 * Shorthand for role-based conditional rendering
 */
export const RoleGate: React.FC<{
  children: React.ReactNode
  roles: string | string[]
  fallback?: React.ReactNode
  requireAll?: boolean
}> = ({
  children,
  roles,
  fallback = null,
  requireAll = false,
}) => {
  return (
    <PermissionGate
      requiredRoles={roles}
      fallback={fallback}
      requireAll={requireAll}
    >
      {children}
    </PermissionGate>
  )
}



export default ProtectedRoute