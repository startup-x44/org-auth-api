import React, { useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import useAuthStore from '@/store/auth'
import LoadingSpinner from '@/components/ui/loading-spinner'
import { Toaster } from '@/components/ui/toaster'

// Auth pages
import Login from '@/pages/auth/Login'
import Register from '@/pages/auth/Register'
import VerifyEmail from '@/pages/auth/VerifyEmail'
import ForgotPassword from '@/pages/auth/ForgotPassword'
import ResetPassword from '@/pages/auth/ResetPassword'
import AcceptInvitation from '@/pages/auth/AcceptInvitation'
import ChooseOrganization from '@/pages/auth/ChooseOrganization'
import CreateOrganization from '@/pages/auth/CreateOrganization'

// User pages
import Dashboard from '@/pages/user/Dashboard'
import Profile from '@/pages/user/Profile'
import Settings from '@/pages/user/Settings'
import Members from '@/pages/user/Members'
import RoleManagement from '@/pages/user/RoleManagement'

// Superadmin pages
import SuperadminDashboard from '@/pages/superadmin/SuperadminDashboard'
import Admin from '@/pages/superadmin/Admin'
import OAuthClientApps from '@/pages/superadmin/OAuthClientApps'
import OAuthAuditLogs from '@/pages/superadmin/OAuthAuditLogs'
import RBACManagement from '@/pages/superadmin/RBACManagement'

// OAuth pages
import OAuthConsent from '@/pages/oauth/OAuthConsent'
import OAuthCallback from '@/pages/oauth/OAuthCallback'

// Developer pages
import DeveloperDocs from '@/pages/DeveloperDocs'
import ApiKeys from '@/pages/dev/ApiKeys'

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

// Router future flags for React Router v7 compatibility
const routerFutureFlags = {
  v7_startTransition: true,
  v7_relativeSplatPath: true,
}

// Protected Route component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, loading } = useAuthStore()

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />
}

// Public Route component (redirects based on user type)
const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, isSuperadmin, loading } = useAuthStore()

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (isAuthenticated) {
    // Redirect superadmin to dedicated dashboard
    return <Navigate to={isSuperadmin ? "/superadmin" : "/dashboard"} replace />
  }

  return <>{children}</>
}

// Dashboard router - redirects based on user type
const DashboardRouter: React.FC = () => {
  const { isAuthenticated, isSuperadmin, loading } = useAuthStore()

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  // Redirect superadmin to their dashboard
  if (isSuperadmin) {
    return <Navigate to="/superadmin" replace />
  }

  // Regular users get the user dashboard
  return <Dashboard />
}

// Smart redirect based on user type
const SmartRedirect: React.FC = () => {
  const { isAuthenticated, isSuperadmin, loading } = useAuthStore()

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Navigate to={isSuperadmin ? "/superadmin" : "/dashboard"} replace />
}

function App() {
  const { initialize } = useAuthStore()

  useEffect(() => {
    initialize()
  }, [initialize])

  return (
    <QueryClientProvider client={queryClient}>
      <Router future={routerFutureFlags}>
        <div className="App">
          <Routes>
            {/* Public routes */}
            <Route
              path="/login"
              element={
                <PublicRoute>
                  <Login />
                </PublicRoute>
              }
            />
            <Route
              path="/register"
              element={
                <PublicRoute>
                  <Register />
                </PublicRoute>
              }
            />
            <Route
              path="/verify-email"
              element={
                <PublicRoute>
                  <VerifyEmail />
                </PublicRoute>
              }
            />
            <Route
              path="/forgot-password"
              element={
                <PublicRoute>
                  <ForgotPassword />
                </PublicRoute>
              }
            />
            <Route
              path="/reset-password"
              element={
                <PublicRoute>
                  <ResetPassword />
                </PublicRoute>
              }
            />
            <Route
              path="/accept-invitation"
              element={<AcceptInvitation />}
            />
            <Route
              path="/choose-organization"
              element={
                <ChooseOrganization />
              }
            />
            <Route
              path="/create-organization"
              element={
                <CreateOrganization />
              }
            />

            {/* Superadmin routes */}
            <Route
              path="/superadmin"
              element={
                <ProtectedRoute>
                  <SuperadminDashboard />
                </ProtectedRoute>
              }
            />

            {/* Protected routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardRouter />
                </ProtectedRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <Profile />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings"
              element={
                <ProtectedRoute>
                  <Settings />
                </ProtectedRoute>
              }
            />
            <Route
              path="/members"
              element={
                <ProtectedRoute>
                  <Members />
                </ProtectedRoute>
              }
            />
            <Route
              path="/roles"
              element={
                <ProtectedRoute>
                  <RoleManagement />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin"
              element={
                <ProtectedRoute>
                  <Admin />
                </ProtectedRoute>
              }
            />
            <Route
              path="/oauth/client-apps"
              element={
                <ProtectedRoute>
                  <OAuthClientApps />
                </ProtectedRoute>
              }
            />
            <Route
              path="/oauth/audit"
              element={
                <ProtectedRoute>
                  <OAuthAuditLogs />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin/rbac"
              element={
                <ProtectedRoute>
                  <RBACManagement />
                </ProtectedRoute>
              }
            />
            <Route
              path="/oauth/authorize"
              element={
                <ProtectedRoute>
                  <OAuthConsent />
                </ProtectedRoute>
              }
            />
            <Route
              path="/oauth/callback"
              element={<OAuthCallback />}
            />
            <Route
              path="/developer/docs"
              element={
                <ProtectedRoute>
                  <DeveloperDocs />
                </ProtectedRoute>
              }
            />
            <Route
              path="/dev/api-keys"
              element={
                <ProtectedRoute>
                  <ApiKeys />
                </ProtectedRoute>
              }
            />

            {/* Default redirect */}
            <Route path="/" element={<SmartRedirect />} />
            <Route path="*" element={<SmartRedirect />} />
          </Routes>
        </div>
        <Toaster />
      </Router>
    </QueryClientProvider>
  )
}

export default App