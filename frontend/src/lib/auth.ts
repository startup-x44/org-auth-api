import { jwtDecode } from 'jwt-decode';
import { makeSecureRequest, clearCSRFToken } from './csrf';

export interface User {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  address?: string;
  phone?: string;
  is_superadmin: boolean;
  is_active: boolean;
  global_role: string;
  created_at: string;
  updated_at: string;
  last_login?: string;
  organizations?: Organization[];
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description?: string;
  status: 'active' | 'suspended' | 'archived';
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  session_id?: string;
  organization_id?: string;
}

export interface JWTPayload {
  sub: string;
  user_id: string;
  email: string;
  global_role: string;
  is_superadmin: boolean;
  permissions: string[];
  organization_id: string;
  session_id: string;
  role_id: string;
  exp: number;
  iat: number;
  iss: string;
  aud: string[];
}

export class AuthService {
  private static readonly ACCESS_TOKEN_KEY = 'access_token';
  private static readonly REFRESH_TOKEN_KEY = 'refresh_token';
  private static readonly API_BASE_URL = process.env.NODE_ENV === 'development' 
    ? 'http://localhost:8080' 
    : '';

  static getAccessToken(): string | null {
    if (typeof window === 'undefined') return null;
    const token = localStorage.getItem(this.ACCESS_TOKEN_KEY);
    console.log('🔍 Getting access token from localStorage:', token ? 'FOUND' : 'NOT FOUND');
    return token;
  }

  static getRefreshToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(this.REFRESH_TOKEN_KEY);
  }

  static setTokens(tokens: AuthTokens): void {
    if (typeof window === 'undefined') {
      console.log('❌ Cannot set tokens: window is undefined (SSR)');
      return;
    }
    
    console.log('💾 Setting tokens to localStorage:', {
      access_token: tokens.access_token ? 'EXISTS' : 'MISSING',
      refresh_token: tokens.refresh_token ? 'EXISTS' : 'MISSING'
    });
    
    localStorage.setItem(this.ACCESS_TOKEN_KEY, tokens.access_token);
    localStorage.setItem(this.REFRESH_TOKEN_KEY, tokens.refresh_token);
    
    // Verify storage worked
    const storedAccess = localStorage.getItem(this.ACCESS_TOKEN_KEY);
    const storedRefresh = localStorage.getItem(this.REFRESH_TOKEN_KEY);
    
    console.log('✅ Tokens stored successfully:', {
      access_stored: !!storedAccess,
      refresh_stored: !!storedRefresh
    });
  }

  static clearTokens(): void {
    if (typeof window === 'undefined') return;
    
    console.log('🗑️ CLEARING TOKENS from localStorage');
    console.trace('clearTokens called from:'); // This will show the call stack
    
    localStorage.removeItem(this.ACCESS_TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_TOKEN_KEY);
    
    console.log('🗑️ Tokens cleared from localStorage');
  }

  static decodeToken(token?: string): JWTPayload | null {
    const accessToken = token || this.getAccessToken();
    if (!accessToken) return null;
    
    try {
      return jwtDecode<JWTPayload>(accessToken);
    } catch {
      return null;
    }
  }

  static isTokenExpired(token?: string): boolean {
    const payload = this.decodeToken(token);
    if (!payload) return true;
    
    return Date.now() >= payload.exp * 1000;
  }

  static isSuperAdmin(token?: string): boolean {
    const payload = this.decodeToken(token);
    return payload?.is_superadmin === true || payload?.global_role === 'superadmin';
  }

  static getUserFromToken(token?: string): Partial<User> | null {
    const payload = this.decodeToken(token);
    if (!payload) return null;

    return {
      id: payload.sub,
      email: payload.email,
      is_superadmin: payload.is_superadmin,
      global_role: payload.global_role as 'user' | 'admin'
    };
  }

  static async apiCall(endpoint: string, options: RequestInit = {}): Promise<Response> {
    const token = this.getAccessToken();
    const method = options.method?.toUpperCase() || 'GET';
    
    // Check if this endpoint requires CSRF protection (not in skip list and not a safe method)
    // Only skip CSRF for login, register, and refresh endpoints
    const requiresCSRF = !['GET', 'HEAD', 'OPTIONS'].includes(method) && 
                        !endpoint.match(/\/(oauth\/token|auth\/(login|register|refresh))$/);
    
    console.log(`API Call: ${method} ${endpoint}, Token exists: ${!!token}, Token expired: ${token ? this.isTokenExpired(token) : 'no token'}, Requires CSRF: ${requiresCSRF}`);
    
    const shouldIncludeAuth = token && !this.isTokenExpired(token);
    console.log('🔑 Should include auth header:', shouldIncludeAuth);
    
    const headers = {
      'Content-Type': 'application/json',
      ...(shouldIncludeAuth && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    };
    
    console.log('📋 Final headers being sent:', JSON.stringify(headers, null, 2));

    let response: Response;

    if (requiresCSRF) {
      // Use secure request with CSRF token
      console.log('Making secure request with CSRF token');
      response = await makeSecureRequest(`${this.API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });
    } else {
      // Regular fetch for endpoints that don't need CSRF
      console.log('Making regular fetch request, headers:', headers);
      response = await fetch(`${this.API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });
    }
    
    console.log(`Response status: ${response.status} for ${method} ${endpoint}`);

    // If token expired or unauthorized, try to refresh first
    if (response.status === 401) {
      if (token) {
        // Try to refresh token first
        const refreshed = await this.refreshToken();
        if (refreshed) {
          // Retry with new token
          const newHeaders = {
            ...headers,
            Authorization: `Bearer ${this.getAccessToken()}`,
          };
          
          if (requiresCSRF) {
            return makeSecureRequest(`${this.API_BASE_URL}${endpoint}`, {
              ...options,
              headers: newHeaders,
            });
          } else {
            return fetch(`${this.API_BASE_URL}${endpoint}`, {
              ...options,
              headers: newHeaders,
            });
          }
        }
      }
      
      // Refresh failed or no token, clear auth state but DON'T redirect (for debugging)
      console.log('401 Unauthorized - NOT redirecting (debugging mode)');
      console.log('Token exists:', !!token);
      console.log('Refresh token exists:', !!this.getRefreshToken());
      // this.clearTokens();
      // clearCSRFToken();
      // if (typeof window !== 'undefined') {
      //   console.log('401 Unauthorized - Redirecting to login');
      //   window.location.href = '/auth/login?reason=unauthorized';
      // }
    }

    return response;
  }

  static async login(email: string, password: string): Promise<{ user: User; tokens: AuthTokens; organizations?: any[]; needsOrgSelection?: boolean } | null> {
    try {
      console.log('AuthService.login called with:', email);
      const response = await fetch(`${this.API_BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      console.log('Login response status:', response.status);
      
      if (!response.ok) {
        let errorMessage = 'Login failed. Please try again.';
        
        try {
          const errorData = await response.json();
          console.error('Login failed with error data:', errorData);
          
          // Handle specific error messages from backend
          if (errorData.message) {
            errorMessage = errorData.message;
          } else if (errorData.error) {
            errorMessage = errorData.error;
          }
        } catch (parseError) {
          // If JSON parsing fails, try to get text
          try {
            const errorText = await response.text();
            console.error('Login failed with error text:', errorText);
            if (errorText.trim()) {
              errorMessage = errorText;
            }
          } catch (textError) {
            console.error('Failed to parse error response');
          }
        }
        
        throw new Error(errorMessage);
      }

      const data = await response.json();
      console.log('Login response data:', JSON.stringify(data, null, 2));
      console.log('Response structure analysis:');
      console.log('- data.success:', data.success);
      console.log('- data.data exists:', !!data.data);
      console.log('- data.data.user exists:', !!(data.data && data.data.user));
      console.log('- data.data.token exists:', !!(data.data && data.data.token));
      console.log('- data.data.organizations exists:', !!(data.data && data.data.organizations));
      
      if (data.data && data.data.token) {
        console.log('Token structure:', JSON.stringify(data.data.token, null, 2));
      }
      
      // Backend returns: { data: { user: {...}, token: {...}, organizations: [...] }, message: "...", success: true }
      if (data.data && data.data.user) {
        // Check if user needs organization selection
        const organizations = data.data.organizations || [];
        const needsOrgSelection = organizations.length === 0 || data.message?.includes('select an organization');
        
        // Set tokens if provided (only for superadmin users)
        if (data.data.token) {
          console.log('✅ Login included token (superadmin user)');
          console.log('Setting tokens:', data.data.token);
          this.setTokens(data.data.token);
          
          // Verify token was stored
          const storedToken = this.getAccessToken();
          console.log('Stored token:', storedToken);
          
          // Decode and check token
          const payload = storedToken ? this.decodeToken(storedToken) : null;
          console.log('Token payload:', payload);
        } else {
          console.log('ℹ️ No token in login response (regular user - will get token after org selection)');
        }
        
        // Return login result with organization info
        return {
          user: data.data.user,
          tokens: data.data.token || null,
          organizations,
          needsOrgSelection
        };
      }
      
      return null;
    } catch (error) {
      console.error('Login error:', error);
      // Re-throw the error so the calling code can handle it properly
      throw error;
    }
  }

  static async refreshToken(): Promise<boolean> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) return false;

    try {
      const response = await fetch(`${this.API_BASE_URL}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        this.clearTokens();
        return false;
      }

      const data = await response.json();
      this.setTokens(data.tokens);
      return true;
    } catch {
      this.clearTokens();
      return false;
    }
  }

  static async switchOrganization(userId: string, organizationId: string): Promise<{ success: boolean; user?: User; organization?: any; tokens?: AuthTokens }> {
    try {
      console.log('Switching organization:', { userId, organizationId });
      
      const response = await this.apiCall('/api/v1/auth/select-organization', {
        method: 'POST',
        body: JSON.stringify({
          user_id: userId,
          organization_id: organizationId
        }),
      });

      if (!response.ok) {
        console.error('Failed to switch organization:', response.status);
        return { success: false };
      }

      const data = await response.json();
      console.log('Switch organization response:', data);

      // Update tokens with new org-scoped token
      if (data.data && data.data.token) {
        console.log('Setting new org-scoped tokens');
        this.setTokens(data.data.token);
      }

      return {
        success: true,
        user: data.data?.user,
        organization: data.data?.organization,
        tokens: data.data?.token
      };
    } catch (error) {
      console.error('Error switching organization:', error);
      return { success: false };
    }
  }

  static async logout(): Promise<void> {
    try {
      await this.apiCall('/api/v1/user/logout', { method: 'POST' });
    } catch {
      // Ignore errors on logout
    } finally {
      this.clearTokens();
      clearCSRFToken();
      if (typeof window !== 'undefined') {
        window.location.href = '/auth/login';
      }
    }
  }

  static async getCurrentUser(): Promise<User | null> {
    const token = this.getAccessToken();
    console.log('📡 getCurrentUser - Token exists:', !!token);
    
    if (!token || this.isTokenExpired(token)) {
      console.log('❌ getCurrentUser - No token or expired');
      return null;
    }

    try {
      // Use appropriate endpoint based on role
      const isSuperAdmin = this.isSuperAdmin(token);
      const endpoint = isSuperAdmin
        ? '/api/v1/admin/profile' 
        : '/api/v1/user/profile';
      
      console.log('📡 getCurrentUser - Calling endpoint:', endpoint, 'isSuperAdmin:', isSuperAdmin);

      const response = await this.apiCall(endpoint);
      console.log('📡 getCurrentUser - Response status:', response.status);
      
      if (!response.ok) {
        console.log('❌ getCurrentUser - Response not ok:', response.status);
        return null;
      }

      const data = await response.json();
      console.log('📡 getCurrentUser - Response data:', data);
      
      // Handle different response formats
      const user = data.data?.user || data.user || data;
      console.log('👤 getCurrentUser - Parsed user:', user);
      
      return user;
    } catch (error) {
      console.error('❌ getCurrentUser - Error:', error);
      return null;
    }
  }
}