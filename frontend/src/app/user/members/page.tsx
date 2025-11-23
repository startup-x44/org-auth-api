'use client';

import { useAuth } from '../../../contexts/auth-context';
import { useState, useEffect } from 'react';
import { AuthService } from '../../../lib/auth';
import { Card } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Label } from '../../../components/ui/label';
import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectTrigger,
    SelectValue
} from '../../../components/ui/select';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter
} from '../../../components/ui/dialog';
import { Users, UserPlus, Mail, Shield, AlertCircle, Search, Crown } from 'lucide-react';
import { useToast } from '../../../hooks/use-toast';

interface Member {
    id: string;
    email: string;
    first_name?: string;
    last_name?: string;
    role: string;
    status: string;
    joined_at: string;
}

interface Role {
    id: string;
    name: string;
    description?: string;
}

interface Organization {
    organization_id: string;
    organization_name: string;
    role: string;
}

export default function MembersPage() {
    const { user } = useAuth();
    const { toast } = useToast();

    const [members, setMembers] = useState<Member[]>([]);
    const [roles, setRoles] = useState<Role[]>([]);
    const [organizations, setOrganizations] = useState<Organization[]>([]);
    const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');

    // Invite modal state
    const [inviteModalOpen, setInviteModalOpen] = useState(false);
    const [inviteEmail, setInviteEmail] = useState('');
    const [inviteRole, setInviteRole] = useState('');
    const [inviting, setInviting] = useState(false);

    useEffect(() => {
        loadOrganizations();
    }, []);

    useEffect(() => {
        if (selectedOrg) {
            loadMembers();
            loadRoles();
        }
    }, [selectedOrg]);

    const loadOrganizations = async () => {
        try {
            const response = await AuthService.apiCall('/api/v1/user/organizations');
            if (response.ok) {
                const data = await response.json();
                const orgs = data.data || data.organizations || [];
                setOrganizations(orgs);

                // Set first owned organization as selected
                const ownedOrg = orgs.find((org: Organization) => org.role === 'owner');
                if (ownedOrg) {
                    setSelectedOrg(ownedOrg);
                }
            }
        } catch (error) {
            console.error('Failed to load organizations:', error);
        }
    };

    const loadMembers = async () => {
        if (!selectedOrg) return;

        try {
            setLoading(true);
            const response = await AuthService.apiCall(`/api/v1/organizations/${selectedOrg.organization_id}/members`);
            if (response.ok) {
                const data = await response.json();
                setMembers(data.data || data.members || []);
            }
        } catch (error) {
            console.error('Failed to load members:', error);
            toast({
                title: "Failed to load members",
                description: "Could not fetch organization members.",
                variant: "destructive",
            });
        } finally {
            setLoading(false);
        }
    };

    const loadRoles = async () => {
        if (!selectedOrg) return;

        try {
            const response = await AuthService.apiCall(`/api/v1/organizations/${selectedOrg.organization_id}/roles`);
            if (response.ok) {
                const data = await response.json();
                setRoles(data.data || data.roles || []);
            }
        } catch (error) {
            console.error('Failed to load roles:', error);
        }
    };

    const handleInvite = async () => {
        if (!inviteEmail || !inviteRole) {
            toast({
                title: "Missing information",
                description: "Please provide email and select a role.",
                variant: "destructive",
            });
            return;
        }

        if (roles.length === 0) {
            toast({
                title: "No roles available",
                description: "Please create at least one role before inviting members.",
                variant: "destructive",
            });
            return;
        }

        try {
            setInviting(true);
            const response = await AuthService.apiCall(
                `/api/v1/organizations/${selectedOrg?.organization_id}/members`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        email: inviteEmail,
                        role: inviteRole,
                    }),
                }
            );

            if (response.ok) {
                toast({
                    title: "Invitation sent",
                    description: `Invitation sent to ${inviteEmail}`,
                });
                setInviteModalOpen(false);
                setInviteEmail('');
                setInviteRole('');
                loadMembers();
            } else {
                const error = await response.json();
                toast({
                    title: "Failed to send invitation",
                    description: error.message || "Could not send invitation.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to invite member:', error);
            toast({
                title: "Error",
                description: "An error occurred while sending the invitation.",
                variant: "destructive",
            });
        } finally {
            setInviting(false);
        }
    };

    const filteredMembers = members.filter(member =>
        member.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
        `${member.first_name} ${member.last_name}`.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const isOwner = selectedOrg?.role === 'owner';

    if (loading && !selectedOrg) {
        return (
            <div className="space-y-6">
                <div className="h-8 w-48 bg-muted rounded animate-pulse"></div>
                <div className="h-96 bg-muted rounded-lg animate-pulse"></div>
            </div>
        );
    }

    if (!isOwner) {
        return (
            <div className="space-y-6">
                <div>
                    <h1 className="text-2xl font-bold text-foreground">Members</h1>
                    <p className="text-sm text-muted-foreground mt-1">Manage organization members</p>
                </div>
                <Card className="p-12 border-none shadow-sm text-center bg-destructive/10">
                    <AlertCircle className="h-16 w-16 text-destructive mx-auto mb-4" />
                    <h2 className="text-xl font-bold text-foreground mb-2">Access Denied</h2>
                    <p className="text-muted-foreground">
                        Only organization owners can access member management.
                    </p>
                </Card>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Page Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-foreground">Members</h1>
                    <p className="text-sm text-muted-foreground mt-1">
                        Manage members for {selectedOrg?.organization_name}
                    </p>
                </div>
                <Button
                    className="w-full sm:w-auto"
                    onClick={() => {
                        if (roles.length === 0) {
                            toast({
                                title: "No roles available",
                                description: "Please create at least one role in Roles & Permissions before inviting members.",
                                variant: "destructive",
                            });
                        } else {
                            setInviteModalOpen(true);
                        }
                    }}
                >
                    <UserPlus className="h-4 w-4 mr-2" />
                    Invite Member
                </Button>
            </div>

            {/* No Roles Warning */}
            {roles.length === 0 && (
                <Card className="p-4 border-destructive/30 bg-destructive/10">
                    <div className="flex items-start space-x-3">
                        <AlertCircle className="h-5 w-5 text-destructive flex-shrink-0 mt-0.5" />
                        <div>
                            <h3 className="font-semibold text-destructive">No roles available</h3>
                            <p className="text-sm text-destructive/80 mt-1">
                                You need to create at least one role before you can invite members.
                                Go to <strong>Roles & Permissions</strong> to create roles.
                            </p>
                        </div>
                    </div>
                </Card>
            )}

            {/* Search */}
            <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                    type="text"
                    placeholder="Search members..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                />
            </div>

            {/* Members List */}
            {loading ? (
                <div className="space-y-4">
                    {[1, 2, 3].map((i) => (
                        <div key={i} className="h-20 bg-muted rounded-lg animate-pulse"></div>
                    ))}
                </div>
            ) : filteredMembers.length > 0 ? (
                <div className="space-y-4">
                    {filteredMembers.map((member, index) => (
                        <Card key={`member-${member.id}-${index}`} className="p-4 border-none shadow-sm hover:shadow-md transition-shadow">
                            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                                <div className="flex items-center space-x-4">
                                    <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center flex-shrink-0">
                                        <Users className="h-6 w-6 text-primary" />
                                    </div>
                                    <div>
                                        <h3 className="font-semibold text-foreground">
                                            {member.first_name} {member.last_name}
                                        </h3>
                                        <p className="text-sm text-muted-foreground flex items-center">
                                            <Mail className="h-3 w-3 mr-1" />
                                            {member.email}
                                        </p>
                                    </div>
                                </div>
                                <div className="flex items-center justify-between sm:justify-end space-x-4 w-full sm:w-auto">
                                    <div className="text-right">
                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${member.role === 'owner'
                                            ? 'bg-accent text-accent-foreground'
                                            : 'bg-secondary text-secondary-foreground'
                                            }`}>
                                            {member.role === 'owner' && <Crown className="h-3 w-3 mr-1" />}
                                            {member.role}
                                        </span>
                                        <p className="text-xs text-muted-foreground mt-1">
                                            Joined {new Date(member.joined_at).toLocaleDateString()}
                                        </p>
                                    </div>
                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${member.status === 'active'
                                        ? 'bg-success/10 text-success'
                                        : 'bg-muted text-muted-foreground'
                                        }`}>
                                        {member.status}
                                    </span>
                                </div>
                            </div>
                        </Card>
                    ))}
                </div>
            ) : (
                <Card className="p-12 border-none shadow-sm text-center">
                    <Users className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-semibold text-foreground mb-2">No members found</h3>
                    <p className="text-sm text-muted-foreground mb-6">
                        {searchTerm ? 'Try a different search term' : 'Invite members to get started'}
                    </p>
                </Card>
            )}

            {/* Invite Modal */}
            <Dialog open={inviteModalOpen} onOpenChange={(open) => {
                setInviteModalOpen(open);
                if (!open) {
                    setInviteEmail('');
                    setInviteRole('');
                }
            }}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>Invite Member</DialogTitle>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="email">Email Address</Label>
                            <Input
                                id="email"
                                type="email"
                                placeholder="member@example.com"
                                value={inviteEmail}
                                onChange={(e) => setInviteEmail(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="role">Role</Label>
                            <Select value={inviteRole} onValueChange={setInviteRole}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select a role" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectGroup>
                                        {roles.map((role) => (
                                            <SelectItem key={role.id} value={role.name}>
                                                {role.name}
                                            </SelectItem>
                                        ))}
                                    </SelectGroup>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => {
                                setInviteModalOpen(false);
                                setInviteEmail('');
                                setInviteRole('');
                            }}
                        >
                            Cancel
                        </Button>
                        <Button onClick={handleInvite} disabled={inviting}>
                            {inviting ? 'Sending...' : 'Send Invitation'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
