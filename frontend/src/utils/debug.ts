import { useEffect, useRef } from 'react'

/**
 * Development debugging utilities
 * Only active in development mode
 */

// Hook to track component re-renders
export const useRenderTracker = (componentName: string) => {
  const renderCount = useRef(0)
  
  useEffect(() => {
    if (import.meta.env.DEV) {
      renderCount.current++
      console.log(`🔄 ${componentName} rendered ${renderCount.current} times`)
    }
  })
}

// Hook to track prop changes
export const usePropTracker = (componentName: string, props: Record<string, any>) => {
  const prevProps = useRef<Record<string, any> | undefined>(undefined)
  
  useEffect(() => {
    if (import.meta.env.DEV) {
      if (prevProps.current) {
        const changedProps: Record<string, { from: any; to: any }> = {}
        
        Object.keys(props).forEach(key => {
          if (prevProps.current![key] !== props[key]) {
            changedProps[key] = {
              from: prevProps.current![key],
              to: props[key]
            }
          }
        })
        
        if (Object.keys(changedProps).length > 0) {
          console.log(`📝 ${componentName} props changed:`, changedProps)
        }
      }
      
      prevProps.current = { ...props }
    }
  })
}

// Hook to track state changes
export const useStateTracker = (componentName: string, state: Record<string, any>) => {
  const prevState = useRef<Record<string, any> | undefined>(undefined)
  
  useEffect(() => {
    if (import.meta.env.DEV) {
      if (prevState.current) {
        const changedState: Record<string, { from: any; to: any }> = {}
        
        Object.keys(state).forEach(key => {
          if (prevState.current![key] !== state[key]) {
            changedState[key] = {
              from: prevState.current![key],
              to: state[key]
            }
          }
        })
        
        if (Object.keys(changedState).length > 0) {
          console.log(`🏪 ${componentName} state changed:`, changedState)
        }
      }
      
      prevState.current = { ...state }
    }
  })
}

// Performance monitoring
export const usePerformanceTracker = (componentName: string) => {
  const startTime = useRef<number | undefined>(undefined)
  
  useEffect(() => {
    if (import.meta.env.DEV) {
      startTime.current = performance.now()
      
      return () => {
        if (startTime.current) {
          const endTime = performance.now()
          const renderTime = endTime - startTime.current
          
          if (renderTime > 16) { // More than one frame (16.67ms at 60fps)
            console.warn(`⚠️ ${componentName} took ${renderTime.toFixed(2)}ms to render (> 16ms)`)
          } else {
            console.log(`⚡ ${componentName} rendered in ${renderTime.toFixed(2)}ms`)
          }
        }
      }
    }
    
    return undefined
  })
}

// Store state debugging
export const debugStoreState = (storeName: string, state: any) => {
  if (import.meta.env.DEV) {
    console.group(`🏬 ${storeName} State`)
    console.log('Current state:', state)
    console.groupEnd()
  }
}

// API call debugging
export const debugApiCall = (method: string, url: string, data?: any, response?: any, error?: any) => {
  if (import.meta.env.DEV) {
    console.group(`🌐 API ${method.toUpperCase()} ${url}`)
    
    if (data) {
      console.log('Request data:', data)
    }
    
    if (response) {
      console.log('Response:', response)
    }
    
    if (error) {
      console.error('Error:', error)
    }
    
    console.groupEnd()
  }
}

// Memory usage tracking
export const trackMemoryUsage = (componentName: string) => {
  if (import.meta.env.DEV && 'memory' in performance) {
    const memory = (performance as any).memory
    console.log(`🧠 ${componentName} Memory Usage:`, {
      used: `${(memory.usedJSHeapSize / 1024 / 1024).toFixed(2)} MB`,
      total: `${(memory.totalJSHeapSize / 1024 / 1024).toFixed(2)} MB`,
      limit: `${(memory.jsHeapSizeLimit / 1024 / 1024).toFixed(2)} MB`
    })
  }
}

// React development tools helpers
export const enableReactDevtools = () => {
  if (import.meta.env.DEV && typeof window !== 'undefined') {
    // Add React DevTools helper
    ;(window as any).__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
      ...((window as any).__REACT_DEVTOOLS_GLOBAL_HOOK__ || {}),
      checkDCE: () => {}
    }
  }
}

// Global error tracking for development
export const setupDevelopmentErrorTracking = () => {
  if (import.meta.env.DEV) {
    // Track unhandled errors
    window.addEventListener('error', (event) => {
      console.error('🚨 Unhandled Error:', {
        message: event.message,
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        error: event.error
      })
    })

    // Track unhandled promise rejections
    window.addEventListener('unhandledrejection', (event) => {
      console.error('🚨 Unhandled Promise Rejection:', {
        reason: event.reason,
        promise: event.promise
      })
    })

    // Track React errors (if available)
    const originalConsoleError = console.error
    console.error = (...args) => {
      if (args[0]?.includes?.('React')) {
        console.error('🚨 React Error:', ...args)
      }
      originalConsoleError.apply(console, args)
    }
  }
}

// Component lifecycle debugging
export const useLifecycleDebugger = (componentName: string) => {
  useEffect(() => {
    if (import.meta.env.DEV) {
      console.log(`🟢 ${componentName} mounted`)
      
      return () => {
        console.log(`🔴 ${componentName} unmounted`)
      }
    }
    
    return undefined
  }, [componentName])
}

// Export for global access in development
if (import.meta.env.DEV) {
  ;(window as any).debugUtils = {
    debugStoreState,
    debugApiCall,
    trackMemoryUsage,
    enableReactDevtools,
    setupDevelopmentErrorTracking
  }
}