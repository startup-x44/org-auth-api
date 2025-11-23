'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../../contexts/auth-context';
import { AuthService } from '../../../lib/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { useToast } from '../../../hooks/use-toast';
import { useAuthError } from '../../../hooks/use-auth-error';
import {
  Building2,
  Plus,
  Users,
  AlertCircle,
  CheckCircle,
  ArrowRight,
  User
} from 'lucide-react';
import { Alert, AlertDescription } from '../../../components/ui/alert';

interface Organization {
  id?: string;
  name?: string;
  slug?: string;
  description?: string;
  status: 'active' | 'suspended' | 'archived';
  created_by?: string;
  created_at?: string;
  updated_at?: string;
  // Backend organization membership structure
  organization_id: string;
  organization_name: string;
  organization_slug: string;
  role: string;
  joined_at: string;
}

export default function SelectOrganizationPage() {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState('');
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newOrgName, setNewOrgName] = useState('');
  const [newOrgDescription, setNewOrgDescription] = useState('');

  const router = useRouter();
  const { user, refreshUser } = useAuth();
  const { toast } = useToast();
  const { checkResponseForAuth } = useAuthError();

  // Redirect if not logged in
  useEffect(() => {
    // Check if we have a valid token first (for users switching from dashboard)
    const token = AuthService.getAccessToken();
    const hasValidToken = token && !AuthService.isTokenExpired(token);
    
    // If we have a valid token, allow access even if user context hasn't loaded yet
    if (!user && !hasValidToken) {
      router.push('/auth/login');
      return;
    }

    // If superadmin, go directly to dashboard
    if (user && user.is_superadmin) {
      router.push('/superadmin/dashboard');
      return;
    }

    // Load user's organizations (works with or without user context if token is valid)
    loadOrganizations();
  }, [user, router]);

  const loadOrganizations = async () => {
    try {
      setLoading(true);

      // Debug: Check current token status
      const token = AuthService.getAccessToken();
      console.log('🔑 Current access token:', token ? 'EXISTS' : 'NOT FOUND');
      console.log('🔑 Token value (first 20 chars):', token ? token.substring(0, 20) + '...' : 'none');
      console.log('🔑 Token expired:', token ? AuthService.isTokenExpired(token) : 'no token');

      // Try to get organizations from session storage first (from login response)
      const storedOrgs = sessionStorage.getItem('user_organizations');
      console.log('📦 Organizations from session storage:', storedOrgs);

      if (storedOrgs && storedOrgs !== 'undefined') {
        try {
          const orgs = JSON.parse(storedOrgs);
          console.log('📋 Parsed organizations from session:', orgs);
          setOrganizations(orgs || []);

          // If user has organizations, show them
          if (orgs && orgs.length > 0) {
            console.log('✅ User has organizations, showing selection UI');
            setShowCreateForm(false);
          } else {
            console.log('ℹ️ No organizations found, showing create form');
            setShowCreateForm(true);
          }

          // Keep the stored data for now (don't clear) in case we need to debug
          // sessionStorage.removeItem('user_organizations');

          // Since we have organizations from login, don't make API call
          return;
        } catch (parseError) {
          console.error('Failed to parse stored organizations:', parseError);
          console.log('📋 Invalid session data, showing create form');
          // Fall through to show create form
        }
      } else {
        // No stored organizations, check if we have a token to make API call
        const token = AuthService.getAccessToken();
        console.log('No stored organizations. Current token exists:', !!token);
        console.log('Token expired:', token ? AuthService.isTokenExpired(token) : 'no token');

        if (token && !AuthService.isTokenExpired(token)) {
          // We have a valid token, try API call
          try {
            console.log('Making API call to fetch organizations...');
            const response = await AuthService.apiCall('/api/v1/user/organizations');
            console.log('Organizations API response status:', response.status);

            if (response.ok) {
              const data = await response.json();
              console.log('Organizations API response data:', data);
              setOrganizations(data.data || []);

              // If user has organizations, show them
              if (data.data && data.data.length > 0) {
                setShowCreateForm(false);
              } else {
                // No organizations, show create form by default
                setShowCreateForm(true);
              }
            } else {
              // Check for 401 and handle redirect
              if (await checkResponseForAuth(response, 'Session expired. Please login to view organizations.')) {
                return;
              }

              // Other error, assume no organizations and show create form
              console.log('API call failed with status:', response.status);
              setOrganizations([]);
              setShowCreateForm(true);
            }
          } catch (err) {
            console.log('API call error, showing create form:', err);
            // API call failed, assume no organizations and show create form
            setOrganizations([]);
            setShowCreateForm(true);
          }
        } else {
          // No token or expired token - this is expected for regular users who haven't selected org yet
          console.log('No valid token - this is expected for users who need to create/select organization');
          setOrganizations([]);
          setShowCreateForm(true);
        }
      }
    } catch (err) {
      console.error('Failed to load organizations:', err);
      // Default to create form
      setOrganizations([]);
      setShowCreateForm(true);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectOrganization = async (orgId: string) => {
    try {
      setActionLoading(true);
      setError('');

      // Get user ID from multiple sources
      let userId: string | undefined;
      
      // 1. Try from user context (when switching from dashboard)
      if (user && user.id) {
        userId = user.id;
        console.log('Using user ID from context:', userId);
      } 
      // 2. Try from JWT token (when context hasn't loaded yet)
      else {
        const token = AuthService.getAccessToken();
        const tokenData = token ? AuthService.decodeToken(token) : null;
        if (tokenData && tokenData.user_id) {
          userId = tokenData.user_id;
          console.log('Using user ID from JWT token:', userId);
        }
        // 3. Fallback to session storage (from login)
        else {
          const storedUser = sessionStorage.getItem('pending_user');
          console.log('Stored user data for select org:', storedUser);

          if (!storedUser) {
            setError('User session expired. Please login again.');
            setTimeout(() => router.push('/auth/login'), 2000);
            return;
          }

          const userData = JSON.parse(storedUser);
          userId = userData.id;
          console.log('Using user ID from session storage:', userId);
        }
      }

      if (!userId) {
        setError('Unable to identify user. Please login again.');
        setTimeout(() => router.push('/auth/login'), 2000);
        return;
      }

      const requestBody = {
        user_id: userId,
        organization_id: orgId
      };
      console.log('Selecting organization with request:', requestBody);

      const response = await AuthService.apiCall('/api/v1/auth/select-organization', {
        method: 'POST',
        body: JSON.stringify(requestBody),
      });

      // Check for 401 and handle redirect
      if (await checkResponseForAuth(response, 'Session expired. Please login to select an organization.')) {
        return;
      }

      if (response.ok) {
        const data = await response.json();

        // Update tokens if provided
        if (data.data && data.data.token) {
          AuthService.setTokens(data.data.token);
        }

        // Clear stored user data
        sessionStorage.removeItem('pending_user');
        sessionStorage.removeItem('user_organizations');

        toast({
          title: "Organization selected!",
          description: "Redirecting to your dashboard...",
          variant: "success",
        });

        // Refresh user data and redirect
        await refreshUser();

        setTimeout(() => {
          router.push('/user/dashboard');
        }, 1000);
      } else {
        const errorData = await response.json();
        setError(errorData.message || 'Failed to select organization');
      }
    } catch (err) {
      console.error('Failed to select organization:', err);
      setError('Failed to select organization');
    } finally {
      setActionLoading(false);
    }
  };

  const handleCreateOrganization = async () => {
    if (!newOrgName.trim()) {
      setError('Organization name is required');
      return;
    }

    try {
      setActionLoading(true);
      setError('');

      // Get user ID from multiple sources
      let userId: string | undefined;
      
      // 1. Try from user context
      if (user && user.id) {
        userId = user.id;
        console.log('Using user ID from context for create org:', userId);
      } 
      // 2. Try from JWT token
      else {
        const token = AuthService.getAccessToken();
        const tokenData = token ? AuthService.decodeToken(token) : null;
        if (tokenData && tokenData.user_id) {
          userId = tokenData.user_id;
          console.log('Using user ID from JWT token for create org:', userId);
        }
        // 3. Fallback to session storage
        else {
          const storedUser = sessionStorage.getItem('pending_user');
          console.log('Stored user data for create org:', storedUser);

          if (!storedUser) {
            setError('User session expired. Please login again.');
            setTimeout(() => router.push('/auth/login'), 2000);
            return;
          }

          const userData = JSON.parse(storedUser);
          userId = userData.id;
          console.log('Using user ID from session storage for create org:', userId);
        }
      }

      if (!userId) {
        setError('Unable to identify user. Please login again.');
        setTimeout(() => router.push('/auth/login'), 2000);
        return;
      }

      // Generate slug from organization name
      const generateSlug = (name: string): string => {
        return name
          .toLowerCase()
          .trim()
          .replace(/[^a-z0-9\s-]/g, '') // Remove special characters
          .replace(/\s+/g, '-') // Replace spaces with hyphens
          .replace(/-+/g, '-') // Replace multiple hyphens with single hyphen
          .replace(/^-|-$/g, ''); // Remove leading/trailing hyphens
      };

      const orgSlug = generateSlug(newOrgName.trim());
      console.log('Generated slug:', orgSlug, 'from name:', newOrgName.trim());

      const requestBody = {
        user_id: userId,
        name: newOrgName.trim(),
        slug: orgSlug,
        ...(newOrgDescription.trim() && { description: newOrgDescription.trim() })
      };

      console.log('Creating organization with request:', requestBody);

      const response = await AuthService.apiCall('/api/v1/auth/create-organization', {
        method: 'POST',
        body: JSON.stringify(requestBody),
      });

      console.log('Create organization response status:', response.status);

      // Check for 401 and handle redirect
      if (await checkResponseForAuth(response, 'Session expired. Please login to create an organization.')) {
        return;
      }

      if (response.ok) {
        const data = await response.json();
        console.log('Create organization success response:', data);

        // Update tokens if provided
        if (data.data && data.data.token) {
          AuthService.setTokens(data.data.token);
        }

        // Clear stored user data
        sessionStorage.removeItem('pending_user');
        sessionStorage.removeItem('user_organizations');

        toast({
          title: "Organization created!",
          description: "Welcome to your new organization. Redirecting to dashboard...",
          variant: "success",
        });

        // Refresh user data and redirect
        await refreshUser();

        setTimeout(() => {
          router.push('/user/dashboard');
        }, 1000);
      } else {
        const errorData = await response.json();
        console.error('Create organization error response:', errorData);
        setError(errorData.message || 'Failed to create organization');
      }
    } catch (err) {
      console.error('Failed to create organization:', err);
      setError('Failed to create organization');
    } finally {
      setActionLoading(false);
    }
  };

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="text-center">
          <LoadingSpinner size="lg" variant="primary" />
          <p className="mt-4 text-gray-600">Redirecting to login...</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="text-center">
          <LoadingSpinner size="lg" variant="primary" />
          <p className="mt-4 text-gray-600">Loading organizations...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-2xl mx-auto">
        <Card className="p-8 shadow-xl border-0 bg-white backdrop-blur">
          {/* Header */}
          <div className="text-center mb-8">
            <div className="flex justify-center mb-4">
              <div className="p-3 bg-primary/10 rounded-full">
                <Building2 className="h-8 w-8 text-primary" />
              </div>
            </div>
            <h1 className="text-2xl font-bold text-foreground mb-2">
              Welcome, {user.first_name || user.email}!
            </h1>
            <p className="text-gray-600">
              {organizations.length > 0
                ? "Select an organization to continue"
                : "Create your first organization to get started"
              }
            </p>
          </div>

          {/* Error Alert */}
          {error && (
            <div className="mb-6">
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  {error}
                </AlertDescription>
              </Alert>
            </div>
          )}

          {/* Existing Organizations */}
          {organizations.length > 0 && !showCreateForm && (
            <div className="space-y-4 mb-8">
              <h2 className="text-lg font-semibold text-gray-900 flex items-center">
                <Users className="h-5 w-5 mr-2" />
                Your Organizations
              </h2>

              <div className="space-y-3">
                {organizations.map((org) => (
                  <div
                    key={org.organization_id || org.id || `org-${org.organization_slug}`}
                    className="p-4 border border-gray-200 rounded-lg hover:border-primary/50 hover:bg-gray-50 transition-colors"
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="font-medium text-gray-900">{org.organization_name}</h3>
                        <p className="text-sm text-gray-600 mt-1">@{org.organization_slug}</p>
                        <div className="flex items-center mt-2 space-x-2">
                          <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${org.status === 'active'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                            }`}>
                            {org.status}
                          </span>
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                            {org.role}
                          </span>
                        </div>
                      </div>
                      <Button
                        onClick={() => handleSelectOrganization(org.organization_id)}
                        disabled={actionLoading}
                        className="ml-4"
                      >
                        <span>Select</span>
                        <ArrowRight className="ml-2 h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Create Organization Form */}
          {showCreateForm && (
            <div className="space-y-6 mb-8">
              <h2 className="text-lg font-semibold text-gray-900 flex items-center">
                <Plus className="h-5 w-5 mr-2" />
                Create New Organization
              </h2>

              <div className="space-y-4">
                <div>
                  <Label htmlFor="orgName" className="mb-2">
                    Organization Name *
                  </Label>
                  <Input
                    id="orgName"
                    type="text"
                    value={newOrgName}
                    onChange={(e) => setNewOrgName(e.target.value)}
                    placeholder="Enter organization name"
                    disabled={actionLoading}
                  />
                </div>

                <div>
                  <Label htmlFor="orgDescription" className="mb-2">
                    Description (Optional)
                  </Label>
                  <Input
                    id="orgDescription"
                    type="text"
                    value={newOrgDescription}
                    onChange={(e) => setNewOrgDescription(e.target.value)}
                    placeholder="Brief description of your organization"
                    disabled={actionLoading}
                  />
                </div>

                <Button
                  onClick={handleCreateOrganization}
                  disabled={actionLoading || !newOrgName.trim()}
                  className="w-full"
                >
                  {actionLoading ? (
                    <div className="flex items-center">
                      <LoadingSpinner size="sm" variant="secondary" />
                      <span className="ml-2">Creating...</span>
                    </div>
                  ) : (
                    <div className="flex items-center justify-center">
                      <Plus className="mr-2 h-5 w-5" />
                      <span>Create Organization</span>
                    </div>
                  )}
                </Button>
              </div>
            </div>
          )}

          {/* Toggle between views */}
          {organizations.length > 0 && (
            <div className="border-t border-gray-200 pt-6">
              <div className="text-center">
                {showCreateForm ? (
                  <Button
                    variant="outline"
                    onClick={() => setShowCreateForm(false)}
                    disabled={actionLoading}
                  >
                    <Users className="mr-2 h-4 w-4" />
                    View My Organizations
                  </Button>
                ) : (
                  <Button
                    variant="outline"
                    onClick={() => setShowCreateForm(true)}
                    disabled={actionLoading}
                  >
                    <Plus className="mr-2 h-4 w-4" />
                    Create New Organization
                  </Button>
                )}
              </div>
            </div>
          )}

          {/* Footer */}
          <div className="mt-8 pt-6 border-t border-gray-200 text-center">
            <p className="text-xs text-gray-500">
              Organizations help you manage teams and permissions in NILOAUTH
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
}