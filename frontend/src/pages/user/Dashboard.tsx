import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { 
  User, 
  Settings, 
  Shield, 
  LogOut, 
  Building2, 
  Users, 
  Crown,
  Mail,
  Calendar,
  Activity,
  ArrowUpRight,
  ChevronDown,
  Sparkles,
  LayoutDashboard
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import useAuthStore from '@/store/auth'
import { RequirePermission } from '@/components/auth/PermissionGate'

const Dashboard = () => {
  const navigate = useNavigate()
  const { user, organization, logout, isAuthenticated } = useAuthStore()

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { replace: true })
    }
  }, [isAuthenticated, navigate])

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  const switchOrganization = () => {
    navigate('/choose-organization')
  }

  if (!isAuthenticated || !user) {
    return null
  }

  const getInitials = (firstName?: string, lastName?: string, email?: string) => {
    if (firstName && lastName) {
      return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase()
    }
    if (email) {
      return email.substring(0, 2).toUpperCase()
    }
    return 'U'
  }

  const getRoleBadge = (role?: string) => {
    const roleConfig = {
      owner: { color: 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400', icon: Crown },
      admin: { color: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400', icon: Shield },
      member: { color: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400', icon: Users },
    }

    const config = roleConfig[role?.toLowerCase() as keyof typeof roleConfig] || roleConfig.member
    const Icon = config.icon

    return (
      <Badge className={`${config.color} flex items-center gap-1.5 px-3 py-1`}>
        <Icon className="h-3.5 w-3.5" />
        <span className="capitalize font-medium">{role || 'Member'}</span>
      </Badge>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-purple-50/20">
      {/* Floating orbs background */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-0 -left-4 w-72 h-72 bg-purple-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob" />
        <div className="absolute top-0 -right-4 w-72 h-72 bg-blue-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-2000" />
        <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-4000" />
      </div>

      {/* Header */}
      <header className="relative bg-white/80 backdrop-blur-xl border-b border-gray-200/50 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl shadow-lg">
                  <Building2 className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h1 className="text-lg font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                    {organization?.organization_name || 'Dashboard'}
                  </h1>
                  <p className="text-xs text-gray-500">
                    {organization?.organization_slug || 'workspace'}
                  </p>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-3">
              {/* Organization Switcher */}
              <Button
                variant="outline"
                size="sm"
                onClick={switchOrganization}
                className="flex items-center gap-2 h-9"
              >
                <Building2 className="h-4 w-4" />
                <span className="hidden sm:inline">Switch Org</span>
                <ChevronDown className="h-3 w-3" />
              </Button>

              {/* User Menu */}
              <div className="flex items-center gap-2">
                <Avatar className="h-9 w-9 ring-2 ring-purple-100">
                  <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-sm font-semibold">
                    {getInitials(user.first_name, user.last_name, user.email)}
                  </AvatarFallback>
                </Avatar>
              </div>

              <Button
                variant="ghost"
                size="sm"
                onClick={handleLogout}
                className="flex items-center gap-2 h-9"
              >
                <LogOut className="h-4 w-4" />
                <span className="hidden sm:inline">Sign out</span>
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="space-y-8"
        >
          {/* Welcome Hero */}
          <div className="relative overflow-hidden bg-gradient-to-br from-blue-600 via-purple-600 to-pink-600 rounded-2xl shadow-xl p-8">
            <div className="absolute inset-0 bg-grid-white/10" />
            <div className="absolute -right-4 -top-4 w-64 h-64 bg-white/10 rounded-full blur-3xl" />
            <div className="relative">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-4">
                  <Avatar className="h-20 w-20 ring-4 ring-white/20">
                    <AvatarFallback className="text-2xl font-bold bg-white/20 text-white backdrop-blur">
                      {getInitials(user.first_name, user.last_name, user.email)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="text-white">
                    <div className="flex items-center gap-2 mb-2">
                      <h2 className="text-3xl font-bold">
                        Welcome back, {user.first_name || 'User'}!
                      </h2>
                      <Sparkles className="h-6 w-6 text-yellow-300" />
                    </div>
                    <p className="text-blue-100 text-lg mb-3">{user.email}</p>
                    <div className="flex items-center gap-3">
                      {getRoleBadge(organization?.role)}
                      <Badge className="bg-white/20 text-white backdrop-blur border-white/30">
                        <Calendar className="h-3 w-3 mr-1" />
                        Joined {organization?.joined_at ? new Date(organization.joined_at).toLocaleDateString() : 'Recently'}
                      </Badge>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Stats Overview */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <motion.div whileHover={{ y: -4 }} transition={{ duration: 0.2 }}>
              <Card className="border-0 shadow-lg bg-gradient-to-br from-blue-50 to-blue-100/50">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-blue-600">Organization</p>
                      <p className="text-2xl font-bold text-blue-900 mt-1">Active</p>
                    </div>
                    <div className="p-3 bg-blue-600 rounded-xl">
                      <Building2 className="h-6 w-6 text-white" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>

            <motion.div whileHover={{ y: -4 }} transition={{ duration: 0.2 }}>
              <Card className="border-0 shadow-lg bg-gradient-to-br from-purple-50 to-purple-100/50">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-purple-600">Your Role</p>
                      <p className="text-2xl font-bold text-purple-900 mt-1 capitalize">{organization?.role || 'Member'}</p>
                    </div>
                    <div className="p-3 bg-purple-600 rounded-xl">
                      <Shield className="h-6 w-6 text-white" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>

            <motion.div whileHover={{ y: -4 }} transition={{ duration: 0.2 }}>
              <Card className="border-0 shadow-lg bg-gradient-to-br from-green-50 to-green-100/50">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-green-600">Status</p>
                      <p className="text-2xl font-bold text-green-900 mt-1">Active</p>
                    </div>
                    <div className="p-3 bg-green-600 rounded-xl">
                      <Activity className="h-6 w-6 text-white" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>

            <motion.div whileHover={{ y: -4 }} transition={{ duration: 0.2 }}>
              <Card className="border-0 shadow-lg bg-gradient-to-br from-amber-50 to-amber-100/50">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-amber-600">Sessions</p>
                      <p className="text-2xl font-bold text-amber-900 mt-1">1</p>
                    </div>
                    <div className="p-3 bg-amber-600 rounded-xl">
                      <LayoutDashboard className="h-6 w-6 text-white" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          </div>

          {/* Quick Actions */}
          <div>
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
                <Card className="cursor-pointer hover:shadow-xl transition-all duration-300 border-0 bg-white/80 backdrop-blur">
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2 text-gray-900">
                      <div className="p-2 bg-blue-100 rounded-lg">
                        <User className="h-5 w-5 text-blue-600" />
                      </div>
                      <span>Profile Settings</span>
                    </CardTitle>
                    <CardDescription>
                      Update your personal information and preferences
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button
                      onClick={() => navigate('/profile')}
                      className="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800"
                    >
                      Manage Profile
                      <ArrowUpRight className="h-4 w-4 ml-2" />
                    </Button>
                  </CardContent>
                </Card>
              </motion.div>

              <RequirePermission permission="member:view">
                <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
                  <Card className="cursor-pointer hover:shadow-xl transition-all duration-300 border-0 bg-white/80 backdrop-blur">
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2 text-gray-900">
                        <div className="p-2 bg-green-100 rounded-lg">
                          <Users className="h-5 w-5 text-green-600" />
                        </div>
                        <span>Team Members</span>
                      </CardTitle>
                      <CardDescription>
                        Manage team members and invitations
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <Button
                        onClick={() => navigate('/members')}
                        className="w-full bg-gradient-to-r from-green-600 to-green-700 hover:from-green-700 hover:to-green-800"
                      >
                        View Members
                        <ArrowUpRight className="h-4 w-4 ml-2" />
                      </Button>
                    </CardContent>
                  </Card>
                </motion.div>
              </RequirePermission>

              <RequirePermission permission="role:view">
                <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
                  <Card className="cursor-pointer hover:shadow-xl transition-all duration-300 border-0 bg-white/80 backdrop-blur">
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2 text-gray-900">
                        <div className="p-2 bg-indigo-100 rounded-lg">
                          <Shield className="h-5 w-5 text-indigo-600" />
                        </div>
                        <span>Roles & Permissions</span>
                      </CardTitle>
                      <CardDescription>
                        Manage custom roles and access control
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <Button
                        onClick={() => navigate('/roles')}
                        className="w-full bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800"
                      >
                        Manage Roles
                        <ArrowUpRight className="h-4 w-4 ml-2" />
                      </Button>
                    </CardContent>
                  </Card>
                </motion.div>
              </RequirePermission>

              <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
                <Card className="cursor-pointer hover:shadow-xl transition-all duration-300 border-0 bg-white/80 backdrop-blur">
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2 text-gray-900">
                      <div className="p-2 bg-purple-100 rounded-lg">
                        <Settings className="h-5 w-5 text-purple-600" />
                      </div>
                      <span>Account Settings</span>
                    </CardTitle>
                    <CardDescription>
                      Change password and security settings
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button
                      onClick={() => navigate('/settings')}
                      variant="outline"
                      className="w-full border-purple-200 hover:bg-purple-50"
                    >
                      Account Settings
                      <ArrowUpRight className="h-4 w-4 ml-2" />
                    </Button>
                  </CardContent>
                </Card>
              </motion.div>
            </div>
          </div>

          {/* Recent Activity */}
          <Card className="border-0 shadow-lg bg-white/80 backdrop-blur">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-gray-900">Recent Activity</CardTitle>
                  <CardDescription>
                    Your recent actions in {organization?.organization_name}
                  </CardDescription>
                </div>
                <Badge variant="outline" className="text-green-600 border-green-200">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse" />
                  Live
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <motion.div 
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  className="flex items-start gap-4 p-4 bg-gradient-to-r from-green-50 to-transparent rounded-xl border border-green-100"
                >
                  <div className="p-2 bg-green-500 rounded-lg">
                    <Activity className="h-4 w-4 text-white" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-semibold text-gray-900">Successful login</p>
                    <p className="text-xs text-gray-500 mt-1">
                      Logged into {organization?.organization_name}
                    </p>
                    <p className="text-xs text-gray-400 mt-1">
                      {new Date().toLocaleString()}
                    </p>
                  </div>
                </motion.div>

                <motion.div 
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: 0.1 }}
                  className="flex items-start gap-4 p-4 bg-gradient-to-r from-blue-50 to-transparent rounded-xl border border-blue-100"
                >
                  <div className="p-2 bg-blue-500 rounded-lg">
                    <Building2 className="h-4 w-4 text-white" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-semibold text-gray-900">Organization selected</p>
                    <p className="text-xs text-gray-500 mt-1">
                      Switched to {organization?.organization_name}
                    </p>
                    <p className="text-xs text-gray-400 mt-1">
                      Just now
                    </p>
                  </div>
                </motion.div>

                <motion.div 
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: 0.2 }}
                  className="flex items-start gap-4 p-4 bg-gradient-to-r from-purple-50 to-transparent rounded-xl border border-purple-100"
                >
                  <div className="p-2 bg-purple-500 rounded-lg">
                    <Mail className="h-4 w-4 text-white" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-semibold text-gray-900">Welcome email sent</p>
                    <p className="text-xs text-gray-500 mt-1">
                      Confirmation email delivered to {user.email}
                    </p>
                    <p className="text-xs text-gray-400 mt-1">
                      {organization?.joined_at ? new Date(organization.joined_at).toLocaleString() : 'Recently'}
                    </p>
                  </div>
                </motion.div>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </main>
    </div>
  )
}

export default Dashboard