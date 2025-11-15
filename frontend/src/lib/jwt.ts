/**
 * JWT Utilities
 * Helper functions for decoding and validating JWT tokens
 */

export interface JWTClaims {
  user_id: string
  email: string
  organization_id?: string
  session_id?: string
  global_role?: string
  organization_role?: string
  permissions?: string[]
  is_superadmin?: boolean
  current_org_id?: string
  exp: number
  iat: number
}

/**
 * Decode JWT token without verification
 * Note: This only decodes the payload - signature verification happens on the backend
 */
export function decodeJWT(token: string): JWTClaims | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      console.error('Invalid JWT format')
      return null
    }

    // Decode the payload (second part)
    const payload = parts[1]
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
    const claims = JSON.parse(decoded) as JWTClaims

    return claims
  } catch (error) {
    console.error('Failed to decode JWT:', error)
    return null
  }
}

/**
 * Check if JWT token is expired
 */
export function isTokenExpired(token: string): boolean {
  const claims = decodeJWT(token)
  if (!claims) return true

  const now = Math.floor(Date.now() / 1000)
  return claims.exp < now
}

/**
 * Get remaining time until token expiration (in seconds)
 */
export function getTokenExpiresIn(token: string): number {
  const claims = decodeJWT(token)
  if (!claims) return 0

  const now = Math.floor(Date.now() / 1000)
  return Math.max(0, claims.exp - now)
}

/**
 * Extract permissions from JWT token
 */
export function extractPermissions(token: string): string[] {
  const claims = decodeJWT(token)
  return claims?.permissions || []
}

/**
 * Extract organization ID from JWT token
 */
export function extractOrganizationId(token: string): string | null {
  const claims = decodeJWT(token)
  return claims?.organization_id || claims?.current_org_id || null
}

/**
 * Extract role information from JWT token
 */
export function extractRoleInfo(token: string): { role: string | null; isSuperadmin: boolean } {
  const claims = decodeJWT(token)
  return {
    role: claims?.organization_role || claims?.global_role || null,
    isSuperadmin: claims?.is_superadmin || false,
  }
}
