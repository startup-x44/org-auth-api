import { useEffect, useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import LoadingSpinner from '@/components/ui/loading-spinner'
import { exchangeCodeForTokens, parseOAuthError } from '@/lib/oauth'
import useAuthStore from '@/store/auth'

const OAuthCallback = () => {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { loginWithOAuthTokens } = useAuthStore()

  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [message, setMessage] = useState('')

  const code = searchParams.get('code')
  const state = searchParams.get('state')

  useEffect(() => {
    const handleCallback = async () => {
      try {
        // Check for OAuth error in URL parameters
        const oauthError = parseOAuthError(searchParams)
        if (oauthError.error) {
          setStatus('error')
          setMessage(oauthError.errorDescription || oauthError.error || 'Authorization failed')
          return
        }

        // Check for authorization code
        if (!code) {
          setStatus('error')
          setMessage('Invalid callback - missing authorization code')
          return
        }

        setStatus('loading')
        setMessage('Exchanging authorization code for tokens...')

        // Exchange code for tokens
        // Note: Using a default client_id here - in production this should come from config/env
        const clientId = import.meta.env.VITE_OAUTH_CLIENT_ID || 'default-client-id'
        const redirectUri = `${window.location.origin}/oauth/callback`

        const tokenResponse = await exchangeCodeForTokens({
          code,
          clientId,
          redirectUri
        })

        // Store tokens using the existing auth store
        if (loginWithOAuthTokens) {
          await loginWithOAuthTokens({
            accessToken: tokenResponse.access_token,
            refreshToken: tokenResponse.refresh_token
          })
        } else {
          // Fallback: store tokens directly in localStorage if method doesn't exist
          localStorage.setItem('access_token', tokenResponse.access_token)
          if (tokenResponse.refresh_token) {
            localStorage.setItem('refresh_token', tokenResponse.refresh_token)
          }
        }

        setStatus('success')
        setMessage('Authentication successful! Redirecting...')

        // Redirect after successful token exchange
        setTimeout(() => {
          // Check if state contains a return_to URL
          if (state) {
            try {
              const stateData = JSON.parse(decodeURIComponent(state))
              if (stateData.return_to) {
                navigate(stateData.return_to, { replace: true })
                return
              }
            } catch {
              // If state parsing fails, just continue with default redirect
            }
          }
          
          // Default redirect to dashboard
          navigate('/dashboard', { replace: true })
        }, 1500)

      } catch (error: any) {
        console.error('OAuth callback error:', error)
        setStatus('error')
        setMessage(error.message || 'Failed to complete authentication')
      }
    }

    handleCallback()
  }, [code, state, searchParams, navigate, loginWithOAuthTokens])

  const handleReturn = () => {
    navigate('/dashboard')
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-md"
      >
        <Card>
          <CardHeader className="text-center">
            <div className="flex justify-center mb-4">
              {status === 'loading' && (
                <LoadingSpinner size="lg" />
              )}
              {status === 'success' && (
                <div className="p-4 bg-green-100 dark:bg-green-900/20 rounded-full">
                  <CheckCircle className="h-12 w-12 text-green-600 dark:text-green-400" />
                </div>
              )}
              {status === 'error' && (
                <div className="p-4 bg-red-100 dark:bg-red-900/20 rounded-full">
                  <XCircle className="h-12 w-12 text-red-600 dark:text-red-400" />
                </div>
              )}
            </div>
            <CardTitle className="text-2xl">
              {status === 'loading' && 'Processing Authorization...'}
              {status === 'success' && 'Authorization Successful'}
              {status === 'error' && 'Authorization Failed'}
            </CardTitle>
            <CardDescription>
              {status === 'loading' && 'Please wait while we process your authorization'}
              {status === 'success' && 'You have successfully authorized the application'}
              {status === 'error' && 'There was a problem with the authorization'}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Message */}
            {message && status !== 'loading' && (
              <Alert variant={status === 'error' ? 'destructive' : 'default'}>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{message}</AlertDescription>
              </Alert>
            )}

            {/* Authorization Code (for development/debugging) */}
            {status === 'success' && code && import.meta.env.DEV && (
              <div className="bg-muted rounded-lg p-4">
                <div className="text-sm text-muted-foreground mb-2">Authorization Code (Development)</div>
                <code className="text-xs break-all">{code}</code>
              </div>
            )}

            {/* State Parameter (if present) */}
            {state && import.meta.env.DEV && (
              <div className="bg-muted rounded-lg p-4">
                <div className="text-sm text-muted-foreground mb-2">State (Development)</div>
                <code className="text-xs break-all">{state}</code>
              </div>
            )}

            {/* Instructions */}
            {status === 'success' && (
              <div className="text-center space-y-2">
                <p className="text-sm text-muted-foreground">
                  The application now has access to your account with the permissions you authorized.
                </p>
                <p className="text-sm text-muted-foreground">
                  You can safely close this window and return to the application.
                </p>
              </div>
            )}

            {/* Action Button */}
            {status !== 'loading' && (
              <Button
                onClick={handleReturn}
                className="w-full"
                variant={status === 'error' ? 'outline' : 'default'}
              >
                Return to Dashboard
              </Button>
            )}

            {/* Security Notice */}
            {status === 'success' && (
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription className="text-xs">
                  If you did not initiate this authorization, please revoke access immediately from your account settings.
                </AlertDescription>
              </Alert>
            )}
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}

export default OAuthCallback
