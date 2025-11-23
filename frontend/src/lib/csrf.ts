// CSRF token management utility

const API_BASE_URL = process.env.NODE_ENV === 'development' 
  ? 'http://localhost:8080' 
  : '';

/**
 * Fetches a new CSRF token from the server
 * The server provides CSRF tokens in response headers for GET requests
 */
export async function getCSRFToken(): Promise<string | null> {
  try {
    // Make a GET request to any endpoint to get a CSRF token
    // Using the health endpoint as it's lightweight
    const response = await fetch(`${API_BASE_URL}/health`, {
      method: 'GET',
      credentials: 'include', // Include cookies for session tracking
    });

    // Get the CSRF token from the response header
    const csrfToken = response.headers.get('X-CSRF-Token');
    
    if (csrfToken) {
      // Store token for future use (optional)
      sessionStorage.setItem('csrf_token', csrfToken);
      return csrfToken;
    }

    return null;
  } catch (error) {
    console.error('Failed to fetch CSRF token:', error);
    return null;
  }
}

/**
 * Gets the cached CSRF token or fetches a new one
 */
export async function getValidCSRFToken(): Promise<string | null> {
  // Try to get cached token first
  const cachedToken = sessionStorage.getItem('csrf_token');
  
  if (cachedToken) {
    return cachedToken;
  }

  // Fetch new token if no cached token
  return await getCSRFToken();
}

/**
 * Initializes CSRF token on app startup
 * Call this when the app loads to pre-fetch the CSRF token
 */
export async function initializeCSRFToken(): Promise<void> {
  try {
    await getCSRFToken();
    console.log('✅ CSRF token initialized successfully');
  } catch (error) {
    console.error('❌ Failed to initialize CSRF token:', error);
  }
}

/**
 * Makes an authenticated request with CSRF token
 */
export async function makeSecureRequest(url: string, options: RequestInit = {}): Promise<Response> {
  let csrfToken = await getValidCSRFToken();
  
  // If we still don't have a token, try a few more times with delay
  if (!csrfToken) {
    console.warn('⚠️ No CSRF token found, retrying...');
    
    for (let attempt = 1; attempt <= 3; attempt++) {
      await new Promise(resolve => setTimeout(resolve, 100 * attempt)); // 100ms, 200ms, 300ms delays
      csrfToken = await getCSRFToken();
      
      if (csrfToken) {
        console.log(`✅ CSRF token obtained on attempt ${attempt}`);
        break;
      }
      
      console.warn(`❌ CSRF token attempt ${attempt} failed`);
    }
  }
  
  if (!csrfToken) {
    throw new Error('Failed to obtain CSRF token after multiple attempts');
  }

  // Ensure headers exist
  const headers = new Headers(options.headers);
  
  // Add CSRF token to headers
  headers.set('X-CSRF-Token', csrfToken);
  
  // Add Content-Type if not present and we're sending JSON
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include', // Include cookies for session tracking
  });

  // Handle 401 unauthorized responses
  if (response.status === 401) {
    // Clear CSRF token on unauthorized
    clearCSRFToken();
    
    // Only redirect if we're not already on the login page
    if (typeof window !== 'undefined' && !window.location.pathname.includes('/auth/login')) {
      console.log('401 Unauthorized - Redirecting to login from CSRF request');
      window.location.href = '/auth/login?reason=unauthorized';
    }
  }

  // If the response includes a new CSRF token, update our cache
  const newCSRFToken = response.headers.get('X-CSRF-Token');
  if (newCSRFToken) {
    sessionStorage.setItem('csrf_token', newCSRFToken);
  }

  return response;
}

/**
 * Clear cached CSRF token (useful on logout)
 */
export function clearCSRFToken(): void {
  sessionStorage.removeItem('csrf_token');
}