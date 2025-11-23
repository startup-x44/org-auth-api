'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { AuthService } from '../lib/auth';

interface AuthGuardProps {
  children: React.ReactNode;
  requireSuperAdmin?: boolean;
}

export function AuthGuard({ children, requireSuperAdmin = false }: AuthGuardProps) {
  const router = useRouter();

  useEffect(() => {
    const token = AuthService.getAccessToken();
    
    if (!token || AuthService.isTokenExpired(token)) {
      router.replace('/(auth)/login');
      return;
    }

    const isSuperAdmin = AuthService.isSuperAdmin(token);

    if (requireSuperAdmin && !isSuperAdmin) {
      router.replace('/user/dashboard');
      return;
    }

    if (!requireSuperAdmin && isSuperAdmin) {
      router.replace('/superadmin/dashboard');
      return;
    }
  }, [router, requireSuperAdmin]);

  return <>{children}</>;
}