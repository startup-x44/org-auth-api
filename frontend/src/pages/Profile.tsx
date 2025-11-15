import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { ArrowLeft, User, Mail, Phone, Save, AlertCircle, CheckCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import LoadingSpinner from '@/components/ui/loading-spinner'
import useAuthStore from '@/store/auth'
import type { UpdateProfileRequest } from '@/lib/types'

const Profile = () => {
  const navigate = useNavigate()
  const { user, updateProfile, loading, error } = useAuthStore()

  const [formData, setFormData] = useState<UpdateProfileRequest>({
    first_name: '',
    last_name: '',
    phone: ''
  })
  const [localError, setLocalError] = useState('')
  const [success, setSuccess] = useState(false)

  // Initialize form with current user data
  useEffect(() => {
    if (user) {
      setFormData({
        first_name: user.first_name || '',
        last_name: user.last_name || '',
        phone: user.phone || ''
      })
    }
  }, [user])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
    if (localError) setLocalError('')
    if (success) setSuccess(false)
  }

  const getInitials = (firstName?: string, lastName?: string) => {
    const first = firstName || ''
    const last = lastName || ''
    return `${first.charAt(0)}${last.charAt(0)}`.toUpperCase()
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLocalError('')
    setSuccess(false)

    // Validate required fields
    if (!formData.first_name?.trim() || !formData.last_name?.trim()) {
      setLocalError('First name and last name are required')
      return
    }

    const result = await updateProfile(formData)
    if (result.success) {
      setSuccess(true)
    } else {
      setLocalError(result.message || 'Failed to update profile')
    }
  }

  const handleBack = () => {
    navigate('/dashboard')
  }

  if (!user) {
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
              Profile Settings
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
          {/* Profile Overview */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <User className="h-5 w-5 text-primary" />
                <span>Profile Information</span>
              </CardTitle>
              <CardDescription>
                Update your personal information and contact details
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center space-x-6 mb-6">
                <Avatar className="h-20 w-20">
                  <AvatarFallback className="text-2xl">
                    {getInitials(user.first_name, user.last_name)}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <h3 className="text-lg font-semibold text-foreground">
                    {user.first_name} {user.last_name}
                  </h3>
                  <p className="text-gray-600 dark:text-gray-400">{user.email}</p>
                  <p className="text-sm text-gray-500 dark:text-gray-500 mt-1">
                    Member since {new Date(user.created_at).toLocaleDateString()}
                  </p>
                </div>
              </div>

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
                    Profile updated successfully!
                  </AlertDescription>
                </Alert>
              )}

              {/* Profile Form */}
              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="first_name">First Name *</Label>
                    <div className="relative">
                      <User className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                      <Input
                        id="first_name"
                        name="first_name"
                        type="text"
                        placeholder="Enter your first name"
                        value={formData.first_name}
                        onChange={handleInputChange}
                        className="pl-10"
                        disabled={loading}
                        required
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="last_name">Last Name *</Label>
                    <div className="relative">
                      <User className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                      <Input
                        id="last_name"
                        name="last_name"
                        type="text"
                        placeholder="Enter your last name"
                        value={formData.last_name}
                        onChange={handleInputChange}
                        className="pl-10"
                        disabled={loading}
                        required
                      />
                    </div>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="phone">Phone Number</Label>
                  <div className="relative">
                    <Phone className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="phone"
                      name="phone"
                      type="tel"
                      placeholder="Enter your phone number"
                      value={formData.phone}
                      onChange={handleInputChange}
                      className="pl-10"
                      disabled={loading}
                    />
                  </div>
                </div>

                {/* Read-only fields */}
                <div className="space-y-4 pt-4 border-t">
                  <div className="space-y-2">
                    <Label>Email Address</Label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                      <Input
                        type="email"
                        value={user.email}
                        className="pl-10 bg-gray-50 dark:bg-gray-800"
                        disabled
                      />
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Email address cannot be changed. Contact support if you need to update it.
                    </p>
                  </div>
                </div>

                <div className="flex justify-end space-x-4 pt-6">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleBack}
                    disabled={loading}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={loading}
                    className="flex items-center space-x-2"
                  >
                    <Save className="h-4 w-4" />
                    <span>{loading ? 'Saving...' : 'Save Changes'}</span>
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </motion.div>
      </main>
    </div>
  )
}

export default Profile