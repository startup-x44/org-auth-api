import { Card } from '../ui/card';
import { Button } from '../ui/button';
import { Building2, Crown, UserPlus, Key, Settings, Users, Eye } from 'lucide-react';

interface Organization {
    organization_id: string;
    organization_name: string;
    organization_slug: string;
    role: string;
    status: string;
    joined_at: string;
}

interface OrganizationsListProps {
    organizations: Organization[];
    onInvite: (org: Organization) => void;
    onManageRoles: (org: Organization) => void;
    onSettings: (org: Organization) => void;
    onCreateOrg: () => void;
}

export function OrganizationsList({
    organizations,
    onInvite,
    onManageRoles,
    onSettings,
    onCreateOrg,
}: OrganizationsListProps) {
    return (
        <Card className="p-6 shadow-md border-none bg-card/80 backdrop-blur-sm">
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-3">
                    <div className="p-2 bg-primary/10 rounded-lg">
                        <Building2 className="h-5 w-5 text-primary" />
                    </div>
                    <h3 className="text-lg font-bold text-foreground">My Organizations</h3>
                </div>
                <Button variant="outline" size="sm" onClick={onCreateOrg}>
                    Create New
                </Button>
            </div>

            <div className="space-y-4">
                {organizations.length > 0 ? (
                    organizations.map((org) => (
                        <div key={org.organization_id} className={`group p-4 border rounded-xl transition-all duration-200 hover:shadow-md ${org.role === 'owner' ? 'border-accent/30 bg-gradient-to-r from-accent/10 to-transparent' : 'border-border hover:border-primary/20 bg-card'
                            }`}>
                            <div className="flex flex-col sm:flex-row sm:items-start gap-4">
                                <div className={`p-3 rounded-full w-fit ${org.role === 'owner' ? 'bg-accent' : 'bg-primary/10'
                                    }`}>
                                    {org.role === 'owner' ? (
                                        <Crown className="h-5 w-5 text-accent-foreground" />
                                    ) : (
                                        <Building2 className="h-5 w-5 text-primary" />
                                    )}
                                </div>

                                <div className="flex-1 min-w-0">
                                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                                        <div>
                                            <h4 className="text-base font-semibold text-foreground truncate">{org.organization_name}</h4>
                                            <p className="text-sm text-muted-foreground">@{org.organization_slug}</p>
                                        </div>
                                        <div className="flex items-center space-x-2">
                                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${org.status === 'active'
                                                ? 'bg-success/10 text-success'
                                                : 'bg-muted text-muted-foreground'
                                                }`}>
                                                {org.status}
                                            </span>
                                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${org.role === 'owner'
                                                ? 'bg-accent text-accent-foreground'
                                                : 'bg-secondary/10 text-secondary-foreground'
                                                }`}>
                                                {org.role === 'owner' && <Crown className="h-3 w-3 mr-1" />}
                                                {org.role}
                                            </span>
                                        </div>
                                    </div>

                                    <div className="mt-2 flex items-center text-xs text-muted-foreground">
                                        <span>Joined {new Date(org.joined_at).toLocaleDateString('en-US', {
                                            year: 'numeric',
                                            month: 'short',
                                            day: 'numeric'
                                        })}</span>
                                    </div>

                                    {/* Owner-specific actions */}
                                    {org.role === 'owner' && (
                                        <div className="mt-4 flex flex-wrap gap-2 opacity-100 sm:opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                                            <Button
                                                size="sm"
                                                variant="outline"
                                                className="text-xs h-8 bg-card hover:bg-primary/10 hover:text-primary border-border"
                                                onClick={() => onInvite(org)}
                                            >
                                                <UserPlus className="h-3 w-3 mr-1.5" />
                                                Invite
                                            </Button>
                                            <Button
                                                size="sm"
                                                variant="outline"
                                                className="text-xs h-8 bg-card hover:bg-secondary/10 hover:text-secondary-foreground border-border"
                                                onClick={() => onManageRoles(org)}
                                            >
                                                <Key className="h-3 w-3 mr-1.5" />
                                                Roles
                                            </Button>
                                            <Button
                                                size="sm"
                                                variant="outline"
                                                className="text-xs h-8 bg-card hover:bg-muted hover:text-foreground border-border"
                                                onClick={() => onSettings(org)}
                                            >
                                                <Settings className="h-3 w-3 mr-1.5" />
                                                Settings
                                            </Button>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    ))
                ) : (
                    <div className="text-center py-12 bg-muted/50 rounded-xl border border-dashed border-border">
                        <Users className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                        <h3 className="text-lg font-medium text-foreground">No organizations yet</h3>
                        <p className="text-sm text-muted-foreground mt-1 mb-6">Get started by creating your first organization</p>
                        <Button onClick={onCreateOrg}>
                            Create Organization
                        </Button>
                    </div>
                )}

                {organizations.length > 0 && (
                    <div className="pt-4 border-t border-border">
                        <Button variant="ghost" size="sm" className="w-full text-muted-foreground hover:text-foreground">
                            <Eye className="h-4 w-4 mr-2" />
                            View All Organizations
                        </Button>
                    </div>
                )}
            </div>
        </Card>
    );
}
