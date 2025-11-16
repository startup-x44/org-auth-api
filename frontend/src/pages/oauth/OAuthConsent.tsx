import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Shield, AlertCircle, CheckCircle, Lock, Globe } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import LoadingSpinner from '@/components/ui/loading-spinner'
import useAuthStore from '@/store/auth'

interface ClientAppInfo {
  name: string
  description: string
  redirect_uris: string[]
}

const OAuthConsent = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { user, isAuthenticated } = useAuthStore()

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [clientApp, setClientApp] = useState<ClientAppInfo | null>(null)
  const [staySignedIn, setStaySignedIn] = useState(false)

  // Extract OAuth parameters from URL
  const clientId = searchParams.get('client_id')
  const redirectUri = searchParams.get('redirect_uri')
  const responseType = searchParams.get('response_type')
  const scope = searchParams.get('scope')
  const state = searchParams.get('state')
  const codeChallenge = searchParams.get('code_challenge')
  const codeChallengeMethod = searchParams.get('code_challenge_method')

  useEffect(() => {
    // Redirect to login if not authenticated
    if (!isAuthenticated) {
      const returnUrl = window.location.pathname + window.location.search
      navigate(`/login?return_to=${encodeURIComponent(returnUrl)}`, { replace: true })
      return
    }

    // Validate required parameters
    if (!clientId || !redirectUri || !responseType || !codeChallenge || !codeChallengeMethod) {
      setError('Missing required OAuth parameters')
      setLoading(false)
      return
    }

    if (responseType !== 'code') {
      setError('Invalid response_type. Only "code" is supported.')
      setLoading(false)
      return
    }

    if (codeChallengeMethod !== 'S256') {
      setError('Invalid code_challenge_method. Only "S256" is supported.')
      setLoading(false)
      return
    }

    loadClientApp()
  }, [isAuthenticated, clientId, redirectUri, responseType, codeChallenge, codeChallengeMethod, navigate])

  const loadClientApp = async () => {
    if (!clientId) return

    try {
      setLoading(true)
      setError('')
      
      // In a real implementation, we'd have an endpoint to get public client info
      // For now, we'll create a mock response
      // You'll need to add a backend endpoint: GET /oauth/client-info?client_id=xxx
      
      setClientApp({
        name: 'Third-Party Application',
        description: 'A third-party application requesting access to your account',
        redirect_uris: [redirectUri || ''],
      })
      
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to load client app info'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  const handleAllow = () => {
    // Build the authorization URL and redirect to backend
    const params = new URLSearchParams()
    params.set('client_id', clientId || '')
    params.set('redirect_uri', redirectUri || '')
    params.set('response_type', responseType || '')
    params.set('code_challenge', codeChallenge || '')
    params.set('code_challenge_method', codeChallengeMethod || '')
    if (scope) params.set('scope', scope)
    if (state) params.set('state', state)
    params.set('consent', 'allow')
    if (staySignedIn) params.set('stay_signed_in', 'true')

    // Redirect to backend authorization endpoint - this will issue the authorization code
    const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    window.location.href = `${baseURL}/api/v1/oauth/authorize?${params.toString()}`
  }

  const handleDeny = () => {
    if (redirectUri) {
      // Redirect back with error
      const url = new URL(redirectUri)
      url.searchParams.set('error', 'access_denied')
      url.searchParams.set('error_description', 'User denied the authorization request')
      if (state) url.searchParams.set('state', state)
      window.location.href = url.toString()
    } else {
      navigate('/dashboard')
    }
  }

  const parsedScopes = scope ? scope.split(' ').filter(s => s.length > 0) : ['profile']

  const scopeDescriptions: Record<string, { icon: any; description: string }> = {
    profile: {
      icon: Shield,
      description: 'Access your basic profile information (name, email)',
    },
    email: {
      icon: Shield,
      description: 'Access your email address',
    },
    'org:read': {
      icon: Globe,
      description: 'View your organization memberships',
    },
    'org:write': {
      icon: Globe,
      description: 'Manage your organization memberships',
    },
    openid: {
      icon: Lock,
      description: 'Authenticate your identity',
    },
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <LoadingSpinner size="lg" />
      </div>
    )
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
              <div className="p-4 bg-primary/10 rounded-full">
                <Shield className="h-12 w-12 text-primary" />
              </div>
            </div>
            <CardTitle className="text-2xl">Authorization Request</CardTitle>
            <CardDescription>
              {clientApp?.name || 'An application'} wants to access your account
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Error Message */}
            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {!error && (
              <>
                {/* User Info */}
                <div className="bg-muted rounded-lg p-4">
                  <div className="text-sm text-muted-foreground mb-1">Signed in as</div>
                  <div className="font-medium">{user?.email}</div>
                  {user?.first_name && user?.last_name && (
                    <div className="text-sm text-muted-foreground">
                      {user.first_name} {user.last_name}
                    </div>
                  )}
                </div>

                {/* Client App Info */}
                {clientApp && (
                  <div>
                    <h3 className="font-semibold mb-2">Application Details</h3>
                    <div className="space-y-2">
                      <div>
                        <div className="text-sm font-medium">{clientApp.name}</div>
                        <div className="text-sm text-muted-foreground">{clientApp.description}</div>
                      </div>
                      {redirectUri && (
                        <div className="text-xs text-muted-foreground flex items-center space-x-1">
                          <Globe className="h-3 w-3" />
                          <span>Redirect to: {new URL(redirectUri).origin}</span>
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {/* Requested Permissions */}
                <div>
                  <h3 className="font-semibold mb-3">This application will be able to:</h3>
                  <div className="space-y-3">
                    {parsedScopes.map((scopeItem) => {
                      const scopeInfo = scopeDescriptions[scopeItem] || {
                        icon: Lock,
                        description: `Access ${scopeItem}`,
                      }
                      const Icon = scopeInfo.icon

                      return (
                        <div key={scopeItem} className="flex items-start space-x-3">
                          <div className="p-1 bg-primary/10 rounded">
                            <Icon className="h-4 w-4 text-primary" />
                          </div>
                          <div className="flex-1">
                            <div className="text-sm font-medium capitalize">{scopeItem.replace(':', ' ')}</div>
                            <div className="text-xs text-muted-foreground">
                              {scopeInfo.description}
                            </div>
                          </div>
                          <CheckCircle className="h-4 w-4 text-green-500" />
                        </div>
                      )
                    })}
                  </div>
                </div>

                {/* Stay Signed In Option */}
                <div className="flex items-center space-x-2 py-2">
                  <Checkbox
                    id="stay-signed-in"
                    checked={staySignedIn}
                    onCheckedChange={(checked: boolean) => setStaySignedIn(checked)}
                  />
                  <Label
                    htmlFor="stay-signed-in"
                    className="text-sm font-normal cursor-pointer"
                  >
                    Stay signed in for 30 days
                  </Label>
                </div>

                {/* Action Buttons */}
                <div className="space-y-3 pt-4">
                  <Button
                    onClick={handleAllow}
                    className="w-full"
                    size="lg"
                  >
                    Allow Access
                  </Button>
                  <Button
                    onClick={handleDeny}
                    variant="outline"
                    className="w-full"
                    size="lg"
                  >
                    Deny
                  </Button>
                </div>

                {/* Security Notice */}
                <Alert>
                  <Lock className="h-4 w-4" />
                  <AlertDescription className="text-xs">
                    By clicking "Allow Access", you authorize this application to access your account according to the permissions listed above.
                    You can revoke this access at any time from your account settings.
                  </AlertDescription>
                </Alert>
              </>
            )}
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}

export default OAuthConsent
