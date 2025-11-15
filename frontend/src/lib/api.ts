import api from './axios-instance'
import {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  ProfileResponse,
  UpdateProfileRequest,
  ChangePasswordRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  Organization,
  ApiSuccess,
  SelectOrganizationRequest,
  SelectOrganizationResponse,
  CreateOrganizationRequest,
  CreateOrganizationResponse,
  OrganizationMembership
} from './types'

// Auth API functions
export const authAPI = {
  login: (data: LoginRequest): Promise<LoginResponse> =>
    api.post('/auth/login-global', data).then(res => res.data),

  register: (data: RegisterRequest): Promise<RegisterResponse> =>
    api.post('/auth/register-global', data).then(res => res.data),

  selectOrganization: (data: SelectOrganizationRequest): Promise<SelectOrganizationResponse> =>
    api.post('/auth/select-organization', data).then(res => res.data),

  createOrganization: (data: CreateOrganizationRequest): Promise<CreateOrganizationResponse> =>
    api.post('/auth/create-organization', data).then(res => res.data),

  refreshToken: (data: RefreshTokenRequest): Promise<RefreshTokenResponse> =>
    api.post('/auth/refresh', data).then(res => res.data),

  logout: (): Promise<ApiSuccess> =>
    api.post('/user/logout').then(res => res.data),

  forgotPassword: (data: ForgotPasswordRequest): Promise<ApiSuccess> =>
    api.post('/auth/forgot-password', data).then(res => res.data),

  resetPassword: (data: ResetPasswordRequest): Promise<ApiSuccess> =>
    api.post('/auth/reset-password', data).then(res => res.data),
}

// User API functions
export const userAPI = {
  getProfile: (): Promise<ProfileResponse> =>
    api.get('/user/profile').then(res => res.data),

  updateProfile: (data: UpdateProfileRequest): Promise<ProfileResponse> =>
    api.put('/user/profile', data).then(res => res.data),

  changePassword: (data: ChangePasswordRequest): Promise<ApiSuccess> =>
    api.post('/user/change-password', data).then(res => res.data),

  getMyOrganizations: (): Promise<ApiSuccess<OrganizationMembership[]>> =>
    api.get('/user/organizations').then(res => res.data),
}

// Organization API functions
export const organizationAPI = {
  listOrganizations: (): Promise<ApiSuccess<Organization[]>> =>
    api.get('/organizations').then(res => res.data),

  createOrganization: (data: { name: string; description?: string }): Promise<ApiSuccess<Organization>> =>
    api.post('/organizations', data).then(res => res.data),

  getOrganization: (orgId: string): Promise<ApiSuccess<Organization>> =>
    api.get(`/organizations/${orgId}`).then(res => res.data),

  updateOrganization: (orgId: string, data: { name?: string; description?: string }): Promise<ApiSuccess<Organization>> =>
    api.put(`/organizations/${orgId}`, data).then(res => res.data),

  deleteOrganization: (orgId: string): Promise<ApiSuccess> =>
    api.delete(`/organizations/${orgId}`).then(res => res.data),
}

// Admin API functions
export const adminAPI = {
  listUsers: (params?: { limit?: number; cursor?: string }): Promise<ApiSuccess<any>> =>
    api.get('/admin/users', { params }).then(res => res.data),

  activateUser: (userId: string): Promise<ApiSuccess> =>
    api.put(`/admin/users/${userId}/activate`).then(res => res.data),

  deactivateUser: (userId: string): Promise<ApiSuccess> =>
    api.put(`/admin/users/${userId}/deactivate`).then(res => res.data),

  deleteUser: (userId: string): Promise<ApiSuccess> =>
    api.delete(`/admin/users/${userId}`).then(res => res.data),
}

// Health check
export const healthAPI = {
  check: () => {
    const baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080'
    return api.get(`${baseURL}/health`, {
      withCredentials: true,
      headers: {
        'Accept': 'application/json',
      }
    })
  },
}