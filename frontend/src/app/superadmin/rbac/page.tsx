'use client'

import { useAuth } from '../../../contexts/auth-context'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { SuperAdminLayout } from '../../../components/layout/superadmin-layout'
import { Card } from '../../../components/ui/card'
import { Button } from '../../../components/ui/button'
import { LoadingSpinner } from '../../../components/ui/loading-spinner'
import { 
  Shield, 
  Users, 
  Key, 
  Plus, 
  Edit, 
  Trash2, 
  Search,
  Filter,
  MoreVertical,
  CheckCircle,
  XCircle
} from 'lucide-react'

interface Role {
  id: string
  name: string
  description: string
  permissions: string[]
  is_system: boolean
  organization_id?: string
  created_at: string
  updated_at: string
  user_count: number
}

interface Permission {
  id: string
  name: string
  description: string
  resource: string
  action: string
  is_system: boolean
  organization_isolation: boolean
}

export default function SuperAdminRBACPage() {
  const { user, loading } = useAuth()
  const router = useRouter()
  const [activeTab, setActiveTab] = useState<'roles' | 'permissions'>('roles')
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [loadingData, setLoadingData] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    if (!loading && !user) {
      router.push('/auth/login')
      return
    }

    if (!loading && user && !user.is_superadmin) {
      router.push('/user/dashboard')
      return
    }
  }, [user, loading, router])

  useEffect(() => {
    if (user && user.is_superadmin) {
      loadRBACData()
    }
  }, [user])

  const loadRBACData = async () => {
    setLoadingData(true)
    try {
      // Mock data for now - replace with actual API calls
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      setRoles([
        {
          id: '1',
          name: 'System Administrator',
          description: 'Full system access with all permissions',
          permissions: ['user:create', 'user:read', 'user:update', 'user:delete', 'org:manage', 'system:admin'],
          is_system: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
          user_count: 3
        },
        {
          id: '2',
          name: 'Organization Admin',
          description: 'Manage organization users and settings',
          permissions: ['user:create', 'user:read', 'user:update', 'org:read', 'org:update'],
          is_system: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
          user_count: 15
        },
        {
          id: '3',
          name: 'Regular User',
          description: 'Standard user with basic permissions',
          permissions: ['user:read', 'profile:update'],
          is_system: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
          user_count: 1227
        }
      ])

      setPermissions([
        {
          id: '1',
          name: 'user:create',
          description: 'Create new users',
          resource: 'user',
          action: 'create',
          is_system: true,
          organization_isolation: true
        },
        {
          id: '2',
          name: 'user:read',
          description: 'View user information',
          resource: 'user',
          action: 'read',
          is_system: true,
          organization_isolation: true
        },
        {
          id: '3',
          name: 'user:update',
          description: 'Update user information',
          resource: 'user',
          action: 'update',
          is_system: true,
          organization_isolation: true
        },
        {
          id: '4',
          name: 'user:delete',
          description: 'Delete users',
          resource: 'user',
          action: 'delete',
          is_system: true,
          organization_isolation: true
        },
        {
          id: '5',
          name: 'org:manage',
          description: 'Manage organization settings',
          resource: 'organization',
          action: 'manage',
          is_system: true,
          organization_isolation: false
        },
        {
          id: '6',
          name: 'system:admin',
          description: 'System administration access',
          resource: 'system',
          action: 'admin',
          is_system: true,
          organization_isolation: false
        }
      ])
    } catch (error) {
      console.error('Failed to load RBAC data:', error)
    } finally {
      setLoadingData(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <LoadingSpinner size="lg" variant="primary" />
          <p className="mt-4 text-muted-foreground">Loading...</p>
        </div>
      </div>
    )
  }

  if (!user || !user.is_superadmin) {
    return null
  }

  const filteredRoles = roles.filter(role => 
    role.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    role.description.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const filteredPermissions = permissions.filter(permission => 
    permission.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    permission.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
    permission.resource.toLowerCase().includes(searchTerm.toLowerCase())
  )

  return (
    <SuperAdminLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-3xl font-bold text-foreground">RBAC Management</h1>
            <p className="text-muted-foreground mt-2">
              Manage system roles and permissions. Custom roles are managed by organization owners.
            </p>
          </div>
          <Button disabled className="flex items-center bg-muted text-muted-foreground cursor-not-allowed">
            <Plus className="h-4 w-4 mr-2" />
            Create Role
          </Button>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <Card className="p-6">
            <div className="flex items-center">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Shield className="h-6 w-6 text-primary" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-muted-foreground">Total Roles</p>
                <p className="text-2xl font-bold text-foreground">{roles.length}</p>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center">
              <div className="p-2 bg-success/20 rounded-lg">
                <Key className="h-6 w-6 text-success" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-muted-foreground">Permissions</p>
                <p className="text-2xl font-bold text-foreground">{permissions.length}</p>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Users className="h-6 w-6 text-primary" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-muted-foreground">Assigned Users</p>
                <p className="text-2xl font-bold text-foreground">
                  {roles.reduce((sum, role) => sum + role.user_count, 0)}
                </p>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Shield className="h-6 w-6 text-primary" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-muted-foreground">System Roles</p>
                <p className="text-2xl font-bold text-foreground">
                  {roles.filter(role => role.is_system).length}
                </p>
              </div>
            </div>
          </Card>
        </div>

        {/* Tabs */}
        <div className="border-b border-border">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('roles')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'roles'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <Shield className="h-4 w-4 inline mr-2" />
              Roles
            </button>
            <button
              onClick={() => setActiveTab('permissions')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'permissions'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <Key className="h-4 w-4 inline mr-2" />
              Permissions
            </button>
          </nav>
        </div>

        {/* Search and Filters */}
        <div className="flex justify-between items-center">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <input
              type="text"
              placeholder={`Search ${activeTab}...`}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-white border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
            />
          </div>
          <Button variant="outline" className="flex items-center">
            <Filter className="h-4 w-4 mr-2" />
            Filter
          </Button>
        </div>

        {/* Content */}
        {loadingData ? (
          <Card className="p-12">
            <div className="text-center">
              <LoadingSpinner size="lg" variant="primary" />
              <p className="mt-4 text-muted-foreground">Loading RBAC data...</p>
            </div>
          </Card>
        ) : (
          <>
            {activeTab === 'roles' && (
              <Card className="p-6">
                <div className="space-y-4">
                  {filteredRoles.map((role) => (
                    <div key={role.id} className="border border-border rounded-lg p-4 hover:bg-muted/50">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-3">
                            <h3 className="text-lg font-semibold text-foreground">{role.name}</h3>
                            {role.is_system ? (
                              <span className="px-2 py-1 text-xs font-medium bg-primary/10 text-primary rounded-full">
                                System
                              </span>
                            ) : (
                              <span className="px-2 py-1 text-xs font-medium bg-success/20 text-success rounded-full">
                                Custom
                              </span>
                            )}
                          </div>
                          <p className="text-muted-foreground mt-1">{role.description}</p>
                          <div className="flex items-center space-x-4 mt-3 text-sm text-muted-foreground">
                            <span className="flex items-center">
                              <Users className="h-4 w-4 mr-1" />
                              {role.user_count} users
                            </span>
                            <span className="flex items-center">
                              <Key className="h-4 w-4 mr-1" />
                              {role.permissions.length} permissions
                            </span>
                          </div>
                          <div className="flex flex-wrap gap-1 mt-3">
                            {role.permissions.slice(0, 3).map((permission) => (
                              <span
                                key={permission}
                                className="px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded"
                              >
                                {permission}
                              </span>
                            ))}
                            {role.permissions.length > 3 && (
                              <span className="px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded">
                                +{role.permissions.length - 3} more
                              </span>
                            )}
                          </div>
                        </div>
                        <div className="flex items-center space-x-2">
                          <Button variant="outline" size="sm">
                            <Edit className="h-4 w-4" />
                          </Button>
                          {!role.is_system && (
                            <Button variant="outline" size="sm" className="text-destructive">
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          )}
                          <Button variant="outline" size="sm">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            )}

            {activeTab === 'permissions' && (
              <Card className="p-6">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-border">
                    <thead className="bg-muted/50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Permission
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Resource
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Action
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Type
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Org Isolation
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-border">
                      {filteredPermissions.map((permission) => (
                        <tr key={permission.id} className="hover:bg-muted/50">
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div>
                              <div className="text-sm font-medium text-foreground">
                                {permission.name}
                              </div>
                              <div className="text-sm text-muted-foreground">
                                {permission.description}
                              </div>
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-foreground">
                            <span className="px-2 py-1 text-xs bg-primary/10 text-primary rounded">
                              {permission.resource}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-foreground">
                            {permission.action}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            {permission.is_system ? (
                              <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-primary/10 text-primary">
                                System
                              </span>
                            ) : (
                              <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-success/20 text-success">
                                Custom
                              </span>
                            )}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-foreground">
                            {permission.organization_isolation ? (
                              <CheckCircle className="h-5 w-5 text-success" />
                            ) : (
                              <XCircle className="h-5 w-5 text-muted-foreground" />
                            )}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                            <div className="flex items-center space-x-2">
                              <Button variant="outline" size="sm">
                                <Edit className="h-4 w-4" />
                              </Button>
                              {!permission.is_system && (
                                <Button variant="outline" size="sm" className="text-destructive">
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              )}
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </Card>
            )}
          </>
        )}
      </div>
    </SuperAdminLayout>
  )
}