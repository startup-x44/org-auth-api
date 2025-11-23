'use client';

import { useAuth } from '@/contexts/auth-context';
import { useState, useEffect } from 'react';
import { AuthService } from '@/lib/auth';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Shield, AlertTriangle, Plus } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { CreateRoleModal } from '@/components/roles/CreateRoleModal';
import { CreatePermissionModal } from '@/components/roles/CreatePermissionModal';
import { AssignPermissionsModal, Permission } from '@/components/roles/AssignPermissionsModal';
import { RoleCard, Role } from '@/components/roles/RoleCard';
import { PermissionCard } from '@/components/roles/PermissionCard';

interface Organization {
    organization_id: string;
    organization_name: string;
    organization_slug: string;
    role: string;
    status: string;
    joined_at: string;
}

export default function RolesPage() {
    const { user } = useAuth();
    const { toast } = useToast();

    const [organizations, setOrganizations] = useState<Organization[]>([]);
    const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null);
    const [roles, setRoles] = useState<Role[]>([]);
    const [permissions, setPermissions] = useState<Permission[]>([]);
    const [loading, setLoading] = useState(true);

    // Modal states
    const [createRoleModalOpen, setCreateRoleModalOpen] = useState(false);
    const [createPermModalOpen, setCreatePermModalOpen] = useState(false);
    const [assignPermModalOpen, setAssignPermModalOpen] = useState(false);

    // Loading states
    const [creatingRole, setCreatingRole] = useState(false);
    const [creatingPerm, setCreatingPerm] = useState(false);
    const [assigningPerms, setAssigningPerms] = useState(false);

    // Selected data
    const [selectedRole, setSelectedRole] = useState<Role | null>(null);
    const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);

    useEffect(() => {
        loadOrganizations();
    }, []);

    useEffect(() => {
        if (selectedOrg) {
            loadRoles();
            loadPermissions();
        }
    }, [selectedOrg]);

    const loadOrganizations = async () => {
        try {
            setLoading(true);
            const response = await AuthService.apiCall('/api/v1/user/organizations');
            if (response.ok) {
                const data = await response.json();
                const orgs = data.data || data.organizations || [];
                setOrganizations(orgs);

                // Select first owner organization
                const ownerOrg = orgs.find((org: Organization) => org.role === 'owner');
                if (ownerOrg) {
                    setSelectedOrg(ownerOrg);
                }
            }
        } catch (error) {
            console.error('Failed to load organizations:', error);
        } finally {
            setLoading(false);
        }
    };

    const loadRoles = async (forceFresh = false) => {
        if (!selectedOrg) return;

        try {
            // Add cache busting parameter when forcing fresh data
            const url = `/api/v1/organizations/${selectedOrg.organization_id}/roles${forceFresh ? `?_t=${Date.now()}` : ''}`;
            const response = await AuthService.apiCall(url);
            if (response.ok) {
                const data = await response.json();
                console.log(`Loaded roles (forceFresh: ${forceFresh}):`, data.data || data.roles || []);
                setRoles(data.data || data.roles || []);
            }
        } catch (error) {
            console.error('Failed to load roles:', error);
        }
    };

    const loadPermissions = async () => {
        if (!selectedOrg) return;

        try {
            const response = await AuthService.apiCall(`/api/v1/organizations/${selectedOrg.organization_id}/permissions`);
            if (response.ok) {
                const data = await response.json();
                setPermissions(data.data || data.permissions || []);
            }
        } catch (error) {
            console.error('Failed to load permissions:', error);
        }
    };

    const handleCreateRole = async (name: string, description: string) => {
        if (!name.trim()) {
            toast({
                title: "Missing information",
                description: "Please provide a role name.",
                variant: "destructive",
            });
            return;
        }

        try {
            setCreatingRole(true);
            const response = await AuthService.apiCall(
                `/api/v1/organizations/${selectedOrg?.organization_id}/roles`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: name.toLowerCase().replace(/\s+/g, '_'),
                        display_name: name,
                        description: description,
                    }),
                }
            );

            if (response.ok) {
                toast({
                    title: "Role created",
                    description: `Role "${name}" has been created successfully.`,
                });
                setCreateRoleModalOpen(false);
                loadRoles(true); // Force fresh data after role creation
            } else {
                const error = await response.json();
                toast({
                    title: "Failed to create role",
                    description: error.message || "Could not create role.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to create role:', error);
            toast({
                title: "Error",
                description: "An error occurred while creating the role.",
                variant: "destructive",
            });
        } finally {
            setCreatingRole(false);
        }
    };

    const handleCreatePermission = async (name: string, displayName: string, category: string) => {
        if (!name.trim()) {
            toast({
                title: "Missing information",
                description: "Please provide a permission name.",
                variant: "destructive",
            });
            return;
        }

        // Validate permission name format: action:resource
        const permissionNameRegex = /^[a-z_]+:[a-z_]+$/;
        if (!permissionNameRegex.test(name)) {
            toast({
                title: "Invalid permission name format",
                description: "Permission name must follow the format 'action:resource' (e.g., view:profile, edit:content). Use only lowercase letters and underscores.",
                variant: "destructive",
            });
            return;
        }

        try {
            setCreatingPerm(true);
            const response = await AuthService.apiCall(
                `/api/v1/organizations/${selectedOrg?.organization_id}/permissions`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: name,
                        display_name: displayName,
                        description: '',
                        category: category,
                    }),
                }
            );

            if (response.ok) {
                toast({
                    title: "Permission created",
                    description: `Permission "${name}" has been created successfully.`,
                });
                setCreatePermModalOpen(false);
                loadPermissions();
            } else {
                const error = await response.json();
                toast({
                    title: "Failed to create permission",
                    description: error.message || "Could not create permission.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to create permission:', error);
            toast({
                title: "Error",
                description: "An error occurred while creating the permission.",
                variant: "destructive",
            });
        } finally {
            setCreatingPerm(false);
        }
    };

    const handleManagePermissions = async (role: Role) => {
        setSelectedRole(role);

        try {
            // Always fetch fresh permission data from the backend to ensure accuracy
            const response = await AuthService.apiCall(
                `/api/v1/organizations/${selectedOrg?.organization_id}/roles/${role.id}/permissions`
            );

            if (response.ok) {
                const data = await response.json();
                console.log('Role permissions API response:', data);

                // Handle different response formats
                let rolePermissionNames: string[] = [];

                if (data.success && data.data) {
                    if (Array.isArray(data.data)) {
                        // If data.data is already an array of strings
                        rolePermissionNames = data.data as string[];
                    } else if (typeof data.data === 'object' && data.data.permissions) {
                        // If data.data is an object with a permissions property
                        rolePermissionNames = Array.isArray(data.data.permissions) ? data.data.permissions : [];
                    } else {
                        console.warn('Unexpected role permissions response format:', data.data);
                        rolePermissionNames = [];
                    }
                } else if (data.data && data.data.permissions) {
                    // Handle case where backend returns permissions directly in data.permissions
                    rolePermissionNames = Array.isArray(data.data.permissions) ? data.data.permissions : [];
                } else {
                    console.log('No role permissions found in response');
                    setSelectedPermissions([]);
                    return;
                }

                console.log('Processed role permission names:', rolePermissionNames);

                const rolePermissionIds = permissions
                    .filter(p => rolePermissionNames.includes(p.name))
                    .map(p => p.id);
                setSelectedPermissions(rolePermissionIds);
            } else {
                console.error('Failed to fetch role permissions, response not OK:', response.status);
                setSelectedPermissions([]);
            }
        } catch (error) {
            console.error('Failed to fetch role permissions:', error);
            // Fallback to role object permissions if available
            if (role.permissions && role.permissions.length > 0) {
                const rolePermissionNames = role.permissions as unknown as string[];
                const rolePermissionIds = permissions
                    .filter(p => rolePermissionNames.includes(p.name))
                    .map(p => p.id);
                setSelectedPermissions(rolePermissionIds);
            } else {
                setSelectedPermissions([]);
            }
        }

        setAssignPermModalOpen(true);
    };

    const handlePermissionToggle = (permissionId: string, checked: boolean) => {
        setSelectedPermissions(prev => {
            if (checked) {
                return prev.includes(permissionId) ? prev : [...prev, permissionId];
            } else {
                return prev.filter(p => p !== permissionId);
            }
        });
    };

    const handleAssignPermissions = async () => {
        if (!selectedRole) return;

        try {
            setAssigningPerms(true);

            // Get current permissions by fetching fresh data from the backend
            let currentPerms: string[] = [];

            try {
                const permResponse = await AuthService.apiCall(
                    `/api/v1/organizations/${selectedOrg?.organization_id}/roles/${selectedRole.id}/permissions`
                );

                if (permResponse.ok) {
                    const permData = await permResponse.json();
                    console.log('Current permissions API response:', permData);

                    // Handle different response formats
                    let rolePermissionNames: string[] = [];

                    if (permData.success && permData.data) {
                        if (Array.isArray(permData.data)) {
                            rolePermissionNames = permData.data as string[];
                        } else if (typeof permData.data === 'object' && permData.data.permissions) {
                            rolePermissionNames = Array.isArray(permData.data.permissions) ? permData.data.permissions : [];
                        } else {
                            console.warn('Unexpected current permissions response format:', permData.data);
                            rolePermissionNames = [];
                        }
                    } else if (permData.data && permData.data.permissions) {
                        rolePermissionNames = Array.isArray(permData.data.permissions) ? permData.data.permissions : [];
                    } else {
                        console.warn('No permissions found in response format:', permData);
                        rolePermissionNames = [];
                    }

                    currentPerms = permissions
                        .filter(p => rolePermissionNames.includes(p.name))
                        .map(p => p.id);
                }
            } catch (error) {
                console.warn('Failed to fetch current permissions, using cached data:', error);
                // Fallback to role object permissions if available
                currentPerms = (selectedRole.permissions || []).map((p: any) =>
                    typeof p === 'string' ? permissions.find(perm => perm.name === p)?.id || p : p.id
                ).filter(Boolean);
            }

            // Find permissions to add and remove
            const toAdd = selectedPermissions.filter(p => !currentPerms.includes(p));
            const toRemove = currentPerms.filter((p: string) => !selectedPermissions.includes(p));

            // Convert permission IDs to names for API requests
            const toAddNames = toAdd.map(id => permissions.find(p => p.id === id)?.name).filter(Boolean) as string[];
            const toRemoveNames = toRemove.map(id => permissions.find(p => p.id === id)?.name).filter(Boolean) as string[];

            // Assign new permissions
            if (toAddNames.length > 0) {
                await AuthService.apiCall(
                    `/api/v1/organizations/${selectedOrg?.organization_id}/roles/${selectedRole.id}/permissions`,
                    {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ permissions: toAddNames }),
                    }
                );
            }

            // Revoke removed permissions
            if (toRemoveNames.length > 0) {
                await AuthService.apiCall(
                    `/api/v1/organizations/${selectedOrg?.organization_id}/roles/${selectedRole.id}/permissions`,
                    {
                        method: 'DELETE',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ permissions: toRemoveNames }),
                    }
                );
            }

            toast({
                title: "Permissions updated",
                description: `Permissions for "${selectedRole.display_name}" have been updated.`,
            });
            setAssignPermModalOpen(false);
            setSelectedRole(null);
            setSelectedPermissions([]);

            // Add a delay to ensure database changes are fully committed and logged
            setTimeout(() => {
                console.log('Refreshing roles after permission changes...');
                loadRoles(true); // Force fresh data from backend
            }, 1000);
        } catch (error: any) {
            console.error('Failed to assign permissions:', error);

            // Provide more specific error messages
            let errorMessage = "Failed to update permissions.";

            if (error?.response?.data?.message) {
                const backendMessage = error.response.data.message;
                if (backendMessage.includes("duplicate key")) {
                    errorMessage = "Some permissions are already assigned. Refreshing data...";
                    // Automatically refresh role data to get current state
                    setTimeout(() => {
                        loadRoles(true); // Force fresh data
                        setAssignPermModalOpen(false);
                        setSelectedRole(null);
                        setSelectedPermissions([]);
                    }, 1000);
                } else if (backendMessage.includes("permission") && backendMessage.includes("not found")) {
                    errorMessage = "One or more permissions could not be found. Please refresh and try again.";
                } else if (backendMessage.includes("Insufficient permissions")) {
                    errorMessage = "You don't have permission to manage role permissions.";
                } else {
                    errorMessage = backendMessage;
                }
            }

            toast({
                title: "Error",
                description: errorMessage,
                variant: "destructive",
            });
        } finally {
            setAssigningPerms(false);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
                    <p className="mt-4 text-muted-foreground">Loading...</p>
                </div>
            </div>
        );
    }

    if (!selectedOrg || selectedOrg.role !== 'owner') {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <Card className="p-8 max-w-md">
                    <div className="text-center">
                        <AlertTriangle className="h-12 w-12 text-destructive mx-auto mb-4" />
                        <h2 className="text-xl font-bold text-foreground mb-2">Access Denied</h2>
                        <p className="text-muted-foreground">
                            Only organization owners can manage roles and permissions.
                        </p>
                    </div>
                </Card>
            </div>
        );
    }

    return (
        <div className="p-6 max-w-7xl mx-auto">
            {/* Header */}
            <div className="mb-6">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-foreground flex items-center gap-2">
                            <Shield className="h-8 w-8 text-primary" />
                            Roles & Permissions
                        </h1>
                        <p className="text-muted-foreground mt-1">
                            Manage roles and permissions for {selectedOrg.organization_name}
                        </p>
                    </div>
                </div>
            </div>

            {/* Roles Section */}
            <div className="mb-8">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-4">
                    <h2 className="text-2xl font-bold text-foreground">Roles</h2>
                    <Button onClick={() => setCreateRoleModalOpen(true)} className="w-full sm:w-auto">
                        <Plus className="h-4 w-4 mr-2" />
                        Create Role
                    </Button>
                </div>
                <div className="space-y-3">
                    {roles.filter(role => !role.is_system).length === 0 ? (
                        <Card className="p-8 text-center text-muted-foreground">
                            No custom roles found. Create your first role to get started.
                        </Card>
                    ) : (
                        roles.filter(role => !role.is_system).map(role => (
                            <RoleCard
                                key={role.id}
                                role={role}
                                onManagePermissions={handleManagePermissions}
                            />
                        ))
                    )}
                </div>
            </div>

            {/* Permissions Section */}
            <div>
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-4">
                    <h2 className="text-2xl font-bold text-foreground">Permissions</h2>
                    <Button onClick={() => setCreatePermModalOpen(true)} className="w-full sm:w-auto">
                        <Plus className="h-4 w-4 mr-2" />
                        Create Permission
                    </Button>
                </div>
                <div className="space-y-8">
                    {permissions.filter(p => !p.is_system).length === 0 ? (
                        <Card className="p-8 text-center text-muted-foreground">
                            No custom permissions found. Create your first permission to get started.
                        </Card>
                    ) : (
                        Object.entries(
                            permissions
                                .filter(p => !p.is_system)
                                .reduce((acc, perm) => {
                                    const category = perm.category || 'other';
                                    if (!acc[category]) {
                                        acc[category] = [];
                                    }
                                    acc[category].push(perm);
                                    return acc;
                                }, {} as Record<string, Permission[]>)
                        ).map(([category, categoryPermissions]) => (
                            <div key={category} className="bg-muted/50 p-4 rounded-lg border border-border">
                                <h3 className="text-lg font-bold text-foreground mb-4 capitalize border-b border-border pb-2 flex items-center">
                                    <span className="bg-primary/10 text-primary text-xs font-medium mr-2 px-2.5 py-0.5 rounded border border-primary/20">
                                        {category}
                                    </span>
                                    Permissions ({categoryPermissions.length})
                                </h3>
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                    {categoryPermissions.map(permission => (
                                        <PermissionCard key={permission.id} permission={permission} />
                                    ))}
                                </div>
                            </div>
                        ))
                    )}
                </div>
            </div>

            {/* Modals */}
            <CreateRoleModal
                isOpen={createRoleModalOpen}
                onClose={() => setCreateRoleModalOpen(false)}
                onSubmit={handleCreateRole}
                isLoading={creatingRole}
            />

            <CreatePermissionModal
                isOpen={createPermModalOpen}
                onClose={() => setCreatePermModalOpen(false)}
                onSubmit={handleCreatePermission}
                isLoading={creatingPerm}
            />

            {selectedRole && (
                <AssignPermissionsModal
                    isOpen={assignPermModalOpen}
                    roleName={selectedRole.display_name}
                    permissions={permissions}
                    selectedPermissions={selectedPermissions}
                    onPermissionToggle={handlePermissionToggle}
                    onSave={handleAssignPermissions}
                    onClose={() => {
                        setAssignPermModalOpen(false);
                        setSelectedRole(null);
                        setSelectedPermissions([]);
                    }}
                    isLoading={assigningPerms}
                />
            )}
        </div>
    );
}
