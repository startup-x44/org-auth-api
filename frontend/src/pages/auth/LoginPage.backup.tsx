import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Link, useNavigate, useLocation, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'react-hot-toast'
import { 
  Mail, 
  Lock, 
  Building, 
  Shield, 
  CheckCircle, 
  Users, 
  Zap, 
  ArrowRight,
  Eye,
  EyeOff
} from 'lucide-react'

import { loginSchema, type LoginFormData, AUTH_ERROR_MESSAGES } from '@/schemas/auth'
import { useAuth } from '@/hooks/useAuth'
import { logger } from '@/utils/logger'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Card, CardHeader, CardContent } from '@/components/ui/card'

// Debug imports temporarily disabled
// import {
//   useRenderTracker,
//   useStateTracker,
//   useLifecycleDebugger
// } from '@/utils/debug'

import AuthDebugInfo from '@/components/debug/AuthDebugInfo'

export const LoginPage: React.FC = React.memo(() => {
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams] = useSearchParams()
  
  const { login, clearError, error: authError } = useAuth()
  
  const [isLoading, setIsLoading] = useState(false)
  const [showOrgField, setShowOrgField] = useState(false)
  const [loginError, setLoginError] = useState<string | null>(null)

  // Debug - temporarily disabled to prevent infinite loops
  // useLifecycleDebugger('LoginPage')
  // useRenderTracker('LoginPage')
  // useStateTracker('LoginPage', { isLoading, showOrgField, loginError, authError })

  const redirectTo = useMemo(() =>
    (location.state as any)?.from || searchParams.get('redirect') || '/dashboard',
    [location.state, searchParams]
  )

  const orgSlugFromUrl = useMemo(() => searchParams.get('org') || '', [searchParams])

  const {
    register,
    handleSubmit,
    formState: { errors, isValid },
    watch,
    setValue,
    setError,
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
      organizationSlug: orgSlugFromUrl,
      rememberMe: false,
    },
    mode: 'onChange',
  })

  const emailValue = watch('email')

  useEffect(() => {
    clearError()
    setLoginError(null)
  }, [])

  useEffect(() => {
    if (emailValue && emailValue.includes('@')) {
      const domain = emailValue.split('@')[1]
      const shouldShow = domain && !['gmail.com', 'yahoo.com', 'hotmail.com'].includes(domain)
      setShowOrgField(shouldShow)
    }
  }, [emailValue])

  const onSubmit = useCallback(async (data: LoginFormData) => {
    let timeoutId: NodeJS.Timeout | undefined

    try {
      setIsLoading(true)
      setLoginError(null)
      clearError()

      const creds = {
        email: data.email,
        password: data.password,
        rememberMe: data.rememberMe,
        ...(data.organizationSlug && { organizationSlug: data.organizationSlug })
      }

      timeoutId = setTimeout(() => {
        setIsLoading(false)
        setLoginError('Login timed out. Please try again.')
      }, 10000)

      const result = await login(creds)

      if (timeoutId) clearTimeout(timeoutId)

      if (result.requiresMFA) {
        navigate('/auth/mfa', {
          state: {
            email: data.email,
            from: redirectTo
          }
        })
        return
      }

      toast.success('Welcome back!')
      navigate(redirectTo, { replace: true })

    } catch (error: any) {
      logger.error('Login failed', error)

      const errorCode = error.code as keyof typeof AUTH_ERROR_MESSAGES
      const message = AUTH_ERROR_MESSAGES[errorCode] || error.message || 'Login failed'

      setLoginError(message)

      if (errorCode === 'INVALID_CREDENTIALS') {
        setError('password', { message: 'Invalid email or password' })
      }

      toast.error(message)
    } finally {
      if (timeoutId) clearTimeout(timeoutId)
      setIsLoading(false)
    }
  }, [login, navigate, redirectTo, clearError, setError])

  const toggleOrgField = useCallback(() => {
    setShowOrgField(prev => {
      if (prev) setValue('organizationSlug', '')
      return !prev
    })
  }, [setValue])

  // Add password visibility state
  const [showPassword, setShowPassword] = useState(false)

  return (
    <div className="min-h-screen flex">
      {/* LEFT SIDE - MARKETING */}
      <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-blue-600 via-blue-700 to-indigo-800 relative overflow-hidden">
     
        <div className="absolute inset-0 opacity-10">
          <div className="absolute top-20 left-20 w-32 h-32 bg-white rounded-full animate-pulse"></div>
          <div className="absolute top-60 right-32 w-24 h-24 bg-white rounded-full animate-pulse delay-100"></div>
          <div className="absolute bottom-40 left-1/3 w-40 h-40 bg-white rounded-full animate-pulse delay-200"></div>
        </div>
        
        <div className="relative z-10 flex flex-col justify-center px-12 xl:px-16 text-white">
       
          <div className="flex items-center mb-10">
            <div className="h-16 w-16 bg-white/20 rounded-2xl flex items-center justify-center backdrop-blur-sm">
              <Shield className="h-9 w-9 text-white" />
            </div>
            <div className="ml-4">
              <h1 className="text-3xl font-bold">NILOAUTH</h1>
              <p className="text-blue-200 text-lg">Enterprise Authentication</p>
            </div>
          </div>

        
          <div className="max-w-lg">
            <h2 className="text-5xl font-bold mb-6 leading-tight">
              Secure by
              <span className="block text-transparent bg-clip-text bg-gradient-to-r from-blue-200 to-white">
                Design
              </span>
            </h2>
            
            <p className="text-xl text-blue-100 mb-10 leading-relaxed">
              Enterprise-grade authentication with zero-trust security, built for modern applications.
            </p>

         
            <div className="space-y-6">
              <div className="flex items-start">
                <div className="h-6 w-6 bg-emerald-400 rounded-full flex items-center justify-center mr-4 mt-1 flex-shrink-0">
                  <CheckCircle className="h-4 w-4 text-emerald-900" />
                </div>
                <div>
                  <p className="font-semibold text-lg">Multi-Factor Authentication</p>
                  <p className="text-blue-200 mt-1">TOTP, SMS, and hardware key support</p>
                </div>
              </div>

              <div className="flex items-start">
                <div className="h-6 w-6 bg-emerald-400 rounded-full flex items-center justify-center mr-4 mt-1 flex-shrink-0">
                  <Users className="h-4 w-4 text-emerald-900" />
                </div>
                <div>
                  <p className="font-semibold text-lg">Role-Based Access Control</p>
                  <p className="text-blue-200 mt-1">Granular permissions and organization management</p>
                </div>
              </div>

              <div className="flex items-start">
                <div className="h-6 w-6 bg-emerald-400 rounded-full flex items-center justify-center mr-4 mt-1 flex-shrink-0">
                  <Zap className="h-4 w-4 text-emerald-900" />
                </div>
                <div>
                  <p className="font-semibold text-lg">Enterprise SSO</p>
                  <p className="text-blue-200 mt-1">OAuth 2.0, SAML, and OpenID Connect</p>
                </div>
              </div>
            </div>

           
            <div className="mt-10 pt-6 border-t border-blue-500/30">
              <p className="text-blue-200 text-sm">
                Trusted by 1000+ organizations worldwide
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* RIGHT SIDE - LOGIN FORM */}
      <div className="flex-1 lg:w-1/2 flex items-center justify-center min-h-screen py-12 px-4 sm:px-6 lg:px-8 bg-gray-50 relative">
        {/* Background decoration */}
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute top-[-10%] right-[-10%] w-80 h-80 bg-gradient-to-br from-blue-100 to-indigo-100 rounded-full opacity-20 blur-3xl"></div>
          <div className="absolute bottom-[-10%] left-[-10%] w-60 h-60 bg-gradient-to-tr from-purple-100 to-pink-100 rounded-full opacity-20 blur-3xl"></div>
        </div>

        <Card className="w-full max-w-md relative z-10 shadow-2xl border-0 bg-white/95 backdrop-blur-sm">
          <CardHeader className="space-y-1 pb-6">
            {/* Mobile Logo */}
            <div className="lg:hidden flex items-center justify-center mb-6">
              <div className="h-12 w-12 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-xl flex items-center justify-center shadow-lg">
                <Shield className="h-6 w-6 text-white" />
              </div>
              <div className="ml-3">
                <h1 className="text-xl font-bold text-gray-900">NILOAUTH</h1>
                <p className="text-sm text-gray-600">Enterprise Auth</p>
              </div>
            </div>

            <div className="text-center lg:text-left">
              <h2 className="text-2xl font-bold tracking-tight text-gray-900">
                Welcome back
              </h2>
              <p className="text-sm text-gray-600 mt-2">
                Sign in to your account to continue
              </p>
            </div>
          </CardHeader>

          <CardContent className="space-y-4">
            {/* Error Alert */}
            {(loginError || authError) && (
              <Alert variant="destructive" className="mb-4">
                <AlertDescription>
                  {loginError || authError}
                </AlertDescription>
              </Alert>
            )}

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {/* Email Field */}
              <div className="space-y-2">
                <Label htmlFor="email">Email address</Label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <Input
                    {...register('email')}
                    id="email"
                    type="email"
                    placeholder="Enter your email"
                    className="pl-10 h-11"
                    autoComplete="email"
                    autoFocus
                  />
                </div>
                {errors.email && (
                  <p className="text-sm text-red-600">{errors.email.message}</p>
                )}
              </div>

              {/* Organization Field (conditional) */}
              {showOrgField && (
                <div className="space-y-2 animate-in slide-in-from-top-2 duration-300">
                  <Label htmlFor="organization">Organization (optional)</Label>
                  <div className="relative">
                    <Building className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                    <Input
                      {...register('organizationSlug')}
                      id="organization"
                      type="text"
                      placeholder="your-organization"
                      className="pl-10 h-11"
                    />
                  </div>
                  {errors.organizationSlug && (
                    <p className="text-sm text-red-600">{errors.organizationSlug.message}</p>
                  )}
                  <p className="text-xs text-gray-500">
                    Enter your organization slug if signing into a specific organization
                  </p>
                </div>
              )}

              {/* Toggle Organization Field */}
              <div className="text-center">
                <button
                  type="button"
                  onClick={toggleOrgField}
                  className="text-sm text-blue-600 hover:text-blue-700 transition-colors font-medium inline-flex items-center"
                >
                  {showOrgField ? 'Hide organization field' : 'Sign in to specific organization'}
                  <ArrowRight className="ml-1 h-3 w-3" />
                </button>
              </div>

              {/* Password Field */}
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <Input
                    {...register('password')}
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Enter your password"
                    className="pl-10 pr-10 h-11"
                    autoComplete="current-password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400 hover:text-gray-600"
                  >
                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
                {errors.password && (
                  <p className="text-sm text-red-600">{errors.password.message}</p>
                )}
              </div>

              {/* Remember Me & Forgot Password */}
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Checkbox 
                    id="remember"
                    defaultChecked={false}
                    onCheckedChange={(checked) => setValue('rememberMe', Boolean(checked))}
                  />
                  <Label htmlFor="remember" className="text-sm font-normal">
                    Remember me
                  </Label>
                </div>
                <Link
                  to="/auth/forgot-password"
                  className="text-sm text-blue-600 hover:text-blue-700 transition-colors font-medium"
                >
                  Forgot password?
                </Link>
              </div>

              {/* Submit Button */}
              <Button
                type="submit"
                className="w-full h-11 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200"
                disabled={!isValid || isLoading}
              >
                {isLoading ? 'Signing in...' : 'Sign in'}
              </Button>
            </form>

            {/* Sign Up Link */}
            <div className="text-center pt-4 border-t">
              <p className="text-sm text-gray-600">
                Don't have an account?{' '}
                <Link
                  to="/auth/register"
                  className="font-medium text-blue-600 hover:text-blue-700 transition-colors"
                >
                  Sign up here
                </Link>
              </p>
            </div>

            {/* Footer Links */}
            <div className="text-center pt-4 space-y-2">
              <p className="text-xs text-gray-500">
                Need help?{' '}
                <Link to="/support" className="text-blue-600 hover:text-blue-700 transition-colors">
                  Contact Support
                </Link>
              </p>
              <p className="text-xs text-gray-400">
                By signing in, you agree to our{' '}
                <Link to="/terms" className="text-blue-600 hover:text-blue-700 transition-colors">
                  Terms
                </Link>{' '}
                and{' '}
                <Link to="/privacy" className="text-blue-600 hover:text-blue-700 transition-colors">
                  Privacy Policy
                </Link>
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Debug info in development */}
      <AuthDebugInfo />
    </div>
  )
})

LoginPage.displayName = 'LoginPage'
export default LoginPage
