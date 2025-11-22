/**
 * User Profile Page - Account settings, security, preferences
 * Handles profile updates, password changes, MFA setup, session management
 */

import React, { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useAuthStore } from '../stores/authStore'
import { useTenantStore } from '../stores/tenantStore'
import LoadingSpinner from '../components/ui/loading-spinner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Label } from '../components/ui/label'
import { Badge } from '../components/ui/badge'
import { Alert, AlertDescription } from '../components/ui/alert'
import { Separator } from '../components/ui/separator'
import { Switch } from '../components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs'
import { 
  User, 
  Shield, 
  Key, 
  Smartphone,
  LogOut,
  AlertTriangle,
  CheckCircle,
  Eye,
  EyeOff,
  Download
} from 'lucide-react'
import { toast } from 'react-hot-toast'

// Form schemas
const profileSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name too long'),
  email: z.string().email('Invalid email address'),
  phone: z.string().optional(),
  timezone: z.string().optional(),
})

const passwordSchema = z.object({
  currentPassword: z.string().min(1, 'Current password is required'),
  newPassword: z.string().min(8, 'Password must be at least 8 characters'),
  confirmPassword: z.string().min(1, 'Please confirm your password'),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

type ProfileFormData = z.infer<typeof profileSchema>
type PasswordFormData = z.infer<typeof passwordSchema>

interface UserSession {
  id: string
  device: string
  location: string
  lastActive: string
  current: boolean
  ipAddress: string
}

export function ProfilePage() {
  const { user, updateProfile, changePassword, enableMFA, disableMFA, logout } = useAuthStore()
  const { currentTenant } = useTenantStore()
  const [loading, setLoading] = useState(false)
  const [mfaLoading, setMfaLoading] = useState(false)
  const [showCurrentPassword, setShowCurrentPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [sessions, setSessions] = useState<UserSession[]>([])
  const [mfaQrCode, setMfaQrCode] = useState<string | null>(null)

  // Profile form
  const profileForm = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      name: user?.name || '',
      email: user?.email || '',
      phone: user?.phone || '',
      timezone: user?.timezone || 'UTC',
    },
  })

  // Password form
  const passwordForm = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
    defaultValues: {
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    },
  })

  useEffect(() => {
    loadUserSessions()
  }, [])

  const loadUserSessions = async () => {
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500))
      
      setSessions([
        {
          id: '1',
          device: 'Chrome on macOS',
          location: 'San Francisco, CA',
          lastActive: '2024-01-20T10:30:00Z',
          current: true,
          ipAddress: '192.168.1.100'
        },
        {
          id: '2',
          device: 'Safari on iPhone',
          location: 'San Francisco, CA',
          lastActive: '2024-01-19T16:20:00Z',
          current: false,
          ipAddress: '192.168.1.101'
        },
      ])
    } catch (error) {
      toast.error('Failed to load sessions')
    }
  }

  const onProfileSubmit = async (data: ProfileFormData) => {
    try {
      setLoading(true)
      await updateProfile(data)
      toast.success('Profile updated successfully')
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to update profile')
    } finally {
      setLoading(false)
    }
  }

  const onPasswordSubmit = async (data: PasswordFormData) => {
    try {
      setLoading(true)
      await changePassword(data.currentPassword, data.newPassword)
      passwordForm.reset()
      toast.success('Password changed successfully')
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to change password')
    } finally {
      setLoading(false)
    }
  }

  const handleEnableMFA = async () => {
    try {
      setMfaLoading(true)
      const qrCode = await enableMFA()
      setMfaQrCode(qrCode)
      toast.success('MFA enabled successfully')
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to enable MFA')
    } finally {
      setMfaLoading(false)
    }
  }

  const handleDisableMFA = async () => {
    try {
      setMfaLoading(true)
      await disableMFA()
      setMfaQrCode(null)
      toast.success('MFA disabled successfully')
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to disable MFA')
    } finally {
      setMfaLoading(false)
    }
  }

  const handleLogoutSession = async (sessionId: string) => {
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      setSessions(sessions.filter(s => s.id !== sessionId))
      toast.success('Session terminated')
    } catch (error) {
      toast.error('Failed to terminate session')
    }
  }

  const downloadSecurityReport = async () => {
    try {
      // Simulate generating security report
      toast.success('Security report downloaded')
    } catch (error) {
      toast.error('Failed to download security report')
    }
  }

  const formatLastActive = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    
    if (diffHours < 1) return 'Just now'
    if (diffHours < 24) return `${diffHours} hours ago`
    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays} days ago`
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Profile Settings</h1>
        <p className="text-gray-500 mt-1">
          Manage your account settings and security preferences
        </p>
      </div>

      <Tabs defaultValue="profile" className="space-y-6">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="profile" className="flex items-center gap-2">
            <User className="h-4 w-4" />
            Profile
          </TabsTrigger>
          <TabsTrigger value="security" className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            Security
          </TabsTrigger>
          <TabsTrigger value="sessions" className="flex items-center gap-2">
            <Key className="h-4 w-4" />
            Sessions
          </TabsTrigger>
          <TabsTrigger value="mfa" className="flex items-center gap-2">
            <Smartphone className="h-4 w-4" />
            2FA
          </TabsTrigger>
        </TabsList>

        <TabsContent value="profile">
          <Card>
            <CardHeader>
              <CardTitle>Profile Information</CardTitle>
              <CardDescription>
                Update your personal information and preferences
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={profileForm.handleSubmit(onProfileSubmit)} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="name">Full Name</Label>
                    <Input
                      id="name"
                      {...profileForm.register('name')}
                      error={profileForm.formState.errors.name?.message}
                    />
                  </div>
                  
                  <div>
                    <Label htmlFor="email">Email Address</Label>
                    <Input
                      id="email"
                      type="email"
                      {...profileForm.register('email')}
                      error={profileForm.formState.errors.email?.message}
                    />
                  </div>
                  
                  <div>
                    <Label htmlFor="phone">Phone Number</Label>
                    <Input
                      id="phone"
                      type="tel"
                      {...profileForm.register('phone')}
                      placeholder="+1 (555) 123-4567"
                    />
                  </div>
                  
                  <div>
                    <Label htmlFor="timezone">Timezone</Label>
                    <select
                      id="timezone"
                      {...profileForm.register('timezone')}
                      className="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="UTC">UTC</option>
                      <option value="America/New_York">Eastern Time</option>
                      <option value="America/Chicago">Central Time</option>
                      <option value="America/Denver">Mountain Time</option>
                      <option value="America/Los_Angeles">Pacific Time</option>
                    </select>
                  </div>
                </div>

                <Separator />

                <div className="space-y-2">
                  <h3 className="text-sm font-medium">Organization Membership</h3>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">{currentTenant?.name || 'Personal'}</Badge>
                    <Badge variant="secondary">{user?.role || 'Member'}</Badge>
                  </div>
                </div>

                <div className="flex justify-end">
                  <Button type="submit" disabled={loading}>
                    {loading && <LoadingSpinner size="small" className="mr-2" />}
                    Save Changes
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="security">
          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Change Password</CardTitle>
                <CardDescription>
                  Update your password to keep your account secure
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={passwordForm.handleSubmit(onPasswordSubmit)} className="space-y-4">
                  <div>
                    <Label htmlFor="currentPassword">Current Password</Label>
                    <div className="relative">
                      <Input
                        id="currentPassword"
                        type={showCurrentPassword ? 'text' : 'password'}
                        {...passwordForm.register('currentPassword')}
                        error={passwordForm.formState.errors.currentPassword?.message}
                      />
                      <button
                        type="button"
                        onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-700"
                      >
                        {showCurrentPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="newPassword">New Password</Label>
                    <div className="relative">
                      <Input
                        id="newPassword"
                        type={showNewPassword ? 'text' : 'password'}
                        {...passwordForm.register('newPassword')}
                        error={passwordForm.formState.errors.newPassword?.message}
                      />
                      <button
                        type="button"
                        onClick={() => setShowNewPassword(!showNewPassword)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-700"
                      >
                        {showNewPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="confirmPassword">Confirm New Password</Label>
                    <Input
                      id="confirmPassword"
                      type="password"
                      {...passwordForm.register('confirmPassword')}
                      error={passwordForm.formState.errors.confirmPassword?.message}
                    />
                  </div>

                  <div className="flex justify-end">
                    <Button type="submit" disabled={loading}>
                      {loading && <LoadingSpinner size="small" className="mr-2" />}
                      Update Password
                    </Button>
                  </div>
                </form>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Security Report</CardTitle>
                <CardDescription>
                  Download a report of your account security activity
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Button variant="outline" onClick={downloadSecurityReport}>
                  <Download className="h-4 w-4 mr-2" />
                  Download Security Report
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="sessions">
          <Card>
            <CardHeader>
              <CardTitle>Active Sessions</CardTitle>
              <CardDescription>
                Manage your active sessions across all devices
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {sessions.map((session) => (
                  <div
                    key={session.id}
                    className="flex items-center justify-between p-4 border rounded-lg"
                  >
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium">{session.device}</h4>
                        {session.current && (
                          <Badge variant="outline" className="text-green-600 border-green-200">
                            Current
                          </Badge>
                        )}
                      </div>
                      <p className="text-sm text-gray-500">{session.location}</p>
                      <p className="text-sm text-gray-500">
                        Last active: {formatLastActive(session.lastActive)}
                      </p>
                      <p className="text-xs text-gray-400">IP: {session.ipAddress}</p>
                    </div>
                    {!session.current && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleLogoutSession(session.id)}
                      >
                        <LogOut className="h-4 w-4 mr-1" />
                        End Session
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="mfa">
          <Card>
            <CardHeader>
              <CardTitle>Two-Factor Authentication</CardTitle>
              <CardDescription>
                Add an extra layer of security to your account
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-blue-100 rounded-lg">
                      <Smartphone className="h-5 w-5 text-blue-600" />
                    </div>
                    <div>
                      <h4 className="font-medium">Authenticator App</h4>
                      <p className="text-sm text-gray-500">
                        Use an app like Google Authenticator or Authy
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {user?.mfaEnabled ? (
                      <>
                        <Badge variant="outline" className="text-green-600 border-green-200">
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Enabled
                        </Badge>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={handleDisableMFA}
                          disabled={mfaLoading}
                        >
                          {mfaLoading && <LoadingSpinner size="small" className="mr-2" />}
                          Disable
                        </Button>
                      </>
                    ) : (
                      <Button
                        onClick={handleEnableMFA}
                        disabled={mfaLoading}
                      >
                        {mfaLoading && <LoadingSpinner size="small" className="mr-2" />}
                        Enable
                      </Button>
                    )}
                  </div>
                </div>

                {mfaQrCode && (
                  <Alert>
                    <CheckCircle className="h-4 w-4" />
                    <AlertDescription>
                      Scan this QR code with your authenticator app to complete setup.
                      <div className="mt-2 p-2 bg-white border rounded">
                        <img src={mfaQrCode} alt="MFA QR Code" className="mx-auto" />
                      </div>
                    </AlertDescription>
                  </Alert>
                )}

                {!user?.mfaEnabled && !mfaQrCode && (
                  <Alert>
                    <AlertTriangle className="h-4 w-4" />
                    <AlertDescription>
                      Two-factor authentication is not enabled. Your account may be vulnerable to unauthorized access.
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}