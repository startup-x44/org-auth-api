'use client'

import { useAuth } from '../../../contexts/auth-context'
import { useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { SuperAdminLayout } from '../../../components/layout/superadmin-layout'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../../components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '../../../components/ui/alert'
import { Badge } from '../../../components/ui/badge'
import { Skeleton } from '../../../components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
import { Users, Building2, Shield, Activity, TrendingUp, AlertTriangle } from 'lucide-react'

export default function SuperAdminDashboard() {
  const { user, loading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    console.log('SuperAdmin Dashboard - User:', user)
    console.log('SuperAdmin Dashboard - Loading:', loading)
    
    if (!loading && !user) {
      console.log('No user found, redirecting to login...')
      router.push('/auth/login')
      return
    }

    if (!loading && user && !user.is_superadmin) {
      console.log('User is not super admin, redirecting to user dashboard...')
      router.push('/user')
      return
    }
  }, [user, loading, router])

  if (loading) {
    return (
      <SuperAdminLayout>
        <div className="space-y-6">
          <div>
            <Skeleton className="h-8 w-64" />
            <Skeleton className="h-4 w-96 mt-2" />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {[1, 2, 3, 4].map((i) => (
              <Card key={i}>
                <CardContent className="pt-6">
                  <Skeleton className="h-20 w-full" />
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </SuperAdminLayout>
    )
  }

  if (!user || !user.is_superadmin) {
    return null
  }

  return (
    <SuperAdminLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold">Dashboard</h1>
          <p className="text-muted-foreground mt-2">Welcome back, {user.email}</p>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Users</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">1,245</div>
              <p className="text-xs text-muted-foreground">+20.1% from last month</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Organizations</CardTitle>
              <Building2 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">89</div>
              <p className="text-xs text-muted-foreground">+12 new this month</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Sessions</CardTitle>
              <Shield className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">342</div>
              <p className="text-xs text-muted-foreground">+15% from last week</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">System Health</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">99.9%</div>
              <p className="text-xs text-muted-foreground">All systems operational</p>
            </CardContent>
          </Card>
        </div>

        {/* Charts and Activity */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Activity Chart */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                User Activity
                <TrendingUp className="h-5 w-5 text-muted-foreground" />
              </CardTitle>
              <CardDescription>Overview of user registrations and logins</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-64 flex items-center justify-center border-2 border-dashed border-muted rounded-lg">
                <p className="text-muted-foreground">Activity Chart Placeholder</p>
              </div>
            </CardContent>
          </Card>

          {/* Recent Alerts */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                Recent Alerts
                <AlertTriangle className="h-5 w-5 text-muted-foreground" />
              </CardTitle>
              <CardDescription>System notifications and warnings</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <Alert variant="default">
                <AlertTriangle className="h-4 w-4" />
                <AlertTitle>High CPU Usage</AlertTitle>
                <AlertDescription>Server load at 85%</AlertDescription>
              </Alert>
              
              <Alert variant="default">
                <Activity className="h-4 w-4" />
                <AlertTitle>New User Registration</AlertTitle>
                <AlertDescription>15 new users today</AlertDescription>
              </Alert>
              
              <Alert variant="default">
                <Shield className="h-4 w-4" />
                <AlertTitle>Security Scan Complete</AlertTitle>
                <AlertDescription>No vulnerabilities found</AlertDescription>
              </Alert>
            </CardContent>
          </Card>
        </div>

        {/* Recent Activity Table */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>Latest user actions and system events</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Action</TableHead>
                  <TableHead>Time</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow>
                  <TableCell className="font-medium">john.doe@example.com</TableCell>
                  <TableCell>Login</TableCell>
                  <TableCell>2 minutes ago</TableCell>
                  <TableCell>
                    <Badge variant="default">Success</Badge>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell className="font-medium">jane.smith@example.com</TableCell>
                  <TableCell>Create Organization</TableCell>
                  <TableCell>5 minutes ago</TableCell>
                  <TableCell>
                    <Badge variant="default">Success</Badge>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell className="font-medium">admin@example.com</TableCell>
                  <TableCell>Failed Login</TableCell>
                  <TableCell>8 minutes ago</TableCell>
                  <TableCell>
                    <Badge variant="destructive">Failed</Badge>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </SuperAdminLayout>
  )
}