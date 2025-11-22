// Logger utility with configurable levels and structured logging

export type LogLevel = 'debug' | 'info' | 'warn' | 'error'

export interface LogEntry {
  timestamp: string
  level: LogLevel
  message: string
  data?: any
  correlationId?: string
  userId?: string
  organizationId?: string
  sessionId?: string
}

class Logger {
  private logLevel: LogLevel
  private isDevelopment: boolean
  private correlationId?: string

  constructor() {
    this.logLevel = (import.meta.env.VITE_LOG_LEVEL as LogLevel) || 'info'
    this.isDevelopment = import.meta.env.DEV
  }

  private shouldLog(level: LogLevel): boolean {
    const levels: Record<LogLevel, number> = {
      debug: 0,
      info: 1,
      warn: 2,
      error: 3,
    }
    
    return levels[level] >= levels[this.logLevel]
  }

  private formatMessage(level: LogLevel, message: string, data?: any): LogEntry {
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      message,
    }

    if (data) {
      entry.data = data
    }

    // Add context if available
    if (this.correlationId) {
      entry.correlationId = this.correlationId
    }

    // Add user context from auth store if available
    try {
      const authState = JSON.parse(localStorage.getItem('auth-store') || '{}')
      if (authState.state?.user?.id) {
        entry.userId = authState.state.user.id
      }
      if (authState.state?.currentOrganization?.id) {
        entry.organizationId = authState.state.currentOrganization.id
      }
    } catch {
      // Ignore errors when getting auth context
    }

    return entry
  }

  private log(level: LogLevel, message: string, data?: any): void {
    if (!this.shouldLog(level)) {
      return
    }

    const entry = this.formatMessage(level, message, data)
    
    if (this.isDevelopment) {
      // Pretty console logging for development
      const style = this.getConsoleStyle(level)
      const timestamp = new Date().toLocaleTimeString()
      
      console.group(`%c[${level.toUpperCase()}] ${timestamp} ${message}`, style)
      
      if (data) {
        console.log('Data:', data)
      }
      
      if (entry.correlationId) {
        console.log('Correlation ID:', entry.correlationId)
      }
      
      if (entry.userId) {
        console.log('User ID:', entry.userId)
      }
      
      if (entry.organizationId) {
        console.log('Organization ID:', entry.organizationId)
      }
      
      console.groupEnd()
    } else {
      // Structured JSON logging for production
      const logMethod = level === 'error' ? console.error : 
                       level === 'warn' ? console.warn : 
                       console.log
      
      logMethod(JSON.stringify(entry))
    }

    // Send to external logging service in production
    if (import.meta.env.PROD && level === 'error') {
      this.sendToExternalLogger(entry)
    }
  }

  private getConsoleStyle(level: LogLevel): string {
    const styles = {
      debug: 'color: #888; font-weight: normal',
      info: 'color: #007ACC; font-weight: bold',
      warn: 'color: #FFA500; font-weight: bold',
      error: 'color: #FF0000; font-weight: bold; background: #FFE6E6; padding: 2px 4px',
    }
    
    return styles[level]
  }

  private async sendToExternalLogger(entry: LogEntry): Promise<void> {
    try {
      // Send to Sentry, LogRocket, or other logging service
      if (window.Sentry) {
        window.Sentry.addBreadcrumb({
          message: entry.message,
          level: entry.level,
          data: entry.data,
          timestamp: Date.now() / 1000,
        })
        
        if (entry.level === 'error') {
          window.Sentry.captureException(new Error(entry.message), {
            extra: entry.data,
            tags: {
              correlationId: entry.correlationId,
              userId: entry.userId,
              organizationId: entry.organizationId,
            },
          })
        }
      }
      
      // Could also send to custom logging endpoint
      // await fetch('/api/logs', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify(entry),
      // })
    } catch (error) {
      console.error('Failed to send log to external service:', error)
    }
  }

  // Public methods
  debug(message: string, data?: any): void {
    this.log('debug', message, data)
  }

  info(message: string, data?: any): void {
    this.log('info', message, data)
  }

  warn(message: string, data?: any): void {
    this.log('warn', message, data)
  }

  error(message: string, data?: any): void {
    this.log('error', message, data)
  }

  // Set correlation ID for request tracing
  setCorrelationId(id: string): void {
    this.correlationId = id
  }

  clearCorrelationId(): void {
    delete this.correlationId
  }

  // Performance logging
  time(label: string): void {
    if (this.isDevelopment) {
      console.time(label)
    }
  }

  timeEnd(label: string): void {
    if (this.isDevelopment) {
      console.timeEnd(label)
    }
  }

  // Group logging
  group(label: string): void {
    if (this.isDevelopment) {
      console.group(label)
    }
  }

  groupEnd(): void {
    if (this.isDevelopment) {
      console.groupEnd()
    }
  }

  // Change log level at runtime
  setLogLevel(level: LogLevel): void {
    this.logLevel = level
    this.info(`Log level changed to: ${level}`)
  }

  getLogLevel(): LogLevel {
    return this.logLevel
  }
}

// Create and export singleton instance
export const logger = new Logger()

// Global error handler
window.addEventListener('error', (event) => {
  logger.error('Global error caught', {
    message: event.error?.message || event.message,
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    stack: event.error?.stack,
  })
})

// Unhandled promise rejection handler
window.addEventListener('unhandledrejection', (event) => {
  logger.error('Unhandled promise rejection', {
    reason: event.reason,
    promise: event.promise,
  })
})

// Export for global access
if (import.meta.env.DEV) {
  (window as any).logger = logger
}

export default logger

// Extend window type for Sentry
declare global {
  interface Window {
    Sentry?: {
      addBreadcrumb: (breadcrumb: any) => void
      captureException: (error: Error, options?: any) => void
    }
  }
}