// Core authentication and user types for the application

export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  avatar?: string
  roles: Role[]
  permissions: Permission[]
  isActive: boolean
  lastLoginAt?: string
  mfaEnabled: boolean
  emailVerified: boolean
  createdAt: string
  updatedAt: string
  // Superadmin fields
  global_role?: string
  is_superadmin?: boolean
}

export interface Role {
  id: string
  name: string
  description?: string
  permissions: Permission[]
  isSystem: boolean
  organizationId?: string
}

export interface Permission {
  id: string
  name: string
  resource: string
  action: string
  description?: string
  isSystem: boolean
  organizationId?: string
}

export interface Organization {
  id: string
  name: string
  slug: string
  domain?: string
  logo?: string
  settings: OrganizationSettings
  subscription: SubscriptionInfo
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface OrganizationSettings {
  allowSelfRegistration: boolean
  requireEmailVerification: boolean
  enforceMFA: boolean
  sessionTimeout: number
  passwordPolicy: PasswordPolicy
  brandingColors?: {
    primary: string
    secondary: string
  }
}

export interface PasswordPolicy {
  minLength: number
  requireUppercase: boolean
  requireLowercase: boolean
  requireNumbers: boolean
  requireSpecialChars: boolean
  preventReuse: number
  expirationDays?: number
}

export interface SubscriptionInfo {
  plan: 'free' | 'pro' | 'enterprise'
  status: 'active' | 'inactive' | 'canceled'
  expiresAt?: string
  limits: {
    maxUsers: number
    maxOrganizations: number
    features: string[]
  }
}

export interface AuthTokens {
  accessToken: string
  refreshToken: string
  expiresAt: number
  tokenType: 'Bearer'
}

export interface MFAChallenge {
  challengeId: string
  method: 'totp' | 'sms' | 'email'
  maskedTarget?: string // e.g., "****@example.com" or "***-***-1234"
}

export interface LoginCredentials {
  email: string
  password: string
  organizationSlug?: string
  rememberMe?: boolean
}

export interface MFAVerification {
  challengeId: string
  code: string
}

export interface AuthState {
  // User & Session
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  
  // Multi-tenant
  currentOrganization: Organization | null
  availableOrganizations: Organization[]
  
  // Auth flow
  requiresMFA: boolean
  mfaChallenge: MFAChallenge | null
  
  // Errors
  error: string | null
  
  // Session metadata
  sessionExpiresAt: number | null
  lastActivity: number
}

export interface AuthActions {
  // Authentication flow
  login: (credentials: LoginCredentials) => Promise<{ requiresMFA?: boolean }>
  verifyMFA: (verification: MFAVerification) => Promise<void>
  logout: () => Promise<void>
  
  // Token management
  refreshTokens: () => Promise<void>
  
  // User management
  updateProfile: (updates: Partial<User>) => Promise<void>
  changePassword: (currentPassword: string, newPassword: string) => Promise<void>
  
  // Organization management
  switchOrganization: (organizationId: string) => Promise<void>
  
  // Session management
  initialize: () => Promise<void>
  updateLastActivity: () => void
  
  // RBAC methods
  hasRole: (role: string) => boolean
  hasPermission: (permission: string) => boolean
  canAccess: (resource: string) => boolean
  
  // Error handling
  clearError: () => void
  setError: (error: string) => void
  forceResetLoading: () => void // Debug function
}

export type AuthStore = AuthState & AuthActions

// API Response types
export interface LoginResponse {
  tokens?: AuthTokens
  user?: User
  organizations?: Organization[]
  requiresMFA?: boolean
  mfaChallenge?: MFAChallenge
}

export interface RefreshTokenResponse {
  tokens: AuthTokens
}

export interface UserProfileResponse {
  user: User
}

export interface OrganizationsResponse {
  organizations: Organization[]
}

// Error types
export interface AuthError {
  code: string
  message: string
  details?: Record<string, unknown>
}

// Route protection types
export type PermissionCheck = string | string[] | ((user: User, org: Organization | null) => boolean)

export interface RouteGuardProps {
  children: React.ReactNode
  requiredPermissions?: PermissionCheck
  requiredRoles?: string[]
  fallback?: React.ReactNode
  redirectTo?: string
}