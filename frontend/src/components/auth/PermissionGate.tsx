/**
 * Permission-based Component Wrappers
 * Components for conditional rendering based on user permissions
 */

import React, { ReactNode } from 'react'
import { usePermission, useAnyPermission, useAllPermissions, useIsSuperadmin } from '@/hooks/usePermission'

interface PermissionGateProps {
  children: ReactNode
  fallback?: ReactNode
}

interface RequirePermissionProps extends PermissionGateProps {
  permission: string
}

interface RequireAnyPermissionProps extends PermissionGateProps {
  permissions: string[]
}

interface RequireAllPermissionsProps extends PermissionGateProps {
  permissions: string[]
}

/**
 * Render children only if user has the specified permission
 * 
 * @example
 * <RequirePermission permission="organization.members.edit">
 *   <EditMemberButton />
 * </RequirePermission>
 */
export const RequirePermission: React.FC<RequirePermissionProps> = ({ 
  permission, 
  children, 
  fallback = null 
}) => {
  const hasPermission = usePermission(permission)
  return <>{hasPermission ? children : fallback}</>
}

/**
 * Render children only if user has ANY of the specified permissions
 * 
 * @example
 * <RequireAnyPermission permissions={['organization.members.edit', 'organization.roles.manage']}>
 *   <ManagementPanel />
 * </RequireAnyPermission>
 */
export const RequireAnyPermission: React.FC<RequireAnyPermissionProps> = ({ 
  permissions, 
  children, 
  fallback = null 
}) => {
  const hasAnyPermission = useAnyPermission(...permissions)
  return <>{hasAnyPermission ? children : fallback}</>
}

/**
 * Render children only if user has ALL of the specified permissions
 * 
 * @example
 * <RequireAllPermissions permissions={['organization.members.edit', 'organization.roles.manage']}>
 *   <FullAccessPanel />
 * </RequireAllPermissions>
 */
export const RequireAllPermissions: React.FC<RequireAllPermissionsProps> = ({ 
  permissions, 
  children, 
  fallback = null 
}) => {
  const hasAllPermissions = useAllPermissions(...permissions)
  return <>{hasAllPermissions ? children : fallback}</>
}

/**
 * Render children only if user is a superadmin
 * 
 * @example
 * <RequireSuperadmin>
 *   <GlobalAdminPanel />
 * </RequireSuperadmin>
 */
export const RequireSuperadmin: React.FC<PermissionGateProps> = ({ 
  children, 
  fallback = null 
}) => {
  const isSuperadmin = useIsSuperadmin()
  return <>{isSuperadmin ? children : fallback}</>
}

/**
 * Render different content based on permission check
 * 
 * @example
 * <PermissionSwitch
 *   permission="organization.members.edit"
 *   granted={<EditButton />}
 *   denied={<ViewOnlyMessage />}
 * />
 */
interface PermissionSwitchProps {
  permission: string
  granted: ReactNode
  denied?: ReactNode
}

export const PermissionSwitch: React.FC<PermissionSwitchProps> = ({
  permission,
  granted,
  denied = null
}) => {
  const hasPermission = usePermission(permission)
  return <>{hasPermission ? granted : denied}</>
}
