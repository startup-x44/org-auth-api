import React, { useState, useEffect, useRef } from 'react'
import { useNavigate, useLocation, Link } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'react-hot-toast'
import { Shield, ArrowLeft, RefreshCw } from 'lucide-react'

import { mfaSchema, type MFAFormData, AUTH_ERROR_MESSAGES } from '@/schemas/auth'
import { useAuth } from '@/hooks/useAuth'
import { logger } from '@/utils/logger'
import { Button, Alert } from '@/components/ui/forms'

interface MFAPageProps {}

export const MFAPage: React.FC<MFAPageProps> = () => {
  const navigate = useNavigate()
  const location = useLocation()
  
  const { verifyMFA, mfaChallenge, clearError, error: authError } = useAuth()
  
  const [isLoading, setIsLoading] = useState(false)
  const [mfaError, setMfaError] = useState<string | null>(null)
  const [timeLeft, setTimeLeft] = useState(300) // 5 minutes
  
  const inputRefs = useRef<(HTMLInputElement | null)[]>([])

  // Get state from location (email, challengeId, etc.)
  const state = location.state as {
    email?: string
    from?: string
  } | null

  const redirectTo = state?.from || '/dashboard'

  const {
    handleSubmit,
    formState: { isValid },
    setValue,
    watch,
    setError,
  } = useForm<MFAFormData>({
    resolver: zodResolver(mfaSchema),
    defaultValues: {
      code: '',
      challengeId: mfaChallenge?.challengeId || '',
    },
  })

  const codeValue = watch('code')

  // Redirect if no MFA challenge
  useEffect(() => {
    if (!mfaChallenge) {
      navigate('/auth/login', { replace: true })
    }
  }, [mfaChallenge, navigate])

  // Update challenge ID when it changes
  useEffect(() => {
    if (mfaChallenge?.challengeId) {
      setValue('challengeId', mfaChallenge.challengeId)
    }
  }, [mfaChallenge, setValue])

  // Timer countdown
  useEffect(() => {
    if (timeLeft <= 0) return

    const timer = setInterval(() => {
      setTimeLeft((prev) => prev - 1)
    }, 1000)

    return () => clearInterval(timer)
  }, [timeLeft])

  // Auto-focus on first input
  useEffect(() => {
    inputRefs.current[0]?.focus()
  }, [])

  const handleCodeChange = (value: string, index: number) => {
    // Only allow digits
    const sanitizedValue = value.replace(/\D/g, '')
    
    if (sanitizedValue.length > 1) {
      // If pasting multiple digits, distribute them
      const digits = sanitizedValue.slice(0, 6).split('')
      
      digits.forEach((digit, i) => {
        if (inputRefs.current[i]) {
          inputRefs.current[i]!.value = digit
        }
      })
      
      // Focus on the next empty input or last input
      const nextIndex = Math.min(digits.length, 5)
      inputRefs.current[nextIndex]?.focus()
      
      // Update form value
      setValue('code', digits.join(''))
    } else {
      // Single digit input
      if (inputRefs.current[index]) {
        inputRefs.current[index]!.value = sanitizedValue
      }
      
      // Auto-advance to next input
      if (sanitizedValue && index < 5) {
        inputRefs.current[index + 1]?.focus()
      }
      
      // Update form value
      const fullCode = inputRefs.current
        .map(input => input?.value || '')
        .join('')
      setValue('code', fullCode)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>, index: number) => {
    if (e.key === 'Backspace' && !e.currentTarget.value && index > 0) {
      // Move to previous input on backspace
      inputRefs.current[index - 1]?.focus()
    } else if (e.key === 'ArrowLeft' && index > 0) {
      inputRefs.current[index - 1]?.focus()
    } else if (e.key === 'ArrowRight' && index < 5) {
      inputRefs.current[index + 1]?.focus()
    }
  }

  const onSubmit = async (data: MFAFormData) => {
    try {
      setIsLoading(true)
      setMfaError(null)
      clearError()

      logger.info('Verifying MFA code')

      await verifyMFA({
        challengeId: data.challengeId,
        code: data.code,
      })

      // MFA verification successful
      toast.success('Verification successful!')
      navigate(redirectTo, { replace: true })
      
    } catch (error: any) {
      logger.error('MFA verification failed', error)
      
      const errorCode = error.code as keyof typeof AUTH_ERROR_MESSAGES
      const errorMessage = AUTH_ERROR_MESSAGES[errorCode] || error.message || 'Verification failed'
      
      setMfaError(errorMessage)
      
      if (errorCode === 'INVALID_MFA_CODE') {
        setError('code', { message: 'Invalid code' })
        // Clear the code inputs
        inputRefs.current.forEach(input => {
          if (input) input.value = ''
        })
        setValue('code', '')
        inputRefs.current[0]?.focus()
      }
      
      toast.error(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }

  const handleResendCode = async () => {
    try {
      // TODO: Implement resend MFA code API call
      toast.success('New code sent!')
      setTimeLeft(300) // Reset timer
    } catch (error) {
      toast.error('Failed to send new code')
    }
  }

  const formatTime = (seconds: number) => {
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = seconds % 60
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
  }

  if (!mfaChallenge) {
    return null // Will redirect
  }

  return (
    <div className="min-h-screen flex flex-col justify-center py-12 sm:px-6 lg:px-8 bg-gray-50">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        {/* Logo */}
        <div className="flex justify-center">
          <div className="h-12 w-12 bg-green-600 rounded-lg flex items-center justify-center">
            <Shield className="h-7 w-7 text-white" />
          </div>
        </div>

        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Two-Factor Authentication
        </h2>
        
        <p className="mt-2 text-center text-sm text-gray-600">
          Enter the 6-digit code from your authenticator app
        </p>
        
        {mfaChallenge.maskedTarget && (
          <p className="mt-1 text-center text-sm text-gray-500">
            Code sent to: {mfaChallenge.maskedTarget}
          </p>
        )}
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          
          {/* Error Alert */}
          {(mfaError || authError) && (
            <Alert variant="error" className="mb-6">
              {mfaError || authError}
            </Alert>
          )}

          {/* MFA Form */}
          <form onSubmit={handleSubmit(onSubmit)}>
            {/* Code Input */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-3 text-center">
                Authentication Code
              </label>
              
              <div className="flex justify-center space-x-2">
                {[0, 1, 2, 3, 4, 5].map((index) => (
                  <input
                    key={index}
                    ref={(el) => { inputRefs.current[index] = el }}
                    type="text"
                    maxLength={1}
                    className="w-12 h-12 text-center border border-gray-300 rounded-lg text-lg font-semibold focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    onChange={(e) => handleCodeChange(e.target.value, index)}
                    onKeyDown={(e) => handleKeyDown(e, index)}
                    autoComplete="off"
                  />
                ))}
              </div>
              
              {codeValue.length === 6 && (
                <p className="text-center text-sm text-green-600 mt-2">
                  Code entered successfully
                </p>
              )}
            </div>

            {/* Timer */}
            {timeLeft > 0 && (
              <p className="text-center text-sm text-gray-500 mb-4">
                Code expires in {formatTime(timeLeft)}
              </p>
            )}

            {/* Submit Button */}
            <Button
              type="submit"
              className="w-full mb-4"
              size="lg"
              loading={isLoading}
              disabled={!isValid || codeValue.length !== 6 || isLoading}
            >
              Verify Code
            </Button>

            {/* Resend Code */}
            <div className="text-center">
              <button
                type="button"
                onClick={handleResendCode}
                disabled={timeLeft > 0}
                className="inline-flex items-center text-sm text-blue-600 hover:text-blue-500 disabled:text-gray-400 disabled:cursor-not-allowed"
              >
                <RefreshCw className="h-4 w-4 mr-1" />
                {timeLeft > 0 ? 'Resend code available in' : 'Resend code'}
              </button>
            </div>
          </form>

          {/* Back to Login */}
          <div className="mt-6 pt-6 border-t border-gray-200">
            <div className="text-center">
              <Link
                to="/auth/login"
                className="inline-flex items-center text-sm text-gray-600 hover:text-gray-500"
              >
                <ArrowLeft className="h-4 w-4 mr-1" />
                Back to sign in
              </Link>
            </div>
          </div>

          {/* Help */}
          <div className="mt-4 text-center">
            <p className="text-xs text-gray-500">
              Having trouble?{' '}
              <Link to="/support" className="text-blue-600 hover:text-blue-500">
                Contact Support
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default MFAPage