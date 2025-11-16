import axios, { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'

const baseURL = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/v1`

// Public API instance (no auth, no interceptors for redirects)
export const publicAPI: AxiosInstance = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
})

// Function to get CSRF token
const getCSRFToken = async (): Promise<string | null> => {
  try {
    const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const response: AxiosResponse = await axios.get(`${baseURL}/health`, {
      withCredentials: true,
      headers: {
        'Accept': 'application/json',
      }
    })
    const token = response.headers['x-csrf-token']
    return token || null
  } catch (error) {
    console.error('Failed to get CSRF token:', error)
    return null
  }
}

// Add CSRF to public API
publicAPI.interceptors.request.use(
  async (config: InternalAxiosRequestConfig): Promise<InternalAxiosRequestConfig> => {
    console.log('🟢 publicAPI REQUEST:', config.url)
    if (config.method && ['post', 'put', 'delete', 'patch'].includes(config.method.toLowerCase())) {
      const csrfToken = await getCSRFToken()
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken
      }
    }
    return config
  },
  (error) => {
    console.log('🔴 publicAPI REQUEST ERROR:', error)
    return Promise.reject(error)
  }
)

// Add response interceptor to public API for debugging
publicAPI.interceptors.response.use(
  (response: AxiosResponse) => {
    console.log('🟢 publicAPI RESPONSE:', response.status, response.config.url)
    return response
  },
  (error) => {
    console.log('🔴 publicAPI RESPONSE ERROR:', error.response?.status, error.config?.url, error.message)
    return Promise.reject(error)
  }
)

// Authenticated API instance (with auth token and interceptors)
const api: AxiosInstance = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
})

// Request interceptor to add auth token and CSRF token
api.interceptors.request.use(
  async (config: InternalAxiosRequestConfig): Promise<InternalAxiosRequestConfig> => {
    // Get organization ID from localStorage
    const organizationId = localStorage.getItem('organization_id')
    
    // Add auth token - check org-specific token first, fallback to global
    let accessToken = null
    if (organizationId) {
      accessToken = localStorage.getItem(`access_token_${organizationId}`)
    }
    if (!accessToken) {
      accessToken = localStorage.getItem('access_token')
    }
    
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`
    }

    // Add organization ID header
    if (organizationId) {
      config.headers['X-Organization-ID'] = organizationId
    }

    // Add CSRF token for POST, PUT, DELETE, PATCH requests
    if (config.method && ['post', 'put', 'delete', 'patch'].includes(config.method.toLowerCase())) {
      const csrfToken = await getCSRFToken()
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken
      }
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle token refresh
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  async (error) => {
    const originalRequest = error.config

    // Handle 401 errors
    if (error.response?.status === 401) {
      // Don't logout on login/register endpoints
      const isAuthEndpoint = error.config?.url?.includes('/auth/login') || 
                            error.config?.url?.includes('/auth/register')
      
      if (isAuthEndpoint) {
        // Just return the error for auth endpoints (let component handle it)
        return Promise.reject(error)
      }
      
      // For protected routes, clear everything and logout
      console.log('🔴 401 Unauthorized - Clearing localStorage and redirecting to login')
      
      // Clear all authentication data
      localStorage.removeItem('auth-storage') // Zustand persisted state
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('organization_id')
      localStorage.removeItem('user_global')
      localStorage.removeItem('organizations_temp')
      
      // Clear all org-specific tokens
      Object.keys(localStorage).forEach(key => {
        if (key.startsWith('access_token_') || key.startsWith('refresh_token_')) {
          localStorage.removeItem(key)
        }
      })
      
      // Clear session storage
      sessionStorage.clear()
      
      // Redirect to login
      window.location.href = '/login'
      
      return Promise.reject(error)
    }

    return Promise.reject(error)
  }
)

export default api