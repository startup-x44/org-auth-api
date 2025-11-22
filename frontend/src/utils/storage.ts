import Cookies from 'js-cookie'
import { AuthTokens, Organization } from '@/types/auth'

// Security configuration
const COOKIE_OPTIONS = {
  secure: import.meta.env.PROD, // Only secure cookies in production
  sameSite: 'strict' as const,
  expires: 7, // 7 days
}

const SENSITIVE_COOKIE_OPTIONS = {
  ...COOKIE_OPTIONS,
  httpOnly: false, // We need to access these from JS
  secure: import.meta.env.PROD,
}

// Storage keys
const STORAGE_KEYS = {
  ACCESS_TOKEN: 'auth_access_token',
  REFRESH_TOKEN: 'auth_refresh_token',
  TOKEN_EXPIRES_AT: 'auth_token_expires_at',
  CSRF_TOKEN: 'csrf_token',
  CURRENT_ORG: 'current_organization',
  USER_PREFERENCES: 'user_preferences',
} as const

/**
 * Secure storage utility for handling sensitive authentication data
 * Uses a combination of localStorage, sessionStorage, and secure cookies
 * based on the sensitivity and requirements of the data
 */
export const secureStorage = {
  // Token management
  setTokens: (tokens: AuthTokens): void => {
    try {
      // Store access token in memory/sessionStorage for this session
      sessionStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, tokens.accessToken)
      
      // Store refresh token in secure httpOnly cookie (if possible) or encrypted localStorage
      if (tokens.refreshToken) {
        // For now, use secure cookie - in production, consider server-side only storage
        Cookies.set(STORAGE_KEYS.REFRESH_TOKEN, tokens.refreshToken, {
          ...SENSITIVE_COOKIE_OPTIONS,
          expires: 30, // 30 days for refresh token
        })
      }
      
      // Store expiration time
      localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRES_AT, tokens.expiresAt.toString())
      
      console.debug('Tokens stored securely')
    } catch (error) {
      console.error('Failed to store tokens:', error)
      throw new Error('Failed to store authentication tokens')
    }
  },

  getTokens: (): AuthTokens | null => {
    try {
      const accessToken = sessionStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)
      const refreshToken = Cookies.get(STORAGE_KEYS.REFRESH_TOKEN)
      const expiresAtStr = localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRES_AT)
      
      if (!accessToken || !expiresAtStr) {
        return null
      }
      
      const expiresAt = parseInt(expiresAtStr, 10)
      
      return {
        accessToken,
        refreshToken: refreshToken || '',
        expiresAt,
        tokenType: 'Bearer',
      }
    } catch (error) {
      console.error('Failed to retrieve tokens:', error)
      return null
    }
  },

  clearTokens: (): void => {
    try {
      // Clear from all storage locations
      sessionStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN)
      localStorage.removeItem(STORAGE_KEYS.TOKEN_EXPIRES_AT)
      Cookies.remove(STORAGE_KEYS.REFRESH_TOKEN)
      Cookies.remove(STORAGE_KEYS.CSRF_TOKEN)
      
      console.debug('Tokens cleared from storage')
    } catch (error) {
      console.error('Failed to clear tokens:', error)
    }
  },

  // CSRF token management
  setCSRFToken: (token: string): void => {
    try {
      Cookies.set(STORAGE_KEYS.CSRF_TOKEN, token, {
        ...COOKIE_OPTIONS,
        expires: 1, // 1 day
      })
    } catch (error) {
      console.error('Failed to store CSRF token:', error)
    }
  },

  getCSRFToken: (): string | null => {
    try {
      return Cookies.get(STORAGE_KEYS.CSRF_TOKEN) || null
    } catch (error) {
      console.error('Failed to retrieve CSRF token:', error)
      return null
    }
  },

  // Organization context
  setCurrentOrganization: (organization: Organization): void => {
    try {
      localStorage.setItem(STORAGE_KEYS.CURRENT_ORG, JSON.stringify(organization))
    } catch (error) {
      console.error('Failed to store current organization:', error)
    }
  },

  getCurrentOrganization: (): Organization | null => {
    try {
      const orgStr = localStorage.getItem(STORAGE_KEYS.CURRENT_ORG)
      return orgStr ? JSON.parse(orgStr) : null
    } catch (error) {
      console.error('Failed to retrieve current organization:', error)
      return null
    }
  },

  clearCurrentOrganization: (): void => {
    try {
      localStorage.removeItem(STORAGE_KEYS.CURRENT_ORG)
    } catch (error) {
      console.error('Failed to clear current organization:', error)
    }
  },

  // User preferences (non-sensitive)
  setUserPreferences: (preferences: Record<string, any>): void => {
    try {
      localStorage.setItem(STORAGE_KEYS.USER_PREFERENCES, JSON.stringify(preferences))
    } catch (error) {
      console.error('Failed to store user preferences:', error)
    }
  },

  getUserPreferences: (): Record<string, any> => {
    try {
      const prefsStr = localStorage.getItem(STORAGE_KEYS.USER_PREFERENCES)
      return prefsStr ? JSON.parse(prefsStr) : {}
    } catch (error) {
      console.error('Failed to retrieve user preferences:', error)
      return {}
    }
  },

  // Storage health check
  isStorageAvailable: (): { localStorage: boolean; sessionStorage: boolean; cookies: boolean } => {
    const checkStorage = (storage: Storage): boolean => {
      try {
        const testKey = '__storage_test__'
        storage.setItem(testKey, 'test')
        storage.removeItem(testKey)
        return true
      } catch {
        return false
      }
    }

    const checkCookies = (): boolean => {
      try {
        const testKey = '__cookie_test__'
        Cookies.set(testKey, 'test')
        const retrieved = Cookies.get(testKey)
        Cookies.remove(testKey)
        return retrieved === 'test'
      } catch {
        return false
      }
    }

    return {
      localStorage: checkStorage(localStorage),
      sessionStorage: checkStorage(sessionStorage),
      cookies: checkCookies(),
    }
  },

  // Clear all application data (for complete logout/reset)
  clearAll: (): void => {
    try {
      // Clear tokens
      secureStorage.clearTokens()
      
      // Clear organization context
      secureStorage.clearCurrentOrganization()
      
      // Clear user preferences
      localStorage.removeItem(STORAGE_KEYS.USER_PREFERENCES)
      
      console.debug('All application data cleared')
    } catch (error) {
      console.error('Failed to clear all application data:', error)
    }
  },
}

// Storage encryption utilities (for future enhancement)
export const storageEncryption = {
  // Simple XOR encryption for localStorage (not cryptographically secure, but better than plain text)
  encrypt: (text: string, key: string): string => {
    let result = ''
    for (let i = 0; i < text.length; i++) {
      result += String.fromCharCode(text.charCodeAt(i) ^ key.charCodeAt(i % key.length))
    }
    return btoa(result)
  },

  decrypt: (encryptedText: string, key: string): string => {
    try {
      const text = atob(encryptedText)
      let result = ''
      for (let i = 0; i < text.length; i++) {
        result += String.fromCharCode(text.charCodeAt(i) ^ key.charCodeAt(i % key.length))
      }
      return result
    } catch {
      return ''
    }
  },
}

// Initialize storage health check on load
const storageHealth = secureStorage.isStorageAvailable()
if (!storageHealth.localStorage || !storageHealth.sessionStorage) {
  console.warn('Storage not fully available:', storageHealth)
}

export default secureStorage