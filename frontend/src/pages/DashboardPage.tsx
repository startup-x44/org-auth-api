/**
 * Main Dashboard Page - Tenant-aware with role-based UI
 * Displays organization overview, recent activity, quick actions
 */

import { useEffect, useState } from 'react'
import { useAuthStore } from '../stores/authStore'
import { useUserName, useHasPermission } from '../hooks/useAuth'
import { useTenantStore } from '../stores/tenantStore'
import LoadingSpinner from '../components/ui/loading-spinner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Badge } from '../components/ui/badge'
import { Alert, AlertDescription } from '../components/ui/alert'
import { 
  Users, 
  Shield, 
  Activity, 
  Settings, 
  Plus,
  TrendingUp,
  AlertTriangle,
  CheckCircle
} from 'lucide-react'
import { performanceTracker } from '../utils/monitoring'

interface DashboardStats {
  totalUsers: number
  activeUsers: number
  totalRoles: number
  recentLogins: number
  securityAlerts: number
}

interface RecentActivity {
  id: string
  type: 'user_login' | 'user_created' | 'role_assigned' | 'security_event'
  description: string
  timestamp: string
  user?: string
  severity?: 'low' | 'medium' | 'high'
}

export function DashboardPage() {
  const { user } = useAuthStore()
  const userName = useUserName()
  const hasPermission = useHasPermission()
  const { currentTenant, members, loadMembers } = useTenantStore()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadDashboardData()
  }, [currentTenant?.id])

  const loadDashboardData = async () => {
    if (!currentTenant?.id) return

    try {
      setLoading(true)
      setError(null)

      await performanceTracker.measureAsync('dashboard-load', async () => {
        // Load members if user has permission
        if (hasPermission('users:read')) {
          await loadMembers()
        }

        // Simulate API calls for dashboard data
        // In real app, these would be actual API calls
        await Promise.all([
          loadDashboardStats(),
          loadRecentActivity()
        ])
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard data')
    } finally {
      setLoading(false)
    }
  }

  const loadDashboardStats = async (): Promise<void> => {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 500))
    
    setStats({
      totalUsers: members?.length || 12,
      activeUsers: Math.floor((members?.length || 12) * 0.8),
      totalRoles: 5,
      recentLogins: 8,
      securityAlerts: 2
    })
  }

  const loadRecentActivity = async (): Promise<void> => {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 300))
    
    setRecentActivity([
      {
        id: '1',
        type: 'user_login',
        description: 'John Doe logged in successfully',
        timestamp: '2024-01-20T10:30:00Z',
        user: 'john.doe@example.com'
      },
      {
        id: '2',
        type: 'role_assigned',
        description: 'Admin role assigned to Jane Smith',
        timestamp: '2024-01-20T09:15:00Z',
        user: 'jane.smith@example.com'
      },
      {
        id: '3',
        type: 'security_event',
        description: 'Failed login attempt detected',
        timestamp: '2024-01-20T08:45:00Z',
        severity: 'medium'
      },
      {
        id: '4',
        type: 'user_created',
        description: 'New user Bob Wilson was created',
        timestamp: '2024-01-19T16:20:00Z',
        user: 'bob.wilson@example.com'
      }
    ])
  }

  const getActivityIcon = (type: RecentActivity['type']) => {
    switch (type) {
      case 'user_login':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'user_created':
        return <Users className="h-4 w-4 text-blue-500" />
      case 'role_assigned':
        return <Shield className="h-4 w-4 text-purple-500" />
      case 'security_event':
        return <AlertTriangle className="h-4 w-4 text-orange-500" />
    }
  }

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString()
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive" className="m-6">
        <AlertTriangle className="h-4 w-4" />
        <AlertDescription>
          Failed to load dashboard: {error}
          <Button variant="outline" size="sm" className="ml-2" onClick={loadDashboardData}>
            Retry
          </Button>
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            Welcome back, {userName || user?.email}
          </h1>
          <p className="text-gray-500 mt-1">
            {currentTenant?.name || 'Personal Workspace'} Dashboard
          </p>
        </div>
        <div className="flex gap-2">
          {hasPermission('users:create') && (
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Invite User
            </Button>
          )}
          {hasPermission('organization:update') && (
            <Button variant="outline">
              <Settings className="h-4 w-4 mr-2" />
              Settings
            </Button>
          )}
        </div>
      </div>

      {/* Stats Grid */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Users</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalUsers}</div>
              <p className="text-xs text-muted-foreground">
                {stats.activeUsers} active this month
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Sessions</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.activeUsers}</div>
              <p className="text-xs text-muted-foreground">
                {stats.recentLogins} recent logins
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Roles & Permissions</CardTitle>
              <Shield className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalRoles}</div>
              <p className="text-xs text-muted-foreground">
                Active role configurations
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Security Alerts</CardTitle>
              <AlertTriangle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-orange-600">{stats.securityAlerts}</div>
              <p className="text-xs text-muted-foreground">
                Require attention
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Activity */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Recent Activity
            </CardTitle>
            <CardDescription>
              Latest events in your organization
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {recentActivity.map((activity) => (
                <div key={activity.id} className="flex items-start gap-3 p-3 rounded-lg bg-gray-50">
                  {getActivityIcon(activity.type)}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900">
                      {activity.description}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      <p className="text-xs text-gray-500">
                        {formatTimestamp(activity.timestamp)}
                      </p>
                      {activity.severity && (
                        <Badge 
                          variant={activity.severity === 'high' ? 'destructive' : 'outline'}
                          className="text-xs"
                        >
                          {activity.severity}
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              Quick Actions
            </CardTitle>
            <CardDescription>
              Common tasks for your role
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {hasPermission('users:create') && (
                <Button variant="outline" className="w-full justify-start" size="sm">
                  <Users className="h-4 w-4 mr-2" />
                  Invite New User
                </Button>
              )}
              
              {hasPermission('roles:read') && (
                <Button variant="outline" className="w-full justify-start" size="sm">
                  <Shield className="h-4 w-4 mr-2" />
                  Manage Roles
                </Button>
              )}
              
              <Button variant="outline" className="w-full justify-start" size="sm">
                <Activity className="h-4 w-4 mr-2" />
                View Security Logs
              </Button>
              
              {hasPermission('organization:update') && (
                <Button variant="outline" className="w-full justify-start" size="sm">
                  <Settings className="h-4 w-4 mr-2" />
                  Organization Settings
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Security Alerts Section */}
      {stats && stats.securityAlerts > 0 && (
        <Card className="border-orange-200">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-orange-700">
              <AlertTriangle className="h-5 w-5" />
              Security Alerts
            </CardTitle>
            <CardDescription>
              {stats.securityAlerts} security events require your attention
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                Multiple failed login attempts detected. Consider reviewing security settings.
                <Button variant="link" className="h-auto p-0 ml-1">
                  View Details
                </Button>
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      )}
    </div>
  )
}