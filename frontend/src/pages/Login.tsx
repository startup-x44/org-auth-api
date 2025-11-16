import React, { useState, useEffect } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { Eye, EyeOff, Mail, Lock, AlertCircle, LogIn, Sparkles, ArrowRight, CheckCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import LoadingSpinner from '@/components/ui/loading-spinner'
import useAuthStore from '@/store/auth'

const Login = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { login, error, clearError } = useAuthStore() // Removed 'loading' from store

  const emailFromParams = searchParams.get('email')
  const isFromRegistration = searchParams.get('registered') === 'true'

  const [formData, setFormData] = useState({
    email: emailFromParams || '',
    password: ''
  })
  const [showPassword, setShowPassword] = useState(false)
  const [localError, setLocalError] = useState('')
  const [isLoading, setIsLoading] = useState(false) // Local loading state

  // Clear errors when component mounts
  useEffect(() => {
    console.log('🟣 Login component MOUNTED')
    if (clearError) clearError()
    setLocalError('')
    
    return () => {
      console.log('🟣 Login component UNMOUNTING')
    }
  }, [])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))

    // Clear errors when typing
    if (localError) setLocalError('')
    if (error && clearError) clearError()
  }

  const handleSubmit = async (e: React.FormEvent) => {
    console.log('🔵 handleSubmit START')
    e.preventDefault()
    e.stopPropagation()
    console.log('✅ preventDefault called')
    
    // Clear any previous errors
    setLocalError('')
    if (clearError) clearError()

    // Basic validation
    if (!formData.email || !formData.password) {
      console.log('🔴 Validation failed - empty fields')
      setLocalError('Please fill in all fields')
      return
    }

    try {
      console.log('🔵 Setting isLoading to true (local state, no Zustand trigger)')
      setIsLoading(true)
      
      console.log('🔵 Calling login API...')
      const result = await login(formData.email, formData.password)
      console.log('🔵 Login API result:', result)
      
      if (!result || result.success !== true) {
        console.log('🔴 Login failed:', result?.message)
        setLocalError(result?.message || 'Invalid email or password')
        console.log('🔵 Error set, returning without navigation')
        return
      }

      // Successful login
      console.log('✅ Login successful, navigating...')
      if (result.organizations && result.organizations.length > 0) {
        navigate('/choose-organization', { replace: true })
      } else {
        navigate('/create-organization', { replace: true })
      }

    } catch (err) {
      console.error('🔴 Unexpected login error:', err)
      setLocalError('An unexpected error occurred. Please try again.')
    } finally {
      console.log('🔵 Setting isLoading to false')
      setIsLoading(false)
    }

    console.log('🔵 handleSubmit END')
    return false
  }

  const togglePasswordVisibility = () => {
    setShowPassword(prev => !prev)
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 dark:from-slate-900 dark:via-blue-900 dark:to-slate-900 p-4">
      <div className="absolute inset-0 bg-grid-slate-200/50 [mask-image:radial-gradient(ellipse_at_center,white,transparent)] dark:bg-grid-slate-700/30" />

      <div className="absolute top-20 left-20 w-72 h-72 bg-blue-400/30 rounded-full blur-3xl opacity-40" />
      <div className="absolute bottom-20 right-20 w-96 h-96 bg-purple-400/20 rounded-full blur-3xl opacity-30" />

      <div className="w-full max-w-md relative z-10">
        <Card className="shadow-2xl border border-slate-200/50 dark:border-slate-700/50 bg-white/90 dark:bg-slate-800/90 backdrop-blur-xl overflow-hidden">

          {/* Header */}
          <div className="bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 p-8 text-center">
            <div className="mx-auto w-20 h-20 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center shadow-lg mb-4">
              <LogIn className="h-10 w-10 text-white" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-white mb-2">Welcome Back</h1>
              <p className="text-blue-100">Sign in to access your workspace</p>
            </div>
          </div>

          <CardContent className="p-8">
            <form onSubmit={handleSubmit} noValidate className="space-y-6">

              {isFromRegistration && (
                <Alert className="border-green-200 bg-green-50 dark:bg-green-950/50">
                  <CheckCircle className="h-4 w-4 text-green-600" />
                  <AlertDescription className="font-medium text-green-800 dark:text-green-200">
                    Registration successful! Please sign in to create your workspace.
                  </AlertDescription>
                </Alert>
              )}

              <div style={{ minHeight: (error || localError) ? 'auto' : 0 }}>
                {(error || localError) && (
                  <Alert variant="destructive" className="border-red-200 bg-red-50 dark:bg-red-950/50">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription className="font-medium">
                      {localError || error}
                    </AlertDescription>
                  </Alert>
                )}
              </div>

              {/* Email */}
              <div className="space-y-2">
                <Label htmlFor="email" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                  Email Address
                </Label>
                <div className="relative group">
                  <Mail className="absolute left-3 top-3.5 h-5 w-5 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
                  <Input
                    id="email"
                    name="email"
                    type="email"
                    placeholder="you@example.com"
                    value={formData.email}
                    onChange={handleInputChange}
                    disabled={isLoading}
                    className="pl-11 h-12 text-lg"
                  />
                </div>
              </div>

              {/* Password */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="password" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                    Password
                  </Label>
                  <Link to="/forgot-password" className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400">
                    Forgot?
                  </Link>
                </div>

                <div className="relative group">
                  <Lock className="absolute left-3 top-3.5 h-5 w-5 text-slate-400 group-focus-within:text-blue-500 transition-colors" />

                  <Input
                    id="password"
                    name="password"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Enter your password"
                    value={formData.password}
                    onChange={handleInputChange}
                    disabled={isLoading}
                    className="pl-11 pr-11 h-12 text-lg"
                  />

                  <button
                    type="button"
                    onClick={togglePasswordVisibility}
                    className="absolute right-3 top-3.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300"
                    tabIndex={-1}
                    disabled={isLoading}
                  >
                    {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                  </button>
                </div>
              </div>

              {/* Submit */}
              <Button
                type="submit"
                disabled={isLoading}
                className="w-full h-12 bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 text-white font-semibold"
              >
                {isLoading ? (
                  <>
                    <LoadingSpinner size="sm" className="mr-2" />
                    Signing in...
                  </>
                ) : (
                  <>
                    Sign In
                    <ArrowRight className="ml-2 h-5 w-5" />
                  </>
                )}
              </Button>
            </form>

            {/* Footer */}
            <div className="mt-8">
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-slate-300 dark:border-slate-600" />
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-4 bg-white dark:bg-slate-800 text-slate-500">
                    New to our platform?
                  </span>
                </div>
              </div>

              <div className="mt-6 text-center">
                <Link
                  to="/register"
                  className="inline-flex items-center gap-2 text-blue-600 hover:text-blue-700 dark:text-blue-400"
                >
                  <Sparkles className="h-4 w-4" />
                  Create your account
                  <ArrowRight className="h-4 w-4" />
                </Link>
              </div>
            </div>
          </CardContent>
        </Card>

        <div className="mt-6 text-center">
          <p className="text-sm text-slate-600 dark:text-slate-400">
            Protected by enterprise-grade security
          </p>
        </div>
      </div>
    </div>
  )
}

export default Login
