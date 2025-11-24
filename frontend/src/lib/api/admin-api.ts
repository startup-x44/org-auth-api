import axios from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  }
  return config
})

// Handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      
      try {
        const refreshToken = localStorage.getItem('refresh_token')
        if (!refreshToken) throw new Error('No refresh token')
        
        const { data } = await axios.post(`${API_URL}/api/v1/auth/refresh`, {
          refresh_token: refreshToken,
        })
        
        localStorage.setItem('access_token', data.access_token)
        originalRequest.headers.Authorization = `Bearer ${data.access_token}`
        
        return api(originalRequest)
      } catch (refreshError) {
        localStorage.clear()
        window.location.href = '/auth/login'
        return Promise.reject(refreshError)
      }
    }
    
    return Promise.reject(error)
  }
)

export interface DashboardStats {
  total_users: number
  total_organizations: number
  active_sessions: number
  total_roles: number
  total_permissions: number
  api_requests_today: number
}

export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  status: string
  is_superadmin: boolean
  created_at: string
  updated_at: string
  last_login_at: string | null
}

export interface Organization {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
  member_count: number
  owner_id: string
}

export interface Role {
  id: string
  name: string
  description: string
  organization_id: string | null
  is_system: boolean
  created_at: string
  updated_at: string
}

export interface Permission {
  id: string
  name: string
  description: string
  resource: string
  action: string
  is_system: boolean
  created_at: string
}

export const adminApi = {
  // Dashboard
  getDashboardStats: async (): Promise<DashboardStats> => {
    const { data } = await api.get('/api/v1/admin/stats')
    return data.data
  },

  // Users
  getUsers: async (params?: { page?: number; limit?: number; search?: string }) => {
    const { data } = await api.get('/api/v1/admin/users', { params })
    return data
  },

  getUserById: async (id: string): Promise<User> => {
    const { data } = await api.get(`/api/v1/admin/users/${id}`)
    return data.data
  },

  updateUser: async (id: string, updates: Partial<User>): Promise<User> => {
    const { data } = await api.put(`/api/v1/admin/users/${id}`, updates)
    return data.data
  },

  deleteUser: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/admin/users/${id}`)
  },

  activateUser: async (id: string): Promise<void> => {
    await api.post(`/api/v1/admin/users/${id}/activate`)
  },

  deactivateUser: async (id: string): Promise<void> => {
    await api.post(`/api/v1/admin/users/${id}/deactivate`)
  },

  // Organizations
  getOrganizations: async (params?: { page?: number; limit?: number; search?: string }) => {
    const { data } = await api.get('/api/v1/admin/organizations', { params })
    return data
  },

  getOrganizationById: async (id: string): Promise<Organization> => {
    const { data } = await api.get(`/api/v1/admin/organizations/${id}`)
    return data.data
  },

  updateOrganization: async (id: string, updates: Partial<Organization>): Promise<Organization> => {
    const { data } = await api.put(`/api/v1/admin/organizations/${id}`, updates)
    return data.data
  },

  deleteOrganization: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/admin/organizations/${id}`)
  },

  // Roles
  getRoles: async (params?: { page?: number; limit?: number }) => {
    const { data } = await api.get('/api/v1/admin/roles', { params })
    return data
  },

  createRole: async (role: { name: string; description: string; organization_id?: string }) => {
    const { data } = await api.post('/api/v1/admin/roles', role)
    return data.data
  },

  updateRole: async (id: string, updates: Partial<Role>): Promise<Role> => {
    const { data } = await api.put(`/api/v1/admin/roles/${id}`, updates)
    return data.data
  },

  deleteRole: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/admin/roles/${id}`)
  },

  // Permissions
  getPermissions: async (params?: { page?: number; limit?: number }) => {
    const { data } = await api.get('/api/v1/admin/permissions', { params })
    return data
  },

  createPermission: async (permission: { name: string; description: string; resource: string; action: string }) => {
    const { data } = await api.post('/api/v1/admin/permissions', permission)
    return data.data
  },

  deletePermission: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/admin/permissions/${id}`)
  },

  assignPermissionsToRole: async (roleId: string, permissionIds: string[]): Promise<void> => {
    await api.post(`/api/v1/admin/roles/${roleId}/permissions`, { permission_ids: permissionIds })
  },

  getRolePermissions: async (roleId: string) => {
    const { data } = await api.get(`/api/v1/admin/roles/${roleId}/permissions`)
    return data.data
  },
}

export default adminApi
