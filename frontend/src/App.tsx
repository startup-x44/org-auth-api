import { Suspense, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ErrorBoundary } from 'react-error-boundary'
import { Toaster } from 'react-hot-toast'

// Store imports
import { useAuthStore } from './stores/authStore'
import { useAuth } from './hooks/useAuth'
import { useTenantStore } from './stores/tenantStore'

// Layout components
import { AppLayout } from './components/layout/AppLayout'
import { AuthLayout } from './components/layout/AuthLayout'

// Auth pages
import { LoginPage } from './pages/auth/LoginPage'
import { MFAPage } from './pages/auth/MFAPage'

// Main pages
import { DashboardPage } from './pages/DashboardPage'
import { ProfilePage } from './pages/ProfilePage'
import { AdminPage } from './pages/AdminPage'
import SuperadminDashboard from './pages/superadmin/SuperadminDashboard'

// Guard components
import { ProtectedRoute, AdminRoute, PublicRoute } from './components/guards'
import { RoleBasedRoute, UserOnlyRoute, AdminOnlyRoute, SuperAdminOnlyRoute } from './components/guards/RoleBasedRoute'
import { DashboardRedirect } from './components/DashboardRedirect'

// Error components
import { ErrorFallback } from './components/errors/ErrorFallback'
import { UnauthorizedPage } from './pages/UnauthorizedPage'
import LoadingSpinner from './components/ui/loading-spinner'
import AuthDebugInfo from './components/debug/AuthDebugInfo'

// Utilities
import { initializeApp } from './utils/app'
import { initSentry } from './utils/monitoring'
import { setupDevelopmentErrorTracking } from './utils/debug'

// Initialize monitoring
if (import.meta.env.PROD) {
  initSentry()
} else {
  // Development error tracking
  setupDevelopmentErrorTracking()
}

// Router future flags for React Router v7 compatibility  
const routerFutureFlags = {
  v7_startTransition: true,
  v7_relativeSplatPath: true,
}

function App() {
  const { initialize } = useAuthStore()
  const { isLoading: authLoading, isAuthenticated } = useAuth()
  const { initialize: initializeTenant } = useTenantStore()

  useEffect(() => {
    // Prevent double initialization in React StrictMode
    let isInitialized = false

    const init = async () => {
      if (isInitialized) return
      isInitialized = true

      try {
        console.log('🚀 Starting app initialization...')
        
        // Initialize app utilities
        await initializeApp()
        console.log('✅ App utilities initialized')
        
        // Initialize auth store with faster timeout
        console.log('🔐 Starting auth initialization...')
        const initPromise = initialize()
        const timeoutPromise = new Promise((_, reject) => 
          setTimeout(() => reject(new Error('Auth initialization timeout')), 5000) // Reduced from 15s to 5s
        )
        
        await Promise.race([initPromise, timeoutPromise])
        console.log('✅ Auth initialization completed')
        
        console.log(' App initialization complete!')
      } catch (error) {
        console.error('❌ App initialization failed:', error)
        
        // Force reset loading state if initialization fails
        const authStore = useAuthStore.getState()
        if (authStore.forceResetLoading) {
          console.warn('🔧 Force resetting loading state due to initialization failure')
          authStore.forceResetLoading()
        }
      }
    }

    init()

    return () => {
      isInitialized = false
    }
  }, []) // Run only once on mount

  // Separate effect for tenant initialization when user becomes authenticated
  useEffect(() => {
    if (isAuthenticated) {
      console.log('🏢 Initializing tenant store...')
      const { initialize: initTenant } = useTenantStore.getState()
      const { user } = useAuthStore.getState()
      initTenant(user || undefined).then(() => {
        console.log('✅ Tenant store initialized')
      }).catch((error) => {
        console.error('❌ Tenant initialization failed:', error)
      })
    }
  }, [isAuthenticated])

  // Force loading to stop after timeout
  useEffect(() => {
    if (authLoading) {
      const forceLoadingTimeout = setTimeout(() => {
        console.warn('⚡ Auth loading timeout - forcing app to continue')
        const authStore = useAuthStore.getState()
        // Force set loading to false directly
        authStore.forceResetLoading?.()
        // Also update state directly if method doesn't exist
        useAuthStore.setState({ isLoading: false })
      }, 2000) // Reduced to 2 seconds

      return () => clearTimeout(forceLoadingTimeout)
    }
  }, [authLoading])

  // Show loading state
  if (authLoading) {
    return (
      <div className="min-h-screen w-full flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 p-4">
        <div className="text-center max-w-md w-full mx-auto">
          <div className="relative mb-8">
            <div className="w-20 h-20 border-4 border-blue-300 border-t-blue-600 rounded-full animate-spin mx-auto"></div>
          </div>
          <div className="space-y-4">
            <h1 className="text-2xl sm:text-3xl font-bold text-gray-800">
              Welcome to NILOAUTH
            </h1>
            <p className="text-base sm:text-lg text-gray-600">
              Initializing secure authentication...
            </p>
            {import.meta.env.DEV && (
              <div className="mt-6 p-3 bg-white/50 rounded-lg">
                <p className="text-sm text-gray-500">
                  Debug: {authLoading ? 'loading' : 'ready'}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    )
  }

  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={(error, errorInfo) => {
        console.error('App Error:', error, errorInfo)
        // Send to monitoring service
        if (import.meta.env.PROD) {
          // Sentry will auto-capture
        }
      }}
      onReset={() => window.location.reload()}
    >
      <BrowserRouter future={routerFutureFlags}>
        <div className="min-h-screen bg-gray-50">
          <Suspense fallback={<LoadingSpinner size="lg" fullScreen />}>
            <Routes>
              {/* Public Auth Routes */}
              <Route path="/auth/*" element={
                <PublicRoute>
                  <AuthLayout>
                    <Routes>
                      <Route path="login" element={<LoginPage />} />
                      <Route path="mfa" element={<MFAPage />} />
                      <Route path="*" element={<Navigate to="/auth/login" replace />} />
                    </Routes>
                  </AuthLayout>
                </PublicRoute>
              } />

              {/* Error Pages */}
              <Route path="/unauthorized" element={<UnauthorizedPage />} />

              {/* Protected App Routes */}
              <Route path="/*" element={
                <ProtectedRoute>
                  <AppLayout>
                    <Routes>
                      {/* Dashboard Redirects */}
                      <Route path="/" element={<DashboardRedirect />} />
                      <Route path="/dashboard" element={<DashboardRedirect />} />
                      
                      {/* User/Member Strict Routes - SuperAdmin BLOCKED */}
                      <Route path="/user/*" element={
                        <UserOnlyRoute>
                          <Routes>
                            <Route path="dashboard" element={<DashboardPage />} />
                            <Route path="profile" element={<ProfilePage />} />
                            <Route path="*" element={<Navigate to="/user/dashboard" replace />} />
                          </Routes>
                        </UserOnlyRoute>
                      } />

                      {/* Admin Routes - SuperAdmin allowed */}
                      <Route path="/admin/*" element={
                        <AdminOnlyRoute>
                          <Routes>
                            <Route path="dashboard" element={<AdminPage />} />
                            <Route path="*" element={<Navigate to="/admin/dashboard" replace />} />
                          </Routes>
                        </AdminOnlyRoute>
                      } />

                      {/* SuperAdmin Only Routes */}
                      <Route path="/superadmin/*" element={
                        <SuperAdminOnlyRoute>
                          <Routes>
                            <Route path="dashboard" element={<SuperadminDashboard />} />
                            <Route path="*" element={<Navigate to="/superadmin/dashboard" replace />} />
                          </Routes>
                        </SuperAdminOnlyRoute>
                      } />
                      
                      {/* Legacy routes for backward compatibility */}
                      <Route path="/profile" element={<Navigate to="/user/profile" replace />} />
                      
                      {/* Catch-all - redirect to role-based dashboard */}
                      <Route path="*" element={<DashboardRedirect />} />
                    </Routes>
                  </AppLayout>
                </ProtectedRoute>
              } />
            </Routes>
          </Suspense>

          {/* Debug info in development */}
          <AuthDebugInfo />

          {/* Global Toast Notifications */}
          <Toaster
            position="top-right"
            toastOptions={{
              duration: 4000,
              style: {
                background: '#363636',
                color: '#fff',
              },
              success: {
                duration: 3000,
                iconTheme: {
                  primary: '#10b981',
                  secondary: '#fff',
                },
              },
              error: {
                duration: 5000,
                iconTheme: {
                  primary: '#ef4444',
                  secondary: '#fff',
                },
              },
            }}
          />
        </div>
      </BrowserRouter>
    </ErrorBoundary>
  )
}

export default App