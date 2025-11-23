'use client';

import { useState, useEffect } from 'react';
import { AuthService } from '../../lib/auth';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Card } from '../ui/card';
import { Label } from '../ui/label';
import { Checkbox } from '../ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../ui/dialog';
import { LoadingSpinner } from '../ui/loading-spinner';
import { useToast } from '../../hooks/use-toast';
import {
  X,
  Key,
  Shield,
  Users,
  Plus,
  Edit,
  Trash2,
  Check,
  Save
} from 'lucide-react';

interface RoleManagementModalProps {
  isOpen: boolean;
  onClose: () => void;
  organizationId: string;
  organizationName: string;
}

interface Role {
  id: string;
  name: string;
  description: string;
  permissions: string[];
  is_system: boolean;
  member_count?: number;
}

interface Permission {
  id: string;
  name: string;
  description: string;
  category: string;
}

const defaultPermissions = [
  { id: 'read_users', name: 'Read Users', description: 'View user information', category: 'Users' },
  { id: 'manage_users', name: 'Manage Users', description: 'Create, update, and delete users', category: 'Users' },
  { id: 'read_roles', name: 'Read Roles', description: 'View role information', category: 'Roles' },
  { id: 'manage_roles', name: 'Manage Roles', description: 'Create and modify custom roles', category: 'Roles' },
  { id: 'read_settings', name: 'Read Settings', description: 'View organization settings', category: 'Settings' },
  { id: 'manage_settings', name: 'Manage Settings', description: 'Modify organization settings', category: 'Settings' },
];

export function RoleManagementModal({ isOpen, onClose, organizationId, organizationName }: RoleManagementModalProps) {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>(defaultPermissions);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [editingRole, setEditingRole] = useState<string | null>(null);
  const [newRole, setNewRole] = useState({ name: '', description: '', permissions: [] as string[] });
  const { toast } = useToast();

  useEffect(() => {
    if (isOpen) {
      loadRoles();
    }
  }, [isOpen, organizationId]);

  if (!isOpen) return null;

  const loadRoles = async () => {
    try {
      setLoading(true);

      // Load roles from backend
      const response = await AuthService.apiCall(`/api/v1/organizations/${organizationId}/roles`);

      if (response.ok) {
        const data = await response.json();
        console.log('Roles response:', data);

        // Map backend response to our Role interface
        const backendRoles = data.data || [];
        const mappedRoles: Role[] = await Promise.all(
          backendRoles.map(async (role: any) => {
            // Get permissions for each role
            let rolePermissions: string[] = [];
            try {
              const permResponse = await AuthService.apiCall(`/api/v1/organizations/${organizationId}/roles/${role.id}/permissions`);
              if (permResponse.ok) {
                const permData = await permResponse.json();
                console.log(`Permissions for role ${role.name}:`, permData);
                rolePermissions = permData.data || [];
              }
            } catch (error) {
              console.error(`Failed to load permissions for role ${role.id}:`, error);
            }

            return {
              id: role.id,
              name: role.display_name || role.name,
              description: role.description || '',
              permissions: rolePermissions,
              is_system: role.is_system || false,
              member_count: 0 // TODO: Get actual member count
            };
          })
        );

        setRoles(mappedRoles);
      } else {
        console.error('Failed to load roles:', response.status);
        toast({
          title: "Failed to load roles",
          description: "Could not fetch roles from server.",
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error('Failed to load roles:', error);
      toast({
        title: "Failed to load roles",
        description: "Please try again later.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRole = async () => {
    if (!newRole.name.trim()) {
      toast({
        title: "Role name required",
        description: "Please enter a name for the role.",
        variant: "destructive",
      });
      return;
    }

    setCreating(true);
    try {
      // Create role
      const roleResponse = await AuthService.apiCall(`/api/v1/organizations/${organizationId}/roles`, {
        method: 'POST',
        body: JSON.stringify({
          name: newRole.name.toLowerCase().replace(/\s+/g, '_'), // Convert to snake_case for backend
          display_name: newRole.name,
          description: newRole.description,
        }),
      });

      if (!roleResponse.ok) {
        const errorData = await roleResponse.json();
        throw new Error(errorData.message || 'Failed to create role');
      }

      const createdRoleData = await roleResponse.json();
      const createdRole = createdRoleData.data;
      console.log('Created role:', createdRole);

      // Assign permissions to the new role
      if (newRole.permissions.length > 0) {
        const permissionResponse = await AuthService.apiCall(
          `/api/v1/organizations/${organizationId}/roles/${createdRole.id}/permissions`,
          {
            method: 'POST',
            body: JSON.stringify({
              permissions: newRole.permissions
            }),
          }
        );

        if (!permissionResponse.ok) {
          console.error('Failed to assign permissions to role');
          // Don't throw error here, role was created successfully
        }
      }

      // Reload roles to get updated data
      await loadRoles();

      // Reset form
      setNewRole({ name: '', description: '', permissions: [] });

      toast({
        title: "Role created!",
        description: `Successfully created ${newRole.name} role with ${newRole.permissions.length} permission${newRole.permissions.length !== 1 ? 's' : ''}.`,
        variant: "success",
      });
    } catch (error) {
      console.error('Create role error:', error);
      toast({
        title: "Failed to create role",
        description: error instanceof Error ? error.message : "Please try again later.",
        variant: "destructive",
      });
    } finally {
      setCreating(false);
    }
  };

  const togglePermission = (permissionId: string) => {
    const current = newRole.permissions.includes(permissionId);
    const updated = current
      ? newRole.permissions.filter(p => p !== permissionId)
      : [...newRole.permissions, permissionId];

    setNewRole({ ...newRole, permissions: updated });
  };

  const toggleCategoryPermissions = (category: string) => {
    const categoryPermissions = groupedPermissions[category].map(p => p.id);
    const allSelected = categoryPermissions.every(permId => newRole.permissions.includes(permId));

    let updated: string[];
    if (allSelected) {
      // Deselect all permissions in this category
      updated = newRole.permissions.filter(permId => !categoryPermissions.includes(permId));
    } else {
      // Select all permissions in this category
      const newPermissions = categoryPermissions.filter(permId => !newRole.permissions.includes(permId));
      updated = [...newRole.permissions, ...newPermissions];
    }

    setNewRole({ ...newRole, permissions: updated });
  };

  const isCategoryFullySelected = (category: string): boolean => {
    const categoryPermissions = groupedPermissions[category].map(p => p.id);
    return categoryPermissions.every(permId => newRole.permissions.includes(permId));
  };

  const isCategoryPartiallySelected = (category: string): boolean => {
    const categoryPermissions = groupedPermissions[category].map(p => p.id);
    return categoryPermissions.some(permId => newRole.permissions.includes(permId)) &&
      !categoryPermissions.every(permId => newRole.permissions.includes(permId));
  };

  const groupedPermissions = permissions.reduce((acc, permission) => {
    if (!acc[permission.category]) {
      acc[permission.category] = [];
    }
    acc[permission.category].push(permission);
    return acc;
  }, {} as Record<string, Permission[]>);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-hidden p-0 gap-0">
        <DialogHeader className="p-6 border-b">
          <div className="flex items-center">
            <div className="p-2 bg-secondary/10 rounded-lg mr-3">
              <Key className="h-5 w-5 text-secondary-foreground" />
            </div>
            <div>
              <DialogTitle className="text-lg font-semibold text-foreground">Role Management</DialogTitle>
              <p className="text-sm text-muted-foreground">Manage roles and permissions for {organizationName}</p>
            </div>
          </div>
        </DialogHeader>

        <div className="flex h-96">
          {/* Existing Roles */}
          <div className="w-1/2 p-6 border-r overflow-y-auto">
            <h4 className="text-md font-semibold text-foreground mb-4">Existing Roles</h4>

            {loading ? (
              <div className="flex items-center justify-center py-8">
                <LoadingSpinner size="sm" />
              </div>
            ) : (
              <div className="space-y-3">
                {roles.map((role) => (
                  <div key={role.id} className="p-4 border border-border rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center">
                        <Shield className="h-4 w-4 text-muted-foreground mr-2" />
                        <span className="font-medium text-foreground">{role.name}</span>
                        {role.is_system && (
                          <span className="ml-2 inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-muted text-muted-foreground">
                            System
                          </span>
                        )}
                      </div>
                      <div className="flex items-center space-x-2">
                        <Users className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm text-muted-foreground">{role.member_count}</span>
                      </div>
                    </div>
                    <p className="text-sm text-muted-foreground mb-2">{role.description}</p>
                    <p className="text-xs text-muted-foreground">
                      {role.permissions.length} permission{role.permissions.length !== 1 ? 's' : ''}
                    </p>

                    {!role.is_system && (
                      <div className="flex items-center space-x-2 mt-3">
                        <Button size="sm" variant="outline" className="text-xs">
                          <Edit className="h-3 w-3 mr-1" />
                          Edit
                        </Button>
                        <Button size="sm" variant="outline" className="text-xs text-red-600">
                          <Trash2 className="h-3 w-3 mr-1" />
                          Delete
                        </Button>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Create New Role */}
          <div className="w-1/2 p-6 overflow-y-auto">
            <h4 className="text-md font-semibold text-foreground mb-4">Create New Role</h4>

            <div className="space-y-4">
              <div>
                <Label className="block text-sm font-medium text-foreground mb-2">
                  Role Name *
                </Label>
                <Input
                  type="text"
                  placeholder="e.g., Project Manager"
                  value={newRole.name}
                  onChange={(e) => setNewRole({ ...newRole, name: e.target.value })}
                />
              </div>

              <div>
                <Label className="block text-sm font-medium text-foreground mb-2">
                  Description
                </Label>
                <Input
                  type="text"
                  placeholder="Brief description of this role"
                  value={newRole.description}
                  onChange={(e) => setNewRole({ ...newRole, description: e.target.value })}
                />
              </div>

              <div>
                <Label className="block text-sm font-medium text-foreground mb-3">
                  Permissions ({newRole.permissions.length} selected)
                </Label>

                <div className="space-y-4 max-h-48 overflow-y-auto">
                  {Object.entries(groupedPermissions).map(([category, perms]) => (
                    <div key={category} className="border border-border rounded-lg p-3">
                      {/* Category Header with Select All Checkbox */}
                      <div className="flex items-center mb-3 pb-2 border-b border-border">
                        <Checkbox
                          id={`cat-${category}`}
                          checked={isCategoryFullySelected(category)}
                          onCheckedChange={() => toggleCategoryPermissions(category)}
                          className="mr-3"
                        />
                        <div className="flex-1">
                          <Label htmlFor={`cat-${category}`} className="text-sm font-medium text-foreground cursor-pointer">
                            {category}
                          </Label>
                          <p className="text-xs text-muted-foreground">
                            {perms.filter(p => newRole.permissions.includes(p.id)).length} of {perms.length} selected
                          </p>
                        </div>
                      </div>

                      {/* Individual Permissions */}
                      <div className="space-y-2 ml-6">
                        {perms.map((permission) => (
                          <div key={permission.id} className="flex items-start">
                            <Checkbox
                              id={permission.id}
                              checked={newRole.permissions.includes(permission.id)}
                              onCheckedChange={() => togglePermission(permission.id)}
                              className="mt-1 mr-3"
                            />
                            <div>
                              <Label htmlFor={permission.id} className="text-sm font-medium text-foreground cursor-pointer">
                                {permission.name}
                              </Label>
                              <p className="text-xs text-muted-foreground">{permission.description}</p>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter className="flex items-center justify-between p-6 border-t bg-muted/50 sm:justify-between">
          <div className="text-sm text-muted-foreground">
            {newRole.permissions.length} permission{newRole.permissions.length !== 1 ? 's' : ''} selected for new role
          </div>
          <div className="flex items-center space-x-3">
            <Button variant="outline" onClick={onClose}>
              Close
            </Button>
            <Button onClick={handleCreateRole} disabled={creating || !newRole.name.trim()}>
              {creating ? (
                <div className="flex items-center">
                  <LoadingSpinner size="sm" variant="secondary" />
                  <span className="ml-2">Creating...</span>
                </div>
              ) : (
                <div className="flex items-center">
                  <Save className="h-4 w-4 mr-2" />
                  <span>Create Role</span>
                </div>
              )}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}