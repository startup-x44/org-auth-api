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
      // If already retried, redirect to login
      if (originalRequest._retry) {
        localStorage.clear()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      // Try to refresh token
      originalRequest._retry = true

      try {
        const organizationId = localStorage.getItem('organization_id')
        let refreshToken = null
        
        if (organizationId) {
          refreshToken = localStorage.getItem(`refresh_token_${organizationId}`)
        }
        if (!refreshToken) {
          refreshToken = localStorage.getItem('refresh_token')
        }
        
        if (refreshToken) {
          const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
          const response: AxiosResponse = await axios.post(
            `${baseURL}/api/v1/auth/refresh`,
            { refresh_token: refreshToken }
          )

          const { access_token, refresh_token: new_refresh_token } = response.data.data.token

          if (organizationId) {
            localStorage.setItem(`access_token_${organizationId}`, access_token)
            localStorage.setItem(`refresh_token_${organizationId}`, new_refresh_token)
          } else {
            localStorage.setItem('access_token', access_token)
            localStorage.setItem('refresh_token', new_refresh_token)
          }

          originalRequest.headers.Authorization = `Bearer ${access_token}`
          return api(originalRequest)
        } else {
          localStorage.clear()
          window.location.href = '/login'
          return Promise.reject(error)
        }
      } catch (refreshError) {
        localStorage.clear()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

export default api