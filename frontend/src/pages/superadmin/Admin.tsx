import  { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { ArrowLeft, Users, Building, Shield, Search, UserCheck, AlertCircle, Key, Book, Activity } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import LoadingSpinner from '@/components/ui/loading-spinner'
import useAuthStore from '@/store/auth'
import { adminAPI, organizationAPI } from '@/lib/api'
import type { User, Organization } from '@/lib/types'

const Admin = () => {
  const navigate = useNavigate()
  const { isSuperadmin } = useAuthStore()

  const [users, setUsers] = useState<User[]>([])
  const [organizations, setOrganizations] = useState<Organization[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [activeTab, setActiveTab] = useState('users')

  // Redirect if not superadmin
  useEffect(() => {
    if (!isSuperadmin) {
      navigate('/dashboard', { replace: true })
    }
  }, [isSuperadmin, navigate])

  // Load data
  useEffect(() => {
    if (isSuperadmin) {
      loadData()
    }
  }, [isSuperadmin])

  const loadData = async () => {
    try {
      setLoading(true)
      setError('')

      const [usersResponse, orgsResponse] = await Promise.all([
        adminAPI.listUsers(),
        adminAPI.listOrganizations()
      ])

      console.log('Admin.tsx - API responses:', { usersResponse, orgsResponse })

      // Ensure we have arrays - API returns { users: [...], total: ... } structure for users
      const usersResponseData = usersResponse as any
      const usersData = Array.isArray(usersResponseData.users) ? usersResponseData.users :
                       (usersResponseData.data?.users || [])
      const orgsData = Array.isArray(orgsResponse.data) ? orgsResponse.data : []

      console.log('Admin.tsx - Processed data:', { usersData, orgsData })

      setUsers(usersData)
      setOrganizations(orgsData)
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to load admin data'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  const handleBack = () => {
    navigate('/dashboard')
  }

  const getInitials = (firstName?: string, lastName?: string) => {
    const first = firstName || ''
    const last = lastName || ''
    return `${first.charAt(0)}${last.charAt(0)}`.toUpperCase()
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString()
  }

  const filteredUsers = (users || []).filter(user =>
    user.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
    user.first_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    user.last_name?.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const filteredOrganizations = (organizations || []).filter(org =>
    org.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    org.slug.toLowerCase().includes(searchTerm.toLowerCase())
  )

  if (!isSuperadmin) {
    return null // Will redirect in useEffect
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      {/* Header */}
      <header className="bg-white dark:bg-slate-800 shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              <Button
                variant="ghost"
                size="sm"
                onClick={handleBack}
                className="flex items-center space-x-2"
              >
                <ArrowLeft className="h-4 w-4" />
                <span>Back to Dashboard</span>
              </Button>
            </div>
            <div className="flex items-center space-x-4">
              <Shield className="h-6 w-6 text-primary" />
              <h1 className="text-xl font-semibold text-foreground">
                Admin Panel
              </h1>
            </div>
            <div className="w-24"></div> {/* Spacer for centering */}
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="space-y-8"
        >
          {/* Error Message */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {error}
              </AlertDescription>
            </Alert>
          )}

          {/* Search */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center space-x-4">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Search users, organizations..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Users</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{users.length}</div>
                <p className="text-xs text-muted-foreground">
                  Across all organizations
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Organizations</CardTitle>
                <Building className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{organizations.length}</div>
                <p className="text-xs text-muted-foreground">
                  Active organizations
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Active Users</CardTitle>
                <UserCheck className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {users.filter(u => u.is_active).length}
                </div>
                <p className="text-xs text-muted-foreground">
                  Currently active
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
              <CardDescription>Superadmin tools and configurations</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                <Button
                  variant="outline"
                  className="justify-start h-auto py-4 px-6"
                  onClick={() => navigate('/oauth/client-apps')}
                >
                  <div className="flex items-start space-x-4">
                    <div className="p-2 bg-primary/10 rounded-lg">
                      <Key className="h-6 w-6 text-primary" />
                    </div>
                    <div className="text-left">
                      <div className="font-semibold">OAuth Client Apps</div>
                      <div className="text-sm text-muted-foreground">
                        Manage OAuth2 client applications
                      </div>
                    </div>
                  </div>
                </Button>
                <Button
                  variant="outline"
                  className="justify-start h-auto py-4 px-6"
                  onClick={() => navigate('/oauth/audit')}
                >
                  <div className="flex items-start space-x-4">
                    <div className="p-2 bg-primary/10 rounded-lg">
                      <Activity className="h-6 w-6 text-primary" />
                    </div>
                    <div className="text-left">
                      <div className="font-semibold">OAuth Audit Logs</div>
                      <div className="text-sm text-muted-foreground">
                        Monitor authorization flows and tokens
                      </div>
                    </div>
                  </div>
                </Button>
                <Button
                  variant="outline"
                  className="justify-start h-auto py-4 px-6"
                  onClick={() => navigate('/developer/docs')}
                >
                  <div className="flex items-start space-x-4">
                    <div className="p-2 bg-primary/10 rounded-lg">
                      <Book className="h-6 w-6 text-primary" />
                    </div>
                    <div className="text-left">
                      <div className="font-semibold">Developer Documentation</div>
                      <div className="text-sm text-muted-foreground">
                        OAuth2.1 integration guides and SDK docs
                      </div>
                    </div>
                  </div>
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Main Content Tabs */}
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="users" className="flex items-center space-x-2">
                <Users className="h-4 w-4" />
                <span>Users</span>
              </TabsTrigger>
              <TabsTrigger value="organizations" className="flex items-center space-x-2">
                <Building className="h-4 w-4" />
                <span>Organizations</span>
              </TabsTrigger>
            </TabsList>

            <TabsContent value="users" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>User Management</CardTitle>
                  <CardDescription>
                    View and manage all users across organizations
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {filteredUsers.length === 0 ? (
                      <div className="text-center py-8 text-muted-foreground">
                        No users found
                      </div>
                    ) : (
                      filteredUsers.map((user) => (
                        <div
                          key={user.id}
                          className="flex items-center justify-between p-4 border rounded-lg"
                        >
                          <div className="flex items-center space-x-4">
                            <Avatar className="h-10 w-10">
                              <AvatarFallback>
                                {getInitials(user.first_name, user.last_name)}
                              </AvatarFallback>
                            </Avatar>
                            <div>
                              <div className="font-medium">
                                {user.first_name} {user.last_name}
                              </div>
                              <div className="text-sm text-muted-foreground">
                                {user.email}
                              </div>
                              <div className="flex items-center space-x-2 mt-1">
                                <Badge variant={user.is_active ? "default" : "secondary"}>
                                  {user.is_active ? "Active" : "Inactive"}
                                </Badge>
                                {user.is_superadmin && (
                                  <Badge variant="destructive">
                                    Super Admin
                                  </Badge>
                                )}
                              </div>
                            </div>
                          </div>
                          <div className="text-right text-sm text-muted-foreground">
                            <div>Joined {formatDate(user.created_at)}</div>
                            <div>Tenant: {user.tenant_id}</div>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="organizations" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Organization Management</CardTitle>
                  <CardDescription>
                    View and manage all organizations
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {filteredOrganizations.length === 0 ? (
                      <div className="text-center py-8 text-muted-foreground">
                        No organizations found
                      </div>
                    ) : (
                      filteredOrganizations.map((org) => (
                        <div
                          key={org.id}
                          className="flex items-center justify-between p-4 border rounded-lg"
                        >
                          <div className="flex items-center space-x-4">
                            <div className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                              <Building className="h-5 w-5 text-primary" />
                            </div>
                            <div>
                              <div className="font-medium">{org.name}</div>
                              <div className="text-sm text-muted-foreground">
                                {org.slug}
                              </div>
                              <div className="flex items-center space-x-2 mt-1">
                                <Badge variant={org.status === 'active' ? "default" : "secondary"}>
                                  {org.status}
                                </Badge>
                                <Badge variant="outline" className="text-xs">
                                  {org.member_count} member{org.member_count !== 1 ? 's' : ''}
                                </Badge>
                              </div>
                              {org.owner && (
                                <div className="text-xs text-muted-foreground mt-1">
                                  Owner: {org.owner.first_name} {org.owner.last_name} ({org.owner.email})
                                </div>
                              )}
                            </div>
                          </div>
                          <div className="text-right text-sm text-muted-foreground">
                            <div>Created {formatDate(org.created_at)}</div>
                            <div>Updated {formatDate(org.updated_at)}</div>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </motion.div>
      </main>
    </div>
  )
}

export default Admin