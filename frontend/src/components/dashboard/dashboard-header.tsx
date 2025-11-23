'use client';

import { RefreshCw, LogOut, Building2, ChevronDown, Menu } from 'lucide-react';
import { Button } from '../ui/button';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

interface Organization {
    organization_id: string;
    organization_name: string;
    organization_slug: string;
    role: string;
    status: string;
    joined_at: string;
}

interface DashboardHeaderProps {
    user: {
        first_name?: string;
        last_name?: string;
        email: string;
    };
    profile: {
        first_name?: string;
        last_name?: string;
        email: string;
    } | null;
    organizations: Organization[];
    selectedOrganization: Organization | null;
    onOrganizationChange: (org: Organization) => void;
    refreshing: boolean;
    onRefresh: () => void;
    onLogout: () => void;
    onMenuToggle?: () => void;
}

export function DashboardHeader({
    user,
    profile,
    organizations,
    selectedOrganization,
    onOrganizationChange,
    refreshing,
    onRefresh,
    onLogout,
    onMenuToggle,
}: DashboardHeaderProps) {
    const [orgDropdownOpen, setOrgDropdownOpen] = useState(false);
    const router = useRouter();

    const handleSwitchOrganization = () => {
        setOrgDropdownOpen(false);
        router.push('/auth/select-organization');
    };

    return (
        <nav className="bg-background shadow-sm border-b border-border sticky top-0 z-50">
            <div className="px-4 sm:px-6 lg:px-8">
                <div className="flex justify-between h-16">
                    <div className="flex items-center space-x-4">
                        {/* Mobile menu toggle */}
                        {onMenuToggle && (
                            <button
                                onClick={onMenuToggle}
                                className="lg:hidden p-2 hover:bg-muted rounded-lg transition-colors"
                            >
                                <Menu className="h-5 w-5 text-muted-foreground" />
                            </button>
                        )}

                        {/* Organization Selector */}
                        <div className="relative">
                            <button
                                onClick={() => setOrgDropdownOpen(!orgDropdownOpen)}
                                className="flex items-center space-x-2 px-3 py-2 rounded-lg hover:bg-muted/50 transition-colors border border-border"
                            >
                                <div className="p-1.5 bg-primary/10 rounded-md">
                                    <Building2 className="h-4 w-4 text-primary" />
                                </div>
                                <div className="hidden sm:block text-left">
                                    <div className="text-sm font-semibold text-foreground">
                                        {selectedOrganization?.organization_name || 'Select Organization'}
                                    </div>
                                    {selectedOrganization && (
                                        <div className="text-xs text-muted-foreground">
                                            {selectedOrganization.role}
                                        </div>
                                    )}
                                </div>
                                <ChevronDown className={`h-4 w-4 text-muted-foreground transition-transform ${orgDropdownOpen ? 'rotate-180' : ''
                                    }`} />
                            </button>

                            {/* Dropdown */}
                            {orgDropdownOpen && (
                                <>
                                    <div
                                        className="fixed inset-0 z-10"
                                        onClick={() => setOrgDropdownOpen(false)}
                                    />
                                    <div className="absolute left-0 mt-2 w-72 bg-popover rounded-lg shadow-lg border border-border py-2 z-20">
                                        <div className="px-3 py-2 border-b border-border">
                                            <p className="text-xs font-semibold text-muted-foreground uppercase">
                                                Current Organization
                                            </p>
                                        </div>
                                        <div className="px-3 py-3">
                                            {selectedOrganization ? (
                                                <div className="flex items-center space-x-3 mb-3">
                                                    <div className="p-2 rounded-md bg-primary/10">
                                                        <Building2 className="h-4 w-4 text-primary" />
                                                    </div>
                                                    <div className="flex-1 min-w-0">
                                                        <div className="text-sm font-medium text-foreground truncate">
                                                            {selectedOrganization.organization_name}
                                                        </div>
                                                        <div className="text-xs text-muted-foreground">
                                                            @{selectedOrganization.organization_slug} • {selectedOrganization.role}
                                                        </div>
                                                    </div>
                                                </div>
                                            ) : (
                                                <div className="text-sm text-muted-foreground mb-3">
                                                    No organization selected
                                                </div>
                                            )}
                                            <Button
                                                onClick={handleSwitchOrganization}
                                                variant="outline"
                                                size="sm"
                                                className="w-full"
                                            >
                                                Switch Organization
                                            </Button>
                                        </div>
                                    </div>
                                </>
                            )}
                        </div>
                    </div>

                    <div className="flex items-center space-x-2 sm:space-x-3">
                        <Button
                            onClick={onRefresh}
                            variant="ghost"
                            size="sm"
                            disabled={refreshing}
                            className="flex items-center hover:bg-muted rounded-lg transition-colors"
                        >
                            <RefreshCw className={`h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
                            <span className="hidden sm:inline ml-2">Refresh</span>
                        </Button>

                        <div className="hidden md:flex items-center px-3 py-1.5 bg-muted/50 rounded-lg border border-border">
                            <div className="text-right">
                                <p className="text-sm font-semibold text-foreground">
                                    {profile?.first_name || user.first_name} {profile?.last_name || user.last_name}
                                </p>
                                <p className="text-xs text-muted-foreground">{profile?.email || user.email}</p>
                            </div>
                        </div>

                        <Button
                            onClick={onLogout}
                            variant="outline"
                            size="sm"
                            className="flex items-center border-border hover:bg-destructive/10 hover:text-destructive hover:border-destructive/20 transition-all duration-200"
                        >
                            <LogOut className="h-4 w-4" />
                            <span className="hidden sm:inline ml-2">Logout</span>
                        </Button>
                    </div>
                </div>
            </div>
        </nav>
    );
}
