// Route guards and protection components
export { default as ProtectedRoute } from './ProtectedRoute'
export { 
  AdminRoute, 
  SuperAdminRoute, 
  PublicRoute, 
  ConditionalRoute,
  PermissionGate,
  RoleGate
} from './ProtectedRoute'
export { usePermissions } from '../../hooks/usePermissions'
export type { 
  ProtectedRouteProps, 
  AdminRouteProps 
} from './ProtectedRoute'