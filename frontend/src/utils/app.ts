/**
 * App initialization utilities for production-ready setup
 * Handles CSP, security headers, performance monitoring setup
 */

import { logger } from './logger'

// Track if already initialized to prevent double initialization
let isAppInitialized = false

/**
 * Initialize application with production-ready setup
 */
export async function initializeApp(): Promise<void> {
  if (isAppInitialized) {
    logger.debug('Application already initialized, skipping')
    return
  }

  try {
    logger.info('Initializing application...')
    
    // Set up CSP headers if running in production
    if (import.meta.env.PROD) {
      setupSecurityHeaders()
    }
    
    // Initialize performance monitoring
    setupPerformanceMonitoring()
    
    // Initialize error handling
    setupGlobalErrorHandling()
    
    // Show React DevTools recommendation in development
    if (import.meta.env.DEV && !(window as any).__REACT_DEVTOOLS_GLOBAL_HOOK__) {
      console.info(
        '%c💡 React DevTools not detected',
        'color: #61dafb; font-weight: bold;',
        '\n📥 Install React DevTools for better debugging: https://reactjs.org/link/react-devtools'
      )
    }
    
    isAppInitialized = true
    logger.info('Application initialized successfully')
  } catch (error) {
    logger.error('Failed to initialize application:', error)
    throw error
  }
}

/**
 * Setup Content Security Policy and other security headers
 */
function setupSecurityHeaders(): void {
  // Add security meta tags if they don't exist
  if (!document.querySelector('meta[http-equiv="Content-Security-Policy"]')) {
    const cspMeta = document.createElement('meta')
    cspMeta.httpEquiv = 'Content-Security-Policy'
    cspMeta.content = [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net",
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
      "font-src 'self' https://fonts.gstatic.com",
      "img-src 'self' data: https:",
      "connect-src 'self' " + import.meta.env.VITE_API_URL,
      "frame-ancestors 'none'",
      "base-uri 'self'"
    ].join('; ')
    document.head.appendChild(cspMeta)
  }

  // Add X-Frame-Options
  if (!document.querySelector('meta[name="x-frame-options"]')) {
    const frameMeta = document.createElement('meta')
    frameMeta.name = 'x-frame-options'
    frameMeta.content = 'DENY'
    document.head.appendChild(frameMeta)
  }

  // Add X-Content-Type-Options
  if (!document.querySelector('meta[name="x-content-type-options"]')) {
    const contentTypeMeta = document.createElement('meta')
    contentTypeMeta.name = 'x-content-type-options'
    contentTypeMeta.content = 'nosniff'
    document.head.appendChild(contentTypeMeta)
  }
}

/**
 * Setup performance monitoring
 */
function setupPerformanceMonitoring(): void {
  // Monitor Core Web Vitals
  if ('PerformanceObserver' in window) {
    // Largest Contentful Paint
    const lcpObserver = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        logger.info('LCP:', entry.startTime)
        // Send to analytics in production
        if (import.meta.env.PROD && window.gtag) {
          window.gtag('event', 'web_vitals', {
            name: 'LCP',
            value: Math.round(entry.startTime),
            event_category: 'performance'
          })
        }
      }
    })
    lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] })

    // First Input Delay
    const fidObserver = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        const fid = (entry as any).processingStart - entry.startTime
        logger.info('FID:', fid)
        if (import.meta.env.PROD && window.gtag) {
          window.gtag('event', 'web_vitals', {
            name: 'FID',
            value: Math.round(fid),
            event_category: 'performance'
          })
        }
      }
    })
    fidObserver.observe({ entryTypes: ['first-input'] })

    // Cumulative Layout Shift
    let clsValue = 0
    let clsEntries: PerformanceEntry[] = []
    const clsObserver = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        if (!(entry as any).hadRecentInput) {
          clsEntries.push(entry)
          clsValue += (entry as any).value
        }
      }
    })
    clsObserver.observe({ entryTypes: ['layout-shift'] })

    // Report CLS on page unload
    window.addEventListener('beforeunload', () => {
      logger.info('CLS:', clsValue)
      if (import.meta.env.PROD && window.gtag) {
        window.gtag('event', 'web_vitals', {
          name: 'CLS',
          value: Math.round(clsValue * 1000),
          event_category: 'performance'
        })
      }
    })
  }

  // Monitor navigation timing
  window.addEventListener('load', () => {
    setTimeout(() => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
      if (navigation) {
        const metrics = {
          dns: navigation.domainLookupEnd - navigation.domainLookupStart,
          tcp: navigation.connectEnd - navigation.connectStart,
          request: navigation.responseStart - navigation.requestStart,
          response: navigation.responseEnd - navigation.responseStart,
          dom: navigation.domContentLoadedEventEnd - navigation.responseEnd,
          load: navigation.loadEventEnd - navigation.loadEventStart,
          total: navigation.loadEventEnd - navigation.fetchStart
        }
        
        logger.info('Navigation metrics:', metrics)
      }
    }, 0)
  })
}

/**
 * Setup global error handling
 */
function setupGlobalErrorHandling(): void {
  // Handle unhandled promise rejections
  window.addEventListener('unhandledrejection', (event) => {
    logger.error('Unhandled promise rejection:', event.reason)
    
    // Prevent the default browser behavior
    event.preventDefault()
    
    // Send to monitoring service in production
    if (import.meta.env.PROD) {
      // Sentry will auto-capture this
    }
  })

  // Handle global errors
  window.addEventListener('error', (event) => {
    logger.error('Global error:', {
      message: event.message,
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      error: event.error
    })
    
    // Send to monitoring service in production
    if (import.meta.env.PROD) {
      // Sentry will auto-capture this
    }
  })

  // Handle resource loading errors
  window.addEventListener('error', (event) => {
    if (event.target !== window) {
      logger.error('Resource loading error:', {
        element: event.target,
        source: (event.target as HTMLElement)?.getAttribute?.('src') || (event.target as HTMLElement)?.getAttribute?.('href')
      })
    }
  }, true)
}

/**
 * Check if the app is running in development mode
 */
export function isDevelopment(): boolean {
  return import.meta.env.DEV
}

/**
 * Check if the app is running in production mode
 */
export function isProduction(): boolean {
  return import.meta.env.PROD
}

/**
 * Get environment variable with fallback
 */
export function getEnvVar(key: string, fallback?: string): string | undefined {
  return import.meta.env[key] || fallback
}

/**
 * Validate required environment variables
 */
export function validateEnvironment(): void {
  const required = [
    'VITE_API_URL',
    'VITE_APP_NAME',
    'VITE_APP_VERSION'
  ]

  const missing = required.filter(key => !import.meta.env[key])
  
  if (missing.length > 0) {
    throw new Error(`Missing required environment variables: ${missing.join(', ')}`)
  }
}

// Declare global types for analytics
declare global {
  interface Window {
    gtag?: (...args: any[]) => void
  }
}