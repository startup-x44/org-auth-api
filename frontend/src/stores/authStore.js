import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { authAPI } from '../services/api';
import useTenantStore from './tenantStore';

const useAuthStore = create(
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
      setUser: (user) => set({
        user,
        isSuperadmin: user?.is_superadmin || false,
        tenantId: user?.tenant_id || get().tenantId
      }),
      setTokens: (accessToken, refreshTokenValue) => set({ accessToken, refreshToken: refreshTokenValue }),
      setTenantId: (tenantId) => {
        set({ tenantId });
        if (tenantId) {
          localStorage.setItem('tenant_id', tenantId);
        } else {
          localStorage.removeItem('tenant_id');
        }
      },
      setAuthenticated: (isAuthenticated) => set({ isAuthenticated }),
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),

      // Auth actions
      login: async (email, password) => {
        try {
          set({ loading: true, error: null });
          const tenantStore = useTenantStore.getState();
          const tenantId = tenantStore.tenantId || tenantStore.resolveTenantFromEmail(email);

          const response = await authAPI.login({ email, password });
          const { user, token } = response.data.data;

          // Validate tenant access
          if (tenantId && user.tenant_id && user.tenant_id !== tenantId) {
            throw new Error('Access denied: User does not belong to this tenant');
          }

          set({
            user,
            accessToken: token.access_token,
            refreshToken: token.refresh_token,
            tenantId: user.tenant_id || tenantId,
            isAuthenticated: true,
            isSuperadmin: user.is_superadmin || false,
            loading: false,
            error: null,
          });

          return { success: true };
        } catch (error) {
          const errorMessage = error.response?.data?.message || error.message || 'Login failed';
          set({ loading: false, error: errorMessage });
          return { success: false, message: errorMessage };
        }
      },

      register: async (userData) => {
        try {
          set({ loading: true, error: null });
          const tenantStore = useTenantStore.getState();
          const tenantId = tenantStore.tenantId || tenantStore.resolveTenantFromEmail(userData.email);

          // Add tenant information to registration data
          const registrationData = {
            ...userData,
            tenant_id: tenantId,
          };

          const response = await authAPI.register(registrationData);
          const { user, token } = response.data.data;

          set({
            user,
            accessToken: token.access_token,
            refreshToken: token.refresh_token,
            tenantId: user.tenant_id || tenantId,
            isAuthenticated: true,
            isSuperadmin: user.is_superadmin || false,
            loading: false,
            error: null,
          });

          return { success: true };
        } catch (error) {
          const errorMessage = error.response?.data?.message || error.message || 'Registration failed';
          set({ loading: false, error: errorMessage });
          return { success: false, message: errorMessage };
        }
      },

      logout: async () => {
        try {
          await authAPI.logout();
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            tenantId: null,
            isAuthenticated: false,
            isSuperadmin: false,
            error: null,
          });
          // Also clear tenant store
          useTenantStore.getState().clearTenant();
        }
      },

      performTokenRefresh: async () => {
        try {
          const { refreshToken } = get();
          if (!refreshToken) throw new Error('No refresh token');

          const response = await authAPI.refreshToken({ refresh_token: refreshToken });
          const { access_token, refresh_token } = response.data.data.token;

          set({
            accessToken: access_token,
            refreshToken: refresh_token,
          });

          return { success: true };
        } catch (error) {
          // If refresh fails, logout
          get().logout();
          return { success: false };
        }
      },

      updateProfile: async (profileData) => {
        try {
          set({ loading: true, error: null });
          const response = await authAPI.updateProfile(profileData);
          const updatedUser = response.data.data;

          set({
            user: updatedUser,
            loading: false,
            error: null,
          });

          return { success: true };
        } catch (error) {
          const errorMessage = error.response?.data?.message || 'Profile update failed';
          set({ loading: false, error: errorMessage });
          return { success: false, message: errorMessage };
        }
      },

      changePassword: async (passwordData) => {
        try {
          set({ loading: true, error: null });
          await authAPI.changePassword(passwordData);
          set({ loading: false, error: null });
          return { success: true };
        } catch (error) {
          const errorMessage = error.response?.data?.message || 'Password change failed';
          set({ loading: false, error: errorMessage });
          return { success: false, message: errorMessage };
        }
      },

      forgotPassword: async (email) => {
        try {
          set({ loading: true, error: null });
          await authAPI.forgotPassword({ email });
          set({ loading: false, error: null });
          return { success: true };
        } catch (error) {
          const errorMessage = error.response?.data?.message || 'Failed to send reset email';
          set({ loading: false, error: errorMessage });
          return { success: false, message: errorMessage };
        }
      },

      resetPassword: async (token, password, confirmPassword) => {
        try {
          set({ loading: true, error: null });
          await authAPI.resetPassword({ token, password, confirm_password: confirmPassword });
          set({ loading: false, error: null });
          return { success: true };
        } catch (error) {
          const errorMessage = error.response?.data?.message || 'Password reset failed';
          set({ loading: false, error: errorMessage });
          return { success: false, message: errorMessage };
        }
      },

      // Initialize auth state
      initialize: () => {
        const { accessToken, user } = get();
        if (accessToken && user) {
          set({
            isAuthenticated: true,
            isSuperadmin: user.is_superadmin || false,
            loading: false,
          });
        } else {
          set({ loading: false });
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
);

export default useAuthStore;