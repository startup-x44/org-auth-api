import React, { useState, useEffect } from 'react'
import { Plus, Edit, Trash2, Shield, Users, CheckSquare, Square, AlertTriangle, X, Lock, Minus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'  
import { Label } from '@/components/ui/label'
import { organizationAPI } from '@/lib/api'
import useAuthStore from '@/store/auth'
import { useToast } from '@/hooks/useToast'
import ToastContainer from '@/components/ToastContainer'
import { RequirePermission } from '@/components/auth/PermissionGate'

interface Role {
  id: string
  name: string
  display_name: string
  description: string
  is_system: boolean
  permissions: string[]
  member_count: number
}

interface Permission {
  id: string
  name: string
  display_name: string
  description: string
  category: string
  is_system: boolean
  organization_id?: string
}

const RoleManagement: React.FC = () => {
  const { organizationId, hasPermission } = useAuthStore()
  const { toasts, removeToast, showSuccess, showError } = useToast()
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // Check if user can manage roles - only use proper role permissions
  const canManageRoles = hasPermission('role:view')

  // Role form state
  const [isCreating, setIsCreating] = useState(false)
  const [selectedRole, setSelectedRole] = useState<Role | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    display_name: '',
    description: '',
    permissions: [] as string[]
  })

  // Delete confirmation state
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [roleToDelete, setRoleToDelete] = useState<Role | null>(null)

  // Permission management state
  const [showPermissionForm, setShowPermissionForm] = useState(false)
  const [isCreatingPermission, setIsCreatingPermission] = useState(false)
  const [selectedPermission, setSelectedPermission] = useState<Permission | null>(null)
  const [showDeletePermissionModal, setShowDeletePermissionModal] = useState(false)
  const [permissionToDelete, setPermissionToDelete] = useState<Permission | null>(null)
  const [permissionFormData, setPermissionFormData] = useState({
    name: '',
    display_name: '',
    description: '',
    category: ''
  })

  useEffect(() => {
    if (organizationId) {
      fetchRoles()
      fetchPermissions()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [organizationId])

  // Check if user has permission to access role management
  if (!canManageRoles) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <Lock className="mx-auto h-16 w-16 text-gray-400 mb-4" />
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Access Denied</h2>
          <p className="text-gray-600">You don't have permission to manage roles and permissions.</p>
        </div>
      </div>
    )
  }

  const fetchRoles = async () => {
    if (!organizationId) return
    
    try {
      setError('')
      const response = await organizationAPI.getRoles(organizationId)
      if (response.success) {
        // Backend already filters out system roles for non-superadmin users
        setRoles(response.data)
      }
    } catch (error: any) {
      console.error('Failed to fetch roles:', error)
      setError(error.response?.data?.message || 'Failed to load roles. Please check your permissions.')
    }
  }

  const fetchPermissions = async () => {
    if (!organizationId) return
    
    try {
      const response = await organizationAPI.getPermissions(organizationId)
      if (response.success) {
        // Backend should already return only organization-specific permissions
        // (system permissions available for assignment + custom org permissions)
        setPermissions(response.data)
      }
    } catch (error: any) {
      console.error('Failed to fetch permissions:', error)
      setError(error.response?.data?.message || 'Failed to load permissions')
    }
  }

  const handleCreateRole = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!organizationId) return
    
    setLoading(true)
    setError('')
    try {
      await organizationAPI.createRole(organizationId, formData)
      
      showSuccess('Role created successfully')
      setIsCreating(false)
      setFormData({ name: '', display_name: '', description: '', permissions: [] })
      fetchRoles()
    } catch (error: any) {
      showError(error.response?.data?.message || 'Failed to create role')
    } finally {
      setLoading(false)
    }
  }

  const handleUpdateRole = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!organizationId || !selectedRole) return
    
    setLoading(true)
    setError('')
    try {
      await organizationAPI.updateRole(organizationId, selectedRole.id, formData)
      
      showSuccess('Role updated successfully')
      setSelectedRole(null)
      setFormData({ name: '', display_name: '', description: '', permissions: [] })
      fetchRoles()
    } catch (error: any) {
      showError(error.response?.data?.message || 'Failed to update role')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteRole = async (roleId: string) => {
    if (!organizationId) return
    
    setLoading(true)
    setError('')
    try {
      await organizationAPI.deleteRole(organizationId, roleId)
      showSuccess('Role deleted successfully')
      setShowDeleteModal(false)
      setRoleToDelete(null)
      fetchRoles()
    } catch (error: any) {
      showError(error.response?.data?.message || 'Failed to delete role')
      setShowDeleteModal(false)
      setRoleToDelete(null)
    } finally {
      setLoading(false)
    }
  }

  const confirmDeleteRole = (role: Role) => {
    setRoleToDelete(role)
    setShowDeleteModal(true)
  }

  const handleEditRole = (role: Role) => {
    setSelectedRole(role)
    setFormData({
      name: role.name,
      display_name: role.display_name,
      description: role.description,
      permissions: role.permissions || []
    })
    setIsCreating(false)
  }

  const handleTogglePermission = (permissionName: string) => {
    setFormData(prev => ({
      ...prev,
      permissions: prev.permissions.includes(permissionName)
        ? prev.permissions.filter(p => p !== permissionName)
        : [...prev.permissions, permissionName]
    }))
  }

  const handleToggleCategory = (categoryPermissions: Permission[]) => {
    const categoryPermissionNames = categoryPermissions.map(p => p.name)
    
    const allCategorySelected = categoryPermissionNames.every(name => 
      formData.permissions.includes(name)
    )

    if (allCategorySelected) {
      // Uncheck all in this category
      setFormData(prev => ({
        ...prev,
        permissions: prev.permissions.filter(p => !categoryPermissionNames.includes(p))
      }))
    } else {
      // Check all in this category  
      setFormData(prev => {
        const combinedPermissions = [...prev.permissions, ...categoryPermissionNames]
        return {
          ...prev,
          permissions: Array.from(new Set(combinedPermissions))
        }
      })
    }
  }

  const getCategoryCheckState = (categoryPermissions: Permission[]) => {
    const checkedCount = categoryPermissions.filter(p => 
      formData.permissions.includes(p.name)
    ).length
    
    if (checkedCount === 0) return 'unchecked'
    if (checkedCount === categoryPermissions.length) return 'checked'
    return 'indeterminate'
  }

  // Permission management functions
  const handleCreatePermission = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!organizationId) return

    setLoading(true)
    try {
      const permission = {
        name: permissionFormData.name.toLowerCase().replace(/\s+/g, '_'),
        display_name: permissionFormData.display_name,
        description: permissionFormData.description,
        category: permissionFormData.category.toLowerCase().replace(/\s+/g, '_')
      }

      await organizationAPI.createPermission(organizationId, permission)
      showSuccess('Permission created successfully')
      await fetchPermissions()
      setShowPermissionForm(false)
      setPermissionFormData({ name: '', display_name: '', description: '', category: '' })
    } catch (error: any) {
      console.error('Error creating permission:', error)
      showError(error.response?.data?.message || 'Failed to create permission')
    } finally {
      setLoading(false)
    }
  }

  const handleUpdatePermission = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!organizationId || !selectedPermission) return

    setLoading(true)
    try {
      const permission = {
        display_name: permissionFormData.display_name,
        description: permissionFormData.description,
        category: permissionFormData.category
      }

      await organizationAPI.updatePermission(organizationId, selectedPermission.id, permission)
      showSuccess('Permission updated successfully')
      await fetchPermissions()
      setShowPermissionForm(false)
      setSelectedPermission(null)
      setPermissionFormData({ name: '', display_name: '', description: '', category: '' })
    } catch (error: any) {
      console.error('Error updating permission:', error)
      showError(error.response?.data?.message || 'Failed to update permission')
    } finally {
      setLoading(false)
    }
  }

  const handleDeletePermission = (permission: Permission) => {
    if (!organizationId || permission.is_system) return
    setPermissionToDelete(permission)
    setShowDeletePermissionModal(true)
  }

  const confirmDeletePermission = async () => {
    if (!organizationId || !permissionToDelete) return

    setLoading(true)
    try {
      await organizationAPI.deletePermission(organizationId, permissionToDelete.id)
      showSuccess('Permission deleted successfully')
      await fetchPermissions()
      setShowDeletePermissionModal(false)
      setPermissionToDelete(null)
    } catch (error: any) {
      console.error('Error deleting permission:', error)
      showError(error.response?.data?.message || 'Failed to delete permission')
    } finally {
      setLoading(false)
    }
  }

  const handleEditPermission = (permission: Permission) => {
    if (permission.is_system) return
    
    setSelectedPermission(permission)
    setPermissionFormData({
      name: permission.name,
      display_name: permission.display_name,
      description: permission.description,
      category: permission.category
    })
    setIsCreatingPermission(false)
    setShowPermissionForm(true)
  }

  const groupedPermissions = permissions.reduce((acc, permission) => {
    if (!acc[permission.category]) {
      acc[permission.category] = []
    }
    acc[permission.category].push(permission)
    return acc
  }, {} as Record<string, Permission[]>)

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      {/* Toast Container */}
      <ToastContainer toasts={toasts} removeToast={removeToast} />
      
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Role Management</h1>
          <p className="text-gray-600 mt-2">Manage roles and permissions for your organization</p>
        </div>

        {/* Error Alert */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg flex items-start gap-3">
            <AlertTriangle className="h-5 w-5 mt-0.5 flex-shrink-0" />
            <div>
              <p className="font-medium">Error</p>
              <p className="text-sm">{error}</p>
            </div>
            <button
              onClick={() => setError('')}
              className="ml-auto p-1 hover:bg-red-100 rounded"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Roles List */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-gray-900">Roles</h2>
              <RequirePermission permission="role:create">
                <Button
                  onClick={() => {
                    setIsCreating(true)
                    setSelectedRole(null)
                    setFormData({ name: '', display_name: '', description: '', permissions: [] })
                  }}
                  className="bg-blue-600 hover:bg-blue-700"
                >
                  <Plus className="h-4 w-4 mr-2" />
                  New Role
                </Button>
              </RequirePermission>
            </div>

            <div className="space-y-3">
              {roles.map((role) => (
                <div
                  key={role.id}
                  className={`p-4 rounded-lg border-2 transition-all ${
                    selectedRole?.id === role.id
                      ? 'border-blue-600 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <Shield className="h-5 w-5 text-gray-600" />
                        <h3 className="font-semibold text-gray-900">{role.display_name}</h3>
                      </div>
                      <p className="text-sm text-gray-600 mt-1">{role.description}</p>
                      <div className="flex items-center gap-4 mt-2 text-sm text-gray-500">
                        <span className="flex items-center gap-1">
                          <Users className="h-4 w-4" />
                          {role.member_count} members
                        </span>
                        <span>{role.permissions?.length || 0} permissions</span>
                      </div>
                      {role.permissions && role.permissions.length > 0 && (
                        <div className="mt-3">
                          <div className="flex flex-wrap gap-1">
                            {role.permissions.slice(0, 3).map((permission) => {
                              const perm = permissions.find(p => p.name === permission)
                              return (
                                <span
                                  key={permission}
                                  className="px-2 py-1 text-xs bg-gray-100 text-gray-600 rounded"
                                  title={perm?.description}
                                >
                                  {perm?.display_name || permission}
                                </span>
                              )
                            })}
                            {role.permissions.length > 3 && (
                              <span className="px-2 py-1 text-xs bg-gray-100 text-gray-600 rounded">
                                +{role.permissions.length - 3} more
                              </span>
                            )}
                          </div>
                        </div>
                      )}
                    </div>
                    {!role.is_system && (
                      <div className="flex gap-2">
                        <RequirePermission permission="role:update">
                          <button
                            onClick={() => handleEditRole(role)}
                            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                          >
                            <Edit className="h-4 w-4 text-gray-600" />
                          </button>
                        </RequirePermission>
                        <RequirePermission permission="role:delete">
                          <button
                            onClick={() => confirmDeleteRole(role)}
                            className="p-2 hover:bg-red-50 rounded-lg transition-colors"
                          >
                            <Trash2 className="h-4 w-4 text-red-600" />
                          </button>
                        </RequirePermission>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Role Form */}
          {(isCreating || selectedRole) && (
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-6">
                {isCreating ? 'Create New Role' : 'Edit Role'}
              </h2>

              <form onSubmit={isCreating ? handleCreateRole : handleUpdateRole} className="space-y-6">
                <div>
                  <Label htmlFor="name">Role Name (Internal)</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value.toLowerCase().replace(/\s+/g, '_') })}
                    placeholder="e.g., issuer"
                    required
                    disabled={!!selectedRole}
                  />
                  <p className="text-xs text-gray-500 mt-1">Lowercase, no spaces (automatically formatted)</p>
                </div>

                <div>
                  <Label htmlFor="display_name">Display Name</Label>
                  <Input
                    id="display_name"
                    value={formData.display_name}
                    onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                    placeholder="e.g., Issuer"
                    required
                  />
                </div>

                <div>
                  <Label htmlFor="description">Description</Label>
                  <Input
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder="Describe what this role can do"
                  />
                </div>

                <div>
                  <Label className="mb-3 block">Permissions</Label>
                  
                  <div className="max-h-96 overflow-y-auto space-y-4 border rounded-lg p-4">
                    {Object.entries(groupedPermissions).map(([category, perms]) => {
                      // Backend should already return only available permissions for this organization
                      const categoryState = getCategoryCheckState(perms)
                      return (
                        <div key={category}>
                          <div className="flex items-center gap-2 mb-2">
                            <button
                              type="button"
                              onClick={() => handleToggleCategory(perms)}
                              className="flex items-center gap-2 hover:bg-gray-100 p-1 rounded"
                            >
                              {categoryState === 'checked' && <CheckSquare className="h-4 w-4 text-blue-600" />}
                              {categoryState === 'unchecked' && <Square className="h-4 w-4 text-gray-400" />}
                              {categoryState === 'indeterminate' && <Minus className="h-4 w-4 text-blue-600 bg-blue-600 text-white rounded-sm" />}
                              <h4 className="font-medium capitalize text-gray-900">
                                {category}
                              </h4>
                            </button>
                          </div>
                          <div className="space-y-2 ml-6">
                          {perms.map((permission) => (
                            <label
                              key={permission.id}
                              className="flex items-start gap-3 p-2 hover:bg-gray-50 rounded cursor-pointer"
                            >
                              <button
                                type="button"
                                onClick={() => handleTogglePermission(permission.name)}
                                className="mt-0.5"
                              >
                                {formData.permissions.includes(permission.name) ? (
                                  <CheckSquare className="h-5 w-5 text-blue-600" />
                                ) : (
                                  <Square className="h-5 w-5 text-gray-400" />
                                )}
                              </button>
                              <div className="flex-1">
                                <div className="flex items-center gap-2">
                                  <p className="font-medium text-sm text-gray-900">
                                    {permission.display_name}
                                  </p>
                                  {permission.is_system ? (
                                    <span className="px-1.5 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">
                                      System
                                    </span>
                                  ) : (
                                    <span className="px-1.5 py-0.5 text-xs bg-blue-100 text-blue-600 rounded">
                                      Custom
                                    </span>
                                  )}
                                </div>
                                <p className="text-xs text-gray-500">
                                  {permission.description}
                                </p>
                              </div>
                            </label>
                          ))}
                        </div>
                      </div>
                      )
                    })}
                  </div>
                </div>

                <div className="flex gap-3 pt-4">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setIsCreating(false)
                      setSelectedRole(null)
                      setFormData({ name: '', display_name: '', description: '', permissions: [] })
                    }}
                    className="flex-1"
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={loading}
                    className="flex-1 bg-blue-600 hover:bg-blue-700"
                  >
                    {loading ? 'Saving...' : isCreating ? 'Create Role' : 'Update Role'}
                  </Button>
                </div>
              </form>
            </div>
          )}
        </div>

        {/* Permission Management Section */}
        <div className="bg-white rounded-lg shadow p-6 mt-8">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Manage Permissions</h2>
              <p className="text-gray-600 text-sm mt-1">Create and manage custom permissions for your organization</p>
            </div>
            <RequirePermission permission="role:create">
              <Button
                onClick={() => {
                  setIsCreatingPermission(true)
                  setSelectedPermission(null)
                  setPermissionFormData({ name: '', display_name: '', description: '', category: '' })
                  setShowPermissionForm(true)
                }}
                className="bg-green-600 hover:bg-green-700"
              >
                <Plus className="h-4 w-4 mr-2" />
                New Permission
              </Button>
            </RequirePermission>
          </div>

          {!showPermissionForm ? (
            <div className="space-y-4">
              {(() => {
                // Only show truly custom permissions (created by organization) in management section
                const customCategories = Object.entries(groupedPermissions)
                  .map(([category, perms]) => ({
                    category,
                    permissions: perms.filter(p => !p.is_system && p.organization_id)
                  }))
                  .filter(({ permissions }) => permissions.length > 0)
                
                if (customCategories.length === 0) {
                  return (
                    <div className="text-center py-8">
                      <div className="text-gray-500">
                        <Shield className="h-12 w-12 mx-auto mb-3 text-gray-300" />
                        <p className="font-medium">No Custom Permissions</p>
                        <p className="text-sm mt-1">You haven't created any custom permissions for your organization yet.</p>
                        <p className="text-sm text-gray-400 mt-2">System permissions are available for role assignment but cannot be managed here.</p>
                      </div>
                    </div>
                  )
                }
                
                return customCategories.map(({ category, permissions: customPermissions }) => (
                  <div key={category} className="border rounded-lg p-4">
                    <h3 className="font-medium text-gray-900 mb-3 capitalize">{category}</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                      {customPermissions.map((permission) => (
                      <div
                        key={permission.id}
                        className={`p-3 border rounded-lg ${
                          permission.is_system 
                            ? 'bg-gray-50 border-gray-200' 
                            : 'bg-white border-gray-200 hover:border-gray-300'
                        }`}
                      >
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <div className="flex items-center gap-2">
                              <p className={`font-medium text-sm ${permission.is_system ? 'text-gray-600' : 'text-gray-900'}`}>
                                {permission.display_name}
                              </p>
                              {permission.is_system && (
                                <div title="System permission - cannot be modified">
                                  <Lock className="h-3 w-3 text-gray-400" />
                                </div>
                              )}
                            </div>
                            <p className="text-xs text-gray-500 mt-1">{permission.description}</p>
                            <p className="text-xs text-gray-400 mt-1 font-mono">{permission.name}</p>
                          </div>
                          {!permission.is_system && (
                            <div className="flex gap-1">
                              <RequirePermission permission="role:update">
                                <button
                                  onClick={() => handleEditPermission(permission)}
                                  className="p-1 hover:bg-gray-100 rounded transition-colors"
                                >
                                  <Edit className="h-3 w-3 text-gray-600" />
                                </button>
                              </RequirePermission>
                              <RequirePermission permission="role:delete">
                                <button
                                  onClick={() => handleDeletePermission(permission)}
                                  className="p-1 hover:bg-red-100 rounded transition-colors"
                                >
                                  <Trash2 className="h-3 w-3 text-red-600" />
                                </button>
                              </RequirePermission>
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                    </div>
                  </div>
                ))
              })()}
            </div>
          ) : (
            <form onSubmit={isCreatingPermission ? handleCreatePermission : handleUpdatePermission} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="perm-name">Permission Name</Label>
                  <Input
                    id="perm-name"
                    value={permissionFormData.name}
                    onChange={(e) => setPermissionFormData({ ...permissionFormData, name: e.target.value })}
                    placeholder="e.g., custom_action"
                    required
                    disabled={!isCreatingPermission}
                  />
                  <p className="text-xs text-gray-500 mt-1">Will be formatted as category:name (e.g., custom:manage_data)</p>
                </div>

                <div>
                  <Label htmlFor="perm-display-name">Display Name</Label>
                  <Input
                    id="perm-display-name"
                    value={permissionFormData.display_name}
                    onChange={(e) => setPermissionFormData({ ...permissionFormData, display_name: e.target.value })}
                    placeholder="e.g., Manage Custom Data"
                    required
                  />
                  <p className="text-xs text-gray-500 mt-1">Must be unique within the same category</p>
                </div>

                <div>
                  <Label htmlFor="perm-category">Category</Label>
                  <Input
                    id="perm-category"
                    value={permissionFormData.category}
                    onChange={(e) => setPermissionFormData({ ...permissionFormData, category: e.target.value.toLowerCase() })}
                    placeholder="e.g., custom"
                    required
                  />
                  <p className="text-xs text-gray-500 mt-1">Groups related permissions together (e.g., custom, billing, reports)</p>
                </div>

                <div>
                  <Label htmlFor="perm-description">Description</Label>
                  <Input
                    id="perm-description"
                    value={permissionFormData.description}
                    onChange={(e) => setPermissionFormData({ ...permissionFormData, description: e.target.value })}
                    placeholder="Describe what this permission allows"
                  />
                </div>
              </div>

              <div className="flex gap-3 pt-4">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setShowPermissionForm(false)
                    setSelectedPermission(null)
                    setPermissionFormData({ name: '', display_name: '', description: '', category: '' })
                  }}
                  className="flex-1"
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={loading}
                  className="flex-1 bg-green-600 hover:bg-green-700"
                >
                  {loading ? 'Saving...' : isCreatingPermission ? 'Create Permission' : 'Update Permission'}
                </Button>
              </div>
            </form>
          )}
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      {showDeleteModal && roleToDelete && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50"
            onClick={() => {
              setShowDeleteModal(false)
              setRoleToDelete(null)
            }}
          />

          {/* Modal */}
          <div className="fixed inset-0 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full overflow-hidden">
              {/* Header */}
              <div className="bg-gradient-to-r from-red-600 to-red-700 p-6 text-white">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-white/20 rounded-lg">
                      <AlertTriangle className="h-6 w-6" />
                    </div>
                    <div>
                      <h2 className="text-xl font-bold">Delete Role</h2>
                      <p className="text-red-100 text-sm">This action cannot be undone</p>
                    </div>
                  </div>
                  <button
                    onClick={() => {
                      setShowDeleteModal(false)
                      setRoleToDelete(null)
                    }}
                    className="p-1 hover:bg-white/20 rounded-lg transition-colors"
                  >
                    <X className="h-5 w-5" />
                  </button>
                </div>
              </div>

              {/* Content */}
              <div className="p-6">
                <p className="text-gray-700 mb-4">
                  Are you sure you want to delete the role{' '}
                  <span className="font-semibold text-gray-900">"{roleToDelete.display_name}"</span>?
                </p>
                {roleToDelete.member_count > 0 && (
                  <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
                    <p className="text-sm text-yellow-800">
                      <strong>Warning:</strong> This role currently has {roleToDelete.member_count}{' '}
                      {roleToDelete.member_count === 1 ? 'member' : 'members'} assigned to it.
                    </p>
                  </div>
                )}
                <p className="text-sm text-gray-600">
                  This will permanently delete the role and all associated permissions.
                </p>
              </div>

              {/* Actions */}
              <div className="bg-gray-50 px-6 py-4 flex gap-3">
                <Button
                  variant="outline"
                  onClick={() => {
                    setShowDeleteModal(false)
                    setRoleToDelete(null)
                  }}
                  className="flex-1"
                  disabled={loading}
                >
                  Cancel
                </Button>
                <Button
                  onClick={() => handleDeleteRole(roleToDelete.id)}
                  className="flex-1 bg-red-600 hover:bg-red-700"
                  disabled={loading}
                >
                  {loading ? 'Deleting...' : 'Delete Role'}
                </Button>
              </div>
            </div>
          </div>
        </>
      )}

      {/* Delete Permission Confirmation Modal */}
      {showDeletePermissionModal && permissionToDelete && (
        <>
          <div className="fixed inset-0 bg-black bg-opacity-50 z-40" onClick={() => {
            setShowDeletePermissionModal(false)
            setPermissionToDelete(null)
          }} />
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div className="bg-white rounded-lg shadow-xl w-full max-w-md">
              {/* Header */}
              <div className="bg-red-600 text-white p-4 rounded-t-lg">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-white/20 rounded-lg">
                      <AlertTriangle className="h-6 w-6" />
                    </div>
                    <div>
                      <h2 className="text-xl font-bold">Delete Permission</h2>
                      <p className="text-red-100 text-sm">This action cannot be undone</p>
                    </div>
                  </div>
                  <button
                    onClick={() => {
                      setShowDeletePermissionModal(false)
                      setPermissionToDelete(null)
                    }}
                    className="p-1 hover:bg-white/20 rounded-lg transition-colors"
                  >
                    <X className="h-5 w-5" />
                  </button>
                </div>
              </div>

              {/* Content */}
              <div className="p-6">
                <p className="text-gray-700 mb-4">
                  Are you sure you want to delete the permission{' '}
                  <span className="font-semibold text-gray-900">"{permissionToDelete.display_name}"</span>?
                </p>
                <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-4">
                  <p className="text-sm text-gray-600">
                    <strong>Permission Name:</strong> <code className="bg-gray-200 px-1 rounded">{permissionToDelete.name}</code>
                  </p>
                  <p className="text-sm text-gray-600 mt-1">
                    <strong>Category:</strong> {permissionToDelete.category}
                  </p>
                </div>
                <p className="text-sm text-gray-600">
                  This will permanently delete the permission and remove it from any roles that have it assigned.
                </p>
              </div>

              {/* Actions */}
              <div className="bg-gray-50 px-6 py-4 flex gap-3 rounded-b-lg">
                <Button
                  variant="outline"
                  onClick={() => {
                    setShowDeletePermissionModal(false)
                    setPermissionToDelete(null)
                  }}
                  className="flex-1"
                  disabled={loading}
                >
                  Cancel
                </Button>
                <Button
                  onClick={confirmDeletePermission}
                  className="flex-1 bg-red-600 hover:bg-red-700"
                  disabled={loading}
                >
                  {loading ? 'Deleting...' : 'Delete Permission'}
                </Button>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )
}

export default RoleManagement
