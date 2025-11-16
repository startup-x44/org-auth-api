import React, { useState, useEffect } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { Mail, ArrowRight, AlertCircle, CheckCircle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import LoadingSpinner from '@/components/ui/loading-spinner'
import { publicAPI } from '@/lib/axios-instance'

const VerifyEmail = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const emailFromParams = searchParams.get('email') || ''

  const [email] = useState(emailFromParams)
  const [code, setCode] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isResending, setIsResending] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [canResend, setCanResend] = useState(true)
  const [resendTimer, setResendTimer] = useState(0)

  // Countdown timer for resend rate limiting
  useEffect(() => {
    if (resendTimer > 0) {
      const timer = setTimeout(() => setResendTimer(resendTimer - 1), 1000)
      return () => clearTimeout(timer)
    } else {
      setCanResend(true)
    }
  }, [resendTimer])

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/\D/g, '').slice(0, 6) // Only digits, max 6
    setCode(value)
    if (error) setError('')
    if (success) setSuccess('')
  }

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault()
    e.stopPropagation()

    if (!email) {
      setError('Email address is required')
      return
    }

    if (code.length !== 6) {
      setError('Please enter the complete 6-digit verification code')
      return
    }

    setIsLoading(true)
    setError('')
    setSuccess('')

    try {
      const response = await publicAPI.post('/auth/verify-email', { email, code })
      
      if (response.data.success) {
        setSuccess('Email verified successfully! Redirecting to login...')
        setTimeout(() => {
          navigate(`/login?email=${encodeURIComponent(email)}`)
        }, 2000)
      }
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || 'Verification failed. Please try again.'
      setError(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }

  const handleResend = async () => {
    if (!canResend || !email) return

    setIsResending(true)
    setError('')
    setSuccess('')

    try {
      const response = await publicAPI.post('/auth/resend-verification', { email })
      
      if (response.data.success) {
        setSuccess('Verification code sent! Please check your email.')
        setCanResend(false)
        setResendTimer(60) // 60-second cooldown
        setCode('') // Clear the code input
      }
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || 'Failed to resend code. Please try again.'
      
      // Check if it's a rate limit error
      if (err.response?.status === 429) {
        setError('Please wait before requesting another code')
        setCanResend(false)
        setResendTimer(60)
      } else {
        setError(errorMessage)
      }
    } finally {
      setIsResending(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-purple-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-green-500 to-emerald-600 rounded-2xl mb-4 shadow-lg">
            <Mail className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Verify Your Email</h1>
          <p className="text-gray-600">
            We've sent a 6-digit verification code to
          </p>
          <p className="font-semibold text-gray-900 mt-1">{email || 'your email'}</p>
        </div>

        {/* Main Card */}
        <Card className="shadow-xl border-0">
          <CardContent className="p-8">
            <form onSubmit={handleVerify} className="space-y-6">
              {/* Success Alert */}
              {success && (
                <Alert className="bg-green-50 border-green-200">
                  <CheckCircle className="h-4 w-4 text-green-600" />
                  <AlertDescription className="text-green-800 ml-2">
                    {success}
                  </AlertDescription>
                </Alert>
              )}

              {/* Error Alert */}
              {error && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription className="ml-2">{error}</AlertDescription>
                </Alert>
              )}

              {/* Code Input */}
              <div className="space-y-2">
                <Label htmlFor="code" className="text-sm font-medium text-gray-700">
                  Verification Code
                </Label>
                <Input
                  id="code"
                  name="code"
                  type="text"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  value={code}
                  onChange={handleCodeChange}
                  placeholder="000000"
                  maxLength={6}
                  className="text-center text-2xl tracking-widest font-mono"
                  disabled={isLoading || !!success}
                  required
                />
                <p className="text-xs text-gray-500 text-center">
                  Enter the 6-digit code from your email
                </p>
              </div>

              {/* Verify Button */}
              <Button
                type="submit"
                className="w-full bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 text-white shadow-lg transition-all duration-200"
                disabled={isLoading || code.length !== 6 || !!success}
              >
                {isLoading ? (
                  <>
                    <LoadingSpinner size="sm" className="mr-2" />
                    Verifying...
                  </>
                ) : (
                  <>
                    Verify Email
                    <ArrowRight className="ml-2 w-4 h-4" />
                  </>
                )}
              </Button>

              {/* Resend Section */}
              <div className="pt-4 border-t border-gray-200">
                <div className="flex items-center justify-between">
                  <p className="text-sm text-gray-600">Didn't receive the code?</p>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={handleResend}
                    disabled={!canResend || isResending || !!success}
                    className="text-indigo-600 hover:text-indigo-700 hover:bg-indigo-50"
                  >
                    {isResending ? (
                      <>
                        <LoadingSpinner size="sm" className="mr-2" />
                        Sending...
                      </>
                    ) : resendTimer > 0 ? (
                      `Resend in ${resendTimer}s`
                    ) : (
                      <>
                        <RefreshCw className="w-4 h-4 mr-2" />
                        Resend Code
                      </>
                    )}
                  </Button>
                </div>
              </div>

              {/* Back to Login */}
              <div className="text-center pt-4">
                <p className="text-sm text-gray-600">
                  Already verified?{' '}
                  <Link
                    to={`/login${email ? `?email=${encodeURIComponent(email)}` : ''}`}
                    className="font-medium text-indigo-600 hover:text-indigo-700 transition-colors"
                  >
                    Back to Login
                  </Link>
                </p>
              </div>
            </form>
          </CardContent>
        </Card>

        {/* Security Tip */}
        <div className="mt-6 text-center">
          <p className="text-xs text-gray-500">
            🔒 For your security, the code will expire in 15 minutes
          </p>
        </div>
      </div>
    </div>
  )
}

export default VerifyEmail
