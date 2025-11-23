'use client';

import { useAuth } from '../../../contexts/auth-context';
import { useState, useEffect } from 'react';
import { AuthService } from '../../../lib/auth';
import { QuickStats } from '../../../components/dashboard/quick-stats';

interface UserProfile {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  phone?: string;
  address?: string;
  global_role: string;
  is_superadmin: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login?: string;
}

interface UserOrganization {
  organization_id: string;
  organization_name: string;
  organization_slug: string;
  role: string;
  status: string;
  joined_at: string;
}

export default function DashboardPage() {
  const { user } = useAuth();

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [organizations, setOrganizations] = useState<UserOrganization[]>([]);
  const [loading, setLoading] = useState(true);

  // Owner privileges
  const ownedOrganizations = organizations.filter(org => org.role === 'owner');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [profileResponse, orgsResponse] = await Promise.allSettled([
        AuthService.apiCall('/api/v1/user/profile'),
        AuthService.apiCall('/api/v1/user/organizations')
      ]);

      if (profileResponse.status === 'fulfilled' && profileResponse.value.ok) {
        const profileData = await profileResponse.value.json();
        setProfile(profileData.data || profileData.user || profileData);
      }

      if (orgsResponse.status === 'fulfilled' && orgsResponse.value.ok) {
        const orgsData = await orgsResponse.value.json();
        setOrganizations(orgsData.data || orgsData.organizations || []);
      }
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 bg-muted rounded animate-pulse"></div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-32 bg-muted rounded-lg animate-pulse"></div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>
        <p className="text-sm text-muted-foreground mt-1">Welcome back! Here's your overview.</p>
      </div>

      {/* Quick Stats */}
      <QuickStats
        isActive={profile?.is_active ?? user?.is_active ?? false}
        organizationsCount={organizations.length}
        ownedOrganizationsCount={ownedOrganizations.length}
        daysMember={Math.floor((Date.now() - new Date(profile?.created_at || user?.created_at || Date.now()).getTime()) / (1000 * 60 * 60 * 24))}
      />
    </div>
  );
}
