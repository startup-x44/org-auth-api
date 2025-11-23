import { NextRequest, NextResponse } from 'next/server';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Skip middleware for auth pages and static files
  if (pathname.startsWith('/auth') || 
      pathname.startsWith('/_next') || 
      pathname.startsWith('/api') ||
      pathname === '/favicon.ico') {
    return NextResponse.next();
  }

  // For client-side auth check, redirect to login for protected routes
  if (pathname.startsWith('/user') || pathname.startsWith('/superadmin')) {
    // Client-side will handle the auth check since tokens are in localStorage
    return NextResponse.next();
  }

  // Root redirect to login
  if (pathname === '/') {
    return NextResponse.redirect(new URL('/auth/login', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};