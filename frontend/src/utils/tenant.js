// Tenant resolution utilities for multi-tenant SaaS

/**
 * Resolves tenant ID from current hostname (subdomain) or email domain
 * @param {string} email - Optional email address to extract domain from
 * @returns {string} - Resolved tenant identifier
 */
export const resolveTenant = (email = null) => {
  // First, try to resolve from subdomain
  const hostname = window.location.hostname;

  // In development, handle localhost subdomains
  if (hostname === 'localhost' || hostname.startsWith('localhost:')) {
    // For localhost, we might not have subdomains, so return null to use email domain
    return null;
  }

  // Check if it's a subdomain (has more than one dot, or is a subdomain of known domain)
  const parts = hostname.split('.');

  // If we have a subdomain (more than 2 parts, or 2 parts that aren't a TLD)
  if (parts.length > 2 || (parts.length === 2 && !['com', 'org', 'net', 'edu', 'gov'].includes(parts[1]))) {
    // Return the subdomain part
    return parts[0];
  }

  // If no subdomain, try to resolve from email domain
  if (email) {
    const emailDomain = extractDomainFromEmail(email);
    if (emailDomain) {
      return emailDomain;
    }
  }

  // Fallback to default
  return 'default';
};

/**
 * Extracts domain from email address
 * @param {string} email - Email address
 * @returns {string|null} - Domain part or null if invalid
 */
export const extractDomainFromEmail = (email) => {
  if (!email || !email.includes('@')) {
    return null;
  }

  const parts = email.split('@');
  if (parts.length !== 2) {
    return null;
  }

  return parts[1].toLowerCase();
};

/**
 * Gets the full tenant domain for API calls
 * @param {string} tenantId - Tenant identifier
 * @returns {string} - Full domain for API calls
 */
export const getTenantDomain = (tenantId) => {
  if (!tenantId || tenantId === 'default') {
    return 'default.local'; // Default tenant domain
  }

  // In production, this would be tenantId + base domain
  // For now, return the tenantId as-is (assuming it's already a domain)
  return tenantId;
};

/**
 * Checks if current environment supports tenant subdomains
 * @returns {boolean}
 */
export const supportsTenantSubdomains = () => {
  const hostname = window.location.hostname;
  return hostname !== 'localhost' && !hostname.startsWith('localhost:');
};