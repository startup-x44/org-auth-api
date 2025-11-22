import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import LoadingSpinner from './ui/loading-spinner'

export function DashboardRedirect() {
  const { user, isLoading, hasRole } = useAuthStore()
  const location = useLocation()
  
  // Prevent infinite redirect loops
  const currentPath = location.pathname
  if (currentPath.includes('/admin/') || 
      currentPath.includes('/user/') || 
      currentPath.includes('/superadmin/')) {
    return null
  }
  
  // Wait for user data to load
  if (isLoading || !user) {
    return <LoadingSpinner size="sm" />
  }
  
  // Role-based dashboard routing with proper hierarchy
  if (user.is_superadmin || hasRole('superadmin')) {
    return <Navigate to="/admin/dashboard" replace />
  }
  
  if (hasRole('admin')) {
    return <Navigate to="/admin/dashboard" replace />
  }
  
  if (hasRole('user') || hasRole('member')) {
    return <Navigate to="/user/dashboard" replace />
  }
  
  // Fallback for users without clear roles
  return <Navigate to="/user/dashboard" replace />
}