/**
 * Permission Hooks
 * React hooks for permission-based UI rendering
 */

import { useMemo } from 'react'
import useAuthStore from '../store/auth'

/**
 * Hook to check if user has a specific permission
 * 
 * @example
 * const canEditMembers = usePermission('organization.members.edit')
 * if (canEditMembers) {
 *   return <EditButton />
 * }
 */
export function usePermission(permission: string): boolean {
  const hasPermission = useAuthStore(state => state.hasPermission)
  return useMemo(() => hasPermission(permission), [hasPermission, permission])
}

/**
 * Hook to check if user has ANY of the specified permissions (OR logic)
 * 
 * @example
 * const canManage = useAnyPermission('organization.members.edit', 'organization.roles.manage')
 */
export function useAnyPermission(...permissions: string[]): boolean {
  const hasAnyPermission = useAuthStore(state => state.hasAnyPermission)
  const userPermissions = useAuthStore(state => state.permissions)
  return useMemo(
    () => hasAnyPermission(...permissions),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [hasAnyPermission, userPermissions, ...permissions]
  )
}

/**
 * Hook to check if user has ALL of the specified permissions (AND logic)
 * 
 * @example
 * const canFullyManage = useAllPermissions('organization.members.edit', 'organization.roles.manage')
 */
export function useAllPermissions(...permissions: string[]): boolean {
  const hasAllPermissions = useAuthStore(state => state.hasAllPermissions)
  const userPermissions = useAuthStore(state => state.permissions)
  return useMemo(
    () => hasAllPermissions(...permissions),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [hasAllPermissions, userPermissions, ...permissions]
  )
}

/**
 * Hook to get all user permissions
 * 
 * @example
 * const permissions = usePermissions()
 * console.log('User has permissions:', permissions)
 */
export function usePermissions(): string[] {
  return useAuthStore(state => state.permissions)
}

/**
 * Hook to get user's role information
 */
export function useRole(): { roleId: string | null; roleName: string | null } {
  const roleId = useAuthStore(state => state.roleId)
  const roleName = useAuthStore(state => state.roleName)
  return { roleId, roleName }
}

/**
 * Hook to check if user is superadmin
 * Superadmins bypass all permission checks
 */
export function useIsSuperadmin(): boolean {
  return useAuthStore(state => state.isSuperadmin)
}

/**
 * Hook to check if user is organization admin
 * Checks if user has admin-level permissions
 */
export function useIsOrgAdmin(): boolean {
  const hasPermission = useAuthStore(state => state.hasPermission)
  const isSuperadmin = useAuthStore(state => state.isSuperadmin)
  
  return useMemo(
    () => isSuperadmin || hasPermission('organization.settings.manage'),
    [hasPermission, isSuperadmin]
  )
}
