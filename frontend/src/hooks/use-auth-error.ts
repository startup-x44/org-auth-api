import { useRouter } from 'next/navigation';
import { useCallback } from 'react';
import { AuthService } from '../lib/auth';
import { clearCSRFToken } from '../lib/csrf';
import { useToast } from './use-toast';

/**
 * Hook to handle unauthorized (401) responses
 * Provides utilities to check for 401 errors and redirect to login
 */
export function useAuthError() {
  const router = useRouter();
  const { toast } = useToast();

  const handleUnauthorized = useCallback((error?: Error, customMessage?: string) => {
    // Clear auth state - DISABLED for debugging
    console.log('🚨 handleUnauthorized called - NOT clearing tokens (debugging mode)');
    console.log('Error:', error);
    console.log('Custom message:', customMessage);
    
    // AuthService.clearTokens();
    // clearCSRFToken();
    
    // Show error message
    toast({
      title: "401 Unauthorized (Debug Mode)",
      description: customMessage || "Debug: 401 error detected, not redirecting.",
      variant: "destructive",
    });

    // Redirect to login - DISABLED for debugging
    // setTimeout(() => {
    //   router.push('/auth/login?reason=unauthorized');
    // }, 1000);
  }, [router, toast]);

  const checkResponseForAuth = useCallback(async (response: Response, customMessage?: string) => {
    if (response.status === 401) {
      handleUnauthorized(undefined, customMessage);
      return true; // Indicates 401 was handled
    }
    return false; // No 401, continue normal flow
  }, [handleUnauthorized]);

  const isAuthError = useCallback((error: any): boolean => {
    return error?.message?.toLowerCase().includes('unauthorized') ||
           error?.message?.toLowerCase().includes('token') ||
           error?.status === 401;
  }, []);

  return {
    handleUnauthorized,
    checkResponseForAuth,
    isAuthError,
  };
}