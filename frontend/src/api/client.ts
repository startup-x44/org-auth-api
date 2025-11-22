import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig } from 'axios'
import { v4 as uuidv4 } from 'uuid'
import { secureStorage } from '../utils/storage'
import { logger } from '../utils/logger'

// Types
export interface ApiResponse<T = any> {
  data: T
  message?: string
  success: boolean
}

export interface ApiError {
  code: string
  message: string
  details?: Record<string, unknown>
  correlationId?: string
}

// Base API client configuration
const baseURL = import.meta.env.DEV ? '/api' : (import.meta.env.VITE_API_URL || '/api')

// Create axios instance
export const apiClient: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
})

// Request interceptor - Add auth tokens and headers
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Generate correlation ID for logging
    const correlationId = uuidv4()
    config.headers['X-Correlation-ID'] = correlationId
    
    // Add auth token if available
    const tokens = secureStorage.getTokens()
    if (tokens?.accessToken) {
      config.headers.Authorization = `Bearer ${tokens.accessToken}`
    }
    
    // Add CSRF token for state-changing requests
    const csrfToken = secureStorage.getCSRFToken()
    if (csrfToken && ['post', 'put', 'patch', 'delete'].includes(config.method || '')) {
      config.headers['X-CSRF-Token'] = csrfToken
    }
    
    // Add current organization context if available
    const currentOrg = secureStorage.getCurrentOrganization()
    if (currentOrg) {
      config.headers['X-Org-ID'] = currentOrg.id
    }
    
    logger.debug('API Request', {
      method: config.method?.toUpperCase(),
      url: config.url,
      correlationId,
      hasAuth: !!tokens?.accessToken,
      orgId: currentOrg?.id,
    })
    
    return config
  },
  (error) => {
    logger.error('Request interceptor error', error)
    return Promise.reject(error)
  }
)

// Response interceptor - Handle token refresh and errors
apiClient.interceptors.response.use(
  (response) => {
    const correlationId = response.config.headers['X-Correlation-ID']
    
    logger.debug('API Response', {
      method: response.config.method?.toUpperCase(),
      url: response.config.url,
      status: response.status,
      correlationId,
    })
    
    // Handle CSRF token updates
    const newCSRFToken = response.headers['x-csrf-token']
    if (newCSRFToken) {
      secureStorage.setCSRFToken(newCSRFToken)
    }
    
    return response
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }
    const correlationId = originalRequest?.headers?.['X-Correlation-ID']
    
    logger.error('API Error', {
      method: originalRequest?.method?.toUpperCase(),
      url: originalRequest?.url,
      status: error.response?.status,
      correlationId,
      message: error.message,
    })
    
    // Handle 401 Unauthorized - attempt token refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      
      try {
        const tokens = secureStorage.getTokens()
        if (tokens?.refreshToken) {
          logger.debug('Attempting token refresh due to 401')
          
          // Attempt refresh
          const refreshResponse = await axios.post(
            `${baseURL}/v1/auth/refresh`,
            { refreshToken: tokens.refreshToken },
            {
              headers: {
                'Content-Type': 'application/json',
                'X-Correlation-ID': uuidv4(),
              },
            }
          )
          
          const newTokens = refreshResponse.data.tokens
          secureStorage.setTokens(newTokens)
          
          // Update original request with new token
          originalRequest.headers.Authorization = `Bearer ${newTokens.accessToken}`
          
          logger.info('Token refresh successful, retrying original request')
          
          // Retry original request
          return apiClient(originalRequest)
        }
      } catch (refreshError) {
        logger.error('Token refresh failed', refreshError)
        
        // Clear tokens and redirect to login
        secureStorage.clearTokens()
        
        // Dispatch custom event for auth failure
        window.dispatchEvent(new CustomEvent('auth:token-refresh-failed'))
        
        return Promise.reject(error)
      }
    }
    
    // Handle 403 Forbidden - insufficient permissions
    if (error.response?.status === 403) {
      window.dispatchEvent(new CustomEvent('auth:insufficient-permissions', {
        detail: { 
          resource: originalRequest?.url,
          correlationId 
        }
      }))
    }
    
    // Handle 429 Too Many Requests - implement exponential backoff
    if (error.response?.status === 429 && !originalRequest._retry) {
      originalRequest._retry = true
      
      const retryAfter = error.response.headers['retry-after']
      const delay = retryAfter ? parseInt(retryAfter) * 1000 : 1000
      
      logger.warn('Rate limited, retrying after delay', { delay, correlationId })
      
      await new Promise(resolve => setTimeout(resolve, delay))
      return apiClient(originalRequest)
    }
    
    // Create structured error object
    const apiError: ApiError = {
      code: (error.response?.data as any)?.code || (error as any).code || 'UNKNOWN_ERROR',
      message: (error.response?.data as any)?.message || error.message || 'An unexpected error occurred',
      details: (error.response?.data as any)?.details,
      correlationId: correlationId as string,
    }
    
    return Promise.reject(apiError)
  }
)

// Helper function to get CSRF token
export const getCSRFToken = async (): Promise<string | null> => {
  try {
    const response = await apiClient.get('/health')
    const csrfToken = response.headers['x-csrf-token']
    if (csrfToken) {
      secureStorage.setCSRFToken(csrfToken)
      return csrfToken
    }
  } catch (error) {
    logger.error('Failed to get CSRF token', error)
  }
  return null
}

// Helper functions for common API patterns
export const api = {
  // Auth endpoints
  auth: {
    login: (credentials: any) => apiClient.post('/v1/auth/login', credentials),
    logout: () => apiClient.post('/v1/user/logout'),
    refresh: (refreshToken: string) => apiClient.post('/v1/auth/refresh', { refreshToken }),
    me: () => apiClient.get('/v1/user/profile'),
    mfaVerify: (verification: any) => apiClient.post('/v1/auth/mfa/verify', verification),
    changePassword: (data: any) => apiClient.post('/v1/user/change-password', data),
    switchOrganization: (organizationId: string) => apiClient.post('/v1/auth/select-organization', { organizationId }),
  },
  
  // Organization endpoints
  organizations: {
    list: () => apiClient.get('/v1/organizations'),
    get: (id: string) => apiClient.get(`/v1/organizations/${id}`),
    create: (data: any) => apiClient.post('/v1/organizations', data),
    update: (id: string, data: any) => apiClient.patch(`/v1/organizations/${id}`, data),
    delete: (id: string) => apiClient.delete(`/v1/organizations/${id}`),
  },
  
  // User management endpoints
  users: {
    list: (params?: any) => apiClient.get('/v1/admin/users', { params }),
    get: (id: string) => apiClient.get(`/v1/admin/users/${id}`),
    create: (data: any) => apiClient.post('/v1/admin/users', data),
    update: (id: string, data: any) => apiClient.patch(`/v1/admin/users/${id}`, data),
    delete: (id: string) => apiClient.delete(`/v1/admin/users/${id}`),
    invite: (data: any) => apiClient.post('/v1/organizations/invite', data),
    resendInvite: (id: string) => apiClient.post(`/v1/organizations/invitations/${id}/resend`),
  },
  
  // Role and permission endpoints
  roles: {
    list: () => apiClient.get('/v1/admin/rbac/roles'),
    get: (id: string) => apiClient.get(`/v1/admin/rbac/roles/${id}`),
    create: (data: any) => apiClient.post('/v1/admin/rbac/roles', data),
    update: (id: string, data: any) => apiClient.patch(`/v1/admin/rbac/roles/${id}`, data),
    delete: (id: string) => apiClient.delete(`/v1/admin/rbac/roles/${id}`),
  },
  
  permissions: {
    list: () => apiClient.get('/v1/admin/rbac/permissions'),
    check: (resource: string, action: string) => apiClient.get(`/v1/permissions/check`, { 
      params: { resource, action } 
    }),
  }
}

export default apiClient