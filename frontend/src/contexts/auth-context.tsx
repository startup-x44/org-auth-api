'use client';

import { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { AuthService, User } from '../lib/auth';
import { initializeCSRFToken } from '../lib/csrf';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<{ success: boolean; needsOrgSelection?: boolean; organizations?: any[]; user?: User }>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    try {
      console.log('🔄 Refreshing user data...');
      const currentUser = await AuthService.getCurrentUser();
      console.log('👤 User data received:', currentUser);
      
      if (currentUser) {
        setUser(currentUser);
      } else {
        // Fallback: Try to get user data from token
        console.log('🔄 API failed, trying token fallback...');
        const token = AuthService.getAccessToken();
        if (token && !AuthService.isTokenExpired(token)) {
          const tokenUser = AuthService.getUserFromToken(token);
          if (tokenUser) {
            console.log('👤 Using token user data:', tokenUser);
            setUser(tokenUser as User);
          } else {
            setUser(null);
          }
        } else {
          setUser(null);
        }
      }
    } catch (error) {
      console.error('❌ Failed to refresh user:', error);
      
      // Check if it's an auth error (401, etc.)
      if (error instanceof Error && (
        error.message.toLowerCase().includes('unauthorized') ||
        error.message.toLowerCase().includes('token')
      )) {
        console.log('🔒 Auth error detected, clearing user state');
        setUser(null);
        return;
      }
      
      // Fallback: Try to get user data from token
      console.log('🔄 Error occurred, trying token fallback...');
      const token = AuthService.getAccessToken();
      if (token && !AuthService.isTokenExpired(token)) {
        const tokenUser = AuthService.getUserFromToken(token);
        if (tokenUser) {
          console.log('👤 Using token user data as fallback:', tokenUser);
          setUser(tokenUser as User);
        } else {
          setUser(null);
        }
      } else {
        setUser(null);
      }
    }
  }, []);

  const login = useCallback(async (email: string, password: string): Promise<{ success: boolean; needsOrgSelection?: boolean; organizations?: any[]; user?: User }> => {
    try {
      console.log('AuthContext: Calling AuthService.login');
      const result = await AuthService.login(email, password);
      console.log('AuthContext: AuthService.login result:', result);
      if (result) {
        setUser(result.user);
        return {
          success: true,
          needsOrgSelection: result.needsOrgSelection,
          organizations: result.organizations,
          user: result.user
        };
      }
      console.log('AuthContext: No result from AuthService, returning false');
      return { success: false };
    } catch (error) {
      console.error('AuthContext: Login failed with error:', error);
      console.log('AuthContext: Re-throwing error');
      // Re-throw the error so the calling component can handle the specific message
      throw error;
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await AuthService.logout();
    } finally {
      setUser(null);
    }
  }, []);

  useEffect(() => {
    const initAuth = async () => {
      try {
        console.log('🔄 Initializing auth...');
        
        // Initialize CSRF token first
        console.log('🛡️ Initializing CSRF token...');
        await initializeCSRFToken();
        
        const token = AuthService.getAccessToken();
        console.log('🔑 Token found:', !!token);
        
        if (token) {
          const isExpired = AuthService.isTokenExpired(token);
          console.log('⏰ Token expired:', isExpired);
          
          if (!isExpired) {
            console.log('✅ Token valid, refreshing user...');
            await refreshUser();
          } else {
            console.log('❌ Token expired, clearing...');
            AuthService.clearTokens();
            setUser(null);
          }
        } else {
          console.log('❌ No token found');
          setUser(null);
        }
      } catch (error) {
        console.error('❌ Auth initialization failed:', error);
        setUser(null);
      } finally {
        console.log('✅ Auth initialization complete');
        setLoading(false);
      }
    };

    initAuth();
  }, [refreshUser]);

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}