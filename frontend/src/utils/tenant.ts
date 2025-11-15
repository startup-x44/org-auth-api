// Tenant resolution utilities for multi-tenant architecture
// Supports subdomain-based and email domain-based tenant resolution

const TENANT_DOMAINS: Record<string, string> = {
  'sprout.com': 'sprout-tenant-id',
  'acme.com': 'acme-tenant-id',
  'example.com': 'example-tenant-id',
  // Add more tenant domains as needed
}

const DEFAULT_TENANT_ID = 'default-tenant-id'

/**
 * Resolve tenant ID from email domain or subdomain
 * @param email - User email address
 * @returns Tenant ID string
 */
export const resolveTenant = (email?: string): string => {
  // First try to resolve from subdomain
  const hostname = window.location.hostname
  const subdomain = hostname.split('.')[0]

  // If subdomain is not 'www' or localhost, use it as tenant identifier
  if (subdomain && subdomain !== 'www' && subdomain !== 'localhost') {
    return subdomain
  }

  // Fallback to email domain resolution
  if (email) {
    const emailDomain = email.split('@')[1]?.toLowerCase()
    if (emailDomain && TENANT_DOMAINS[emailDomain]) {
      return TENANT_DOMAINS[emailDomain]
    }
  }

  // Return default tenant if no resolution possible
  return DEFAULT_TENANT_ID
}

/**
 * Get current tenant ID from localStorage or resolve from context
 * @returns Current tenant ID
 */
export const getCurrentTenantId = (): string => {
  const stored = localStorage.getItem('tenant_id')
  if (stored) return stored

  // Try to resolve from current context
  return resolveTenant()
}

/**
 * Set tenant ID in localStorage
 * @param tenantId - Tenant ID to store
 */
export const setCurrentTenantId = (tenantId: string): void => {
  localStorage.setItem('tenant_id', tenantId)
}

/**
 * Clear tenant ID from localStorage
 */
export const clearTenantId = (): void => {
  localStorage.removeItem('tenant_id')
}

/**
 * Check if current hostname is a tenant subdomain
 * @returns Boolean indicating if current host is a tenant subdomain
 */
export const isTenantSubdomain = (): boolean => {
  const hostname = window.location.hostname
  const subdomain = hostname.split('.')[0]
  return Boolean(subdomain && subdomain !== 'www' && subdomain !== 'localhost')
}