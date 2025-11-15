// User types
export interface User {
  id: string
  email: string
  first_name?: string
  last_name?: string
  phone?: string
  user_type: string
  tenant_id?: string
  is_active: boolean
  is_superadmin?: boolean
  created_at: string
  updated_at: string
  last_login?: string
}

// Organization types
export interface Organization {
  id: string
  name: string
  slug: string
  description?: string
  status: string
  created_by: string
  created_at: string
  updated_at: string
}

// Organization membership
export interface OrganizationMembership {
  id: string
  organization_id: string
  user_id: string
  role: string
  status: string
  joined_at?: string
  created_at: string
  updated_at: string
  organization?: Organization
}

// API Response types
export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  user: User
  token: TokenPair
}

export interface RegisterRequest {
  email: string
  password: string
  confirm_password: string
  first_name: string
  last_name: string
  phone?: string
}

export interface RegisterResponse {
  user: User
  token: TokenPair
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface RefreshTokenRequest {
  refresh_token: string
}

export interface RefreshTokenResponse {
  token: TokenPair
}

export interface ProfileResponse {
  user: User
  organizations: Organization[]
}

export interface UpdateProfileRequest {
  first_name?: string
  last_name?: string
  phone?: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
  confirm_password: string
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ResetPasswordRequest {
  token: string
  password: string
  confirm_password: string
}

// Auth store state
export interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  tenantId: string | null
  isAuthenticated: boolean
  isSuperadmin: boolean
  loading: boolean
  error: string | null
}

// API Error response
export interface ApiError {
  success: false
  message: string
  errors?: Record<string, string[]>
}

// API Success response
export interface ApiSuccess<T = any> {
  success: true
  data: T
  message?: string
}