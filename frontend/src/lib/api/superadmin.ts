/**
 * Superadmin API Client
 * All endpoints require superadmin privileges
 */

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

interface ApiResponse<T> {
  success: boolean
  data?: T
  message?: string
}

// ============================================================================
// TYPES
// ============================================================================

export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  status: string
  is_superadmin: boolean
  created_at: string
  updated_at: string
  last_login?: string
}

export interface UsersListResponse {
  users: User[]
  next_cursor?: string
  has_more: boolean
}

export interface Organization {
  id: string
  name: string
  slug: string
  description?: string
  created_at: string
  updated_at: string
  member_count?: number
}

export interface Role {
  id: string
  name: string
  description?: string
  is_system: boolean
  permissions: string[]
  created_at: string
  updated_at: string
}

export interface Permission {
  id: string
  name: string
  resource: string
  action: string
  description?: string
  is_system: boolean
  created_at: string
}

export interface RBACStats {
  total_roles: number
  total_permissions: number
  system_roles: number
  custom_roles: number
  system_permissions: number
  custom_permissions: number
}

export interface ClientApp {
  id: string
  name: string
  client_id: string
  is_confidential: boolean
  redirect_uris: string[]
  allowed_origins: string[]
  allowed_scopes: string[]
  organization_id: string
  created_at: string
  updated_at: string
}

// ============================================================================
// API FUNCTIONS
// ============================================================================

async function fetchWithAuth<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('access_token')
  
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': token ? `Bearer ${token}` : '',
      ...options.headers,
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(error.message || `HTTP ${response.status}`)
  }

  return response.json()
}

// ============================================================================
// USERS
// ============================================================================

export interface GetUsersParams {
  search?: string
  limit?: number
  cursor?: string
}

export interface UserListResponse {
  users: User[]
  total: number
  next_cursor?: string
  has_more: boolean
}

export async function listUsers(params: GetUsersParams = {}): Promise<ApiResponse<UsersListResponse>> {
  const searchParams = new URLSearchParams()
  if (params.limit) searchParams.append('limit', params.limit.toString())
  if (params.cursor) searchParams.append('cursor', params.cursor)
  if (params.search) searchParams.append('search', params.search)
  
  return fetchWithAuth(`/api/v1/admin/users?${searchParams}`)
}

export async function activateUser(userId: string): Promise<ApiResponse<void>> {
  return fetchWithAuth(`/api/v1/admin/users/${userId}/activate`, { method: 'PUT' })
}

export async function deactivateUser(userId: string): Promise<ApiResponse<void>> {
  return fetchWithAuth(`/api/v1/admin/users/${userId}/deactivate`, { method: 'PUT' })
}

export async function deleteUser(userId: string): Promise<ApiResponse<void>> {
  return fetchWithAuth(`/api/v1/admin/users/${userId}`, { method: 'DELETE' })
}

// Convenience object for easier imports
export const adminApi = {
  getUsers: async (params?: GetUsersParams) => {
    const response = await listUsers(params)
    return { users: response.data?.users || [], total: response.data?.users.length || 0 }
  },
  deleteUser,
  activateUser,
  deactivateUser,
}

// ============================================================================
// ORGANIZATIONS
// ============================================================================

export async function listOrganizations(): Promise<ApiResponse<Organization[]>> {
  return fetchWithAuth('/api/v1/admin/organizations')
}

// ============================================================================
// RBAC
// ============================================================================

export async function listAllPermissions(): Promise<ApiResponse<Permission[]>> {
  return fetchWithAuth('/api/v1/admin/rbac/permissions')
}

export async function listSystemRoles(): Promise<ApiResponse<Role[]>> {
  return fetchWithAuth('/api/v1/admin/rbac/roles')
}

export async function getSystemRole(roleId: string): Promise<ApiResponse<Role>> {
  return fetchWithAuth(`/api/v1/admin/rbac/roles/${roleId}`)
}

export async function createSystemRole(data: {
  name: string
  description?: string
}): Promise<ApiResponse<Role>> {
  return fetchWithAuth('/api/v1/admin/rbac/roles', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateSystemRole(roleId: string, data: {
  name?: string
  description?: string
}): Promise<ApiResponse<Role>> {
  return fetchWithAuth(`/api/v1/admin/rbac/roles/${roleId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteSystemRole(roleId: string): Promise<ApiResponse<void>> {
  return fetchWithAuth(`/api/v1/admin/rbac/roles/${roleId}`, { method: 'DELETE' })
}

export async function getRolePermissions(roleId: string): Promise<ApiResponse<Permission[]>> {
  return fetchWithAuth(`/api/v1/admin/rbac/roles/${roleId}/permissions`)
}

export async function assignPermissionsToRole(roleId: string, permissionIds: string[]): Promise<ApiResponse<void>> {
  return fetchWithAuth(`/api/v1/admin/rbac/roles/${roleId}/permissions`, {
    method: 'POST',
    body: JSON.stringify({ permission_ids: permissionIds }),
  })
}

export async function revokePermissionsFromRole(roleId: string, permissionIds: string[]): Promise<ApiResponse<void>> {
  return fetchWithAuth(`/api/v1/admin/rbac/roles/${roleId}/permissions`, {
    method: 'DELETE',
    body: JSON.stringify({ permission_ids: permissionIds }),
  })
}

export async function getRBACStats(): Promise<ApiResponse<RBACStats>> {
  return fetchWithAuth('/api/v1/admin/rbac/stats')
}

// ============================================================================
// CLIENT APPS
// ============================================================================

export async function listClientApps(): Promise<ApiResponse<ClientApp[]>> {
  return fetchWithAuth('/api/v1/admin/client-apps')
}

export async function getClientApp(id: string): Promise<ApiResponse<ClientApp>> {
  return fetchWithAuth(`/api/v1/admin/client-apps/${id}`)
}
