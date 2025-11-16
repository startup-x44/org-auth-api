import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Mail, CheckCircle, XCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import useAuthStore from '@/store/auth'
import { authAPI } from '@/lib/api'

const AcceptInvitation = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { isAuthenticated, getMyOrganizations } = useAuthStore()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [invitationDetails, setInvitationDetails] = useState<any>(null)

  const token = searchParams.get('token')
  const email = searchParams.get('email')

  const acceptInvitation = async () => {
    if (!token) {
      setError('Invalid invitation token')
      return
    }

    try {
      setLoading(true)
      setError('')

      const response = await authAPI.acceptInvitation(token)

      if (response.success) {
        setSuccess(true)
        setInvitationDetails(response.data)
        
        // Refresh user's organizations list to include the newly joined org
        await getMyOrganizations()
        
        // Redirect to dashboard after a brief moment
        setTimeout(() => {
          navigate('/dashboard')
        }, 1500)
      }
    } catch (err: any) {
      console.error('Failed to accept invitation:', err)
      setError(err.response?.data?.message || 'Failed to accept invitation')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!token) {
      setError('Invalid invitation link')
      setLoading(false)
      return
    }

    // If not authenticated, redirect to register with token so they can create account
    if (!isAuthenticated) {
      const emailParam = email ? `&email=${encodeURIComponent(email)}` : ''
      navigate(`/register?invitation_token=${token}${emailParam}`)
      return
    }

    // If authenticated, accept the invitation
    acceptInvitation()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, isAuthenticated])

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-purple-50/20 flex items-center justify-center p-4">
      {/* Floating orbs background */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-0 -left-4 w-72 h-72 bg-purple-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob" />
        <div className="absolute top-0 -right-4 w-72 h-72 bg-blue-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-2000" />
        <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-4000" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="relative w-full max-w-md"
      >
        <Card className="border-0 shadow-2xl bg-white/80 backdrop-blur-xl">
          <CardHeader className="text-center space-y-4">
            {loading ? (
              <div className="mx-auto w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-500 rounded-full flex items-center justify-center">
                <Loader2 className="h-8 w-8 text-white animate-spin" />
              </div>
            ) : success ? (
              <div className="mx-auto w-16 h-16 bg-gradient-to-br from-green-500 to-emerald-500 rounded-full flex items-center justify-center">
                <CheckCircle className="h-8 w-8 text-white" />
              </div>
            ) : error ? (
              <div className="mx-auto w-16 h-16 bg-gradient-to-br from-red-500 to-pink-500 rounded-full flex items-center justify-center">
                <XCircle className="h-8 w-8 text-white" />
              </div>
            ) : (
              <div className="mx-auto w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-500 rounded-full flex items-center justify-center">
                <Mail className="h-8 w-8 text-white" />
              </div>
            )}
            
            <div>
              <CardTitle className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                {loading ? 'Accepting Invitation...' : success ? 'Invitation Accepted!' : error ? 'Invitation Failed' : 'Organization Invitation'}
              </CardTitle>
              <CardDescription className="mt-2">
                {loading ? 'Please wait while we process your invitation' : success ? 'You have successfully joined the organization' : error ? 'There was a problem with your invitation' : 'Processing your invitation'}
              </CardDescription>
            </div>
          </CardHeader>

          <CardContent className="space-y-4">
            {error && (
              <div className="bg-red-50 border border-red-200 rounded-xl p-4">
                <p className="text-sm text-red-800">{error}</p>
              </div>
            )}

            {success && invitationDetails && (
              <div className="bg-green-50 border border-green-200 rounded-xl p-4 space-y-2">
                <p className="text-sm text-green-800 font-medium">
                  Welcome to the organization!
                </p>
                <p className="text-sm text-green-700">
                  Redirecting you to your dashboard...
                </p>
              </div>
            )}

            {!loading && error && (
              <div className="space-y-3">
                <Button
                  onClick={() => navigate('/login')}
                  className="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
                >
                  Go to Login
                </Button>
                <Button
                  onClick={() => navigate('/register')}
                  variant="outline"
                  className="w-full"
                >
                  Create Account
                </Button>
              </div>
            )}

            {!token && (
              <div className="text-center">
                <p className="text-sm text-gray-600 mb-4">
                  Invalid invitation link. Please check your email and try again.
                </p>
                <Button
                  onClick={() => navigate('/login')}
                  variant="outline"
                  className="w-full"
                >
                  Go to Login
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}

export default AcceptInvitation
