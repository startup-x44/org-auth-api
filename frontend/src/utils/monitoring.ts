/**
 * Monitoring and observability setup with Sentry integration
 * Handles error reporting, performance monitoring, user session tracking
 */

import React from 'react'
import * as Sentry from '@sentry/react'
import { logger } from './logger'

/**
 * Initialize Sentry for production monitoring
 */
export function initSentry(): void {
  if (!import.meta.env['VITE_SENTRY_DSN']) {
    logger.warn('Sentry DSN not configured, skipping Sentry initialization')
    return
  }

  try {
    Sentry.init({
      dsn: import.meta.env['VITE_SENTRY_DSN'],
      environment: import.meta.env.MODE,
      release: import.meta.env['VITE_APP_VERSION'],
      
      // Performance monitoring
      integrations: [
        Sentry.browserTracingIntegration({
          // Set tracing origins to match your API
          tracePropagationTargets: [
            'localhost',
            import.meta.env['VITE_API_URL'],
            /^https:\/\/.*\.niloauth\.com\/.*$/
          ],
        }),
      ],
      
      // Performance monitoring sample rate
      // Adjust this value in production based on your traffic
      tracesSampleRate: import.meta.env.PROD ? 0.1 : 1.0,
      
      // Session replay for debugging (be mindful of privacy)
      replaysSessionSampleRate: 0.1,
      replaysOnErrorSampleRate: 1.0,
      
      // Configure what gets sent to Sentry
      beforeSend(event, hint) {
        // Filter out non-actionable errors
        if (event.exception) {
          const error = hint.originalException
          
          // Skip network errors that are likely user connectivity issues
          if (error instanceof Error) {
            if (error.message.includes('NetworkError') || 
                error.message.includes('Failed to fetch') ||
                error.message.includes('ERR_NETWORK')) {
              return null
            }
          }
        }
        
        // Filter out development environment noise
        if (!import.meta.env.PROD) {
          // You can add development-specific filtering here
        }
        
        return event
      },
      
      // Configure sensitive data scrubbing
      beforeBreadcrumb(breadcrumb) {
        // Don't log sensitive data in breadcrumbs
        if (breadcrumb.category === 'http' && breadcrumb.data && 'url' in breadcrumb.data) {
          // Remove sensitive query parameters
          const url = new URL(breadcrumb.data['url'] as string)
          url.searchParams.delete('token')
          url.searchParams.delete('password')
          url.searchParams.delete('api_key')
          breadcrumb.data['url'] = url.toString()
        }
        
        return breadcrumb
      },
      
      // Configure user context
      initialScope: {
        tags: {
          component: 'frontend',
          app: import.meta.env.VITE_APP_NAME || 'auth-service-frontend',
        },
      },
    })

    logger.info('Sentry initialized successfully')
  } catch (error) {
    logger.error('Failed to initialize Sentry:', error)
  }
}

/**
 * Set user context for Sentry
 */
export function setSentryUser(user: {
  id: string
  email?: string
  username?: string
  organizationId?: string
}): void {
  Sentry.setUser({
    id: user.id,
    email: user.email || undefined,
    username: user.username || undefined,
    organization_id: user.organizationId || undefined,
  })
}

/**
 * Clear user context from Sentry
 */
export function clearSentryUser(): void {
  Sentry.setUser(null)
}

/**
 * Set additional context for Sentry
 */
export function setSentryContext(key: string, context: Record<string, any>): void {
  Sentry.setContext(key, context)
}

/**
 * Add breadcrumb for debugging
 */
export function addBreadcrumb(message: string, category?: string, level?: 'info' | 'warning' | 'error'): void {
  Sentry.addBreadcrumb({
    message,
    category: category || 'custom',
    level: level || 'info',
    timestamp: Date.now() / 1000,
  })
}

/**
 * Capture exception with additional context
 */
export function captureException(error: Error, context?: Record<string, any>): string {
  if (context) {
    return Sentry.withScope((scope) => {
      scope.setExtra('context', context)
      return Sentry.captureException(error)
    })
  }
  return Sentry.captureException(error)
}

/**
 * Capture message with additional context
 */
export function captureMessage(message: string, level?: 'info' | 'warning' | 'error', context?: Record<string, any>): string {
  if (context) {
    return Sentry.withScope((scope) => {
      scope.setLevel(level || 'info')
      scope.setExtra('context', context)
      return Sentry.captureMessage(message)
    })
  }
  return Sentry.captureMessage(message, level || 'info')
}

/**
 * Start a new transaction for performance monitoring
 */
export function startTransaction(name: string, operation: string): Sentry.Transaction {
  return Sentry.startTransaction({
    name,
    op: operation,
  })
}

/**
 * Higher-order component for Sentry error boundary
 */
export const withSentryErrorBoundary = Sentry.withErrorBoundary

/**
 * React profiler for performance monitoring
 */
export const SentryProfiler = Sentry.Profiler

/**
 * Custom hook for measuring component performance
 */
export function useSentryTransaction(name: string, operation: string = 'navigation') {
  const [transaction, setTransaction] = React.useState<Sentry.Transaction | null>(null)

  React.useEffect(() => {
    const txn = startTransaction(name, operation)
    setTransaction(txn)

    return () => {
      if (txn) {
        txn.finish()
      }
    }
  }, [name, operation])

  const addSpan = React.useCallback((spanName: string, spanOperation: string) => {
    if (transaction) {
      return transaction.startChild({
        op: spanOperation,
        description: spanName,
      })
    }
    return null
  }, [transaction])

  return { transaction, addSpan }
}

// Performance measurement utilities
export class PerformanceTracker {
  private marks: Map<string, number> = new Map()
  
  start(name: string): void {
    this.marks.set(name, performance.now())
  }
  
  end(name: string): number | null {
    const startTime = this.marks.get(name)
    if (startTime === undefined) {
      logger.warn(`Performance mark "${name}" not found`)
      return null
    }
    
    const duration = performance.now() - startTime
    this.marks.delete(name)
    
    // Log performance data
    logger.info(`Performance: ${name} took ${duration.toFixed(2)}ms`)
    
    // Send to Sentry in production
    if (import.meta.env.PROD) {
      addBreadcrumb(`${name} completed in ${duration.toFixed(2)}ms`, 'performance', 'info')
    }
    
    return duration
  }
  
  measure(name: string, fn: () => void): number {
    this.start(name)
    fn()
    return this.end(name) || 0
  }
  
  async measureAsync<T>(name: string, fn: () => Promise<T>): Promise<T> {
    this.start(name)
    try {
      const result = await fn()
      this.end(name)
      return result
    } catch (error) {
      this.end(name)
      throw error
    }
  }
}

export const performanceTracker = new PerformanceTracker()