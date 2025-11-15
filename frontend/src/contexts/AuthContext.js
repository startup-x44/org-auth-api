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
  const [organization, setOrganization] = useState(null);
  const [organizations, setOrganizations] = useState([]);
  const [needsOrgSelection, setNeedsOrgSelection] = useState(false);

  // Helper function to get organization-specific localStorage key
  const getStorageKey = (key, orgId) => {
    return orgId ? `${key}_${orgId}` : key;
  };

  useEffect(() => {
    // Check if user is logged in on app start
    const storedOrgId = localStorage.getItem('organization_id');
    if (storedOrgId) {
      const token = localStorage.getItem(getStorageKey('access_token', storedOrgId));
      const storedUser = localStorage.getItem(getStorageKey('user', storedOrgId));
      const storedOrg = localStorage.getItem(getStorageKey('organization', storedOrgId));

      if (token && storedUser && storedOrg) {
        setUser(JSON.parse(storedUser));
        setOrganization(JSON.parse(storedOrg));
      }
    }

    setLoading(false);
  }, []);

  const login = async (email, password) => {
    try {
      const response = await authAPI.login({
        email,
        password,
      });

      const { user: userData, organizations: userOrgs } = response.data.data;

      // Store user info globally (not org-specific yet)
      localStorage.setItem('user_global', JSON.stringify(userData));
      setUser(userData);
      setOrganizations(userOrgs);

      // If user has organizations, they need to select one
      if (userOrgs && userOrgs.length > 0) {
        setNeedsOrgSelection(true);
        return { success: true, needsOrgSelection: true, organizations: userOrgs };
      }

      // No organizations - user needs to create one
      return { success: true, needsOrgSelection: false, organizations: [] };
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.message || 'Login failed',
      };
    }
  };

  const selectOrganization = async (organizationId) => {
    try {
      const response = await authAPI.selectOrganization({
        user_id: user.id,
        organization_id: organizationId,
      });

      const { token, organization: orgData } = response.data.data;

      // Use organization-specific storage keys
      const tokenKey = getStorageKey('access_token', organizationId);
      const refreshKey = getStorageKey('refresh_token', organizationId);
      const userKey = getStorageKey('user', organizationId);
      const orgKey = getStorageKey('organization', organizationId);

      localStorage.setItem(tokenKey, token.access_token);
      localStorage.setItem(refreshKey, token.refresh_token);
      localStorage.setItem(userKey, JSON.stringify(user));
      localStorage.setItem(orgKey, JSON.stringify(orgData));
      localStorage.setItem('organization_id', organizationId); // Global org reference

      setOrganization(orgData);
      setNeedsOrgSelection(false);

      return { success: true };
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.message || 'Organization selection failed',
      };
    }
  };

  const createOrganization = async (orgData) => {
    try {
      const response = await authAPI.createOrganization({
        user_id: user.id,
        ...orgData,
      });

      const { token, organization: newOrg } = response.data.data;

      // Use organization-specific storage keys
      const tokenKey = getStorageKey('access_token', newOrg.id);
      const refreshKey = getStorageKey('refresh_token', newOrg.id);
      const userKey = getStorageKey('user', newOrg.id);
      const orgKey = getStorageKey('organization', newOrg.id);

      localStorage.setItem(tokenKey, token.access_token);
      localStorage.setItem(refreshKey, token.refresh_token);
      localStorage.setItem(userKey, JSON.stringify(user));
      localStorage.setItem(orgKey, JSON.stringify(newOrg));
      localStorage.setItem('organization_id', newOrg.id); // Global org reference

      setOrganization(newOrg);
      setNeedsOrgSelection(false);

      return { success: true };
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.message || 'Organization creation failed',
      };
    }
  };

  const register = async (userData) => {
    try {
      const response = await authAPI.register({
        email: userData.email,
        password: userData.password,
        confirm_password: userData.confirmPassword,
        first_name: userData.firstName,
        last_name: userData.lastName,
      });
      
      const { user: newUser } = response.data.data;

      // Store user info globally
      localStorage.setItem('user_global', JSON.stringify(newUser));
      setUser(newUser);
      setNeedsOrgSelection(false); // New users need to create an organization

      return { success: true, needsOrgCreation: true };
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
        refresh_token: localStorage.getItem(getStorageKey('refresh_token', organization?.id)),
      });
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Clear organization-specific storage
      if (organization?.id) {
        localStorage.removeItem(getStorageKey('access_token', organization.id));
        localStorage.removeItem(getStorageKey('refresh_token', organization.id));
        localStorage.removeItem(getStorageKey('user', organization.id));
        localStorage.removeItem(getStorageKey('organization', organization.id));
      }
      localStorage.removeItem('organization_id');
      localStorage.removeItem('user_global');
      setUser(null);
      setOrganization(null);
      setOrganizations([]);
      setNeedsOrgSelection(false);
    }
  };

  const updateProfile = async (profileData) => {
    try {
      const response = await authAPI.updateProfile(profileData);
      const updatedUser = response.data.data;

      localStorage.setItem(getStorageKey('user', organization?.id), JSON.stringify(updatedUser));
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

  const switchOrganization = async (organizationId) => {
    // Clear current org storage
    if (organization?.id) {
      localStorage.removeItem(getStorageKey('access_token', organization.id));
      localStorage.removeItem(getStorageKey('refresh_token', organization.id));
      localStorage.removeItem(getStorageKey('user', organization.id));
      localStorage.removeItem(getStorageKey('organization', organization.id));
    }

    // Select new organization
    return await selectOrganization(organizationId);
  };

  const getMyOrganizations = async () => {
    try {
      const response = await authAPI.getMyOrganizations();
      const orgs = response.data.data;
      setOrganizations(orgs);
      return { success: true, organizations: orgs };
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.message || 'Failed to fetch organizations',
      };
    }
  };

  const isAdmin = () => {
    return user?.global_role === 'admin' || user?.is_superadmin;
  };

  const isOrgAdmin = () => {
    return organization?.role === 'admin' || organization?.role === 'owner';
  };

  const isAuthenticated = () => {
    return !!user && !!organization && !!localStorage.getItem(getStorageKey('access_token', organization.id));
  };

  const value = {
    user,
    organization,
    organizations,
    needsOrgSelection,
    loading,
    login,
    register,
    logout,
    selectOrganization,
    createOrganization,
    switchOrganization,
    getMyOrganizations,
    updateProfile,
    changePassword,
    isAdmin,
    isOrgAdmin,
    isAuthenticated,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};