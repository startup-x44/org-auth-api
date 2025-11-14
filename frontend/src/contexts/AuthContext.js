import React, { createContext, useContext, useState, useEffect } from 'react';
import { authAPI } from '../services/api';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [tenantId, setTenantId] = useState(null);

  // Helper function to get tenant-specific localStorage key
  const getStorageKey = (key, tenantId) => {
    return tenantId ? `${key}_${tenantId}` : key;
  };

  useEffect(() => {
    // Check if user is logged in on app start
    const storedTenantId = localStorage.getItem('tenant_id');
    if (storedTenantId) {
      const token = localStorage.getItem(getStorageKey('access_token', storedTenantId));
      const storedUser = localStorage.getItem(getStorageKey('user', storedTenantId));

      if (token && storedUser) {
        setUser(JSON.parse(storedUser));
        setTenantId(storedTenantId);
      }
    }

    setLoading(false);
  }, []);

  const login = async (email, password, tenantId) => {
    try {
      const response = await authAPI.login({
        email,
        password,
        tenant_id: tenantId,
      });

      const { user: userData, token } = response.data.data;

      // Use tenant-specific storage keys
      const tokenKey = getStorageKey('access_token', tenantId);
      const refreshKey = getStorageKey('refresh_token', tenantId);
      const userKey = getStorageKey('user', tenantId);

      localStorage.setItem(tokenKey, token.access_token);
      localStorage.setItem(refreshKey, token.refresh_token);
      localStorage.setItem(userKey, JSON.stringify(userData));
      localStorage.setItem('tenant_id', tenantId); // Global tenant reference

      setUser(userData);
      setTenantId(tenantId);

      return { success: true };
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.message || 'Login failed',
      };
    }
  };

  const register = async (userData) => {
    try {
      const response = await authAPI.register({
        ...userData,
        tenant_id: userData.tenant_id,
      });
      const { user: newUser, token } = response.data.data;

      // Use tenant-specific storage keys
      const tokenKey = getStorageKey('access_token', newUser.tenant_id);
      const refreshKey = getStorageKey('refresh_token', newUser.tenant_id);
      const userKey = getStorageKey('user', newUser.tenant_id);

      localStorage.setItem(tokenKey, token.access_token);
      localStorage.setItem(refreshKey, token.refresh_token);
      localStorage.setItem(userKey, JSON.stringify(newUser));
      localStorage.setItem('tenant_id', newUser.tenant_id); // Global tenant reference

      setUser(newUser);
      setTenantId(newUser.tenant_id);

      return { success: true };
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.message || 'Registration failed',
      };
    }
  };

  const logout = async () => {
    try {
      await authAPI.logout({
        user_id: user?.id,
        refresh_token: localStorage.getItem(getStorageKey('refresh_token', tenantId)),
      });
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Clear tenant-specific storage
      if (tenantId) {
        localStorage.removeItem(getStorageKey('access_token', tenantId));
        localStorage.removeItem(getStorageKey('refresh_token', tenantId));
        localStorage.removeItem(getStorageKey('user', tenantId));
      }
      localStorage.removeItem('tenant_id');
      setUser(null);
      setTenantId(null);
    }
  };

  const updateProfile = async (profileData) => {
    try {
      const response = await authAPI.updateProfile(profileData);
      const updatedUser = response.data.data;

      localStorage.setItem(getStorageKey('user', tenantId), JSON.stringify(updatedUser));
      setUser(updatedUser);

      return { success: true };
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.message || 'Profile update failed',
      };
    }
  };

  const changePassword = async (passwordData) => {
    try {
      await authAPI.changePassword(passwordData);
      return { success: true };
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.message || 'Password change failed',
      };
    }
  };

  const isAdmin = () => {
    return user?.user_type === 'admin';
  };

  const isAuthenticated = () => {
    return !!user && !!tenantId && !!localStorage.getItem(getStorageKey('access_token', tenantId));
  };

  const value = {
    user,
    tenantId,
    loading,
    login,
    register,
    logout,
    updateProfile,
    changePassword,
    isAdmin,
    isAuthenticated,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};