'use client';

import { useState, useEffect, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Shield, Lock, AlertCircle, CheckCircle } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

function AuthorizeContent() {
    const searchParams = useSearchParams();
    const [loading, setLoading] = useState(false);
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [userInfo, setUserInfo] = useState<any>(null);
    const [appInfo, setAppInfo] = useState<any>(null);
    const { toast } = useToast();

    // OAuth2 parameters from URL
    const clientId = searchParams.get('client_id');
    const redirectUri = searchParams.get('redirect_uri');
    const responseType = searchParams.get('response_type');
    const scope = searchParams.get('scope');
    const state = searchParams.get('state');

    // Login form
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');

    const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    // Check if user is already logged in
    useEffect(() => {
        checkAuthStatus();
    }, []);
    
    // Fetch app info after login
    useEffect(() => {
        if (isLoggedIn && clientId) {
            fetchAppInfo();
        }
    }, [isLoggedIn, clientId]);

    const checkAuthStatus = async () => {
        const token = localStorage.getItem('token');
        if (token) {
            try {
                const response = await fetch(`${API_BASE}/api/v1/user/profile`, {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
                if (response.ok) {
                    const data = await response.json();
                    setUserInfo(data);
                    setIsLoggedIn(true);
                }
            } catch (error) {
                console.error('Auth check failed:', error);
            }
        }
    };

    const fetchAppInfo = async () => {
        try {
            const token = localStorage.getItem('token');
            const headers: HeadersInit = {};
            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
            }
            
            const response = await fetch(`${API_BASE}/api/v1/dev/client-apps/${clientId}`, {
                headers
            });
            
            if (response.ok) {
                const data = await response.json();
                setAppInfo(data.data || data);
            } else {
                // If can't fetch app info, just show the client ID
                setAppInfo({ 
                    name: 'OAuth2 Application',
                    client_id: clientId 
                });
            }
        } catch (error) {
            console.error('Failed to fetch app info:', error);
            // Fallback to basic info
            setAppInfo({ 
                name: 'OAuth2 Application',
                client_id: clientId 
            });
        }
    };

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            // Generate PKCE parameters
            const codeVerifier = generateCodeVerifier();
            const codeChallenge = await generateCodeChallenge(codeVerifier);
            
            // Store code_verifier for later token exchange
            localStorage.setItem('pkce_code_verifier', codeVerifier);

            // Use the new OAuth2-specific authorization endpoint
            const response = await fetch(`${API_BASE}/api/v1/oauth/authorize-with-credentials`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    email,
                    password,
                    client_id: clientId,
                    redirect_uri: redirectUri,
                    response_type: responseType,
                    scope: scope || '',
                    state: state || '',
                    code_challenge: codeChallenge,
                    code_challenge_method: 'S256'
                })
            });

            const data = await response.json();

            if (response.ok && data.success && data.redirect_uri) {
                // Success! Backend validated user belongs to org and created auth code
                // Redirect to the callback URL with the code
                window.location.href = data.redirect_uri;
            } else if (response.status === 401) {
                toast({
                    title: "Invalid Credentials",
                    description: data.error_description || "Email or password is incorrect",
                    variant: "destructive"
                });
            } else if (response.status === 403) {
                toast({
                    title: "Access Denied",
                    description: data.error_description || "You don't have access to this application",
                    variant: "destructive"
                });
            } else {
                toast({
                    title: "Authorization Failed",
                    description: data.error_description || data.message || "Could not authorize",
                    variant: "destructive"
                });
            }
        } catch (error) {
            toast({
                title: "Error",
                description: "Network error occurred",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    const selectOrganization = async (orgId: string, tempToken: string) => {
        try {
            const response = await fetch(`${API_BASE}/api/v1/auth/select-organization`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${tempToken}`
                },
                body: JSON.stringify({ organization_id: orgId })
            });

            const data = await response.json();

            if (response.ok && (data.token || data.data?.token)) {
                const token = data.token || data.data?.token;
                const user = data.user || data.data?.user;
                
                localStorage.setItem('token', token);
                setUserInfo(user);
                setIsLoggedIn(true);
                
                // Refetch app info with token
                if (clientId) {
                    fetchAppInfo();
                }
                
                toast({
                    title: "Login Successful",
                    description: "Please review the permissions below.",
                });
            } else {
                toast({
                    title: "Organization Selection Failed",
                    description: data.message || "Could not select organization",
                    variant: "destructive"
                });
            }
        } catch (error) {
            toast({
                title: "Error",
                description: "Organization selection failed",
                variant: "destructive"
            });
        }
    };

    const selectOrganizationWithCredentials = async (orgId: string) => {
        try {
            // First login again to get session
            const loginResponse = await fetch(`${API_BASE}/api/v1/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            if (!loginResponse.ok) {
                throw new Error('Re-login failed');
            }

            // Now select organization
            const response = await fetch(`${API_BASE}/api/v1/auth/select-organization`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    organization_id: orgId,
                    email,
                    password
                })
            });

            const data = await response.json();

            if (response.ok && (data.token || data.data?.token)) {
                const token = data.token || data.data?.token;
                const user = data.user || data.data?.user;
                
                localStorage.setItem('token', token);
                setUserInfo(user);
                setIsLoggedIn(true);
                
                // Refetch app info with token
                if (clientId) {
                    fetchAppInfo();
                }
                
                toast({
                    title: "Login Successful",
                    description: "Please review the permissions below.",
                });
            } else {
                toast({
                    title: "Organization Selection Failed",
                    description: data.message || "Could not select organization",
                    variant: "destructive"
                });
            }
        } catch (error) {
            toast({
                title: "Error",
                description: "Organization selection failed",
                variant: "destructive"
            });
        }
    };

    const handleAuthorize = async () => {
        if (!clientId || !redirectUri || !responseType) {
            toast({
                title: "Invalid Request",
                description: "Missing required OAuth2 parameters",
                variant: "destructive"
            });
            return;
        }

        setLoading(true);

        try {
            const token = localStorage.getItem('token');
            
            // Build the full authorization URL and redirect directly
            // The backend will handle the redirect with the code
            const params = new URLSearchParams({
                client_id: clientId,
                redirect_uri: redirectUri,
                response_type: responseType,
            });
            if (scope) params.append('scope', scope);
            if (state) params.append('state', state);
            
            // Generate PKCE parameters (required by backend)
            const codeVerifier = generateCodeVerifier();
            const codeChallenge = await generateCodeChallenge(codeVerifier);
            
            // Store code_verifier in localStorage for later use
            localStorage.setItem('pkce_code_verifier', codeVerifier);
            
            params.append('code_challenge', codeChallenge);
            params.append('code_challenge_method', 'S256');

            // Redirect to backend OAuth authorize endpoint with token
            // The backend will redirect to redirect_uri with code
            window.location.href = `${API_BASE}/api/v1/oauth/authorize?${params.toString()}`;
            
        } catch (error) {
            toast({
                title: "Error",
                description: "Authorization request failed",
                variant: "destructive"
            });
            setLoading(false);
        }
    };
    
    // PKCE helper functions
    const generateCodeVerifier = () => {
        const array = new Uint8Array(32);
        crypto.getRandomValues(array);
        return base64URLEncode(array);
    };
    
    const generateCodeChallenge = async (verifier: string) => {
        const encoder = new TextEncoder();
        const data = encoder.encode(verifier);
        const hash = await crypto.subtle.digest('SHA-256', data);
        return base64URLEncode(new Uint8Array(hash));
    };
    
    const base64URLEncode = (buffer: Uint8Array) => {
        const base64 = btoa(String.fromCharCode(...buffer));
        return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    };

    const handleDeny = () => {
        if (redirectUri) {
            const url = new URL(redirectUri);
            url.searchParams.append('error', 'access_denied');
            url.searchParams.append('error_description', 'User denied authorization');
            if (state) url.searchParams.append('state', state);
            window.location.href = url.toString();
        }
    };

    // Validation
    if (!clientId || !redirectUri || !responseType) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center p-6">
                <Card className="p-8 max-w-md w-full border-destructive">
                    <div className="text-center">
                        <AlertCircle className="h-12 w-12 text-destructive mx-auto mb-4" />
                        <h1 className="text-2xl font-bold text-foreground mb-2">Invalid Request</h1>
                        <p className="text-muted-foreground">
                            Missing required OAuth2 parameters. Please check your authorization URL.
                        </p>
                    </div>
                </Card>
            </div>
        );
    }

    if (responseType !== 'code') {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center p-6">
                <Card className="p-8 max-w-md w-full border-destructive">
                    <div className="text-center">
                        <AlertCircle className="h-12 w-12 text-destructive mx-auto mb-4" />
                        <h1 className="text-2xl font-bold text-foreground mb-2">Unsupported Response Type</h1>
                        <p className="text-muted-foreground">
                            Only 'authorization_code' flow is supported. Got: {responseType}
                        </p>
                    </div>
                </Card>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background flex items-center justify-center p-6">
            <Card className="p-8 max-w-md w-full">
                {/* Header */}
                <div className="text-center mb-6">
                    <Shield className="h-12 w-12 text-primary mx-auto mb-3" />
                    <h1 className="text-2xl font-bold text-foreground">Authorize Application</h1>
                </div>

                {/* App Info */}
                {appInfo && (
                    <Card className="p-4 mb-6 bg-muted/50">
                        <div className="flex items-center gap-3 mb-2">
                            <Lock className="h-5 w-5 text-primary" />
                            <div>
                                <p className="font-semibold text-foreground">{appInfo.name}</p>
                                <p className="text-xs text-muted-foreground">Client ID: {appInfo.client_id}</p>
                            </div>
                        </div>
                        {scope && (
                            <div className="mt-3 pt-3 border-t border-border">
                                <p className="text-xs font-medium text-foreground mb-2">Requested Permissions:</p>
                                <div className="flex flex-wrap gap-1">
                                    {scope.split(' ').map((s, i) => (
                                        <span key={i} className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded">
                                            {s}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        )}
                    </Card>
                )}

                {/* Login Form or Authorization Buttons */}
                {!isLoggedIn ? (
                    <form onSubmit={handleLogin} className="space-y-4">
                        <div>
                            <p className="text-sm text-muted-foreground mb-4 text-center">
                                Sign in to authorize this application
                            </p>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                Email
                            </label>
                            <Input
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                placeholder="your@email.com"
                                required
                                disabled={loading}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                Password
                            </label>
                            <Input
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="Your password"
                                required
                                disabled={loading}
                            />
                        </div>
                        <Button type="submit" className="w-full" disabled={loading}>
                            {loading ? 'Signing in...' : 'Sign In'}
                        </Button>
                    </form>
                ) : (
                    <div className="space-y-4">
                        <Card className="p-4 bg-primary/10 border-primary/30">
                            <div className="flex items-center gap-2 mb-2">
                                <CheckCircle className="h-5 w-5 text-green-500" />
                                <p className="font-medium text-foreground">Signed in as</p>
                            </div>
                            <p className="text-sm text-muted-foreground">{userInfo?.email}</p>
                        </Card>

                        <div className="bg-muted/50 rounded-lg p-4">
                            <p className="text-sm text-foreground mb-2">
                                <strong>{appInfo?.name || 'This application'}</strong> would like to:
                            </p>
                            <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                                {scope?.split(' ').map((s, i) => (
                                    <li key={i}>
                                        {s === 'read:profile' && 'Read your profile information'}
                                        {s === 'read:organization' && 'Read your organization information'}
                                        {s === 'write:profile' && 'Update your profile information'}
                                        {!['read:profile', 'read:organization', 'write:profile'].includes(s) && s}
                                    </li>
                                ))}
                            </ul>
                        </div>

                        <div className="flex gap-3">
                            <Button
                                variant="outline"
                                onClick={handleDeny}
                                disabled={loading}
                                className="flex-1"
                            >
                                Deny
                            </Button>
                            <Button
                                onClick={handleAuthorize}
                                disabled={loading}
                                className="flex-1"
                            >
                                {loading ? 'Authorizing...' : 'Authorize'}
                            </Button>
                        </div>

                        <p className="text-xs text-muted-foreground text-center mt-4">
                            By authorizing, you allow this application to access your information as described above.
                        </p>
                    </div>
                )}
            </Card>
        </div>
    );
}

export default function AuthorizePage() {
    return (
        <Suspense fallback={
            <div className="min-h-screen bg-background flex items-center justify-center">
                <div className="animate-pulse">Loading...</div>
            </div>
        }>
            <AuthorizeContent />
        </Suspense>
    );
}
