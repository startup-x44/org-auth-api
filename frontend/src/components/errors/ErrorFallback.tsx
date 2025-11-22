/**
 * Error Fallback Component - Used by React Error Boundary
 * Provides user-friendly error display with recovery options
 */

import React from 'react'
import { Button } from '../ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card'
import { Alert, AlertDescription } from '../ui/alert'
import { 
  AlertTriangle, 
  RefreshCw, 
  Bug, 
  Home,
  Copy,
  ExternalLink 
} from 'lucide-react'
import { toast } from 'react-hot-toast'
import { captureException, addBreadcrumb } from '../../utils/monitoring'

interface ErrorInfo {
  componentStack: string
  errorBoundary?: string
  errorBoundaryStack?: string
}

interface ErrorFallbackProps {
  error: Error
  errorInfo?: ErrorInfo
  resetErrorBoundary?: () => void
}

export function ErrorFallback({ error, errorInfo, resetErrorBoundary }: ErrorFallbackProps) {
  const [showDetails, setShowDetails] = React.useState(false)
  const [reportSent, setReportSent] = React.useState(false)

  React.useEffect(() => {
    // Log error details
    console.error('Error boundary caught an error:', error, errorInfo)
    
    // Add breadcrumb for debugging
    addBreadcrumb('Error boundary triggered', 'error', 'error')
    
    // Capture exception in Sentry
    if (import.meta.env.PROD) {
      captureException(error, {
        extra: {
          errorInfo,
          timestamp: new Date().toISOString(),
          userAgent: navigator.userAgent,
          url: window.location.href,
        }
      })
    }
  }, [error, errorInfo])

  const handleCopyError = async () => {
    const errorDetails = {
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo?.componentStack,
      timestamp: new Date().toISOString(),
      url: window.location.href,
      userAgent: navigator.userAgent
    }

    try {
      await navigator.clipboard.writeText(JSON.stringify(errorDetails, null, 2))
      toast.success('Error details copied to clipboard')
    } catch (err) {
      console.error('Failed to copy error details:', err)
      toast.error('Failed to copy error details')
    }
  }

  const handleSendReport = async () => {
    try {
      setReportSent(true)
      
      // In a real app, you would send this to your error reporting service
      // For now, we'll just simulate it
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      toast.success('Error report sent successfully')
    } catch (err) {
      toast.error('Failed to send error report')
      setReportSent(false)
    }
  }

  const handleGoHome = () => {
    window.location.href = '/dashboard'
  }

  const handleReload = () => {
    window.location.reload()
  }

  const isChunkLoadError = error.message.includes('ChunkLoadError') || 
                          error.message.includes('Loading chunk')

  const isNetworkError = error.message.includes('NetworkError') ||
                         error.message.includes('Failed to fetch')

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-2xl w-full">
        <Card className="border-red-200">
          <CardHeader className="text-center">
            <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-4">
              <AlertTriangle className="h-8 w-8 text-red-600" />
            </div>
            <CardTitle className="text-2xl font-bold text-gray-900">
              {isChunkLoadError ? 'Application Update Required' : 
               isNetworkError ? 'Connection Problem' : 
               'Something went wrong'}
            </CardTitle>
            <CardDescription className="text-gray-600">
              {isChunkLoadError ? 
                'The application has been updated. Please refresh your browser to get the latest version.' :
               isNetworkError ?
                'Unable to connect to our servers. Please check your internet connection and try again.' :
                'An unexpected error occurred. Our team has been notified and is working on a fix.'}
            </CardDescription>
          </CardHeader>
          
          <CardContent className="space-y-6">
            {/* Error message */}
            <Alert variant="destructive">
              <Bug className="h-4 w-4" />
              <AlertDescription className="font-mono text-sm">
                {error.message}
              </AlertDescription>
            </Alert>

            {/* Action buttons */}
            <div className="flex flex-col sm:flex-row gap-3">
              {isChunkLoadError ? (
                <Button onClick={handleReload} className="flex-1">
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Refresh Application
                </Button>
              ) : (
                <>
                  {resetErrorBoundary && (
                    <Button onClick={resetErrorBoundary} className="flex-1">
                      <RefreshCw className="h-4 w-4 mr-2" />
                      Try Again
                    </Button>
                  )}
                  <Button onClick={handleGoHome} variant="outline" className="flex-1">
                    <Home className="h-4 w-4 mr-2" />
                    Go to Dashboard
                  </Button>
                </>
              )}
            </div>

            {/* Advanced options */}
            {!isChunkLoadError && (
              <div className="border-t pt-6 space-y-4">
                <div className="flex justify-between items-center">
                  <h3 className="text-sm font-medium text-gray-900">Advanced Options</h3>
                  <button
                    onClick={() => setShowDetails(!showDetails)}
                    className="text-sm text-blue-600 hover:text-blue-700"
                  >
                    {showDetails ? 'Hide' : 'Show'} Details
                  </button>
                </div>

                {showDetails && (
                  <div className="space-y-3">
                    <Alert>
                      <AlertDescription>
                        <div className="space-y-2">
                          <p className="font-semibold">Error Details:</p>
                          <pre className="text-xs bg-gray-100 p-2 rounded overflow-auto max-h-32">
                            {error.stack}
                          </pre>
                          {errorInfo?.componentStack && (
                            <>
                              <p className="font-semibold mt-2">Component Stack:</p>
                              <pre className="text-xs bg-gray-100 p-2 rounded overflow-auto max-h-32">
                                {errorInfo.componentStack}
                              </pre>
                            </>
                          )}
                        </div>
                      </AlertDescription>
                    </Alert>

                    <div className="flex gap-2">
                      <Button
                        onClick={handleCopyError}
                        variant="outline"
                        size="sm"
                        className="flex-1"
                      >
                        <Copy className="h-4 w-4 mr-2" />
                        Copy Error Details
                      </Button>

                      <Button
                        onClick={handleSendReport}
                        variant="outline"
                        size="sm"
                        className="flex-1"
                        disabled={reportSent}
                      >
                        <ExternalLink className="h-4 w-4 mr-2" />
                        {reportSent ? 'Report Sent' : 'Send Report'}
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Contact support */}
            <div className="text-center text-sm text-gray-500">
              <p>
                If this problem persists, please{' '}
                <a 
                  href="mailto:support@niloauth.com" 
                  className="text-blue-600 hover:text-blue-700 underline"
                >
                  contact support
                </a>{' '}
                with the error details above.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}