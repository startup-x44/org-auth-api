/**
 * PKCE (Proof Key for Code Exchange) utilities for OAuth 2.1
 * Implements RFC 7636 with S256 method (SHA-256 + base64url)
 */

/**
 * Generate a cryptographically secure code verifier
 * @param length Length of the verifier (43-128 characters)
 * @returns Base64url encoded random string
 */
export async function generateCodeVerifier(length: number = 43): Promise<string> {
  if (length < 43 || length > 128) {
    throw new Error('Code verifier length must be between 43 and 128 characters')
  }

  // Check if Web Crypto API is available
  if (!window.crypto || !window.crypto.getRandomValues) {
    throw new Error('Web Crypto API is not available in this environment')
  }

  // Generate random bytes
  const array = new Uint8Array(length)
  window.crypto.getRandomValues(array)

  // Convert to base64url (RFC 4648 Section 5)
  return btoa(String.fromCharCode.apply(null, Array.from(array)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '')
    .substring(0, length)
}

/**
 * Generate code challenge from verifier using S256 method
 * @param verifier The code verifier string
 * @returns Base64url encoded SHA-256 hash
 */
export async function generateCodeChallenge(verifier: string): Promise<string> {
  if (!isValidCodeVerifier(verifier)) {
    throw new Error('Invalid code verifier')
  }

  // Check if Web Crypto API is available
  if (!window.crypto || !window.crypto.subtle) {
    throw new Error('Web Crypto API (subtle) is not available in this environment')
  }

  // Convert verifier to bytes
  const encoder = new TextEncoder()
  const data = encoder.encode(verifier)

  // Hash with SHA-256
  const hashBuffer = await window.crypto.subtle.digest('SHA-256', data)
  const hashArray = new Uint8Array(hashBuffer)

  // Convert to base64url
  return btoa(String.fromCharCode.apply(null, Array.from(hashArray)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '')
}

/**
 * Validate a code verifier string
 * @param verifier The code verifier to validate
 * @returns True if valid, false otherwise
 */
export function isValidCodeVerifier(verifier: string): boolean {
  if (!verifier || typeof verifier !== 'string') {
    return false
  }

  // Check length requirements (RFC 7636)
  if (verifier.length < 43 || verifier.length > 128) {
    return false
  }

  // Check character set: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
  const validPattern = /^[A-Za-z0-9\-._~]+$/
  return validPattern.test(verifier)
}

/**
 * Generate a complete PKCE pair (verifier + challenge)
 * @param verifierLength Length for the code verifier
 * @returns Object containing verifier and challenge
 */
export async function generatePKCEPair(verifierLength: number = 43): Promise<{
  codeVerifier: string
  codeChallenge: string
  codeChallengeMethod: string
}> {
  const codeVerifier = await generateCodeVerifier(verifierLength)
  const codeChallenge = await generateCodeChallenge(codeVerifier)

  return {
    codeVerifier,
    codeChallenge,
    codeChallengeMethod: 'S256'
  }
}