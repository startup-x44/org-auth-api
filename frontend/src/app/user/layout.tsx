'use client';

import { useAuth } from '../../contexts/auth-context';
import { useRouter, usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';
import { AuthService } from '../../lib/auth';
import { useToast } from '../../hooks/use-toast';
import { Sidebar } from '../../components/dashboard/sidebar';
import { DashboardHeader } from '../../components/dashboard/dashboard-header';
import { DashboardSkeleton } from '../../components/dashboard/dashboard-skeleton';

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

export default function UserLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const { user, logout, loading } = useAuth();
    const router = useRouter();
    const pathname = usePathname();
    const { toast } = useToast();

    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [organizations, setOrganizations] = useState<UserOrganization[]>([]);
    const [selectedOrganization, setSelectedOrganization] = useState<UserOrganization | null>(null);
    const [dashboardLoading, setDashboardLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [sidebarOpen, setSidebarOpen] = useState(false);
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

    // Owner privileges - check if user is owner of the SELECTED organization
    const hasOwnerPrivileges = selectedOrganization?.role === 'owner';

    useEffect(() => {
        if (!loading && !user) {
            router.push('/auth/login');
            return;
        }

        if (!loading && user && user.is_superadmin) {
            router.push('/superadmin/dashboard');
            return;
        }

        if (!loading && user) {
            loadDashboardData();
        }
    }, [user, loading, router]);

    // Watch for token changes (when user returns from org switching)
    useEffect(() => {
        const token = AuthService.getAccessToken();
        const tokenData = token ? AuthService.decodeToken(token) : null;
        const tokenOrgId = tokenData?.organization_id;

        // If token org doesn't match selected org, reload data
        if (tokenOrgId && selectedOrganization && tokenOrgId !== selectedOrganization.organization_id) {
            console.log('🔄 Token org changed! Reloading dashboard data...');
            console.log('Token org:', tokenOrgId, 'Selected org:', selectedOrganization.organization_id);
            loadDashboardData();
        }
    }, [pathname]); // Run when pathname changes (returning from select-org page)

    const loadDashboardData = async () => {
        try {
            setDashboardLoading(true);

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
                const orgs = orgsData.data || orgsData.organizations || [];
                setOrganizations(orgs);

                // ALWAYS sync selected organization with JWT token's organization_id
                const token = AuthService.getAccessToken();
                const tokenData = token ? AuthService.decodeToken(token) : null;
                const tokenOrgId = tokenData?.organization_id;

                console.log('🔍 Dashboard Data Load - Token organization_id:', tokenOrgId);
                console.log('📋 Dashboard Data Load - Available organizations:', orgs);
                console.log('🎯 Dashboard Data Load - Current selected org:', selectedOrganization);

                if (tokenOrgId && orgs.length > 0) {
                    // Find the organization that matches the token's org ID
                    const matchingOrg = orgs.find((org: UserOrganization) => org.organization_id === tokenOrgId);
                    
                    if (matchingOrg) {
                        // Only update if different from current selection
                        if (!selectedOrganization || selectedOrganization.organization_id !== matchingOrg.organization_id) {
                            console.log('✅ Setting selected org from token:', matchingOrg);
                            setSelectedOrganization(matchingOrg);
                        } else {
                            console.log('✅ Selected org already matches token');
                        }
                    } else {
                        console.warn('⚠️ Token org_id not found in user organizations list. Token may be stale.');
                        console.warn('Token org ID:', tokenOrgId);
                        console.warn('Available org IDs:', orgs.map((o: UserOrganization) => o.organization_id));
                        // Fallback to first org if token org not found
                        setSelectedOrganization(orgs[0]);
                    }
                } else if (orgs.length > 0 && !selectedOrganization) {
                    // No token org ID, use first organization
                    console.log('ℹ️ No token org ID, using first org:', orgs[0]);
                    setSelectedOrganization(orgs[0]);
                }
            }
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            toast({
                title: "Failed to load dashboard data",
                description: "Some information may not be up to date.",
                variant: "destructive",
            });
        } finally {
            setDashboardLoading(false);
        }
    };

    const handleRefresh = async () => {
        setRefreshing(true);
        await loadDashboardData();
        setRefreshing(false);
        toast({
            title: "Dashboard refreshed",
            description: "All data has been updated.",
        });
    };

    const handleLogout = async () => {
        await logout();
        router.push('/auth/login');
    };

    const handleOrganizationChange = (org: UserOrganization) => {
        // Organization switching is now handled via redirect to /auth/select-organization
        // This function is kept for backwards compatibility but does nothing
        console.log('Organization change requested, redirect handled by header component');
    };

    if (loading || dashboardLoading) {
        return <DashboardSkeleton />;
    }

    if (!user) {
        return null;
    }

    return (
        <div className="min-h-screen bg-background">
            {/* Sidebar */}
            <Sidebar
                hasOwnerPrivileges={hasOwnerPrivileges}
                organizationName={selectedOrganization?.organization_name}
                isOpen={sidebarOpen}
                onClose={() => setSidebarOpen(false)}
                collapsed={sidebarCollapsed}
                onCollapsedChange={setSidebarCollapsed}
            />

            {/* Main Content Area */}
            <div className={`transition-all duration-300 ${sidebarCollapsed ? 'lg:pl-16' : 'lg:pl-64'}`}>
                {/* Header */}
                <DashboardHeader
                    user={user}
                    profile={profile}
                    organizations={organizations}
                    selectedOrganization={selectedOrganization}
                    onOrganizationChange={handleOrganizationChange}
                    refreshing={refreshing}
                    onRefresh={handleRefresh}
                    onLogout={handleLogout}
                    onMenuToggle={() => setSidebarOpen(!sidebarOpen)}
                />

                {/* Page Content */}
                <main className="p-4 sm:p-6 lg:p-8">
                    {children}
                </main>
            </div>

            {/* Mobile Sidebar Overlay */}
            {sidebarOpen && (
                <div
                    className="fixed inset-0 bg-black/50 z-30 lg:hidden"
                    onClick={() => setSidebarOpen(false)}
                />
            )}
        </div>
    );
}
