import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { authAPI, userAPI } from '../lib/api'
import { AuthState, User, RegisterRequest, UpdateProfileRequest, ChangePasswordRequest } from '../lib/types'

interface AuthStore extends AuthState {
  // Actions
  login: (email: string, password: string) => Promise<{ success: boolean; message?: string }>
  register: (data: RegisterRequest) => Promise<{ success: boolean; message?: string }>
  logout: () => Promise<void>
  performTokenRefresh: () => Promise<{ success: boolean }>
  updateProfile: (data: UpdateProfileRequest) => Promise<{ success: boolean; message?: string }>
  changePassword: (data: ChangePasswordRequest) => Promise<{ success: boolean; message?: string }>
  forgotPassword: (email: string) => Promise<{ success: boolean; message?: string }>
  resetPassword: (token: string, password: string, confirmPassword: string) => Promise<{ success: boolean; message?: string }>
  initialize: () => void
  clearError: () => void
  setTenantId: (tenantId: string | null) => void
}

const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      // State
      user: null,
      accessToken: null,
      refreshToken: null,
      tenantId: null,
      isAuthenticated: false,
      isSuperadmin: false,
      loading: true,
      error: null,

      // Actions
      setUser: (user: User | null) => set({
        user,
        isSuperadmin: user?.is_superadmin || false,
        tenantId: user?.tenant_id || get().tenantId
      }),

      setTenantId: (tenantId: string | null) => {
        set({ tenantId })
        if (tenantId) {
          localStorage.setItem('tenant_id', tenantId)
        } else {
          localStorage.removeItem('tenant_id')
        }
      },

      login: async (email: string, password: string) => {
        try {
          set({ loading: true, error: null })
          const response = await authAPI.login({ email, password })
          const { user, token } = response

          set({
            user,
            accessToken: token.access_token,
            refreshToken: token.refresh_token,
            tenantId: user.tenant_id || get().tenantId,
            isAuthenticated: true,
            isSuperadmin: user.is_superadmin || false,
            loading: false,
            error: null,
          })

          return { success: true }
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || error.message || 'Login failed'
          set({ loading: false, error: errorMessage })
          return { success: false, message: errorMessage }
        }
      },

      register: async (data: RegisterRequest) => {
        try {
          set({ loading: true, error: null })
          const response = await authAPI.register(data)
          const { user, token } = response

          set({
            user,
            accessToken: token.access_token,
            refreshToken: token.refresh_token,
            tenantId: user.tenant_id || get().tenantId,
            isAuthenticated: true,
            isSuperadmin: user.is_superadmin || false,
            loading: false,
            error: null,
          })

          return { success: true }
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
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            tenantId: null,
            isAuthenticated: false,
            isSuperadmin: false,
            error: null,
          })
          // Clear localStorage
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          localStorage.removeItem('user')
          localStorage.removeItem('tenant_id')
        }
      },

      performTokenRefresh: async () => {
        try {
          const { refreshToken } = get()
          if (!refreshToken) throw new Error('No refresh token')

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
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        tenantId: state.tenantId,
        isAuthenticated: state.isAuthenticated,
        isSuperadmin: state.isSuperadmin,
      }),
    }
  )
)

export default useAuthStore