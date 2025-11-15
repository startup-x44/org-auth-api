import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import useAuthStore from './stores/authStore';
import { NotificationContainer } from './components/shared';
import Login from './components/auth/Login';
import Register from './components/auth/Register';
import ForgotPassword from './components/auth/ForgotPassword';
import ResetPassword from './components/auth/ResetPassword';
import Dashboard from './components/dashboard/Dashboard';
import Profile from './components/profile/Profile';
import AdminUsers from './components/admin/Users';
import AdminTenants from './components/admin/Tenants';
import Layout from './components/layout/Layout';
import './App.css';

// Protected Route component
const ProtectedRoute = ({ children, adminOnly = false, superAdminOnly = false }) => {
  const { isAuthenticated, user, isLoading } = useAuthStore();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (superAdminOnly && user?.user_type !== 'superadmin') {
    return <Navigate to="/dashboard" replace />;
  }

  if (adminOnly && !['admin', 'superadmin'].includes(user?.user_type)) {
    return <Navigate to="/dashboard" replace />;
  }

  return children;
};

// App Routes component
const AppRoutes = () => {
  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/reset-password" element={<ResetPassword />} />

      {/* Protected routes */}
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <Layout>
              <Dashboard />
            </Layout>
          </ProtectedRoute>
        }
      />

      <Route
        path="/profile"
        element={
          <ProtectedRoute>
            <Layout>
              <Profile />
            </Layout>
          </ProtectedRoute>
        }
      />

      {/* Organization routes - placeholder for future implementation */}
      <Route
        path="/organizations"
        element={
          <ProtectedRoute>
            <Layout>
              <div className="p-6">
                <h1 className="text-2xl font-bold text-foreground">Organizations</h1>
                <p className="mt-2 text-gray-600">Organization management coming soon...</p>
              </div>
            </Layout>
          </ProtectedRoute>
        }
      />

      {/* Admin routes */}
      <Route
        path="/admin/users"
        element={
          <ProtectedRoute superAdminOnly>
            <Layout>
              <AdminUsers />
            </Layout>
          </ProtectedRoute>
        }
      />

      <Route
        path="/admin/tenants"
        element={
          <ProtectedRoute superAdminOnly>
            <Layout>
              <AdminTenants />
            </Layout>
          </ProtectedRoute>
        }
      />

      {/* Default redirect */}
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
};

// Main App component
function App() {
  return (
    <Router>
      <div className="App">
        <AppRoutes />
        <NotificationContainer />
      </div>
    </Router>
  );
}

export default App;