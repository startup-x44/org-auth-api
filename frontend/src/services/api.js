import axios from 'axios';

// Create axios instance with base configuration
const api = axios.create({
  baseURL: `${process.env.REACT_APP_API_URL || 'http://localhost:8080'}/api/v1`,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Enable cookies for CSRF
});

// Function to get CSRF token
const getCSRFToken = async () => {
  try {
    const baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
    const response = await axios.get(`${baseURL}/health`, { 
      withCredentials: true,
      headers: {
        'Accept': 'application/json',
      }
    });
    // Axios normalizes headers to lowercase
    const token = response.headers['x-csrf-token'];
    console.log('CSRF Token fetched:', token); // Debug log
    return token;
  } catch (error) {
    console.error('Failed to get CSRF token:', error);
    return null;
  }
};

// Request interceptor to add auth token and CSRF token
api.interceptors.request.use(
  async (config) => {
    // Add tenant ID if available
    const tenantId = localStorage.getItem('tenant_id');
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId;
    }

    // Add auth token using tenant-specific key
    if (tenantId) {
      const token = localStorage.getItem(`access_token_${tenantId}`);
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }

    // Add CSRF token for POST, PUT, DELETE, PATCH requests
    if (['post', 'put', 'delete', 'patch'].includes(config.method.toLowerCase())) {
      const csrfToken = await getCSRFToken();
      console.log('Adding CSRF token to request:', csrfToken); // Debug log
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken;
      }
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle token refresh
api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const tenantId = localStorage.getItem('tenant_id');
        const refreshToken = tenantId ? localStorage.getItem(`refresh_token_${tenantId}`) : null;
        if (refreshToken) {
          const baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
          const response = await axios.post(
            `${baseURL}/api/v1/auth/refresh`,
            { refresh_token: refreshToken }
          );

          const { access_token, refresh_token } = response.data.data.token;

          // Store with tenant-specific keys
          if (tenantId) {
            localStorage.setItem(`access_token_${tenantId}`, access_token);
            localStorage.setItem(`refresh_token_${tenantId}`, refresh_token);
          }

          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return api(originalRequest);
        }
      } catch (refreshError) {
        // Refresh failed, redirect to login
        localStorage.removeItem('tenant_id');
        const tenantId = localStorage.getItem('tenant_id');
        if (tenantId) {
          localStorage.removeItem(`access_token_${tenantId}`);
          localStorage.removeItem(`refresh_token_${tenantId}`);
          localStorage.removeItem(`user_${tenantId}`);
        }
        window.location.href = '/login';
      }
    }

    return Promise.reject(error);
  }
);

// Auth API functions
export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  refreshToken: (data) => api.post('/auth/refresh', data),
  logout: (data) => api.post('/user/logout', data),
  forgotPassword: (data) => api.post('/auth/forgot-password', data),
  resetPassword: (data) => api.post('/auth/reset-password', data),
};

// User API functions
export const userAPI = {
  getProfile: () => api.get('/user/profile'),
  updateProfile: (data) => api.put('/user/profile', data),
  changePassword: (data) => api.post('/user/change-password', data),
};

// Admin API functions
export const adminAPI = {
  listUsers: (params) => api.get('/admin/users', { params }),
  activateUser: (userId) => api.put(`/admin/users/${userId}/activate`),
  deactivateUser: (userId) => api.put(`/admin/users/${userId}/deactivate`),
  deleteUser: (userId) => api.delete(`/admin/users/${userId}`),
  createTenant: (data) => api.post('/admin/tenants', data),
  listTenants: (params) => api.get('/admin/tenants', { params }),
  getTenant: (tenantId) => api.get(`/admin/tenants/${tenantId}`),
  updateTenant: (tenantId, data) => api.put(`/admin/tenants/${tenantId}`, data),
  deleteTenant: (tenantId) => api.delete(`/admin/tenants/${tenantId}`),
};

// Health check
export const healthAPI = {
  check: () => {
    const baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
    return axios.get(`${baseURL}/health`, { 
      withCredentials: true,
      headers: {
        'Accept': 'application/json',
      }
    });
  },
};

export default api;