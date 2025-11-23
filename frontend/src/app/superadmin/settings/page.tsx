'use client'

import { useAuth } from '../../../contexts/auth-context'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { SuperAdminLayout } from '../../../components/layout/superadmin-layout'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Settings, Save, Shield, Database, Mail, Globe, Bell } from 'lucide-react'

export default function SettingsPage() {
  const { user, loading } = useAuth()
  const router = useRouter()
  const [activeTab, setActiveTab] = useState('general')

  useEffect(() => {
    if (!loading && !user) {
      router.push('/auth/login')
      return
    }

    if (!loading && user && !user.is_superadmin) {
      router.push('/user')
      return
    }
  }, [user, loading, router])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  if (!user || !user.is_superadmin) {
    return null
  }

  const tabs = [
    { id: 'general', name: 'General', icon: Settings },
    { id: 'security', name: 'Security', icon: Shield },
    { id: 'database', name: 'Database', icon: Database },
    { id: 'email', name: 'Email', icon: Mail },
    { id: 'api', name: 'API', icon: Globe },
    { id: 'notifications', name: 'Notifications', icon: Bell },
  ]

  return (
    <SuperAdminLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold text-gray-900">System Settings</h1>
          <p className="text-gray-600 mt-2">Configure system-wide settings and preferences</p>
        </div>

        <div className="flex gap-6">
          {/* Sidebar Navigation */}
          <div className="w-64 shrink-0">
            <Card className="p-4">
              <nav className="space-y-2">
                {tabs.map((tab) => {
                  const Icon = tab.icon
                  return (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id)}
                      className={`w-full flex items-center px-3 py-2 rounded-lg text-sm font-medium transition-colors ${activeTab === tab.id
                        ? 'bg-blue-100 text-blue-700'
                        : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                        }`}
                    >
                      <Icon className="h-4 w-4 mr-3" />
                      {tab.name}
                    </button>
                  )
                })}
              </nav>
            </Card>
          </div>

          {/* Content Area */}
          <div className="flex-1">
            {activeTab === 'general' && (
              <Card className="p-6">
                <h2 className="text-xl font-semibold text-gray-900 mb-6">General Settings</h2>
                <div className="space-y-6">
                  <div>
                    <Label className="mb-2 block">
                      Application Name
                    </Label>
                    <Input
                      type="text"
                      defaultValue="NILOAUTH"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      Application URL
                    </Label>
                    <Input
                      type="url"
                      defaultValue="https://auth.example.com"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      Time Zone
                    </Label>
                    <Select defaultValue="UTC">
                      <SelectTrigger>
                        <SelectValue placeholder="Select timezone" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="UTC">UTC</SelectItem>
                        <SelectItem value="America/New_York">America/New_York</SelectItem>
                        <SelectItem value="Europe/London">Europe/London</SelectItem>
                        <SelectItem value="Asia/Tokyo">Asia/Tokyo</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="flex items-center justify-between">
                    <Label htmlFor="maintenance" className="text-sm text-gray-900">
                      Enable maintenance mode
                    </Label>
                    <Switch id="maintenance" />
                  </div>
                </div>
              </Card>
            )}

            {activeTab === 'security' && (
              <Card className="p-6">
                <h2 className="text-xl font-semibold text-gray-900 mb-6">Security Settings</h2>
                <div className="space-y-6">
                  <div>
                    <Label className="mb-2 block">
                      JWT Token Expiration (minutes)
                    </Label>
                    <Input
                      type="number"
                      defaultValue="15"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      Refresh Token Expiration (days)
                    </Label>
                    <Input
                      type="number"
                      defaultValue="7"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      Password Minimum Length
                    </Label>
                    <Input
                      type="number"
                      defaultValue="8"
                    />
                  </div>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <Label htmlFor="requireSpecialChars" className="text-sm text-gray-900">
                        Require special characters in passwords
                      </Label>
                      <Switch id="requireSpecialChars" defaultChecked />
                    </div>
                    <div className="flex items-center justify-between">
                      <Label htmlFor="enable2FA" className="text-sm text-gray-900">
                        Enable two-factor authentication
                      </Label>
                      <Switch id="enable2FA" />
                    </div>
                    <div className="flex items-center justify-between">
                      <Label htmlFor="sessionTimeout" className="text-sm text-gray-900">
                        Enable automatic session timeout
                      </Label>
                      <Switch id="sessionTimeout" defaultChecked />
                    </div>
                  </div>
                </div>
              </Card>
            )}

            {activeTab === 'database' && (
              <Card className="p-6">
                <h2 className="text-xl font-semibold text-gray-900 mb-6">Database Settings</h2>
                <div className="space-y-6">
                  <div>
                    <Label className="mb-2 block">
                      Connection Pool Size
                    </Label>
                    <Input
                      type="number"
                      defaultValue="20"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      Query Timeout (seconds)
                    </Label>
                    <Input
                      type="number"
                      defaultValue="30"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      Backup Retention (days)
                    </Label>
                    <Input
                      type="number"
                      defaultValue="30"
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <Label htmlFor="autoBackup" className="text-sm text-gray-900">
                      Enable automatic backups
                    </Label>
                    <Switch id="autoBackup" defaultChecked />
                  </div>
                </div>
              </Card>
            )}

            {activeTab === 'email' && (
              <Card className="p-6">
                <h2 className="text-xl font-semibold text-gray-900 mb-6">Email Settings</h2>
                <div className="space-y-6">
                  <div>
                    <Label className="mb-2 block">
                      SMTP Server
                    </Label>
                    <Input
                      type="text"
                      defaultValue="smtp.gmail.com"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      SMTP Port
                    </Label>
                    <Input
                      type="number"
                      defaultValue="587"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      From Email
                    </Label>
                    <Input
                      type="email"
                      defaultValue="noreply@example.com"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      From Name
                    </Label>
                    <Input
                      type="text"
                      defaultValue="NILOAUTH"
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <Label htmlFor="enableTLS" className="text-sm text-gray-900">
                      Enable TLS encryption
                    </Label>
                    <Switch id="enableTLS" defaultChecked />
                  </div>
                </div>
              </Card>
            )}

            {activeTab === 'api' && (
              <Card className="p-6">
                <h2 className="text-xl font-semibold text-gray-900 mb-6">API Settings</h2>
                <div className="space-y-6">
                  <div>
                    <Label className="mb-2 block">
                      Rate Limit (requests per minute)
                    </Label>
                    <Input
                      type="number"
                      defaultValue="100"
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      API Version
                    </Label>
                    <Select defaultValue="v1">
                      <SelectTrigger>
                        <SelectValue placeholder="Select version" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="v1">v1</SelectItem>
                        <SelectItem value="v2">v2</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      CORS Origins (one per line)
                    </Label>
                    <Textarea
                      rows={4}
                      defaultValue="http://localhost:3000&#10;https://app.example.com"
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <Label htmlFor="enableAPILogs" className="text-sm text-gray-900">
                      Enable API request logging
                    </Label>
                    <Switch id="enableAPILogs" defaultChecked />
                  </div>
                </div>
              </Card>
            )}

            {activeTab === 'notifications' && (
              <Card className="p-6">
                <h2 className="text-xl font-semibold text-gray-900 mb-6">Notification Settings</h2>
                <div className="space-y-6">
                  <div className="space-y-4">
                    <h3 className="text-lg font-medium text-gray-900">Email Notifications</h3>
                    <div className="flex items-center justify-between">
                      <Label htmlFor="newUserNotification" className="text-sm text-gray-900">
                        New user registrations
                      </Label>
                      <Switch id="newUserNotification" defaultChecked />
                    </div>
                    <div className="flex items-center justify-between">
                      <Label htmlFor="securityAlerts" className="text-sm text-gray-900">
                        Security alerts
                      </Label>
                      <Switch id="securityAlerts" defaultChecked />
                    </div>
                    <div className="flex items-center justify-between">
                      <Label htmlFor="systemAlerts" className="text-sm text-gray-900">
                        System alerts
                      </Label>
                      <Switch id="systemAlerts" defaultChecked />
                    </div>
                  </div>
                  <div>
                    <Label className="mb-2 block">
                      Admin Email Recipients (comma-separated)
                    </Label>
                    <Input
                      type="text"
                      defaultValue="admin@example.com, security@example.com"
                    />
                  </div>
                </div>
              </Card>
            )}

            {/* Save Button */}
            <div className="mt-6">
              <Button className="flex items-center gap-2">
                <Save className="h-4 w-4" />
                Save Settings
              </Button>
            </div>
          </div>
        </div>
      </div>
    </SuperAdminLayout>
  )
}