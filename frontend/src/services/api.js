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
    return token;
  } catch (error) {
    console.error('Failed to get CSRF token:', error);
    return null;
  }
};

// Request interceptor to add auth token and CSRF token
api.interceptors.request.use(
  async (config) => {
    // Add auth token
    const orgId = localStorage.getItem('organization_id');
    const accessToken = orgId 
      ? localStorage.getItem(`access_token_${orgId}`)
      : localStorage.getItem('access_token');
    
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }

    // Add organization ID header
    if (orgId) {
      config.headers['X-Organization-ID'] = orgId;
    }

    // Add CSRF token for POST, PUT, DELETE, PATCH requests
    if (['post', 'put', 'delete', 'patch'].includes(config.method.toLowerCase())) {
      const csrfToken = await getCSRFToken();
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
        const orgId = localStorage.getItem('organization_id');
        const refreshToken = orgId
          ? localStorage.getItem(`refresh_token_${orgId}`)
          : localStorage.getItem('refresh_token');
        
        if (refreshToken) {
          const baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
          const response = await axios.post(
            `${baseURL}/api/v1/auth/refresh`,
            { refresh_token: refreshToken }
          );

          const { access_token, refresh_token } = response.data.data.token;

          // Store tokens with org prefix if applicable
          if (orgId) {
            localStorage.setItem(`access_token_${orgId}`, access_token);
            localStorage.setItem(`refresh_token_${orgId}`, refresh_token);
          } else {
            localStorage.setItem('access_token', access_token);
            localStorage.setItem('refresh_token', refresh_token);
          }

          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return api(originalRequest);
        }
      } catch (refreshError) {
        // Refresh failed, redirect to login
        const orgId = localStorage.getItem('organization_id');
        if (orgId) {
          localStorage.removeItem(`access_token_${orgId}`);
          localStorage.removeItem(`refresh_token_${orgId}`);
          localStorage.removeItem(`user_${orgId}`);
          localStorage.removeItem(`organization_${orgId}`);
        }
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user');
        localStorage.removeItem('organization_id');
        localStorage.removeItem('user_global');
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
  selectOrganization: (data) => api.post('/auth/select-organization', data),
  createOrganization: (data) => api.post('/auth/create-organization', data),
  getMyOrganizations: () => api.get('/user/organizations'),
  refreshToken: (data) => api.post('/auth/refresh', data),
  logout: () => api.post('/user/logout'),
  forgotPassword: (data) => api.post('/auth/forgot-password', data),
  resetPassword: (data) => api.post('/auth/reset-password', data),
};

// User API functions
export const userAPI = {
  getProfile: () => api.get('/user/profile'),
  updateProfile: (data) => api.put('/user/profile', data),
  changePassword: (data) => api.post('/user/change-password', data),
};

// Organization API functions
export const organizationAPI = {
  listOrganizations: () => api.get('/organizations'),
  createOrganization: (data) => api.post('/organizations', data),
  getOrganization: (orgId) => api.get(`/organizations/${orgId}`),
  updateOrganization: (orgId, data) => api.put(`/organizations/${orgId}`, data),
  deleteOrganization: (orgId) => api.delete(`/organizations/${orgId}`),

  // Members
  listMembers: (orgId) => api.get(`/organizations/${orgId}/members`),
  inviteUser: (orgId, data) => api.post(`/organizations/${orgId}/members`, data),
  updateMember: (orgId, userId, data) => api.put(`/organizations/${orgId}/members/${userId}`, data),
  removeMember: (orgId, userId) => api.delete(`/organizations/${orgId}/members/${userId}`),

  // Invitations
  listInvitations: (orgId) => api.get(`/organizations/${orgId}/invitations`),
  cancelInvitation: (orgId, invitationId) => api.delete(`/organizations/${orgId}/invitations/${invitationId}`),
  acceptInvitation: (token) => api.post(`/invitations/${token}/accept`),
  getInvitationDetails: (token) => api.get(`/invitations/${token}`),
};

// Admin API functions
export const adminAPI = {
  listUsers: (params) => api.get('/admin/users', { params }),
  activateUser: (userId) => api.put(`/admin/users/${userId}/activate`),
  deactivateUser: (userId) => api.put(`/admin/users/${userId}/deactivate`),
  deleteUser: (userId) => api.delete(`/admin/users/${userId}`),
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