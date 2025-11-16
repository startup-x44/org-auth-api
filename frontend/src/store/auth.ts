import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { authAPI, userAPI } from '../lib/api'
import { AuthState, User, RegisterRequest, UpdateProfileRequest, ChangePasswordRequest } from '../lib/types'
import { decodeJWT } from '../lib/jwt'

interface AuthStore extends AuthState {
  // Actions
  login: (email: string, password: string) => Promise<{ success: boolean; message?: string; needsOrgSelection?: boolean; organizations?: any[] }>
  selectOrganization: (organizationId: string) => Promise<{ success: boolean; message?: string }>
  createOrganization: (name: string, slug: string) => Promise<{ success: boolean; message?: string }>
  switchOrganization: (organizationId: string) => Promise<{ success: boolean; message?: string }>
  getMyOrganizations: () => Promise<{ success: boolean; organizations?: any[] }>
  register: (data: RegisterRequest) => Promise<{ success: boolean; message?: string; needsOrgCreation?: boolean }>
  logout: () => Promise<void>
  performTokenRefresh: () => Promise<{ success: boolean }>
  updateProfile: (data: UpdateProfileRequest) => Promise<{ success: boolean; message?: string }>
  changePassword: (data: ChangePasswordRequest) => Promise<{ success: boolean; message?: string }>
  forgotPassword: (email: string) => Promise<{ success: boolean; message?: string }>
  resetPassword: (token: string, password: string, confirmPassword: string) => Promise<{ success: boolean; message?: string }>
  initialize: () => void
  clearError: () => void
  setOrganizationId: (organizationId: string | null) => void
  hasPermission: (permission: string) => boolean
  hasAnyPermission: (...permissions: string[]) => boolean
  hasAllPermissions: (...permissions: string[]) => boolean
}

const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      // State
      user: null,
      accessToken: null,
      refreshToken: null,
      tenantId: null,
      organizationId: null,
      organization: null,
      organizations: [],
      permissions: [],
      roleId: null,
      roleName: null,
      needsOrgSelection: false,
      isAuthenticated: false,
      isSuperadmin: false,
      loading: true,
      error: null,

      // Actions
      setUser: (user: User | null) => set({
        user,
        isSuperadmin: user?.is_superadmin || false,
      }),

      setOrganizationId: (organizationId: string | null) => {
        set({ organizationId })
        if (organizationId) {
          localStorage.setItem('organization_id', organizationId)
        } else {
          localStorage.removeItem('organization_id')
        }
      },

      login: async (email: string, password: string) => {
        try {
          // REMOVED: set({ loading: true, error: null }) to prevent Zustand persist triggers
          console.log('🟢 auth.login() called - NOT updating Zustand state to avoid component remount')
          
          const response = await authAPI.login({ email, password })
          
          console.log('Login API response (full):', response) // Debug log
          
          // Backend returns { data: { user, organizations }, success, message }
          // The API wrapper does .then(res => res.data), so we get the outer wrapper
          // We need to extract from response.data
          const user = (response as any).data?.user || response.user
          const organizations = (response as any).data?.organizations || response.organizations
          
          console.log('Extracted user:', user) // Debug log
          console.log('Extracted organizations:', organizations) // Debug log

          if (!user) {
            console.error('No user in response! Response structure:', response)
            const errorMessage = 'Invalid login response - no user data'
            // Only set error state, not loading state
            set({ error: errorMessage })
            return { success: false, message: errorMessage }
          }

          // Store user globally (but don't update Zustand state yet to avoid re-render)
          console.log('Storing user in localStorage:', user)
          localStorage.setItem('user_global', JSON.stringify(user))
          localStorage.setItem('organizations_temp', JSON.stringify(organizations || []))
          
          // REMOVED: Don't update loading state here to prevent Zustand persist triggers
          // Component now manages its own loading state
          console.log('🟢 Login successful - returning result without updating Zustand state')

          console.log('Organizations after login:', organizations) // Debug log

          // Check if user has organizations
          if (organizations && organizations.length > 0) {
            return { success: true, needsOrgSelection: true, organizations, user }
          }

          // No organizations - user needs to create one
          return { success: true, needsOrgSelection: false, organizations: [], user }
        } catch (error: any) {
          console.error('Login error:', error) // Debug log
          const errorMessage = error.response?.data?.message || error.message || 'Login failed'
          // Only set error state, not loading state
          set({ error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      selectOrganization: async (organizationId: string) => {
        try {
          set({ loading: true, error: null })
          const { user } = get()
          
          if (!user?.id) {
            const errorMessage = 'User not authenticated'
            set({ loading: false, error: errorMessage })
            return { success: false, message: errorMessage }
          }
          
          const response = await authAPI.selectOrganization({
            user_id: user.id,
            organization_id: organizationId
          })
          
          console.log('Select organization response:', response)
          
          // Backend returns { success: true, data: { token, organization } }
          // After .then(res => res.data), we get the full response body
          const token = (response as any).data?.token
          const organization = (response as any).data?.organization

          console.log('Extracted token:', token)
          console.log('Extracted organization:', organization)

          if (!token || !organization) {
            console.error('Invalid response structure:', response)
            const errorMessage = 'Invalid response from server'
            set({ loading: false, error: errorMessage })
            return { success: false, message: errorMessage }
          }

          // Decode JWT to extract permissions
          const claims = decodeJWT(token.access_token)
          const permissions = claims?.permissions || []
          const roleId = organization.role_id || null
          const roleName = organization.role_name || organization.role || null

          // Store with org-specific keys
          localStorage.setItem(`access_token_${organizationId}`, token.access_token)
          localStorage.setItem(`refresh_token_${organizationId}`, token.refresh_token)
          localStorage.setItem(`user_${organizationId}`, JSON.stringify(user))
          localStorage.setItem(`organization_${organizationId}`, JSON.stringify(organization))
          localStorage.setItem('organization_id', organizationId)

          set({
            accessToken: token.access_token,
            refreshToken: token.refresh_token,
            organizationId,
            organization,
            permissions,
            roleId,
            roleName,
            isAuthenticated: true,
            needsOrgSelection: false,
            loading: false,
          })

          return { success: true }
        } catch (error: any) {
          console.error('Select organization error:', error)
          const errorMessage = error.response?.data?.message || error.message || 'Organization selection failed'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      createOrganization: async (name: string, slug: string) => {
        try {
          set({ loading: true, error: null })
          let { user } = get()
          
          console.log('CreateOrganization called with:', { name, slug })
          console.log('User from state:', user)
          
          // If no user in state, try to get from localStorage
          if (!user) {
            const userGlobal = localStorage.getItem('user_global')
            console.log('user_global from localStorage (raw):', userGlobal)
            if (userGlobal && userGlobal !== 'undefined') {
              try {
                user = JSON.parse(userGlobal)
                console.log('Retrieved user from localStorage:', user)
              } catch (e) {
                console.error('Failed to parse user_global:', e)
              }
            }
          }
          
          if (!user?.id) {
            console.error('User not authenticated! User object:', user)
            const errorMessage = 'User not authenticated'
            set({ loading: false, error: errorMessage })
            return { success: false, message: errorMessage }
          }
          
          const payload = {
            user_id: user.id,
            name,
            slug
          }
          
          console.log('Creating organization with payload:', payload)
          
          const response = await authAPI.createOrganization(payload)
          
          console.log('Create organization response (full):', response)
          
          // Backend returns { success: true, data: { organization, token } }
          // authAPI.createOrganization does .then(res => res.data), so response = { success, data: {...} }
          const { data } = response as any
          const { token, organization } = data || {}
          
          console.log('Extracted token:', token)
          console.log('Extracted organization:', organization)

          if (!token || !organization) {
            console.error('Invalid response structure. Full response:', response)
            const errorMessage = 'Invalid response from server - missing token or organization'
            set({ loading: false, error: errorMessage })
            return { success: false, message: errorMessage }
          }

          // Decode JWT to extract permissions
          const claims = decodeJWT(token.access_token)
          const permissions = claims?.permissions || []
          const roleId = organization.role_id || null
          const roleName = organization.role_name || organization.role || null

          // Store with org-specific keys
          localStorage.setItem(`access_token_${organization.id}`, token.access_token)
          localStorage.setItem(`refresh_token_${organization.id}`, token.refresh_token)
          localStorage.setItem(`user_${organization.id}`, JSON.stringify(user))
          localStorage.setItem(`organization_${organization.id}`, JSON.stringify(organization))
          localStorage.setItem('organization_id', organization.id)

          set((state) => ({
            ...state,
            user,
            accessToken: token.access_token,
            refreshToken: token.refresh_token,
            organizationId: organization.id,
            organization,
            organizations: [...(state.organizations || []), organization], // Add new organization to list
            permissions,
            roleId,
            roleName,
            isAuthenticated: true,
            needsOrgSelection: false,
            loading: false,
          }))

          return { success: true }
        } catch (error: any) {
          console.error('Organization creation error:', error)
          const errorMessage = error.response?.data?.message || error.message || 'Organization creation failed'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      switchOrganization: async (organizationId: string) => {
        // Clear current org storage
        const { organizationId: currentOrgId } = get()
        if (currentOrgId) {
          localStorage.removeItem(`access_token_${currentOrgId}`)
          localStorage.removeItem(`refresh_token_${currentOrgId}`)
          localStorage.removeItem(`user_${currentOrgId}`)
          localStorage.removeItem(`organization_${currentOrgId}`)
        }

        // Select new organization
        return get().selectOrganization(organizationId)
      },

      getMyOrganizations: async () => {
        try {
          set({ loading: true, error: null })
          const response = await userAPI.getMyOrganizations()
          const organizations = response.data || []

          set({
            organizations,
            loading: false,
            error: null,
          })

          return { success: true, organizations }
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || error.message || 'Failed to fetch organizations'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage, organizations: [] }
        }
      },

      register: async (data: RegisterRequest) => {
        try {
          set({ loading: true, error: null })
          const response = await authAPI.register(data)
          const { user } = response

          // Store user globally
          localStorage.setItem('user_global', JSON.stringify(user))

          set({
            user,
            loading: false,
            error: null,
          })

          // New users need to create an organization
          return { success: true, needsOrgCreation: true }
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || error.message || 'Registration failed'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      logout: async () => {
        try {
          await authAPI.logout()
        } catch (error) {
          console.error('Logout error:', error)
        } finally {
          const { organizationId } = get()
          
          // Clear org-specific storage
          if (organizationId) {
            localStorage.removeItem(`access_token_${organizationId}`)
            localStorage.removeItem(`refresh_token_${organizationId}`)
            localStorage.removeItem(`user_${organizationId}`)
            localStorage.removeItem(`organization_${organizationId}`)
          }
          
          // Clear global storage
          localStorage.removeItem('organization_id')
          localStorage.removeItem('user_global')
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          localStorage.removeItem('user')
          localStorage.removeItem('tenant_id')
          
          // Clear Zustand persist storage
          localStorage.removeItem('auth-storage')
          
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            tenantId: null,
            organizationId: null,
            organization: null,
            organizations: [],
            permissions: [],
            roleId: null,
            roleName: null,
            needsOrgSelection: false,
            isAuthenticated: false,
            isSuperadmin: false,
            error: null,
          })
        }
      },

      performTokenRefresh: async () => {
        try {
          const { refreshToken } = get()
          if (!refreshToken) {
            console.error('No refresh token available')
            get().logout()
            return { success: false }
          }

          const response = await authAPI.refreshToken({ refresh_token: refreshToken })
          const { token } = response

          set({
            accessToken: token.access_token,
            refreshToken: token.refresh_token,
          })

          return { success: true }
        } catch (error) {
          // If refresh fails, logout
          get().logout()
          return { success: false }
        }
      },

      updateProfile: async (data: UpdateProfileRequest) => {
        try {
          set({ loading: true, error: null })
          const response = await userAPI.updateProfile(data)
          const updatedUser = response.user

          set({
            user: updatedUser,
            loading: false,
            error: null,
          })

          return { success: true }
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || error.message || 'Profile update failed'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      changePassword: async (data: ChangePasswordRequest) => {
        try {
          set({ loading: true, error: null })
          await userAPI.changePassword(data)
          set({ loading: false, error: null })
          return { success: true }
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || error.message || 'Password change failed'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      forgotPassword: async (email: string) => {
        try {
          set({ loading: true, error: null })
          await authAPI.forgotPassword({ email })
          set({ loading: false, error: null })
          return { success: true }
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || error.message || 'Failed to send reset email'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      resetPassword: async (token: string, password: string, confirmPassword: string) => {
        try {
          set({ loading: true, error: null })
          await authAPI.resetPassword({ token, password, confirm_password: confirmPassword })
          set({ loading: false, error: null })
          return { success: true }
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || error.message || 'Password reset failed'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      // Initialize auth state
      initialize: () => {
        const { accessToken, user } = get()
        if (accessToken && user) {
          set({
            isAuthenticated: true,
            isSuperadmin: user.is_superadmin || false,
            loading: false,
          })
        } else {
          set({ loading: false })
        }
      },

      // Clear error
      clearError: () => set({ error: null }),

      // Permission checks
      hasPermission: (permission: string) => {
        const { permissions, isSuperadmin } = get()
        if (isSuperadmin) return true
        return permissions.includes(permission)
      },

      hasAnyPermission: (...requiredPermissions: string[]) => {
        const { permissions, isSuperadmin } = get()
        if (isSuperadmin) return true
        return requiredPermissions.some(p => permissions.includes(p))
      },

      hasAllPermissions: (...requiredPermissions: string[]) => {
        const { permissions, isSuperadmin } = get()
        if (isSuperadmin) return true
        return requiredPermissions.every(p => permissions.includes(p))
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        tenantId: state.tenantId,
        organizationId: state.organizationId,
        organization: state.organization,
        organizations: state.organizations,
        permissions: state.permissions,
        roleId: state.roleId,
        roleName: state.roleName,
        needsOrgSelection: state.needsOrgSelection,
        isAuthenticated: state.isAuthenticated,
        isSuperadmin: state.isSuperadmin,
      }),
    }
  )
)

export default useAuthStore