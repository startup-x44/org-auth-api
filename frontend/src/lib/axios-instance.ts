import axios, { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'

// Create axios instance with base configuration
const api: AxiosInstance = axios.create({
  baseURL: `${process.env.REACT_APP_API_URL || 'http://localhost:8080'}/api/v1`,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Enable cookies for CSRF
})

// Function to get CSRF token
const getCSRFToken = async (): Promise<string | null> => {
  try {
    const baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080'
    const response: AxiosResponse = await axios.get(`${baseURL}/health`, {
      withCredentials: true,
      headers: {
        'Accept': 'application/json',
      }
    })
    // Axios normalizes headers to lowercase
    const token = response.headers['x-csrf-token']
    return token || null
  } catch (error) {
    console.error('Failed to get CSRF token:', error)
    return null
  }
}

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

    // Debug log for create-organization requests
    if (config.url?.includes('create-organization')) {
      console.log('Axios interceptor - create-organization request:', {
        url: config.url,
        method: config.method,
        data: config.data,
        headers: config.headers
      })
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
      // If already retried or no refresh token available, redirect to login
      if (originalRequest._retry) {
        // Already tried refresh, clear everything and redirect
        localStorage.clear()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      // Try to refresh token
      originalRequest._retry = true

      try {
        const organizationId = localStorage.getItem('organization_id')
        let refreshToken = null
        
        // Get org-specific refresh token first, fallback to global
        if (organizationId) {
          refreshToken = localStorage.getItem(`refresh_token_${organizationId}`)
        }
        if (!refreshToken) {
          refreshToken = localStorage.getItem('refresh_token')
        }
        
        if (refreshToken) {
          const baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080'
          const response: AxiosResponse = await axios.post(
            `${baseURL}/api/v1/auth/refresh`,
            { refresh_token: refreshToken }
          )

          const { access_token, refresh_token: new_refresh_token } = response.data.data.token

          // Store tokens with org-specific keys
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
          // No refresh token, clear storage and redirect to login
          localStorage.clear()
          window.location.href = '/login'
          return Promise.reject(error)
        }
      } catch (refreshError) {
        // Refresh failed, clear everything and redirect to login
        localStorage.clear()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

export default api