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
    // Add auth token
    const accessToken = localStorage.getItem('access_token')
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`
    }

    // Add tenant ID header
    const tenantId = localStorage.getItem('tenant_id')
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId
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

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        const refreshToken = localStorage.getItem('refresh_token')
        if (refreshToken) {
          const baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080'
          const response: AxiosResponse = await axios.post(
            `${baseURL}/api/v1/auth/refresh`,
            { refresh_token: refreshToken }
          )

          const { access_token, refresh_token } = response.data.data.token

          // Store tokens
          localStorage.setItem('access_token', access_token)
          localStorage.setItem('refresh_token', refresh_token)

          originalRequest.headers.Authorization = `Bearer ${access_token}`
          return api(originalRequest)
        }
      } catch (refreshError) {
        // Refresh failed, redirect to login
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('user')
        localStorage.removeItem('tenant_id')
        window.location.href = '/login'
      }
    }

    return Promise.reject(error)
  }
)

export default api