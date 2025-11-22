import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { jwtDecode } from 'jwt-decode'

import { getCSRFToken } from '../api/client'
import type { 
  AuthStore, 
  AuthState, 
  LoginCredentials, 
  MFAVerification, 
  User, 
  Organization, 
  LoginResponse,
  RefreshTokenResponse 
} from '../types/auth'
import { apiClient } from '../api/client'
import { secureStorage } from '../utils/storage'
import { logger } from '../utils/logger'

// Initial state
const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: true,
  currentOrganization: null,
  availableOrganizations: [],
  requiresMFA: false,
  mfaChallenge: null,
  error: null,
  sessionExpiresAt: null,
  lastActivity: Date.now(),
}

// Token utilities
const REFRESH_TOKEN_BUFFER = 5 * 60 * 1000 // 5 minutes before expiry

const isTokenExpired = (token: string): boolean => {
  try {
    const decoded = jwtDecode(token)
    if (!decoded.exp) return true
    return decoded.exp * 1000 <= Date.now()
  } catch {
    return true
  }
}

const shouldRefreshToken = (expiresAt: number): boolean => {
  return Date.now() >= (expiresAt - REFRESH_TOKEN_BUFFER)
}

// Store implementation
export const useAuthStore = create<AuthStore>()(
  devtools(
    persist(
      immer((set, get) => ({
        ...initialState,

        // Authentication flow
        login: async (credentials: LoginCredentials) => {
          // Prevent multiple simultaneous login attempts
          const { isLoading } = get()
          if (isLoading) {
            logger.warn('Login attempt blocked - already in progress')
            return { requiresMFA: false }
          }

          try {
            set((state) => {
              state.isLoading = true
              state.error = null
              state.requiresMFA = false
              state.mfaChallenge = null
            })

            logger.info('Attempting login', { email: credentials.email })

            // Get CSRF token first if we don't have one
            if (!secureStorage.getCSRFToken()) {
              logger.debug('Getting CSRF token before login')
              try {
                await apiClient.get('/health')
              } catch (csrfError) {
                logger.warn('Failed to get CSRF token, continuing with login', csrfError)
              }
            }

            // Add timeout to prevent hanging  
            const response = await Promise.race([
              apiClient.post<LoginResponse>('/v1/auth/login', credentials),
              new Promise((_, reject) => 
                setTimeout(() => reject(new Error('Login timeout')), 15000)
              )
            ]) as any
            
            if (response.data.requiresMFA) {
              set((state) => {
                state.requiresMFA = true
                state.mfaChallenge = response.data.mfaChallenge || null
                state.isLoading = false
              })
              
              return { requiresMFA: true }
            }

            // Successful login without MFA
            const { token, tokens, user, organizations } = response.data.data
            
            // Handle both token formats (superadmin gets 'token', org selection gets 'tokens')
            let authTokens = tokens || token
            
            if (authTokens && user) {
              // Convert API response format to AuthTokens interface format
              if (authTokens.expires_in && !authTokens.expiresAt) {
                authTokens = {
                  accessToken: authTokens.access_token,
                  refreshToken: authTokens.refresh_token,
                  expiresAt: Date.now() + (authTokens.expires_in * 1000),
                  tokenType: 'Bearer'
                }
              }
              
              // Store tokens securely
              secureStorage.setTokens(authTokens)
              
              // Update state
              set((state) => {
                state.user = user
                state.isAuthenticated = true
                state.availableOrganizations = organizations || []
                state.currentOrganization = organizations?.[0] || null
                state.sessionExpiresAt = authTokens.expiresAt
                state.lastActivity = Date.now()
                state.isLoading = false
                state.requiresMFA = false
                state.mfaChallenge = null
              })

              logger.info('Login successful', { userId: user.id })
            } else {
              // Handle case where response doesn't contain expected data
              set((state) => {
                state.isLoading = false
                state.error = 'Invalid login response'
              })
              throw new Error('Invalid login response')
            }

            return {}
          } catch (error: any) {
            logger.error('Login failed', error)
            
            set((state) => {
              state.error = error.response?.data?.message || error.message || 'Login failed'
              state.isLoading = false
              state.requiresMFA = false
              state.mfaChallenge = null
            })
            
            throw error
          }
        },

        verifyMFA: async (verification: MFAVerification) => {
          try {
            set((state) => {
              state.isLoading = true
              state.error = null
            })

            logger.info('Verifying MFA', { challengeId: verification.challengeId })

            const response = await apiClient.post<LoginResponse>('/v1/auth/mfa/verify', verification)
            const { tokens, user, organizations } = response.data

            if (tokens && user) {
              // Store tokens securely
              secureStorage.setTokens(tokens)
              
              // Update state
              set((state) => {
                state.user = user
                state.isAuthenticated = true
                state.availableOrganizations = organizations || []
                state.currentOrganization = organizations?.[0] || null
                state.sessionExpiresAt = tokens.expiresAt
                state.lastActivity = Date.now()
                state.isLoading = false
                state.requiresMFA = false
                state.mfaChallenge = null
              })

              logger.info('MFA verification successful', { userId: user.id })
            }
          } catch (error: any) {
            logger.error('MFA verification failed', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'MFA verification failed'
              state.isLoading = false
            })
            
            throw error
          }
        },

        logout: async () => {
          try {
            logger.info('Logging out')
            
            // Call logout endpoint
            await apiClient.post('/v1/user/logout')
          } catch (error) {
            logger.warn('Logout API call failed', error)
            // Continue with local cleanup even if API fails
          } finally {
            // Clear tokens and state
            secureStorage.clearTokens()
            
            set((state) => {
              Object.assign(state, {
                ...initialState,
                isLoading: false,
              })
            })
            
            logger.info('Logout completed')
          }
        },

        refreshTokens: async () => {
          try {
            const currentTokens = secureStorage.getTokens()
            if (!currentTokens?.refreshToken) {
              throw new Error('No refresh token available')
            }

            logger.debug('Refreshing tokens')

            const response = await apiClient.post<RefreshTokenResponse>('/v1/auth/refresh', {
              refreshToken: currentTokens.refreshToken
            })

            const { tokens } = response.data
            secureStorage.setTokens(tokens)

            set((state) => {
              state.sessionExpiresAt = tokens.expiresAt
              state.lastActivity = Date.now()
            })

            logger.debug('Tokens refreshed successfully')
          } catch (error: any) {
            logger.error('Token refresh failed', error)
            
            // If refresh fails, logout user
            get().logout()
            throw error
          }
        },

        updateProfile: async (updates: Partial<User>) => {
          try {
            const response = await apiClient.patch<{ user: User }>('/v1/user/profile', updates)
            
            set((state) => {
              if (state.user) {
                Object.assign(state.user, response.data.user)
              }
            })

            logger.info('Profile updated successfully')
          } catch (error: any) {
            logger.error('Profile update failed', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'Profile update failed'
            })
            
            throw error
          }
        },

        changePassword: async (currentPassword: string, newPassword: string) => {
          try {
            await apiClient.post('/v1/user/change-password', {
              currentPassword,
              newPassword
            })

            logger.info('Password changed successfully')
          } catch (error: any) {
            logger.error('Password change failed', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'Password change failed'
            })
            
            throw error
          }
        },

        switchOrganization: async (organizationId: string) => {
          try {
            const organization = get().availableOrganizations.find(org => org.id === organizationId)
            if (!organization) {
              throw new Error('Organization not found')
            }

            // Call API to switch context
            await apiClient.post('/v1/auth/select-organization', { organizationId })

            set((state) => {
              state.currentOrganization = organization
            })

            logger.info('Organization switched', { organizationId, name: organization.name })
          } catch (error: any) {
            logger.error('Organization switch failed', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'Failed to switch organization'
            })
            
            throw error
          }
        },

        initialize: async () => {
          // Prevent multiple simultaneous initializations
          const state = get()
          if (state.isLoading) {
            logger.debug('Auth initialization already in progress, skipping')
            return
          }
          
          if (state.isAuthenticated) {
            logger.debug('Already authenticated, skipping initialization')
            return
          }

          logger.debug('Starting auth initialization...')

          set((state) => {
            state.isLoading = true
            state.error = null
          })

          try {
            const tokens = secureStorage.getTokens()
            logger.debug('Retrieved tokens from storage:', tokens ? 'found' : 'none')
            
            if (!tokens?.accessToken) {
              logger.debug('No access token found, user not authenticated')
              set((state) => {
                state.isLoading = false
              })
              logger.debug('Set loading to false, returning from initialization')
              return
            }

            // Check if token is expired
            if (isTokenExpired(tokens.accessToken)) {
              logger.debug('Access token expired, attempting refresh')
              if (tokens.refreshToken && !isTokenExpired(tokens.refreshToken)) {
                try {
                  await get().refreshTokens()
                  // After refresh, tokens should be updated, continue with profile fetch
                } catch (refreshError) {
                  logger.warn('Token refresh failed during initialization', refreshError)
                  await get().logout()
                  return
                }
              } else {
                logger.debug('Refresh token expired or missing, logging out')
                await get().logout()
                return
              }
            }

            // Try to get current user profile with shorter timeout
            logger.debug('Fetching user profile')
            try {
              const response = await Promise.race([
                apiClient.get<{ user: User; organizations: Organization[] }>('/v1/user/profile'),
                new Promise((_, reject) => 
                  setTimeout(() => reject(new Error('Profile fetch timeout')), 2000) // Reduced to 2 seconds
                )
              ]) as any

              const { user, organizations } = response.data

              set((state) => {
                state.user = user
                state.isAuthenticated = true
                state.availableOrganizations = organizations
                state.currentOrganization = organizations[0] || null
                state.sessionExpiresAt = tokens.expiresAt
                state.isLoading = false
              })
            } catch (profileError) {
              logger.warn('Profile fetch failed, clearing auth state', profileError)
              // Clear tokens and force re-login
              secureStorage.clearTokens()
              set((state) => {
                state.isAuthenticated = false
                state.isLoading = false
                state.user = null
                state.error = null
              })
            }

            logger.info('Auth initialization completed')
          } catch (error: any) {
            logger.error('Auth initialization failed', error)
            
            // Clear invalid tokens and reset state
            secureStorage.clearTokens()
            
            // Ensure loading state is cleared on any error
            set((state) => {
              state.isLoading = false
              state.error = 'Authentication initialization failed'
              state.isAuthenticated = false
              state.user = null
            })
          } finally {
            // Guarantee that loading is always set to false
            const currentState = get()
            if (currentState.isLoading) {
              logger.warn('Loading still true after initialization, forcing to false')
              set((state) => {
                state.isLoading = false
              })
            }
          }
        },

        updateLastActivity: () => {
          const currentTime = Date.now()
          const { lastActivity, isLoading } = get()
          
          // Don't update if already loading or if less than 30 seconds have passed
          if (isLoading || currentTime - lastActivity < 30000) return
          
          set((state) => {
            state.lastActivity = currentTime
          })
        },

        clearError: () => {
          set((state) => {
            state.error = null
          })
        },

        setError: (error: string) => {
          set((state) => {
            state.error = error
          })
        },

        // Debug function to force reset loading state
        forceResetLoading: () => {
          logger.warn('Force resetting loading state - this should only be used for debugging')
          set((state) => {
            state.isLoading = false
          })
        },

        // Role-based access control methods
        hasRole: (role: string) => {
          const { user } = get()
          if (!user) return false
          
          // Check if user has the specific role
          if (user.roles?.some(r => r.name === role)) return true
          
          // Check global role
          if (user.global_role === role) return true
          
          // Check is_superadmin flag for superadmin role
          if (role === 'superadmin' && user.is_superadmin) return true
          
          return false
        },

        hasPermission: (permission: string) => {
          const { user } = get()
          if (!user) return false
          
          // Get user permissions - this comes from JWT claims or user object
          const permissions = user.permissions || []
          
          // Check if user has the specific permission
          return permissions.includes(permission)
        },

        canAccess: (resource: string) => {
          const { user } = get()
          if (!user) return false
          
          // Define resource policies (matching backend logic)
          const adminResources = [
            'admin:', 'system:', 'rbac:', 'client-apps:', 'audit:',
            'users:create', 'users:update', 'users:delete', 'users:activate', 'users:deactivate',
            'organizations:create', 'organizations:update', 'organizations:delete',
            'roles:create', 'roles:update', 'roles:delete', 'roles:assign',
            'permissions:create', 'permissions:update', 'permissions:delete',
          ]
          
          const roleSpecificResources = [
            'role:user', 'role:member', 'role:admin', 'role:superadmin',
            'dashboard:user', 'dashboard:member', 'dashboard:admin',
            'access:user-routes', 'access:member-routes',
          ]
          
          const userResources = [
            'profile:', 'settings:', 'notifications:', 'user:read',
            'member:view', 'organization:view',
          ]
          
          // Check if it's a role-specific resource (strict matching required)
          for (const roleResource of roleSpecificResources) {
            if (resource.startsWith(roleResource)) {
              // For role-specific resources, check exact permission match
              // SuperAdmin CANNOT bypass role-specific permissions
              return get().hasPermission(resource)
            }
          }
          
          // Check if it's an administrative resource (superadmin bypass allowed)
          for (const adminResource of adminResources) {
            if (resource.startsWith(adminResource)) {
              // SuperAdmin can bypass admin resources
              if (user.is_superadmin || get().hasRole('superadmin')) {
                return true
              }
              // Non-superadmin must have the specific permission
              return get().hasPermission(resource)
            }
          }
          
          // Check if it's a user resource (role hierarchy applies)
          for (const userResource of userResources) {
            if (resource.startsWith(userResource)) {
              // Check role hierarchy: superadmin > admin > member > user
              if (user.is_superadmin || get().hasRole('superadmin')) return true
              if (get().hasRole('admin')) return true
              if (get().hasRole('member')) return true
              if (get().hasRole('user')) return true
              
              // Also check exact permission
              return get().hasPermission(resource)
            }
          }
          
          // Default: check exact permission match (no superadmin bypass)
          return get().hasPermission(resource)
        },
      })),
      {
        name: 'auth-store-persist',
        // Only persist minimal data (not tokens - those go in secure storage)
        partialize: (state) => ({
          lastActivity: state.lastActivity,
        }),
      }
    ),
    {
      name: 'auth-store-devtools',
    }
  )
)

// Auto-refresh tokens periodically
let refreshInterval: NodeJS.Timeout | null = null

// Start auto-refresh when store is created
const startAutoRefresh = () => {
  if (refreshInterval) return

  refreshInterval = setInterval(() => {
    const { isAuthenticated, sessionExpiresAt } = useAuthStore.getState()
    
    if (isAuthenticated && sessionExpiresAt && shouldRefreshToken(sessionExpiresAt)) {
      useAuthStore.getState().refreshTokens().catch((error) => {
        logger.error('Periodic token refresh failed', error)
      })
    }
  }, 60000) // Check every minute
}

const stopAutoRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

// Auto-refresh will be started manually when needed
export const startTokenAutoRefresh = () => {
  const { isAuthenticated } = useAuthStore.getState()
  if (isAuthenticated) {
    startAutoRefresh()
  }
}

export const stopTokenAutoRefresh = () => {
  stopAutoRefresh()
}

// Store is already exported above at line 52