export interface User {
  id: string;
  email: string;
  firstname?: string;
  lastname?: string;
  address?: string;
  phone?: string;
  is_superadmin: boolean;
  global_role: GlobalRole;
  status: 'active' | 'suspended' | 'deactivated';
  last_login_at?: string;
  created_at: string;
  updated_at: string;
  organizations?: Organization[];
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description?: string;
  settings: Record<string, any>;
  status: 'active' | 'suspended' | 'archived';
  created_by: string;
  created_at: string;
  updated_at: string;
  creator?: User;
}

export interface OrganizationMembership {
  id: string;
  organization_id: string;
  user_id: string;
  role_id: string;
  status: 'active' | 'invited' | 'pending' | 'suspended';
  invited_by?: string;
  invited_at?: string;
  joined_at?: string;
  last_activity_at?: string;
  created_at: string;
  updated_at: string;
  organization?: Organization;
  user?: User;
  role?: Role;
  inviter?: User;
}

export interface Role {
  id: string;
  name: string;
  display_name: string;
  description?: string;
  is_system: boolean;
  permissions?: Permission[];
}

export interface Permission {
  id: string;
  name: string;
  display_name: string;
  description?: string;
  resource: string;
  action: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface JWTPayload {
  sub: string;
  email: string;
  role: string;
  org_id?: string;
  exp: number;
  iat: number;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
  };
}

export interface SearchParams {
  query?: string;
  page?: number;
  limit?: number;
  sort?: string;
  order?: 'asc' | 'desc';
}

// Simplified role structure
export type GlobalRole = 'superadmin' | 'user';
export type OrganizationRole = 'admin' | 'issuer' | 'rto' | 'student';
export type UserStatus = 'active' | 'suspended' | 'deactivated';
export type OrganizationStatus = 'active' | 'suspended' | 'archived';
export type MembershipStatus = 'active' | 'invited' | 'pending' | 'suspended';