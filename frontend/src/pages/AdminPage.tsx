/**
 * Admin Page - Organization management, user administration, role management
 * Available only to users with admin permissions
 */

import React, { useState, useEffect } from 'react'
import { useAuthStore } from '../stores/authStore'
import { useTenantStore } from '../stores/tenantStore'
import LoadingSpinner from '../components/ui/loading-spinner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Badge } from '../components/ui/badge'
import { Alert, AlertDescription } from '../components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs'
import { 
  Users, 
  Shield, 
  Settings, 
  Activity,
  Plus,
  Edit,
  Trash2,
  Crown,
  AlertTriangle,
  CheckCircle,
  Mail,
  Key
} from 'lucide-react'
import { toast } from 'react-hot-toast'

interface Role {
  id: string
  name: string
  description: string
  permissions: string[]
  userCount: number
  isSystem: boolean
}

interface AuditLog {
  id: string
  action: string
  user: string
  target?: string
  timestamp: string
  ipAddress: string
  details: string
}

export function AdminPage() {
  const { user, hasPermission } = useAuthStore()
  const { currentTenant, members, loadMembers, inviteMember, removeMember, updateMemberRole } = useTenantStore()
  const [loading, setLoading] = useState(false)
  const [roles, setRoles] = useState<Role[]>([])
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [inviteEmail, setInviteEmail] = useState('')
  const [selectedRole, setSelectedRole] = useState('member')

  useEffect(() => {
    loadAdminData()
  }, [currentTenant?.id])

  const loadAdminData = async () => {
    if (!currentTenant?.id) return

    try {
      setLoading(true)
      await Promise.all([
        loadMembers(),
        loadRoles(),
        loadAuditLogs()
      ])
    } catch (error) {
      toast.error('Failed to load admin data')
    } finally {
      setLoading(false)
    }
  }

  const loadRoles = async () => {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 500))
    
    setRoles([
      {
        id: '1',
        name: 'Super Admin',
        description: 'Full system access with all permissions',
        permissions: ['*'],
        userCount: 1,
        isSystem: true
      },
      {
        id: '2',
        name: 'Admin',
        description: 'Administrative access to organization',
        permissions: ['users:*', 'roles:read', 'organization:*'],
        userCount: 2,
        isSystem: true
      },
      {
        id: '3',
        name: 'Manager',
        description: 'Can manage users and view reports',
        permissions: ['users:read', 'users:create', 'users:update', 'reports:read'],
        userCount: 3,
        isSystem: false
      },
      {
        id: '4',
        name: 'Member',
        description: 'Basic access to organization resources',
        permissions: ['profile:read', 'profile:update'],
        userCount: members?.length || 5,
        isSystem: true
      }
    ])
  }

  const loadAuditLogs = async () => {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 300))
    
    setAuditLogs([
      {
        id: '1',
        action: 'User Invited',
        user: 'john.admin@example.com',
        target: 'new.user@example.com',
        timestamp: '2024-01-20T10:30:00Z',
        ipAddress: '192.168.1.100',
        details: 'Invited with Manager role'
      },
      {
        id: '2',
        action: 'Role Updated',
        user: 'john.admin@example.com',
        target: 'jane.doe@example.com',
        timestamp: '2024-01-20T09:15:00Z',
        ipAddress: '192.168.1.100',
        details: 'Changed from Member to Manager'
      },
      {
        id: '3',
        action: 'User Removed',
        user: 'john.admin@example.com',
        target: 'old.user@example.com',
        timestamp: '2024-01-19T16:20:00Z',
        ipAddress: '192.168.1.100',
        details: 'User account deactivated'
      }
    ])
  }

  const handleInviteUser = async () => {
    if (!inviteEmail || !selectedRole) {
      toast.error('Please provide email and role')
      return
    }

    try {
      await inviteMember(inviteEmail, selectedRole)
      setInviteEmail('')
      setSelectedRole('member')
      toast.success('User invited successfully')
      loadMembers()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to invite user')
    }
  }

  const handleRemoveUser = async (userId: string, userName: string) => {
    if (!confirm(`Are you sure you want to remove ${userName}?`)) {
      return
    }

    try {
      await removeMember(userId)
      toast.success('User removed successfully')
      loadMembers()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to remove user')
    }
  }

  const handleUpdateRole = async (userId: string, newRole: string, userName: string) => {
    try {
      await updateMemberRole(userId, newRole)
      toast.success(`Updated ${userName}'s role to ${newRole}`)
      loadMembers()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to update role')
    }
  }

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString()
  }

  const getRoleColor = (role: string) => {
    switch (role.toLowerCase()) {
      case 'super admin':
        return 'bg-red-100 text-red-800 border-red-200'
      case 'admin':
        return 'bg-purple-100 text-purple-800 border-purple-200'
      case 'manager':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  if (!hasPermission('users:read')) {
    return (
      <div className="p-6">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            You don't have permission to access admin features.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Organization Administration</h1>
        <p className="text-gray-500 mt-1">
          Manage users, roles, and organization settings for {currentTenant?.name}
        </p>
      </div>

      <Tabs defaultValue="users" className="space-y-6">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="users" className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            Users
          </TabsTrigger>
          <TabsTrigger value="roles" className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            Roles
          </TabsTrigger>
          <TabsTrigger value="audit" className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Audit Logs
          </TabsTrigger>
          <TabsTrigger value="settings" className="flex items-center gap-2">
            <Settings className="h-4 w-4" />
            Settings
          </TabsTrigger>
        </TabsList>

        <TabsContent value="users">
          <div className="space-y-6">
            {/* Invite User Section */}
            {hasPermission('users:create') && (
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Plus className="h-5 w-5" />
                    Invite New User
                  </CardTitle>
                  <CardDescription>
                    Add a new member to your organization
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex gap-4 items-end">
                    <div className="flex-1">
                      <label htmlFor="invite-email" className="block text-sm font-medium mb-2">
                        Email Address
                      </label>
                      <input
                        id="invite-email"
                        type="email"
                        value={inviteEmail}
                        onChange={(e) => setInviteEmail(e.target.value)}
                        placeholder="user@example.com"
                        className="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label htmlFor="invite-role" className="block text-sm font-medium mb-2">
                        Role
                      </label>
                      <select
                        id="invite-role"
                        value={selectedRole}
                        onChange={(e) => setSelectedRole(e.target.value)}
                        className="p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      >
                        {roles.filter(role => !role.isSystem || role.name === 'Member').map(role => (
                          <option key={role.id} value={role.name.toLowerCase()}>
                            {role.name}
                          </option>
                        ))}
                      </select>
                    </div>
                    <Button onClick={handleInviteUser} disabled={!inviteEmail || !selectedRole}>
                      <Mail className="h-4 w-4 mr-2" />
                      Send Invite
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Users List */}
            <Card>
              <CardHeader>
                <CardTitle>Organization Members</CardTitle>
                <CardDescription>
                  Manage user roles and permissions
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {members?.map((member) => (
                    <div key={member.id} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                          <User className="h-5 w-5 text-blue-600" />
                        </div>
                        <div>
                          <h4 className="font-medium">{member.name || member.email}</h4>
                          <p className="text-sm text-gray-500">{member.email}</p>
                          <div className="flex items-center gap-2 mt-1">
                            <Badge className={getRoleColor(member.role)}>
                              {member.role}
                            </Badge>
                            {member.id === user?.id && (
                              <Badge variant="outline" className="text-green-600 border-green-200">
                                You
                              </Badge>
                            )}
                          </div>
                        </div>
                      </div>
                      
                      {hasPermission('users:update') && member.id !== user?.id && (
                        <div className="flex items-center gap-2">
                          <select
                            value={member.role.toLowerCase()}
                            onChange={(e) => handleUpdateRole(member.id, e.target.value, member.name || member.email)}
                            className="text-sm p-1 border border-gray-300 rounded"
                          >
                            {roles.filter(role => !role.isSystem || role.name === 'Member').map(role => (
                              <option key={role.id} value={role.name.toLowerCase()}>
                                {role.name}
                              </option>
                            ))}
                          </select>
                          
                          {hasPermission('users:delete') && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleRemoveUser(member.id, member.name || member.email)}
                              className="text-red-600 hover:text-red-700"
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="roles">
          <Card>
            <CardHeader>
              <CardTitle>Role Management</CardTitle>
              <CardDescription>
                View and manage organization roles and permissions
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {roles.map((role) => (
                  <div key={role.id} className="p-4 border rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium">{role.name}</h4>
                        {role.isSystem && (
                          <Badge variant="outline" className="text-blue-600 border-blue-200">
                            <Crown className="h-3 w-3 mr-1" />
                            System Role
                          </Badge>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary">{role.userCount} users</Badge>
                        {hasPermission('roles:update') && !role.isSystem && (
                          <Button variant="outline" size="sm">
                            <Edit className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </div>
                    <p className="text-sm text-gray-600 mb-3">{role.description}</p>
                    <div className="flex flex-wrap gap-1">
                      {role.permissions.slice(0, 5).map((permission, index) => (
                        <Badge key={index} variant="outline" className="text-xs">
                          {permission}
                        </Badge>
                      ))}
                      {role.permissions.length > 5 && (
                        <Badge variant="outline" className="text-xs">
                          +{role.permissions.length - 5} more
                        </Badge>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="audit">
          <Card>
            <CardHeader>
              <CardTitle>Audit Logs</CardTitle>
              <CardDescription>
                Track administrative actions and security events
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {auditLogs.map((log) => (
                  <div key={log.id} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <h4 className="font-medium">{log.action}</h4>
                          <Badge variant="outline" className="text-xs">
                            {log.user}
                          </Badge>
                        </div>
                        {log.target && (
                          <p className="text-sm text-gray-600 mt-1">
                            Target: {log.target}
                          </p>
                        )}
                        <p className="text-sm text-gray-500 mt-1">{log.details}</p>
                        <div className="flex items-center gap-4 mt-2 text-xs text-gray-400">
                          <span>{formatTimestamp(log.timestamp)}</span>
                          <span>IP: {log.ipAddress}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="settings">
          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Organization Settings</CardTitle>
                <CardDescription>
                  Configure organization-wide settings and policies
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-3 border rounded-lg">
                    <div>
                      <h4 className="font-medium">Require MFA for all users</h4>
                      <p className="text-sm text-gray-500">
                        Force all organization members to enable two-factor authentication
                      </p>
                    </div>
                    <input type="checkbox" className="toggle" />
                  </div>
                  
                  <div className="flex items-center justify-between p-3 border rounded-lg">
                    <div>
                      <h4 className="font-medium">Session timeout</h4>
                      <p className="text-sm text-gray-500">
                        Automatically log out users after period of inactivity
                      </p>
                    </div>
                    <select className="p-2 border border-gray-300 rounded">
                      <option value="30">30 minutes</option>
                      <option value="60">1 hour</option>
                      <option value="240">4 hours</option>
                      <option value="480">8 hours</option>
                    </select>
                  </div>
                  
                  <div className="flex items-center justify-between p-3 border rounded-lg">
                    <div>
                      <h4 className="font-medium">Password policy</h4>
                      <p className="text-sm text-gray-500">
                        Enforce minimum password requirements
                      </p>
                    </div>
                    <Button variant="outline" size="sm">
                      <Key className="h-4 w-4 mr-2" />
                      Configure
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-red-200">
              <CardHeader>
                <CardTitle className="text-red-700">Danger Zone</CardTitle>
                <CardDescription>
                  Irreversible actions that affect the entire organization
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <Alert variant="destructive">
                    <AlertTriangle className="h-4 w-4" />
                    <AlertDescription>
                      These actions are permanent and cannot be undone. Proceed with caution.
                    </AlertDescription>
                  </Alert>
                  
                  <div className="flex items-center justify-between p-3 border border-red-200 rounded-lg">
                    <div>
                      <h4 className="font-medium text-red-700">Delete Organization</h4>
                      <p className="text-sm text-gray-500">
                        Permanently delete this organization and all associated data
                      </p>
                    </div>
                    <Button variant="destructive" size="sm">
                      Delete Organization
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}