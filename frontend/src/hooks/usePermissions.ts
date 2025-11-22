import { useAuth } from '@/hooks/useAuth'
import { useTenantPermissions } from '@/stores/tenantStore'

/**
 * Hook for permission checking in components
 * Separated from ProtectedRoute to avoid HMR issues
 */
export const usePermissions = () => {
  const { isAuthenticated } = useAuth()
  const { permissions, roles, hasPermission, hasRole } = useTenantPermissions()

  const can = (permission: string): boolean => {
    return isAuthenticated && hasPermission(permission)
  }

  const cannot = (permission: string): boolean => {
    return !can(permission)
  }

  const is = (role: string): boolean => {
    return isAuthenticated && hasRole(role)
  }

  const isNot = (role: string): boolean => {
    return !is(role)
  }

  const canAny = (permissionsList: string[]): boolean => {
    return isAuthenticated && permissionsList.some(permission => hasPermission(permission))
  }

  const canAll = (permissionsList: string[]): boolean => {
    return isAuthenticated && permissionsList.every(permission => hasPermission(permission))
  }

  const hasAnyRole = (rolesList: string[]): boolean => {
    return isAuthenticated && rolesList.some(role => hasRole(role))
  }

  const hasAllRoles = (rolesList: string[]): boolean => {
    return isAuthenticated && rolesList.every(role => hasRole(role))
  }

  return {
    permissions,
    roles,
    can,
    cannot,
    is,
    isNot,
    canAny,
    canAll,
    hasAnyRole,
    hasAllRoles,
    isAuthenticated,
  }
}