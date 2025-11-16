import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { ArrowLeft, Lock, Eye, EyeOff, AlertCircle, CheckCircle, Shield } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import useAuthStore from '@/store/auth'
import type { ChangePasswordRequest } from '@/lib/types'

const Settings = () => {
  const navigate = useNavigate()
  const { changePassword, loading, error } = useAuthStore()

  const [passwordData, setPasswordData] = useState<ChangePasswordRequest>({
    current_password: '',
    new_password: '',
    confirm_password: ''
  })
  const [showPasswords, setShowPasswords] = useState({
    current: false,
    new: false,
    confirm: false
  })
  const [localError, setLocalError] = useState('')
  const [success, setSuccess] = useState(false)
  const [passwordStrength, setPasswordStrength] = useState({
    length: false,
    uppercase: false,
    lowercase: false,
    number: false,
    special: false
  })

  const handlePasswordInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setPasswordData(prev => ({ ...prev, [name]: value }))

    if (name === 'new_password') {
      checkPasswordStrength(value)
    }

    if (localError) setLocalError('')
    if (success) setSuccess(false)
  }

  const checkPasswordStrength = (password: string) => {
    setPasswordStrength({
      length: password.length >= 8,
      uppercase: /[A-Z]/.test(password),
      lowercase: /[a-z]/.test(password),
      number: /\d/.test(password),
      special: /[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]/.test(password)
    })
  }

  const isPasswordValid = Object.values(passwordStrength).every(Boolean)

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLocalError('')
    setSuccess(false)

    // Validate form
    if (!passwordData.current_password || !passwordData.new_password || !passwordData.confirm_password) {
      setLocalError('Please fill in all password fields')
      return
    }

    if (passwordData.new_password !== passwordData.confirm_password) {
      setLocalError('New passwords do not match')
      return
    }

    if (!isPasswordValid) {
      setLocalError('New password does not meet requirements')
      return
    }

    if (passwordData.current_password === passwordData.new_password) {
      setLocalError('New password must be different from current password')
      return
    }

    const result = await changePassword(passwordData)
    if (result.success) {
      setSuccess(true)
      setPasswordData({
        current_password: '',
        new_password: '',
        confirm_password: ''
      })
      setPasswordStrength({
        length: false,
        uppercase: false,
        lowercase: false,
        number: false,
        special: false
      })
    } else {
      setLocalError(result.message || 'Failed to change password')
    }
  }

  const togglePasswordVisibility = (field: keyof typeof showPasswords) => {
    setShowPasswords(prev => ({ ...prev, [field]: !prev[field] }))
  }

  const handleBack = () => {
    navigate('/dashboard')
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      {/* Header */}
      <header className="bg-white dark:bg-slate-800 shadow-sm border-b">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
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
            <h1 className="text-xl font-semibold text-foreground">
              Account Settings
            </h1>
            <div className="w-24"></div> {/* Spacer for centering */}
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="space-y-8"
        >
          {/* Password Change */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Lock className="h-5 w-5 text-primary" />
                <span>Change Password</span>
              </CardTitle>
              <CardDescription>
                Update your password to keep your account secure
              </CardDescription>
            </CardHeader>
            <CardContent>
              {/* Success/Error Messages */}
              {(error || localError) && (
                <Alert variant="destructive" className="mb-6">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>
                    {localError || error}
                  </AlertDescription>
                </Alert>
              )}

              {success && (
                <Alert className="mb-6 border-green-200 bg-green-50 text-green-800 dark:border-green-800 dark:bg-green-900 dark:text-green-200">
                  <CheckCircle className="h-4 w-4" />
                  <AlertDescription>
                    Password changed successfully!
                  </AlertDescription>
                </Alert>
              )}

              <form onSubmit={handlePasswordSubmit} className="space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="current_password">Current Password</Label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="current_password"
                      name="current_password"
                      type={showPasswords.current ? 'text' : 'password'}
                      placeholder="Enter your current password"
                      value={passwordData.current_password}
                      onChange={handlePasswordInputChange}
                      className="pl-10 pr-10"
                      disabled={loading}
                      required
                    />
                    <button
                      type="button"
                      onClick={() => togglePasswordVisibility('current')}
                      className="absolute right-3 top-3 text-muted-foreground hover:text-foreground"
                      disabled={loading}
                    >
                      {showPasswords.current ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="new_password">New Password</Label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="new_password"
                      name="new_password"
                      type={showPasswords.new ? 'text' : 'password'}
                      placeholder="Enter your new password"
                      value={passwordData.new_password}
                      onChange={handlePasswordInputChange}
                      className="pl-10 pr-10"
                      disabled={loading}
                      required
                    />
                    <button
                      type="button"
                      onClick={() => togglePasswordVisibility('new')}
                      className="absolute right-3 top-3 text-muted-foreground hover:text-foreground"
                      disabled={loading}
                    >
                      {showPasswords.new ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  </div>

                  {/* Password strength indicator */}
                  {passwordData.new_password && (
                    <div className="space-y-1 text-xs">
                      <div className={`flex items-center gap-2 ${passwordStrength.length ? 'text-green-600' : 'text-muted-foreground'}`}>
                        {passwordStrength.length ? <CheckCircle className="h-3 w-3" /> : <div className="h-3 w-3 rounded-full border" />}
                        At least 8 characters
                      </div>
                      <div className={`flex items-center gap-2 ${passwordStrength.uppercase ? 'text-green-600' : 'text-muted-foreground'}`}>
                        {passwordStrength.uppercase ? <CheckCircle className="h-3 w-3" /> : <div className="h-3 w-3 rounded-full border" />}
                        One uppercase letter
                      </div>
                      <div className={`flex items-center gap-2 ${passwordStrength.lowercase ? 'text-green-600' : 'text-muted-foreground'}`}>
                        {passwordStrength.lowercase ? <CheckCircle className="h-3 w-3" /> : <div className="h-3 w-3 rounded-full border" />}
                        One lowercase letter
                      </div>
                      <div className={`flex items-center gap-2 ${passwordStrength.number ? 'text-green-600' : 'text-muted-foreground'}`}>
                        {passwordStrength.number ? <CheckCircle className="h-3 w-3" /> : <div className="h-3 w-3 rounded-full border" />}
                        One number
                      </div>
                      <div className={`flex items-center gap-2 ${passwordStrength.special ? 'text-green-600' : 'text-muted-foreground'}`}>
                        {passwordStrength.special ? <CheckCircle className="h-3 w-3" /> : <div className="h-3 w-3 rounded-full border" />}
                        One special character
                      </div>
                    </div>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="confirm_password">Confirm New Password</Label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="confirm_password"
                      name="confirm_password"
                      type={showPasswords.confirm ? 'text' : 'password'}
                      placeholder="Confirm your new password"
                      value={passwordData.confirm_password}
                      onChange={handlePasswordInputChange}
                      className="pl-10 pr-10"
                      disabled={loading}
                      required
                    />
                    <button
                      type="button"
                      onClick={() => togglePasswordVisibility('confirm')}
                      className="absolute right-3 top-3 text-muted-foreground hover:text-foreground"
                      disabled={loading}
                    >
                      {showPasswords.confirm ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  </div>
                </div>

                <Button
                  type="submit"
                  className="w-full"
                  disabled={loading || !isPasswordValid}
                >
                  {loading ? 'Changing Password...' : 'Change Password'}
                </Button>
              </form>
            </CardContent>
          </Card>

          <Separator />

          {/* Account Security */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Shield className="h-5 w-5 text-primary" />
                <span>Account Security</span>
              </CardTitle>
              <CardDescription>
                Additional security settings and information
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between p-4 bg-gray-50 dark:bg-slate-700 rounded-lg">
                  <div>
                    <h4 className="font-medium text-foreground">Two-Factor Authentication</h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Add an extra layer of security to your account
                    </p>
                  </div>
                  <Button variant="outline" size="sm" disabled>
                    Coming Soon
                  </Button>
                </div>

                <div className="flex items-center justify-between p-4 bg-gray-50 dark:bg-slate-700 rounded-lg">
                  <div>
                    <h4 className="font-medium text-foreground">Login Sessions</h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      View and manage your active login sessions
                    </p>
                  </div>
                  <Button variant="outline" size="sm" disabled>
                    Coming Soon
                  </Button>
                </div>

                <div className="flex items-center justify-between p-4 bg-gray-50 dark:bg-slate-700 rounded-lg">
                  <div>
                    <h4 className="font-medium text-foreground">Account Deletion</h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Permanently delete your account and all associated data
                    </p>
                  </div>
                  <Button variant="destructive" size="sm" disabled>
                    Coming Soon
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </main>
    </div>
  )
}

export default Settings