import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  Users,
  Building,
  Key,
  Activity,
  Book,
  TrendingUp,
  Shield,
  Database,
  ArrowRight,
  UserCheck,
  AlertCircle,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import LoadingSpinner from '@/components/ui/loading-spinner'
import useAuthStore from '@/store/auth'
import { adminAPI, organizationAPI } from '@/lib/api'

export default function SuperadminDashboard() {
  const navigate = useNavigate()
  const { isSuperadmin, user } = useAuthStore()

  const [stats, setStats] = useState({
    totalUsers: 0,
    totalOrganizations: 0,
    activeUsers: 0,
    recentUsers: 0,
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Redirect if not superadmin
  useEffect(() => {
    if (!isSuperadmin) {
      navigate('/dashboard', { replace: true })
    }
  }, [isSuperadmin, navigate])

  // Load statistics
  useEffect(() => {
    if (isSuperadmin) {
      loadStats()
    }
  }, [isSuperadmin])

  const loadStats = async () => {
    try {
      setLoading(true)
      setError('')

      const [usersResponse, orgsResponse] = await Promise.all([
        adminAPI.listUsers(),
        organizationAPI.listOrganizations(),
      ])

      const usersResponseData = usersResponse as any
      const users = Array.isArray(usersResponseData.users) ? usersResponseData.users :
                   (usersResponseData.data?.users || [])
      const orgs = orgsResponse.data || []

      // Calculate stats
      const now = new Date()
      const last7Days = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)

      setStats({
        totalUsers: users.length,
        totalOrganizations: orgs.length,
        activeUsers: users.filter((u: any) => u.is_active).length,
        recentUsers: users.filter((u: any) => new Date(u.created_at) > last7Days).length,
      })
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to load statistics'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

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
      <div className="bg-white dark:bg-slate-800 border-b">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold flex items-center gap-2">
                <Shield className="h-8 w-8 text-primary" />
                Superadmin Console
              </h1>
              <p className="text-muted-foreground mt-2">
                System-wide management and monitoring
              </p>
            </div>
            <div className="flex items-center gap-3">
              <div className="text-right">
                <p className="text-sm font-medium">{user?.email}</p>
                <p className="text-xs text-muted-foreground">Superadmin</p>
              </div>
              <Button
                variant="outline"
                onClick={() => {
                  const logout = useAuthStore.getState().logout
                  logout()
                }}
              >
                Logout
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        {/* Error Alert */}
        {error && (
          <Alert variant="destructive" className="mb-6">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Statistics Overview */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Users</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.totalUsers}</div>
                <p className="text-xs text-muted-foreground">
                  {stats.activeUsers} active
                </p>
              </CardContent>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
          >
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Organizations</CardTitle>
                <Building className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.totalOrganizations}</div>
                <p className="text-xs text-muted-foreground">
                  Multi-tenant instances
                </p>
              </CardContent>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
          >
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">New Users</CardTitle>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.recentUsers}</div>
                <p className="text-xs text-muted-foreground">
                  Last 7 days
                </p>
              </CardContent>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
          >
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">System Status</CardTitle>
                <Database className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">Healthy</div>
                <p className="text-xs text-muted-foreground">
                  All services operational
                </p>
              </CardContent>
            </Card>
          </motion.div>
        </div>

        {/* Main Navigation Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          {/* Users Management */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
          >
            <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/admin')}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="p-3 bg-blue-100 dark:bg-blue-900 rounded-lg">
                    <Users className="h-6 w-6 text-blue-600 dark:text-blue-400" />
                  </div>
                  <ArrowRight className="h-5 w-5 text-muted-foreground" />
                </div>
                <CardTitle className="mt-4">User Management</CardTitle>
                <CardDescription>
                  View and manage all system users across organizations
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Total Users</span>
                  <span className="font-bold">{stats.totalUsers}</span>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Organizations Management */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 }}
          >
            <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/admin')}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="p-3 bg-purple-100 dark:bg-purple-900 rounded-lg">
                    <Building className="h-6 w-6 text-purple-600 dark:text-purple-400" />
                  </div>
                  <ArrowRight className="h-5 w-5 text-muted-foreground" />
                </div>
                <CardTitle className="mt-4">Organizations</CardTitle>
                <CardDescription>
                  Manage tenant organizations and their configurations
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Total Orgs</span>
                  <span className="font-bold">{stats.totalOrganizations}</span>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* OAuth Client Apps */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.7 }}
          >
            <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/oauth/client-apps')}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="p-3 bg-green-100 dark:bg-green-900 rounded-lg">
                    <Key className="h-6 w-6 text-green-600 dark:text-green-400" />
                  </div>
                  <ArrowRight className="h-5 w-5 text-muted-foreground" />
                </div>
                <CardTitle className="mt-4">OAuth Client Apps</CardTitle>
                <CardDescription>
                  Manage OAuth2.1 client applications and credentials
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Manage Apps</span>
                  <span className="font-bold">→</span>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* OAuth Audit Logs */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.8 }}
          >
            <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/oauth/audit')}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="p-3 bg-orange-100 dark:bg-orange-900 rounded-lg">
                    <Activity className="h-6 w-6 text-orange-600 dark:text-orange-400" />
                  </div>
                  <ArrowRight className="h-5 w-5 text-muted-foreground" />
                </div>
                <CardTitle className="mt-4">OAuth Audit Logs</CardTitle>
                <CardDescription>
                  Monitor authorization flows and token activity
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">View Logs</span>
                  <span className="font-bold">→</span>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Developer Documentation */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.9 }}
          >
            <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/developer/docs')}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="p-3 bg-indigo-100 dark:bg-indigo-900 rounded-lg">
                    <Book className="h-6 w-6 text-indigo-600 dark:text-indigo-400" />
                  </div>
                  <ArrowRight className="h-5 w-5 text-muted-foreground" />
                </div>
                <CardTitle className="mt-4">Developer Docs</CardTitle>
                <CardDescription>
                  OAuth2.1 integration guides and SDK documentation
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">View Docs</span>
                  <span className="font-bold">→</span>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* API Keys Management */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 1.0 }}
          >
            <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/dev/api-keys')}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="p-3 bg-teal-100 dark:bg-teal-900 rounded-lg">
                    <Key className="h-6 w-6 text-teal-600 dark:text-teal-400" />
                  </div>
                  <ArrowRight className="h-5 w-5 text-muted-foreground" />
                </div>
                <CardTitle className="mt-4">API Keys</CardTitle>
                <CardDescription>
                  Manage API keys for programmatic access
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Developer Tools</span>
                  <span className="font-bold">→</span>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* RBAC Management - Only for Superadmin */}
          {isSuperadmin && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 1.1 }}
            >
              <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/admin/rbac')}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="p-3 bg-red-100 dark:bg-red-900 rounded-lg">
                      <Shield className="h-6 w-6 text-red-600 dark:text-red-400" />
                    </div>
                    <ArrowRight className="h-5 w-5 text-muted-foreground" />
                  </div>
                  <CardTitle className="mt-4">System RBAC Management</CardTitle>
                  <CardDescription>
                    System-wide roles and permissions configuration
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Manage System Roles</span>
                    <span className="font-bold">→</span>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          )}
        </div>

        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common administrative tasks</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-3">
              <Button onClick={() => navigate('/admin')}>
                <UserCheck className="mr-2 h-4 w-4" />
                View All Users
              </Button>
              <Button variant="outline" onClick={() => navigate('/oauth/client-apps')}>
                <Key className="mr-2 h-4 w-4" />
                Create OAuth App
              </Button>
              <Button variant="outline" onClick={() => navigate('/dev/api-keys')}>
                <Key className="mr-2 h-4 w-4" />
                Manage API Keys
              </Button>
              <Button variant="outline" onClick={() => navigate('/oauth/audit')}>
                <Activity className="mr-2 h-4 w-4" />
                View Audit Logs
              </Button>
              <Button variant="outline" onClick={() => loadStats()}>
                <Database className="mr-2 h-4 w-4" />
                Refresh Stats
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
