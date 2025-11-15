import React from 'react';
import useAuthStore from '../../stores/authStore';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '../shared';
import useNotificationStore from '../../stores/notificationStore';

const Layout = ({ children }) => {
  const { user, logout } = useAuthStore();
  const { error: showError } = useNotificationStore();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
    } catch (err) {
      showError('Failed to logout. Please try again.');
    }
  };

  const isSuperAdmin = user?.user_type === 'superadmin';

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-semibold text-foreground">
                Auth Service
              </h1>
            </div>

            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">
                Welcome, {user?.first_name || user?.email}
              </span>

              <Button
                onClick={handleLogout}
                variant="secondary"
                size="sm"
              >
                Logout
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Sidebar */}
      <div className="flex">
        <nav className="w-64 bg-white shadow-sm min-h-screen border-r border-gray-200">
          <div className="p-6">
            <ul className="space-y-2">
              <li>
                <Link
                  to="/dashboard"
                  className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
                >
                  Dashboard
                </Link>
              </li>
              <li>
                <Link
                  to="/profile"
                  className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
                >
                  Profile
                </Link>
              </li>
              <li>
                <Link
                  to="/organizations"
                  className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
                >
                  Organizations
                </Link>
              </li>

              {isSuperAdmin && (
                <>
                  <li className="pt-4">
                    <div className="px-4 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                      Administration
                    </div>
                  </li>
                  <li>
                    <Link
                      to="/admin/users"
                      className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
                    >
                      User Management
                    </Link>
                  </li>
                  <li>
                    <Link
                      to="/admin/tenants"
                      className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
                    >
                      Tenant Management
                    </Link>
                  </li>
                </>
              )}
            </ul>
          </div>
        </nav>

        {/* Main content */}
        <main className="flex-1 p-8">
          {children}
        </main>
      </div>
    </div>
  );
};

export default Layout;