/**
 * OAuth 2.1 flow utilities for frontend
 * Handles authorization URL building, token exchange, and state management
 */

import axios from 'axios'
import { generatePKCEPair } from './pkce'

// SessionStorage keys for PKCE state
const OAUTH_CODE_VERIFIER_KEY = 'oauth_code_verifier'
const OAUTH_STATE_KEY = 'oauth_state'

export interface AuthorizeUrlParams {
  authBaseUrl?: string
  clientId: string
  redirectUri: string
  scope: string
  state?: string
}

export interface StartAuthFlowOptions {
  clientId: string
  redirectUri: string
  scope: string
  state?: string
}

export interface TokenExchangeParams {
  code: string
  clientId: string
  redirectUri: string
  tokenEndpoint?: string
}

export interface TokenResponse {
  access_token: string
  refresh_token?: string
  expires_in?: number
  token_type?: string
  scope?: string
}

/**
 * Build OAuth authorization URL with PKCE parameters
 * Generates and stores code_verifier in sessionStorage
 */
export async function buildAuthorizeUrl(params: AuthorizeUrlParams): Promise<string> {
  const {
    authBaseUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/v1/oauth/authorize`,
    clientId,
    redirectUri,
    scope,
    state
  } = params

  // Generate PKCE pair
  const { codeVerifier, codeChallenge, codeChallengeMethod } = await generatePKCEPair()

  // Store code_verifier in sessionStorage for later token exchange
  sessionStorage.setItem(OAUTH_CODE_VERIFIER_KEY, codeVerifier)
  
  // Optionally store state
  if (state) {
    sessionStorage.setItem(OAUTH_STATE_KEY, state)
  }

  // Build authorization URL
  const authUrl = new URL(authBaseUrl)
  authUrl.searchParams.set('response_type', 'code')
  authUrl.searchParams.set('client_id', clientId)
  authUrl.searchParams.set('redirect_uri', redirectUri)
  authUrl.searchParams.set('scope', scope)
  authUrl.searchParams.set('code_challenge', codeChallenge)
  authUrl.searchParams.set('code_challenge_method', codeChallengeMethod)

  if (state) {
    authUrl.searchParams.set('state', state)
  }

  return authUrl.toString()
}

/**
 * Start OAuth authorization flow by redirecting to authorization server
 */
export async function startAuthFlow(options: StartAuthFlowOptions): Promise<void> {
  const authorizeUrl = await buildAuthorizeUrl(options)
  window.location.href = authorizeUrl
}

/**
 * Exchange authorization code for tokens
 * Uses stored code_verifier from sessionStorage
 */
export async function exchangeCodeForTokens(params: TokenExchangeParams): Promise<TokenResponse> {
  const {
    code,
    clientId,
    redirectUri,
    tokenEndpoint = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/v1/oauth/token`
  } = params

  // Retrieve code_verifier from sessionStorage
  const codeVerifier = sessionStorage.getItem(OAUTH_CODE_VERIFIER_KEY)
  if (!codeVerifier) {
    throw new Error('No code verifier found in session storage. Authorization flow may not have been initiated properly.')
  }

  try {
    // Prepare token exchange request
    const tokenRequest = {
      grant_type: 'authorization_code',
      code,
      client_id: clientId,
      redirect_uri: redirectUri,
      code_verifier: codeVerifier
    }

    // Exchange code for tokens using form-encoded data (as per OAuth spec)
    const response = await axios.post(tokenEndpoint, tokenRequest, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      // Convert object to URLSearchParams for form encoding
      transformRequest: [(data) => {
        const params = new URLSearchParams()
        for (const [key, value] of Object.entries(data)) {
          if (value !== undefined && value !== null) {
            params.append(key, String(value))
          }
        }
        return params.toString()
      }],
      withCredentials: true
    })

    // Clear stored PKCE state on successful exchange
    clearOAuthState()

    return response.data as TokenResponse
  } catch (error: any) {
    // Clear PKCE state on error to prevent reuse
    clearOAuthState()

    // Extract error message from response if available
    const errorMessage = error.response?.data?.error_description || 
                        error.response?.data?.message || 
                        error.message || 
                        'Token exchange failed'
    
    throw new Error(`OAuth token exchange failed: ${errorMessage}`)
  }
}

/**
 * Clear OAuth state from sessionStorage
 * Should be called after successful token exchange or on errors
 */
export function clearOAuthState(): void {
  sessionStorage.removeItem(OAUTH_CODE_VERIFIER_KEY)
  sessionStorage.removeItem(OAUTH_STATE_KEY)
}

/**
 * Get the stored OAuth state (if any)
 */
export function getStoredOAuthState(): {
  codeVerifier: string | null
  state: string | null
} {
  return {
    codeVerifier: sessionStorage.getItem(OAUTH_CODE_VERIFIER_KEY),
    state: sessionStorage.getItem(OAUTH_STATE_KEY)
  }
}

/**
 * Validate that OAuth flow can be initiated (Web Crypto API available)
 */
export function canInitiateOAuthFlow(): boolean {
  return !!(window.crypto && window.crypto.subtle && window.crypto.getRandomValues)
}

/**
 * Parse OAuth error from URL parameters (for error handling in callback)
 */
export function parseOAuthError(searchParams: URLSearchParams): {
  error: string | null
  errorDescription: string | null
  state: string | null
} {
  return {
    error: searchParams.get('error'),
    errorDescription: searchParams.get('error_description'),
    state: searchParams.get('state')
  }
}