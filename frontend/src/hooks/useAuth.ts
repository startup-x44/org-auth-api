import { useAuthStore } from '@/stores/authStore'

// Helper hooks for common use cases
export const useAuth = () => {
  const store = useAuthStore()
  return {
    user: store.user,
    isAuthenticated: store.isAuthenticated,
    isLoading: store.isLoading,
    error: store.error,
    mfaChallenge: store.mfaChallenge,
    requiresMFA: store.requiresMFA,
    login: store.login,
    verifyMFA: store.verifyMFA,
    logout: store.logout,
    clearError: store.clearError,
  }
}

export const useCurrentUser = () => useAuthStore((state) => state.user)
export const useCurrentOrganization = () => useAuthStore((state) => state.currentOrganization)
export const useIsAuthenticated = () => useAuthStore((state) => state.isAuthenticated)
export const useAuthLoading = () => useAuthStore((state) => state.isLoading)

// Helper functions for computed user properties
export const useUserName = () => {
  const user = useAuthStore((state) => state.user)
  return user ? `${user.firstName} ${user.lastName}`.trim() : ''
}

export const useUserRole = () => {
  const user = useAuthStore((state) => state.user)
  return user?.roles?.[0]?.name || 'member'
}

export const useHasPermission = () => {
  const user = useAuthStore((state) => state.user)
  
  return (permission: string): boolean => {
    if (!user) return false
    
    // Super admin has all permissions
    if (user.roles?.some(role => role.name === 'super_admin')) {
      return true
    }
    
    // Check if user has the specific permission
    return user.permissions?.some(p => 
      p.name === permission || p.name === '*'
    ) || false
  }
}