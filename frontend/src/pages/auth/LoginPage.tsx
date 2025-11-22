import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Link, useNavigate, useLocation, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'react-hot-toast'
import { 
  Mail, 
  Lock, 
  Building, 
  ArrowRight,
  Eye,
  EyeOff
} from 'lucide-react'

import { loginSchema, type LoginFormData, AUTH_ERROR_MESSAGES } from '@/schemas/auth'
import { useAuth } from '@/hooks/useAuth'
import { useAuthStore } from '@/stores/authStore'
import { logger } from '@/utils/logger'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Alert, AlertDescription } from '@/components/ui/alert'

import AuthDebugInfo from '@/components/debug/AuthDebugInfo'

export const LoginPage: React.FC = React.memo(() => {
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams] = useSearchParams()
  
  const { login, clearError, error: authError } = useAuth()
  
  const [isLoading, setIsLoading] = useState(false)
  const [showOrgField, setShowOrgField] = useState(false)
  const [loginError, setLoginError] = useState<string | null>(null)

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
    if (emailValue && emailValue.includes('@') && emailValue.length > 3) {
      const domain = emailValue.split('@')[1]
      const shouldShow = Boolean(domain && !['gmail.com', 'yahoo.com', 'hotmail.com'].includes(domain))
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
      
      // Small delay to ensure auth store is updated
      setTimeout(() => {
        // Get fresh user data from auth store
        const user = useAuthStore.getState().user
        
        if (user) {
          // Role-based redirect logic
          if (user.is_superadmin || user.global_role === 'superadmin') {
            navigate('/admin/dashboard', { replace: true })
          } else if (user.roles?.some(r => r.name === 'admin')) {
            navigate('/admin/dashboard', { replace: true })
          } else if (user.roles?.some(r => ['user', 'member'].includes(r.name))) {
            navigate('/user/dashboard', { replace: true })
          } else {
            // Fallback to requested redirect or default dashboard
            navigate(redirectTo, { replace: true })
          }
        } else {
          // Fallback if user data is not available
          navigate(redirectTo, { replace: true })
        }
      }, 100)

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
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center">
        <h2 className="text-2xl font-bold tracking-tight text-gray-900">
          Welcome back
        </h2>
        <p className="text-sm text-gray-600 mt-2">
          Sign in to your account to continue
        </p>
      </div>

      {/* Error Alert */}
      {(loginError || authError) && (
        <Alert variant="destructive">
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

      {/* Debug info in development */}
      <AuthDebugInfo />
    </div>
  )
})

LoginPage.displayName = 'LoginPage'