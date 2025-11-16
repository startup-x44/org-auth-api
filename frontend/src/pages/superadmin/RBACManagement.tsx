import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  ArrowLeft,
  Shield,
  Users,
  Key,
  Plus,
  Edit,
  Trash2,
  Settings,
  Lock,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Checkbox } from '@/components/ui/checkbox'
import LoadingSpinner from '@/components/ui/loading-spinner'
import useAuthStore from '@/store/auth'
import { rbacAPI } from '@/lib/api'

interface Permission {
  id: string
  name: string
  display_name: string
  description: string
  category: string
  is_system: boolean
  organization_id?: string
}

interface Role {
  id: string
  name: string
  display_name: string
  description: string
  is_system: boolean
  organization_id?: string
  created_at: string
  updated_at: string
}

interface RBACStats {
  total_permissions: number
  system_permissions: number
  custom_permissions: number
  system_roles: number
}

export default function RBACManagement() {
  const navigate = useNavigate()
  const { isSuperadmin, user, organizationId } = useAuthStore()

  const [activeTab, setActiveTab] = useState('roles')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Data state
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [stats, setStats] = useState<RBACStats | null>(null)

  // Dialog state
  const [isCreateRoleOpen, setIsCreateRoleOpen] = useState(false)
  const [isEditRoleOpen, setIsEditRoleOpen] = useState(false)
  const [isPermissionsOpen, setIsPermissionsOpen] = useState(false)
  const [selectedRole, setSelectedRole] = useState<Role | null>(null)
  const [rolePermissions, setRolePermissions] = useState<string[]>([])

  // Form state
  const [roleForm, setRoleForm] = useState({
    name: '',
    display_name: '',
    description: '',
    permissions: [] as string[]
  })

  // Check access - allow both superadmin and organization owners
  useEffect(() => {
    // Allow access for superadmin OR users with organization (owners)
    if (!user || (!isSuperadmin && !organizationId)) {
      navigate('/dashboard', { replace: true })
    }
  }, [isSuperadmin, user, organizationId, navigate])

  // Load data
  useEffect(() => {
    if (user && (isSuperadmin || organizationId)) {
      loadData()
    }
  }, [isSuperadmin, user, organizationId])

  const loadData = async () => {
    try {
      setLoading(true)
      setError('')

      if (isSuperadmin) {
        // Superadmin: Load system-wide RBAC data
        const [rolesResponse, permissionsResponse, statsResponse] = await Promise.all([
          rbacAPI.listRoles(),
          rbacAPI.listPermissions(),
          rbacAPI.getStats(),
        ])

        setRoles(rolesResponse.data || [])
        setPermissions(permissionsResponse.data || [])
        setStats(statsResponse.data || null)
      } else if (organizationId) {
        // Organization Owner: Load organization-specific roles and permissions
        // TODO: Create organization-scoped API endpoints
        // For now, filter system roles on the frontend
        const [rolesResponse, permissionsResponse] = await Promise.all([
          rbacAPI.listRoles(),
          rbacAPI.listPermissions(),
        ])

        // Filter to show only organization-specific roles (exclude system roles)
        const orgRoles = (rolesResponse.data || []).filter((role: Role) => 
          !role.is_system && role.organization_id === organizationId
        )
        
        // Filter to show only system permissions (org can use these) and org-specific permissions
        const availablePermissions = (permissionsResponse.data || []).filter((perm: Permission) => 
          perm.is_system || perm.organization_id === organizationId
        )

        setRoles(orgRoles)
        setPermissions(availablePermissions)
        setStats(null) // No stats for org-level view for now
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to load RBAC data'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateRole = async () => {
    try {
      await rbacAPI.createRole(roleForm)
      setIsCreateRoleOpen(false)
      setRoleForm({ name: '', display_name: '', description: '', permissions: [] })
      loadData()
    } catch (error: any) {
      setError(error.response?.data?.message || 'Failed to create role')
    }
  }

  const handleUpdateRole = async () => {
    if (!selectedRole) return

    try {
      await rbacAPI.updateRole(selectedRole.id, {
        display_name: roleForm.display_name,
        description: roleForm.description,
      })
      setIsEditRoleOpen(false)
      setSelectedRole(null)
      setRoleForm({ name: '', display_name: '', description: '', permissions: [] })
      loadData()
    } catch (error: any) {
      setError(error.response?.data?.message || 'Failed to update role')
    }
  }

  const handleDeleteRole = async (role: Role) => {
    if (!confirm(`Are you sure you want to delete the "${role.display_name}" role?`)) return

    try {
      await rbacAPI.deleteRole(role.id)
      loadData()
    } catch (error: any) {
      setError(error.response?.data?.message || 'Failed to delete role')
    }
  }

  const handleManagePermissions = async (role: Role) => {
    try {
      setSelectedRole(role)
      const response = await rbacAPI.getRolePermissions(role.id)
      setRolePermissions(response.data || [])
      setIsPermissionsOpen(true)
    } catch (error: any) {
      setError(error.response?.data?.message || 'Failed to load role permissions')
    }
  }

  const handleUpdatePermissions = async () => {
    if (!selectedRole) return

    try {
      // Get current permissions
      const currentResponse = await rbacAPI.getRolePermissions(selectedRole.id)
      const currentPermissions = currentResponse.data || []

      // Find permissions to add and remove
      const toAdd = rolePermissions.filter(p => !currentPermissions.includes(p))
      const toRemove = currentPermissions.filter(p => !rolePermissions.includes(p))

      // Apply changes
      if (toAdd.length > 0) {
        await rbacAPI.assignPermissions(selectedRole.id, toAdd)
      }
      if (toRemove.length > 0) {
        await rbacAPI.revokePermissions(selectedRole.id, toRemove)
      }

      setIsPermissionsOpen(false)
      setSelectedRole(null)
      setRolePermissions([])
      loadData()
    } catch (error: any) {
      setError(error.response?.data?.message || 'Failed to update permissions')
    }
  }

  const openCreateRole = () => {
    setRoleForm({ name: '', display_name: '', description: '', permissions: [] })
    setIsCreateRoleOpen(true)
  }

  const openEditRole = (role: Role) => {
    setSelectedRole(role)
    setRoleForm({
      name: role.name,
      display_name: role.display_name,
      description: role.description,
      permissions: []
    })
    setIsEditRoleOpen(true)
  }

  const groupPermissionsByCategory = (permissions: Permission[]) => {
    return permissions.reduce((acc, permission) => {
      const category = permission.category || 'other'
      if (!acc[category]) {
        acc[category] = []
      }
      acc[category].push(permission)
      return acc
    }, {} as Record<string, Permission[]>)
  }

  if (!isSuperadmin) {
    return null
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  const permissionsByCategory = groupPermissionsByCategory(permissions)

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      {/* Header */}
      <div className="bg-white dark:bg-slate-800 border-b">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Button
                variant="ghost"
                onClick={() => navigate('/admin')}
                className="flex items-center space-x-2"
              >
                <ArrowLeft className="h-4 w-4" />
                <span>Back to Admin</span>
              </Button>
              <div>
                <h1 className="text-3xl font-bold flex items-center gap-2">
                  <Shield className="h-8 w-8 text-red-600" />
                  {isSuperadmin ? 'System RBAC Management' : 'Role Management'}
                </h1>
                <p className="text-muted-foreground mt-2">
                  {isSuperadmin 
                    ? 'Manage system-wide roles and permissions' 
                    : 'Manage organization roles and permissions'
                  }
                </p>
              </div>
            </div>
            {activeTab === 'roles' && (
              <Button onClick={openCreateRole}>
                <Plus className="mr-2 h-4 w-4" />
                Create Role
              </Button>
            )}
          </div>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        {/* Error Alert */}
        {error && (
          <Alert variant="destructive" className="mb-6">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Statistics - Only for Superadmin */}
        {isSuperadmin && stats && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">System Roles</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.system_roles}</div>
                <p className="text-xs text-muted-foreground">Active system roles</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Permissions</CardTitle>
                <Key className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.total_permissions}</div>
                <p className="text-xs text-muted-foreground">Available permissions</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">System Permissions</CardTitle>
                <Lock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.system_permissions}</div>
                <p className="text-xs text-muted-foreground">Built-in permissions</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Custom Permissions</CardTitle>
                <Settings className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.custom_permissions}</div>
                <p className="text-xs text-muted-foreground">Organization-specific</p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Organization Stats - For Organization Owners */}
        {!isSuperadmin && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Custom Roles</CardTitle>
                <Shield className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{roles.length}</div>
                <p className="text-xs text-muted-foreground">Organization roles</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Available Permissions</CardTitle>
                <Key className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{permissions.length}</div>
                <p className="text-xs text-muted-foreground">Assignable permissions</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">System Permissions</CardTitle>
                <Lock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {permissions.filter(p => p.is_system).length}
                </div>
                <p className="text-xs text-muted-foreground">Built-in permissions</p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Main Content */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="roles" className="flex items-center space-x-2">
              <Users className="h-4 w-4" />
              <span>Roles</span>
            </TabsTrigger>
            <TabsTrigger value="permissions" className="flex items-center space-x-2">
              <Key className="h-4 w-4" />
              <span>Permissions</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="roles" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>
                  {isSuperadmin ? 'System Roles' : 'Organization Roles'}
                </CardTitle>
                <CardDescription>
                  {isSuperadmin 
                    ? 'Manage system-wide roles and their permissions'
                    : 'Manage custom roles for your organization'
                  }
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {roles.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">
                      No roles found
                    </div>
                  ) : (
                    roles.map((role) => (
                      <motion.div
                        key={role.id}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        className="flex items-center justify-between p-4 border rounded-lg"
                      >
                        <div className="flex items-center space-x-4">
                          <div className="w-10 h-10 bg-red-100 dark:bg-red-900 rounded-lg flex items-center justify-center">
                            <Users className="h-5 w-5 text-red-600 dark:text-red-400" />
                          </div>
                          <div>
                            <div className="font-medium">{role.display_name}</div>
                            <div className="text-sm text-muted-foreground">
                              {role.description}
                            </div>
                            <div className="flex items-center space-x-2 mt-1">
                              <Badge variant={role.is_system ? "default" : "secondary"}>
                                {role.is_system ? "System" : "Custom"}
                              </Badge>
                              <span className="text-xs text-muted-foreground">
                                {role.name}
                              </span>
                            </div>
                          </div>
                        </div>
                        <div className="flex items-center space-x-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleManagePermissions(role)}
                          >
                            <Settings className="mr-1 h-3 w-3" />
                            Permissions
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => openEditRole(role)}
                          >
                            <Edit className="mr-1 h-3 w-3" />
                            Edit
                          </Button>
                          {!role.is_system && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleDeleteRole(role)}
                              className="text-red-600 hover:text-red-700"
                            >
                              <Trash2 className="mr-1 h-3 w-3" />
                              Delete
                            </Button>
                          )}
                        </div>
                      </motion.div>
                    ))
                  )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="permissions" className="space-y-4">
            {Object.entries(permissionsByCategory).map(([category, categoryPermissions]) => (
              <Card key={category}>
                <CardHeader>
                  <CardTitle className="capitalize">{category} Permissions</CardTitle>
                  <CardDescription>
                    {categoryPermissions.length} permission{categoryPermissions.length !== 1 ? 's' : ''} in this category
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {categoryPermissions.map((permission) => (
                      <motion.div
                        key={permission.id}
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        className="p-4 border rounded-lg"
                      >
                        <div className="flex items-center justify-between mb-2">
                          <h4 className="font-medium">{permission.display_name}</h4>
                          <Badge variant={permission.is_system ? "default" : "secondary"}>
                            {permission.is_system ? "System" : "Custom"}
                          </Badge>
                        </div>
                        <p className="text-sm text-muted-foreground mb-2">
                          {permission.description}
                        </p>
                        <code className="text-xs bg-muted px-2 py-1 rounded">
                          {permission.name}
                        </code>
                      </motion.div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            ))}
          </TabsContent>
        </Tabs>
      </div>

      {/* Create Role Dialog */}
      <Dialog open={isCreateRoleOpen} onOpenChange={setIsCreateRoleOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create System Role</DialogTitle>
            <DialogDescription>
              Create a new system role with specific permissions
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">Name</label>
              <Input
                placeholder="role-name"
                value={roleForm.name}
                onChange={(e) => setRoleForm({ ...roleForm, name: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Display Name</label>
              <Input
                placeholder="Role Display Name"
                value={roleForm.display_name}
                onChange={(e) => setRoleForm({ ...roleForm, display_name: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Description</label>
              <Textarea
                placeholder="Role description..."
                value={roleForm.description}
                onChange={(e) => setRoleForm({ ...roleForm, description: e.target.value })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsCreateRoleOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateRole}>Create Role</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Role Dialog */}
      <Dialog open={isEditRoleOpen} onOpenChange={setIsEditRoleOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Role</DialogTitle>
            <DialogDescription>
              Update role display name and description
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">Display Name</label>
              <Input
                placeholder="Role Display Name"
                value={roleForm.display_name}
                onChange={(e) => setRoleForm({ ...roleForm, display_name: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Description</label>
              <Textarea
                placeholder="Role description..."
                value={roleForm.description}
                onChange={(e) => setRoleForm({ ...roleForm, description: e.target.value })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsEditRoleOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateRole}>Update Role</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Manage Permissions Dialog */}
      <Dialog open={isPermissionsOpen} onOpenChange={setIsPermissionsOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Manage Permissions - {selectedRole?.display_name}</DialogTitle>
            <DialogDescription>
              Select permissions to assign to this role
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-6">
            {Object.entries(permissionsByCategory).map(([category, categoryPermissions]) => (
              <div key={category}>
                <h4 className="font-medium mb-3 capitalize">{category} Permissions</h4>
                <div className="space-y-2">
                  {categoryPermissions.map((permission) => (
                    <div key={permission.id} className="flex items-center space-x-2 p-2 border rounded">
                      <Checkbox
                        id={permission.id}
                        checked={rolePermissions.includes(permission.name)}
                        onCheckedChange={(checked) => {
                          if (checked) {
                            setRolePermissions([...rolePermissions, permission.name])
                          } else {
                            setRolePermissions(rolePermissions.filter(p => p !== permission.name))
                          }
                        }}
                      />
                      <div className="flex-1">
                        <label htmlFor={permission.id} className="font-medium">
                          {permission.display_name}
                        </label>
                        <p className="text-sm text-muted-foreground">
                          {permission.description}
                        </p>
                        <code className="text-xs bg-muted px-1 py-0.5 rounded">
                          {permission.name}
                        </code>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsPermissionsOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdatePermissions}>Update Permissions</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}