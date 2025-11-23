'use client';

import { useState } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Shield, Key, Send, Copy, CheckCircle, XCircle, Loader2 } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

interface TestResult {
    endpoint: string;
    method: string;
    status: 'success' | 'error' | 'loading';
    statusCode?: number;
    response?: any;
    error?: string;
}

export default function OAuthTestPage() {
    const [clientId, setClientId] = useState('');
    const [clientSecret, setClientSecret] = useState('');
    const [accessToken, setAccessToken] = useState('');
    const [authCode, setAuthCode] = useState('');
    const [callbackUrl, setCallbackUrl] = useState('');
    const [redirectUri, setRedirectUri] = useState('http://localhost:8080/oauth/callback');
    const [scopes, setScopes] = useState('read:profile read:organization');
    const [testResults, setTestResults] = useState<TestResult[]>([]);
    
    const { toast } = useToast();

    // Extract code from callback URL
    const extractCodeFromUrl = (url: string) => {
        try {
            const urlObj = new URL(url);
            const code = urlObj.searchParams.get('code');
            
            // Check if this is the authorization URL (not callback)
            if (urlObj.pathname.includes('/oauth/authorize')) {
                toast({
                    title: "Wrong URL",
                    description: "This is the authorization URL. Please open it in browser first, then paste the callback URL you get after authorization.",
                    variant: "destructive"
                });
                return;
            }
            
            if (code) {
                setAuthCode(code);
                toast({
                    title: "Code Extracted!",
                    description: "Authorization code extracted from URL.",
                });
            } else {
                toast({
                    title: "No Code Found",
                    description: "Could not find 'code' parameter in URL. Make sure you paste the callback URL after authorization.",
                    variant: "destructive"
                });
            }
        } catch (error) {
            toast({
                title: "Invalid URL",
                description: "Please paste the full callback URL.",
                variant: "destructive"
            });
        }
    };

    const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast({
            title: "Copied!",
            description: "Copied to clipboard.",
        });
    };

    const addTestResult = (result: TestResult) => {
        setTestResults(prev => [result, ...prev]);
    };

    const updateTestResult = (index: number, updates: Partial<TestResult>) => {
        setTestResults(prev => {
            const newResults = [...prev];
            newResults[index] = { ...newResults[index], ...updates };
            return newResults;
        });
    };

    // Step 1: Get Authorization URL - Open consent page
    const openAuthorizationPage = () => {
        if (!clientId) {
            toast({
                title: "Missing Client ID",
                description: "Please enter your OAuth2 Client ID first.",
                variant: "destructive"
            });
            return;
        }
        
        // Generate PKCE parameters
        const codeVerifier = generateCodeVerifier();
        generateCodeChallenge(codeVerifier).then(codeChallenge => {
            // Store code_verifier for later token exchange
            localStorage.setItem('pkce_code_verifier', codeVerifier);
            
            const state = generateState();
            
            // Use backend's traditional OAuth2 consent flow (GET shows HTML form, POST processes it)
            const authUrl = `${API_BASE}/api/v1/oauth/authorize?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=${encodeURIComponent(scopes)}&state=${state}&code_challenge=${codeChallenge}&code_challenge_method=S256`;
            
            // Store state for verification
            localStorage.setItem('oauth_state', state);
            
            // Open in new window/tab
            window.open(authUrl, '_blank', 'width=500,height=700');
            
            toast({
                title: "Authorization Page Opened!",
                description: "Login in the new window. After authorization, copy the callback URL and paste it here.",
            });
        });
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
    
    const generateState = () => {
        const array = new Uint8Array(16);
        crypto.getRandomValues(array);
        return base64URLEncode(array);
    };

    // Step 2: Exchange code for token
    const exchangeCodeForToken = async () => {
        const resultIndex = testResults.length;
        addTestResult({
            endpoint: '/api/v1/oauth/token',
            method: 'POST',
            status: 'loading'
        });

        try {
            // Get PKCE code_verifier from localStorage
            const codeVerifier = localStorage.getItem('pkce_code_verifier');
            
            if (!codeVerifier) {
                updateTestResult(resultIndex, {
                    status: 'error',
                    error: 'PKCE code verifier not found. Please go through the authorization flow again.'
                });
                toast({
                    title: "Missing Code Verifier",
                    description: "Please authorize again to get a new code.",
                    variant: "destructive"
                });
                return;
            }

            // Use form data format (application/x-www-form-urlencoded) as required by backend
            const formData = new URLSearchParams();
            formData.append('grant_type', 'authorization_code');
            formData.append('code', authCode);
            formData.append('client_id', clientId);
            formData.append('client_secret', clientSecret);
            formData.append('redirect_uri', redirectUri);
            formData.append('code_verifier', codeVerifier);

            const response = await fetch(`${API_BASE}/api/v1/oauth/token`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: formData.toString()
            });

            const data = await response.json();

            if (response.ok && data.access_token) {
                setAccessToken(data.access_token);
                updateTestResult(resultIndex, {
                    status: 'success',
                    statusCode: response.status,
                    response: data
                });
                toast({
                    title: "Success!",
                    description: "Access token obtained successfully.",
                });
                
                // Clear code verifier after successful exchange
                localStorage.removeItem('pkce_code_verifier');
            } else {
                updateTestResult(resultIndex, {
                    status: 'error',
                    statusCode: response.status,
                    error: data.error || data.error_description || 'Token exchange failed'
                });
            }
        } catch (error) {
            updateTestResult(resultIndex, {
                status: 'error',
                error: error instanceof Error ? error.message : 'Network error'
            });
        }
    };

    // Test API endpoints
    const testEndpoint = async (endpoint: string, method: string = 'GET') => {
        if (!accessToken) {
            toast({
                title: "No Access Token",
                description: "Please obtain an access token first.",
                variant: "destructive"
            });
            return;
        }

        const resultIndex = testResults.length;
        addTestResult({
            endpoint,
            method,
            status: 'loading'
        });

        try {
            const response = await fetch(`${API_BASE}${endpoint}`, {
                method,
                headers: {
                    'Authorization': `Bearer ${accessToken}`,
                    'Content-Type': 'application/json'
                }
            });

            const data = await response.json();

            updateTestResult(resultIndex, {
                status: response.ok ? 'success' : 'error',
                statusCode: response.status,
                response: data,
                error: !response.ok ? data.error || data.message : undefined
            });
        } catch (error) {
            updateTestResult(resultIndex, {
                status: 'error',
                error: error instanceof Error ? error.message : 'Network error'
            });
        }
    };

    const availableEndpoints = [
        { name: 'Get User Profile', endpoint: '/api/v1/user/profile', method: 'GET' },
        { name: 'Get Organizations', endpoint: '/api/v1/user/organizations', method: 'GET' },
        { name: 'List Organizations', endpoint: '/api/v1/organizations', method: 'GET' },
    ];

    return (
        <div className="min-h-screen bg-background p-6">
            <div className="max-w-7xl mx-auto">
                {/* Header */}
                <div className="mb-8">
                    <h1 className="text-3xl font-bold text-foreground flex items-center gap-2">
                        <Shield className="h-8 w-8 text-primary" />
                        OAuth2 Testing Playground
                    </h1>
                    <p className="text-muted-foreground mt-2">
                        Test your OAuth2 integration by pasting your credentials below
                    </p>
                </div>

                <div className="grid lg:grid-cols-2 gap-6">
                    {/* Left Column - Configuration */}
                    <div className="space-y-6">
                        {/* OAuth2 Credentials */}
                        <Card className="p-6">
                            <h2 className="text-xl font-bold text-foreground mb-4 flex items-center gap-2">
                                <Key className="h-5 w-5 text-primary" />
                                OAuth2 Credentials
                            </h2>
                            <div className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-foreground mb-2">
                                        Client ID *
                                    </label>
                                    <Input
                                        value={clientId}
                                        onChange={(e) => setClientId(e.target.value)}
                                        placeholder="Your OAuth2 Client ID"
                                        className="font-mono text-sm"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-foreground mb-2">
                                        Client Secret *
                                    </label>
                                    <Input
                                        type="password"
                                        value={clientSecret}
                                        onChange={(e) => setClientSecret(e.target.value)}
                                        placeholder="Your OAuth2 Client Secret"
                                        className="font-mono text-sm"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-foreground mb-2">
                                        Redirect URI
                                    </label>
                                    <Input
                                        value={redirectUri}
                                        onChange={(e) => setRedirectUri(e.target.value)}
                                        placeholder="http://localhost:3001/callback"
                                        className="font-mono text-sm"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-foreground mb-2">
                                        Scopes (space-separated)
                                    </label>
                                    <Input
                                        value={scopes}
                                        onChange={(e) => setScopes(e.target.value)}
                                        placeholder="read:profile read:organization"
                                        className="font-mono text-sm"
                                    />
                                    <p className="text-xs text-muted-foreground mt-1">
                                        Common scopes: read:profile, read:organization, write:profile
                                    </p>
                                </div>
                            </div>
                        </Card>

                        {/* Step 1: Authorization */}
                        <Card className="p-6">
                            <h3 className="text-lg font-bold text-foreground mb-3">
                                Step 1: Authorize with OAuth2
                            </h3>
                            <p className="text-sm text-muted-foreground mb-4">
                                Click to open authorization page in a new window. Login and authorize, then copy the callback URL.
                            </p>
                            <Button 
                                onClick={openAuthorizationPage}
                                disabled={!clientId}
                                className="w-full"
                            >
                                <Shield className="h-4 w-4 mr-2" />
                                Open OAuth2 Login
                            </Button>
                        </Card>

                        {/* Step 2: Exchange Code */}
                        <Card className="p-6">
                            <h3 className="text-lg font-bold text-foreground mb-3">
                                Step 2: Exchange Code for Token
                            </h3>
                            <p className="text-sm text-muted-foreground mb-4">
                                After authorization, paste the full callback URL here and we'll extract the code.
                            </p>
                            <div className="space-y-3">
                                <div>
                                    <label className="block text-xs font-medium text-muted-foreground mb-1">
                                        Paste Callback URL (e.g., http://localhost:3001/callback?code=xyz)
                                    </label>
                                    <div className="flex gap-2">
                                        <Input
                                            value={callbackUrl}
                                            onChange={(e) => setCallbackUrl(e.target.value)}
                                            placeholder="http://localhost:3001/callback?code=..."
                                            className="font-mono text-sm flex-1"
                                        />
                                        <Button 
                                            onClick={() => extractCodeFromUrl(callbackUrl)}
                                            disabled={!callbackUrl}
                                            variant="outline"
                                        >
                                            Extract
                                        </Button>
                                    </div>
                                </div>
                                
                                {authCode && (
                                    <div>
                                        <label className="block text-xs font-medium text-muted-foreground mb-1">
                                            Authorization Code
                                        </label>
                                        <Input
                                            value={authCode}
                                            onChange={(e) => setAuthCode(e.target.value)}
                                            placeholder="Authorization code"
                                            className="font-mono text-sm"
                                        />
                                    </div>
                                )}
                                
                                <Button 
                                    onClick={exchangeCodeForToken}
                                    disabled={!authCode || !clientId || !clientSecret}
                                    className="w-full"
                                >
                                    <Send className="h-4 w-4 mr-2" />
                                    Exchange for Token
                                </Button>
                            </div>
                        </Card>

                        {/* Access Token Display */}
                        {accessToken && (
                            <Card className="p-6 bg-primary/10 border-primary/30">
                                <h3 className="text-lg font-bold text-foreground mb-3 flex items-center gap-2">
                                    <CheckCircle className="h-5 w-5 text-green-500" />
                                    Access Token Obtained!
                                </h3>
                                <div className="space-y-2">
                                    <div className="bg-slate-950 rounded p-3 overflow-x-auto">
                                        <code className="text-xs text-white break-all">
                                            {accessToken}
                                        </code>
                                    </div>
                                    <Button 
                                        variant="outline" 
                                        size="sm"
                                        onClick={() => copyToClipboard(accessToken)}
                                        className="w-full"
                                    >
                                        <Copy className="h-4 w-4 mr-2" />
                                        Copy Token
                                    </Button>
                                </div>
                            </Card>
                        )}

                        {/* Step 3: Test Endpoints */}
                        <Card className="p-6">
                            <h3 className="text-lg font-bold text-foreground mb-3">
                                Step 3: Test API Endpoints
                            </h3>
                            <p className="text-sm text-muted-foreground mb-4">
                                Test authenticated API requests with your access token.
                            </p>
                            <div className="space-y-2">
                                {availableEndpoints.map((endpoint, index) => (
                                    <Button
                                        key={index}
                                        variant="outline"
                                        onClick={() => testEndpoint(endpoint.endpoint, endpoint.method)}
                                        disabled={!accessToken}
                                        className="w-full justify-start"
                                    >
                                        <Badge variant="secondary" className="mr-2 font-mono text-xs">
                                            {endpoint.method}
                                        </Badge>
                                        {endpoint.name}
                                    </Button>
                                ))}
                            </div>
                        </Card>
                    </div>

                    {/* Right Column - Test Results */}
                    <div>
                        <Card className="p-6 sticky top-6">
                            <h2 className="text-xl font-bold text-foreground mb-4">
                                Test Results
                            </h2>
                            
                            {testResults.length === 0 ? (
                                <div className="text-center py-12 text-muted-foreground">
                                    <p>No tests run yet. Start by getting your authorization URL!</p>
                                </div>
                            ) : (
                                <div className="space-y-4 max-h-[calc(100vh-12rem)] overflow-y-auto">
                                    {testResults.map((result, index) => (
                                        <Card key={index} className={`p-4 ${
                                            result.status === 'success' ? 'border-green-500/50 bg-green-500/5' :
                                            result.status === 'error' ? 'border-destructive/50 bg-destructive/5' :
                                            'border-primary/50 bg-primary/5'
                                        }`}>
                                            <div className="flex items-start justify-between mb-2">
                                                <div className="flex items-center gap-2">
                                                    {result.status === 'loading' && (
                                                        <Loader2 className="h-4 w-4 animate-spin text-primary" />
                                                    )}
                                                    {result.status === 'success' && (
                                                        <CheckCircle className="h-4 w-4 text-green-500" />
                                                    )}
                                                    {result.status === 'error' && (
                                                        <XCircle className="h-4 w-4 text-destructive" />
                                                    )}
                                                    <Badge variant="secondary" className="font-mono text-xs">
                                                        {result.method}
                                                    </Badge>
                                                </div>
                                                {result.statusCode && (
                                                    <Badge variant={result.status === 'success' ? 'default' : 'destructive'}>
                                                        {result.statusCode}
                                                    </Badge>
                                                )}
                                            </div>
                                            
                                            <div className="text-sm font-mono text-foreground mb-2">
                                                {result.endpoint}
                                            </div>

                                            {result.status === 'loading' && (
                                                <p className="text-sm text-muted-foreground">Loading...</p>
                                            )}

                                            {result.error && (
                                                <div className="mt-2 p-2 bg-destructive/10 rounded text-sm text-destructive">
                                                    {result.error}
                                                </div>
                                            )}

                                            {result.response && (
                                                <div className="mt-2">
                                                    <div className="flex items-center justify-between mb-1">
                                                        <span className="text-xs text-muted-foreground">Response:</span>
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => copyToClipboard(JSON.stringify(result.response, null, 2))}
                                                        >
                                                            <Copy className="h-3 w-3" />
                                                        </Button>
                                                    </div>
                                                    <div className="bg-slate-950 rounded p-3 overflow-x-auto max-h-64">
                                                        <pre className="text-xs text-white">
                                                            {JSON.stringify(result.response, null, 2)}
                                                        </pre>
                                                    </div>
                                                </div>
                                            )}
                                        </Card>
                                    ))}
                                </div>
                            )}

                            {testResults.length > 0 && (
                                <Button 
                                    variant="outline" 
                                    onClick={() => setTestResults([])}
                                    className="w-full mt-4"
                                >
                                    Clear Results
                                </Button>
                            )}
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
}
