'use client';

import { useState, useEffect } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Shield, Key, CheckCircle, ExternalLink, Copy, AlertCircle } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

/**
 * Modern OAuth2 Flow Demo
 * 
 * This demonstrates how a Single Page Application (SPA) uses OAuth2
 * with the Authorization Code + PKCE flow.
 * 
 * Flow:
 * 1. User clicks "Login with OAuth2"
 * 2. Generate PKCE code_verifier and code_challenge
 * 3. Redirect to backend authorization endpoint
 * 4. User logs in on backend consent page
 * 5. Backend redirects back with authorization code
 * 6. SPA exchanges code for access token (using PKCE, no client secret)
 * 7. SPA stores token and makes authenticated API calls
 */

interface UserProfile {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    organization?: {
        id: string;
        name: string;
    };
}

export default function ModernOAuthDemo() {
    const [clientId, setClientId] = useState('');
    const [clientSecret, setClientSecret] = useState('');
    const [scope, setScope] = useState('');
    const [redirectUri, setRedirectUri] = useState('');
    const [accessToken, setAccessToken] = useState<string | null>(null);
    const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    
    // Initialize redirect URI on mount
    useEffect(() => {
        if (typeof window !== 'undefined' && !redirectUri) {
            setRedirectUri(`${window.location.origin}/oauth-modern-demo`);
        }
    }, []);

    // Check for authorization code in URL on mount
    useEffect(() => {
        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get('code');
        const state = urlParams.get('state');
        const error = urlParams.get('error');

        if (error) {
            toast({
                title: "Authorization Failed",
                description: urlParams.get('error_description') || error,
                variant: "destructive"
            });
            // Clean URL
            window.history.replaceState({}, '', window.location.pathname);
            return;
        }

        if (code) {
            // Verify state
            const savedState = sessionStorage.getItem('oauth_state');
            if (state !== savedState) {
                toast({
                    title: "Security Error",
                    description: "State mismatch - possible CSRF attack",
                    variant: "destructive"
                });
                window.history.replaceState({}, '', window.location.pathname);
                return;
            }

            // Exchange code for token
            exchangeCodeForToken(code);
            
            // Clean URL
            window.history.replaceState({}, '', window.location.pathname);
        }
    }, []);

    // Load saved token on mount
    useEffect(() => {
        const savedToken = sessionStorage.getItem('access_token');
        if (savedToken) {
            setAccessToken(savedToken);
            fetchUserProfile(savedToken);
        }
    }, []);

    // Generate PKCE code verifier (random string)
    const generateCodeVerifier = (): string => {
        const array = new Uint8Array(32);
        crypto.getRandomValues(array);
        return base64URLEncode(array);
    };

    // Generate PKCE code challenge (SHA-256 hash of verifier)
    const generateCodeChallenge = async (verifier: string): Promise<string> => {
        const encoder = new TextEncoder();
        const data = encoder.encode(verifier);
        const hash = await crypto.subtle.digest('SHA-256', data);
        return base64URLEncode(new Uint8Array(hash));
    };

    // Base64 URL encoding (without padding)
    const base64URLEncode = (buffer: Uint8Array): string => {
        const base64 = btoa(String.fromCharCode(...buffer));
        return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    };

    // Generate random state for CSRF protection
    const generateState = (): string => {
        const array = new Uint8Array(16);
        crypto.getRandomValues(array);
        return base64URLEncode(array);
    };

    // Step 1: Start OAuth2 Flow
    const startOAuthFlow = async () => {
        if (!clientId) {
            toast({
                title: "Missing Client ID",
                description: "Please enter your OAuth2 Client ID",
                variant: "destructive"
            });
            return;
        }

        if (!redirectUri) {
            toast({
                title: "Missing Redirect URI",
                description: "Please enter your Redirect URI",
                variant: "destructive"
            });
            return;
        }

        try {
            // Generate PKCE parameters
            const codeVerifier = generateCodeVerifier();
            const codeChallenge = await generateCodeChallenge(codeVerifier);
            const state = generateState();

            // Store for later use
            sessionStorage.setItem('pkce_code_verifier', codeVerifier);
            sessionStorage.setItem('oauth_state', state);
            sessionStorage.setItem('oauth_client_id', clientId);
            sessionStorage.setItem('oauth_client_secret', clientSecret);
            sessionStorage.setItem('oauth_redirect_uri', redirectUri);

            // Build authorization URL
            const authUrl = new URL(`${API_BASE}/api/v1/oauth/authorize`);
            authUrl.searchParams.set('client_id', clientId);
            authUrl.searchParams.set('redirect_uri', redirectUri);
            authUrl.searchParams.set('response_type', 'code');
            authUrl.searchParams.set('scope', scope);
            authUrl.searchParams.set('state', state);
            authUrl.searchParams.set('code_challenge', codeChallenge);
            authUrl.searchParams.set('code_challenge_method', 'S256');

            // Redirect to authorization endpoint
            window.location.href = authUrl.toString();
        } catch (error) {
            toast({
                title: "Error",
                description: "Failed to start OAuth flow",
                variant: "destructive"
            });
        }
    };

    // Step 2: Exchange authorization code for access token
    const exchangeCodeForToken = async (code: string) => {
        setLoading(true);

        try {
            const codeVerifier = sessionStorage.getItem('pkce_code_verifier');
            
            if (!codeVerifier) {
                throw new Error('Code verifier not found');
            }

            // Get saved client ID, secret, and redirect URI
            const savedClientId = sessionStorage.getItem('oauth_client_id') || clientId;
            const savedClientSecret = sessionStorage.getItem('oauth_client_secret') || '';
            const savedRedirectUri = sessionStorage.getItem('oauth_redirect_uri') || redirectUri;

            // Exchange code for token
            const formData = new URLSearchParams();
            formData.append('grant_type', 'authorization_code');
            formData.append('code', code);
            formData.append('client_id', savedClientId);
            formData.append('redirect_uri', savedRedirectUri);
            formData.append('code_verifier', codeVerifier);
            
            // Add client secret if provided (for confidential clients)
            if (savedClientSecret) {
                formData.append('client_secret', savedClientSecret);
            }

            const response = await fetch(`${API_BASE}/api/v1/oauth/token`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: formData.toString()
            });

            const data = await response.json();

            if (response.ok && data.access_token) {
                // Store access token (in production, use secure storage)
                setAccessToken(data.access_token);
                sessionStorage.setItem('access_token', data.access_token);
                
                if (data.refresh_token) {
                    sessionStorage.setItem('refresh_token', data.refresh_token);
                }

                // Clean up PKCE data
                sessionStorage.removeItem('pkce_code_verifier');
                sessionStorage.removeItem('oauth_state');

                toast({
                    title: "Login Successful!",
                    description: "Access token obtained successfully",
                });

                // Fetch user profile
                await fetchUserProfile(data.access_token);
            } else {
                throw new Error(data.error_description || 'Token exchange failed');
            }
        } catch (error) {
            toast({
                title: "Token Exchange Failed",
                description: error instanceof Error ? error.message : "Unknown error",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    // Step 3: Fetch user profile with access token
    const fetchUserProfile = async (token: string) => {
        try {
            const response = await fetch(`${API_BASE}/api/v1/user/profile`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                setUserProfile(data.data || data);
            } else {
                throw new Error('Failed to fetch profile');
            }
        } catch (error) {
            console.error('Profile fetch error:', error);
        }
    };

    // Logout
    const handleLogout = () => {
        setAccessToken(null);
        setUserProfile(null);
        sessionStorage.removeItem('access_token');
        sessionStorage.removeItem('refresh_token');
        toast({
            title: "Logged Out",
            description: "You have been logged out successfully",
        });
    };

    // Copy token
    const copyToken = () => {
        if (accessToken) {
            navigator.clipboard.writeText(accessToken);
            toast({
                title: "Copied!",
                description: "Access token copied to clipboard",
            });
        }
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-purple-50 to-blue-50 p-6">
            <div className="max-w-4xl mx-auto">
                {/* Header */}
                <div className="mb-8 text-center">
                    <div className="flex items-center justify-center gap-2 mb-4">
                        <Shield className="h-10 w-10 text-purple-600" />
                        <h1 className="text-4xl font-bold text-gray-900">Modern OAuth2 Demo</h1>
                    </div>
                    <p className="text-gray-600 text-lg">
                        Single Page Application (SPA) using Authorization Code + PKCE Flow
                    </p>
                    <Badge variant="outline" className="mt-2">
                        <Key className="h-3 w-3 mr-1" />
                        Public Client (No Client Secret)
                    </Badge>
                </div>

                {/* Info Card */}
                <Card className="p-6 mb-6 bg-blue-50 border-blue-200">
                    <div className="flex items-start gap-3">
                        <AlertCircle className="h-5 w-5 text-blue-600 mt-0.5" />
                        <div>
                            <h3 className="font-semibold text-blue-900 mb-1">How This Works</h3>
                            <ol className="text-sm text-blue-800 space-y-1 list-decimal list-inside">
                                <li>Generate PKCE code_verifier and code_challenge in browser</li>
                                <li>Redirect to backend authorization endpoint with code_challenge</li>
                                <li>User logs in on backend consent page</li>
                                <li>Backend redirects back with authorization code</li>
                                <li>SPA exchanges code + code_verifier for access token (no client secret!)</li>
                                <li>Use access token to call protected APIs</li>
                            </ol>
                        </div>
                    </div>
                </Card>

                {!accessToken ? (
                    /* Login View */
                    <Card className="p-8">
                        <h2 className="text-2xl font-bold text-gray-900 mb-6">Step 1: Start OAuth2 Flow</h2>
                        
                        <div className="space-y-4 mb-6">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    OAuth2 Client ID *
                                </label>
                                <Input
                                    value={clientId}
                                    onChange={(e) => setClientId(e.target.value)}
                                    placeholder="Enter your OAuth2 Client ID"
                                    className="font-mono"
                                />
                                <p className="text-xs text-gray-500 mt-1">
                                    Get this from your OAuth2 client app registration
                                </p>
                            </div>

                            <div>
                                <label className="text-sm font-medium text-gray-700 block mb-2">
                                    Client Secret (Optional)
                                </label>
                                <Input
                                    type="password"
                                    value={clientSecret}
                                    onChange={(e) => setClientSecret(e.target.value)}
                                    placeholder="Leave empty for public clients (SPAs)"
                                    className="font-mono"
                                />
                                <div className="flex items-start gap-2 mt-1">
                                    <AlertCircle className="h-3 w-3 text-amber-600 mt-0.5 flex-shrink-0" />
                                    <p className="text-xs text-amber-600">
                                        <strong>Warning:</strong> Only use for confidential clients (server-side). Never expose secrets in browser/SPA apps!
                                    </p>
                                </div>
                            </div>

                            <div>
                                <label className="text-sm font-medium text-gray-700 block mb-2">
                                    Redirect URI *
                                </label>
                                <Input
                                    value={redirectUri}
                                    onChange={(e) => setRedirectUri(e.target.value)}
                                    placeholder="http://localhost:3000/oauth-modern-demo"
                                    className="font-mono text-sm"
                                />
                                <p className="text-xs text-gray-500 mt-1">
                                    Must match exactly with your registered redirect URI
                                </p>
                            </div>

                            <div>
                                <label className="text-sm font-medium text-gray-700 block mb-2">
                                    Scopes
                                </label>
                                <Input
                                    value={scope}
                                    onChange={(e) => setScope(e.target.value)}
                                    placeholder="read:profile read:organization"
                                    className="font-mono text-sm"
                                />
                                <p className="text-xs text-gray-500 mt-1">
                                    Space-separated list of scopes to request
                                </p>
                            </div>

                            <div className="bg-gray-50 p-4 rounded-lg">
                                <p className="text-sm font-medium text-gray-700 mb-2">OAuth2 Configuration:</p>
                                <div className="text-xs text-gray-600 space-y-1 font-mono">
                                    <div><strong>Response Type:</strong> code</div>
                                    <div><strong>PKCE Method:</strong> S256</div>
                                    <div><strong>Grant Type:</strong> authorization_code</div>
                                </div>
                            </div>
                        </div>

                        <Button 
                            onClick={startOAuthFlow}
                            disabled={!clientId || !redirectUri || loading}
                            className="w-full"
                            size="lg"
                        >
                            <ExternalLink className="h-4 w-4 mr-2" />
                            {loading ? 'Processing...' : 'Login with OAuth2'}
                        </Button>

                        <p className="text-xs text-gray-500 mt-4 text-center">
                            You'll be redirected to the authorization server to log in
                        </p>
                    </Card>
                ) : (
                    /* Authenticated View */
                    <div className="space-y-6">
                        {/* Access Token Card */}
                        <Card className="p-6 bg-green-50 border-green-200">
                            <div className="flex items-start justify-between mb-4">
                                <div className="flex items-center gap-2">
                                    <CheckCircle className="h-6 w-6 text-green-600" />
                                    <h2 className="text-xl font-bold text-green-900">Authenticated!</h2>
                                </div>
                                <Button variant="outline" size="sm" onClick={copyToken}>
                                    <Copy className="h-4 w-4 mr-2" />
                                    Copy Token
                                </Button>
                            </div>
                            <div className="bg-white rounded p-3 mb-4">
                                <p className="text-xs text-gray-500 mb-1">Access Token:</p>
                                <code className="text-xs text-gray-900 break-all block">
                                    {accessToken}
                                </code>
                            </div>
                            <div className="flex gap-2">
                                <Badge variant="outline" className="text-green-700 border-green-300">
                                    <Key className="h-3 w-3 mr-1" />
                                    Bearer Token
                                </Badge>
                                <Badge variant="outline" className="text-green-700 border-green-300">
                                    Valid
                                </Badge>
                            </div>
                        </Card>

                        {/* User Profile Card */}
                        {userProfile && (
                            <Card className="p-6">
                                <h2 className="text-xl font-bold text-gray-900 mb-4">User Profile</h2>
                                <div className="space-y-3">
                                    <div>
                                        <label className="text-xs text-gray-500">Email</label>
                                        <p className="font-medium">{userProfile.email}</p>
                                    </div>
                                    <div>
                                        <label className="text-xs text-gray-500">Name</label>
                                        <p className="font-medium">
                                            {userProfile.first_name} {userProfile.last_name}
                                        </p>
                                    </div>
                                    {userProfile.organization && (
                                        <div>
                                            <label className="text-xs text-gray-500">Organization</label>
                                            <p className="font-medium">{userProfile.organization.name}</p>
                                        </div>
                                    )}
                                    <div>
                                        <label className="text-xs text-gray-500">User ID</label>
                                        <p className="font-mono text-sm text-gray-600">{userProfile.id}</p>
                                    </div>
                                </div>
                            </Card>
                        )}

                        {/* Actions */}
                        <div className="flex gap-4">
                            <Button variant="outline" onClick={handleLogout} className="flex-1">
                                Logout
                            </Button>
                            <Button 
                                onClick={() => fetchUserProfile(accessToken)} 
                                className="flex-1"
                                variant="secondary"
                            >
                                Refresh Profile
                            </Button>
                        </div>
                    </div>
                )}

                {/* Code Example */}
                <Card className="p-6 mt-6 bg-gray-50">
                    <h3 className="font-semibold text-gray-900 mb-3">Implementation Code:</h3>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-xs">
{`// 1. Generate PKCE and redirect to auth endpoint
const codeVerifier = generateCodeVerifier();
const codeChallenge = await sha256(codeVerifier);
sessionStorage.setItem('code_verifier', codeVerifier);

window.location.href = \`\${API_BASE}/api/v1/oauth/authorize?
  client_id=\${clientId}&
  redirect_uri=\${redirectUri}&
  response_type=code&
  code_challenge=\${codeChallenge}&
  code_challenge_method=S256\`;

// 2. After redirect, exchange code for token
const code = new URLSearchParams(window.location.search).get('code');
const response = await fetch(\`\${API_BASE}/api/v1/oauth/token\`, {
  method: 'POST',
  body: new URLSearchParams({
    grant_type: 'authorization_code',
    code: code,
    client_id: clientId,
    redirect_uri: redirectUri,
    code_verifier: sessionStorage.getItem('code_verifier')
    // NO CLIENT SECRET! This is a public client
  })
});

const { access_token } = await response.json();

// 3. Use access token
const profile = await fetch(\`\${API_BASE}/api/v1/user/profile\`, {
  headers: { 'Authorization': \`Bearer \${access_token}\` }
});`}
                    </pre>
                </Card>
            </div>
        </div>
    );
}
