import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { ArrowLeft, Copy, Terminal, Book, CheckCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useToast } from '@/hooks/use-toast'

import { startAuthFlow } from '@/lib/oauth'

const DeveloperDocs = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [copiedCode, setCopiedCode] = useState<string | null>(null)

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    setCopiedCode(label)
    setTimeout(() => setCopiedCode(null), 2000)
    toast({
      title: 'Copied!',
      description: `${label} copied to clipboard`,
    })
  }

  const testOAuthFlow = async () => {
    try {
      // Use the OAuth helper to start the flow
      await startAuthFlow({
        clientId: import.meta.env.VITE_OAUTH_CLIENT_ID || 'demo-client-id',
        redirectUri: `${window.location.origin}/oauth/callback`,
        scope: 'profile email',
        state: JSON.stringify({ return_to: '/developer/docs' })
      })
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to start OAuth flow',
        variant: 'destructive'
      })
    }
  }

  const authServerUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      {/* Header */}
      <header className="bg-white dark:bg-slate-800 shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => navigate('/dashboard')}
                className="flex items-center space-x-2"
              >
                <ArrowLeft className="h-4 w-4" />
                <span>Back to Dashboard</span>
              </Button>
            </div>
            <div className="flex items-center space-x-4">
              <Book className="h-6 w-6 text-primary" />
              <h1 className="text-xl font-semibold text-foreground">
                Developer Documentation
              </h1>
            </div>
            <div className="w-24"></div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="space-y-8"
        >
          <div>
            <h2 className="text-3xl font-bold mb-2">OAuth2.1 Integration Guide</h2>
            <p className="text-muted-foreground">
              Learn how to integrate OAuth2.1 authentication into your application
            </p>
          </div>

          <Tabs defaultValue="quickstart" className="w-full">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="quickstart">Quick Start</TabsTrigger>
              <TabsTrigger value="javascript">JavaScript</TabsTrigger>
              <TabsTrigger value="nodejs">Node.js</TabsTrigger>
              <TabsTrigger value="react">React</TabsTrigger>
            </TabsList>

            {/* Quick Start */}
            <TabsContent value="quickstart" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Getting Started</CardTitle>
                  <CardDescription>
                    Set up OAuth2.1 authentication in 3 simple steps
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  {/* Step 1 */}
                  <div>
                    <div className="flex items-center space-x-2 mb-3">
                      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground text-sm font-bold">
                        1
                      </div>
                      <h3 className="text-lg font-semibold">Create a Client Application</h3>
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">
                      Contact your administrator to create an OAuth client app and get your credentials.
                    </p>
                    <div className="bg-muted p-4 rounded-lg">
                      <div className="space-y-2">
                        <div>
                          <span className="text-sm font-medium">Client ID:</span>
                          <code className="ml-2 text-xs">YOUR_CLIENT_ID</code>
                        </div>
                        <div>
                          <span className="text-sm font-medium">Client Secret:</span>
                          <code className="ml-2 text-xs">YOUR_CLIENT_SECRET</code>
                        </div>
                        <div>
                          <span className="text-sm font-medium">Redirect URI:</span>
                          <code className="ml-2 text-xs">https://your-app.com/callback</code>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Step 2 */}
                  <div>
                    <div className="flex items-center space-x-2 mb-3">
                      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground text-sm font-bold">
                        2
                      </div>
                      <h3 className="text-lg font-semibold">Implement PKCE Flow</h3>
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">
                      Generate PKCE challenge and redirect users to the authorization endpoint.
                    </p>
                    <CodeBlock
                      language="javascript"
                      code={`// Generate PKCE code verifier and challenge
const codeVerifier = generateRandomString(128);
const codeChallenge = await sha256(codeVerifier);

// Store code verifier for later use
sessionStorage.setItem('code_verifier', codeVerifier);

// Redirect to authorization endpoint
const authUrl = new URL('${authServerUrl}/oauth/authorize');
authUrl.searchParams.set('client_id', 'YOUR_CLIENT_ID');
authUrl.searchParams.set('redirect_uri', 'https://your-app.com/callback');
authUrl.searchParams.set('response_type', 'code');
authUrl.searchParams.set('scope', 'profile email');
authUrl.searchParams.set('code_challenge', codeChallenge);
authUrl.searchParams.set('code_challenge_method', 'S256');
authUrl.searchParams.set('state', generateRandomString(32));

window.location.href = authUrl.toString();`}
                      onCopy={(code) => copyToClipboard(code, 'PKCE code')}
                      copied={copiedCode === 'PKCE code'}
                    />
                  </div>

                  {/* Step 3 */}
                  <div>
                    <div className="flex items-center space-x-2 mb-3">
                      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground text-sm font-bold">
                        3
                      </div>
                      <h3 className="text-lg font-semibold">Exchange Code for Token</h3>
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">
                      Handle the callback and exchange authorization code for access token.
                    </p>
                    <CodeBlock
                      language="javascript"
                      code={`// In your callback handler
const urlParams = new URLSearchParams(window.location.search);
const code = urlParams.get('code');
const codeVerifier = sessionStorage.getItem('code_verifier');

// Exchange code for token
const response = await fetch('${authServerUrl}/oauth/token', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
  },
  body: new URLSearchParams({
    grant_type: 'authorization_code',
    code: code,
    redirect_uri: 'https://your-app.com/callback',
    client_id: 'YOUR_CLIENT_ID',
    code_verifier: codeVerifier,
  }),
});

const { access_token, refresh_token, id_token } = await response.json();

// Store tokens securely
sessionStorage.setItem('access_token', access_token);
sessionStorage.setItem('refresh_token', refresh_token);`}
                      onCopy={(code) => copyToClipboard(code, 'Token exchange code')}
                      copied={copiedCode === 'Token exchange code'}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* JavaScript SDK */}
            <TabsContent value="javascript" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>JavaScript Client</CardTitle>
                  <CardDescription>
                    Pure JavaScript implementation for web applications
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="text-lg font-semibold">PKCE Helper Functions</h3>
                      <Button onClick={testOAuthFlow} size="sm" className="flex items-center space-x-2">
                        <Terminal className="h-4 w-4" />
                        <span>Test OAuth Flow</span>
                      </Button>
                    </div>
                    <CodeBlock
                      language="javascript"
                      code={`// Import from our utilities (frontend/src/lib/pkce.ts)
import { generatePKCEPair } from '@/lib/pkce'
import { startAuthFlow } from '@/lib/oauth'

// Generate PKCE pair
const { codeVerifier, codeChallenge, codeChallengeMethod } = await generatePKCEPair()

// Start OAuth flow (this handles PKCE automatically)
await startAuthFlow({
  clientId: 'your-client-id',
  redirectUri: window.location.origin + '/oauth/callback',
  scope: 'profile email'
})

// Legacy helper functions (for reference)
function generateRandomString(length) {
  const charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
  const randomValues = new Uint8Array(length);
  crypto.getRandomValues(randomValues);
  return Array.from(randomValues)
    .map(v => charset[v % charset.length])
    .join('');
}

async function sha256(plain) {
  const encoder = new TextEncoder();
  const data = encoder.encode(plain);
  const hash = await crypto.subtle.digest('SHA-256', data);
  return base64UrlEncode(hash);
}

function base64UrlEncode(buffer) {
  const bytes = new Uint8Array(buffer);
  const binary = String.fromCharCode(...bytes);
  const base64 = btoa(binary);
  return base64
    .replace(/\\+/g, '-')
    .replace(/\\//g, '_')
    .replace(/=/g, '');
}`}
                      onCopy={(code) => copyToClipboard(code, 'PKCE helpers')}
                      copied={copiedCode === 'PKCE helpers'}
                    />
                  </div>

                  <div>
                    <h3 className="text-lg font-semibold mb-3">Complete Auth Client</h3>
                    <CodeBlock
                      language="javascript"
                      code={`class AuthClient {
  constructor(config) {
    this.clientId = config.clientId;
    this.redirectUri = config.redirectUri;
    this.authServerUrl = '${authServerUrl}';
  }

  async login(options = {}) {
    const codeVerifier = generateRandomString(128);
    const codeChallenge = await sha256(codeVerifier);
    const state = generateRandomString(32);

    // Store for later use
    sessionStorage.setItem('code_verifier', codeVerifier);
    sessionStorage.setItem('oauth_state', state);

    const authUrl = new URL(\`\${this.authServerUrl}/oauth/authorize\`);
    authUrl.searchParams.set('client_id', this.clientId);
    authUrl.searchParams.set('redirect_uri', this.redirectUri);
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('scope', options.scope || 'profile email');
    authUrl.searchParams.set('code_challenge', codeChallenge);
    authUrl.searchParams.set('code_challenge_method', 'S256');
    authUrl.searchParams.set('state', state);

    window.location.href = authUrl.toString();
  }

  async handleCallback() {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const state = urlParams.get('state');
    const storedState = sessionStorage.getItem('oauth_state');
    const codeVerifier = sessionStorage.getItem('code_verifier');

    if (state !== storedState) {
      throw new Error('State mismatch - possible CSRF attack');
    }

    const response = await fetch(\`\${this.authServerUrl}/oauth/token\`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        code,
        redirect_uri: this.redirectUri,
        client_id: this.clientId,
        code_verifier: codeVerifier,
      }),
    });

    if (!response.ok) {
      throw new Error('Token exchange failed');
    }

    const tokens = await response.json();
    
    // Store tokens
    sessionStorage.setItem('access_token', tokens.access_token);
    sessionStorage.setItem('refresh_token', tokens.refresh_token);
    
    // Clean up
    sessionStorage.removeItem('code_verifier');
    sessionStorage.removeItem('oauth_state');

    return tokens;
  }

  async refreshToken() {
    const refreshToken = sessionStorage.getItem('refresh_token');
    
    const response = await fetch(\`\${this.authServerUrl}/oauth/token\`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
        client_id: this.clientId,
      }),
    });

    const tokens = await response.json();
    sessionStorage.setItem('access_token', tokens.access_token);
    sessionStorage.setItem('refresh_token', tokens.refresh_token);
    
    return tokens;
  }

  logout() {
    sessionStorage.removeItem('access_token');
    sessionStorage.removeItem('refresh_token');
  }
}`}
                      onCopy={(code) => copyToClipboard(code, 'Auth client')}
                      copied={copiedCode === 'Auth client'}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* Node.js */}
            <TabsContent value="nodejs" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Node.js Backend Integration</CardTitle>
                  <CardDescription>
                    Server-side OAuth2.1 implementation for Node.js
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div>
                    <h3 className="text-lg font-semibold mb-3">Install Dependencies</h3>
                    <CodeBlock
                      language="bash"
                      code={`npm install express axios crypto`}
                      onCopy={(code) => copyToClipboard(code, 'npm install')}
                      copied={copiedCode === 'npm install'}
                    />
                  </div>

                  <div>
                    <h3 className="text-lg font-semibold mb-3">Express Server Example</h3>
                    <CodeBlock
                      language="javascript"
                      code={`const express = require('express');
const crypto = require('crypto');
const axios = require('axios');

const app = express();

const config = {
  clientId: import.meta.env.VITE_OAUTH_CLIENT_ID,
  clientSecret: import.meta.env.VITE_OAUTH_CLIENT_SECRET,
  redirectUri: 'http://localhost:3000/callback',
  authServerUrl: '${authServerUrl}',
};

// Login route
app.get('/login', (req, res) => {
  const codeVerifier = crypto.randomBytes(64).toString('base64url');
  const codeChallenge = crypto
    .createHash('sha256')
    .update(codeVerifier)
    .digest('base64url');
  const state = crypto.randomBytes(32).toString('base64url');

  // Store in session (use express-session in production)
  req.session = { codeVerifier, state };

  const authUrl = new URL(\`\${config.authServerUrl}/oauth/authorize\`);
  authUrl.searchParams.set('client_id', config.clientId);
  authUrl.searchParams.set('redirect_uri', config.redirectUri);
  authUrl.searchParams.set('response_type', 'code');
  authUrl.searchParams.set('scope', 'profile email');
  authUrl.searchParams.set('code_challenge', codeChallenge);
  authUrl.searchParams.set('code_challenge_method', 'S256');
  authUrl.searchParams.set('state', state);

  res.redirect(authUrl.toString());
});

// Callback route
app.get('/callback', async (req, res) => {
  const { code, state } = req.query;
  const { codeVerifier, state: storedState } = req.session;

  if (state !== storedState) {
    return res.status(400).send('State mismatch');
  }

  try {
    const response = await axios.post(
      \`\${config.authServerUrl}/oauth/token\`,
      new URLSearchParams({
        grant_type: 'authorization_code',
        code,
        redirect_uri: config.redirectUri,
        client_id: config.clientId,
        code_verifier: codeVerifier,
      }),
      {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      }
    );

    const { access_token, refresh_token, id_token } = response.data;
    
    // Store tokens in session
    req.session.accessToken = access_token;
    req.session.refreshToken = refresh_token;

    res.redirect('/dashboard');
  } catch (error) {
    res.status(500).send('Token exchange failed');
  }
});

app.listen(3000, () => {
  console.log('Server running on http://localhost:3000');
});`}
                      onCopy={(code) => copyToClipboard(code, 'Express server')}
                      copied={copiedCode === 'Express server'}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* React */}
            <TabsContent value="react" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>React Integration</CardTitle>
                  <CardDescription>
                    OAuth2.1 authentication in React applications
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div>
                    <h3 className="text-lg font-semibold mb-3">Auth Hook</h3>
                    <CodeBlock
                      language="typescript"
                      code={`import { useState, useEffect } from 'react';

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [accessToken, setAccessToken] = useState<string | null>(null);

  useEffect(() => {
    const token = sessionStorage.getItem('access_token');
    if (token) {
      setAccessToken(token);
      setIsAuthenticated(true);
    }
  }, []);

  const login = async () => {
    const codeVerifier = generateRandomString(128);
    const codeChallenge = await sha256(codeVerifier);
    const state = generateRandomString(32);

    sessionStorage.setItem('code_verifier', codeVerifier);
    sessionStorage.setItem('oauth_state', state);

    const authUrl = new URL('${authServerUrl}/oauth/authorize');
    authUrl.searchParams.set('client_id', 'YOUR_CLIENT_ID');
    authUrl.searchParams.set('redirect_uri', window.location.origin + '/callback');
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('scope', 'profile email');
    authUrl.searchParams.set('code_challenge', codeChallenge);
    authUrl.searchParams.set('code_challenge_method', 'S256');
    authUrl.searchParams.set('state', state);

    window.location.href = authUrl.toString();
  };

  const logout = () => {
    sessionStorage.removeItem('access_token');
    sessionStorage.removeItem('refresh_token');
    setAccessToken(null);
    setIsAuthenticated(false);
  };

  return { isAuthenticated, accessToken, login, logout };
}`}
                      onCopy={(code) => copyToClipboard(code, 'React hook')}
                      copied={copiedCode === 'React hook'}
                    />
                  </div>

                  <div>
                    <h3 className="text-lg font-semibold mb-3">Callback Component</h3>
                    <CodeBlock
                      language="typescript"
                      code={`import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

export function OAuthCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  useEffect(() => {
    const handleCallback = async () => {
      const code = searchParams.get('code');
      const state = searchParams.get('state');
      const storedState = sessionStorage.getItem('oauth_state');
      const codeVerifier = sessionStorage.getItem('code_verifier');

      if (state !== storedState) {
        throw new Error('State mismatch');
      }

      const response = await fetch('${authServerUrl}/oauth/token', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'authorization_code',
          code: code!,
          redirect_uri: window.location.origin + '/callback',
          client_id: 'YOUR_CLIENT_ID',
          code_verifier: codeVerifier!,
        }),
      });

      const tokens = await response.json();
      sessionStorage.setItem('access_token', tokens.access_token);
      sessionStorage.setItem('refresh_token', tokens.refresh_token);
      
      navigate('/dashboard');
    };

    handleCallback();
  }, [searchParams, navigate]);

  return <div>Processing login...</div>;
}`}
                      onCopy={(code) => copyToClipboard(code, 'Callback component')}
                      copied={copiedCode === 'Callback component'}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>

          {/* API Endpoints Reference */}
          <Card>
            <CardHeader>
              <CardTitle>API Endpoints</CardTitle>
              <CardDescription>OAuth2.1 server endpoints reference</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <EndpointCard
                  method="GET"
                  endpoint="/oauth/authorize"
                  description="Authorization endpoint - redirects user to consent screen"
                  params={[
                    { name: 'client_id', required: true, description: 'Your client ID' },
                    { name: 'redirect_uri', required: true, description: 'Registered redirect URI' },
                    { name: 'response_type', required: true, description: 'Must be "code"' },
                    { name: 'scope', required: false, description: 'Space-separated scopes' },
                    { name: 'state', required: false, description: 'CSRF protection token' },
                    { name: 'code_challenge', required: true, description: 'PKCE code challenge (S256)' },
                    { name: 'code_challenge_method', required: true, description: 'Must be "S256"' },
                  ]}
                />

                <EndpointCard
                  method="POST"
                  endpoint="/oauth/token"
                  description="Token endpoint - exchange code for access token"
                  params={[
                    { name: 'grant_type', required: true, description: '"authorization_code" or "refresh_token"' },
                    { name: 'code', required: true, description: 'Authorization code (for authorization_code grant)' },
                    { name: 'redirect_uri', required: true, description: 'Same redirect URI from authorize' },
                    { name: 'client_id', required: true, description: 'Your client ID' },
                    { name: 'code_verifier', required: true, description: 'PKCE code verifier' },
                  ]}
                />

                <EndpointCard
                  method="GET"
                  endpoint="/oauth/userinfo"
                  description="Get authenticated user information"
                  params={[
                    { name: 'Authorization', required: true, description: 'Bearer {access_token}' },
                  ]}
                />
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </main>
    </div>
  )
}

// Code Block Component
const CodeBlock = ({ language, code, onCopy, copied }: {
  language: string
  code: string
  onCopy: (code: string) => void
  copied: boolean
}) => {
  return (
    <div className="relative group">
      <div className="absolute top-2 right-2 z-10">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onCopy(code)}
          className="opacity-0 group-hover:opacity-100 transition-opacity"
        >
          {copied ? <CheckCircle className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
        </Button>
      </div>
      <div className="bg-slate-950 text-slate-50 p-4 rounded-lg overflow-x-auto">
        <div className="flex items-center space-x-2 mb-3 pb-2 border-b border-slate-700">
          <Terminal className="h-4 w-4 text-slate-400" />
          <span className="text-xs text-slate-400 uppercase">{language}</span>
        </div>
        <pre className="text-sm">
          <code>{code}</code>
        </pre>
      </div>
    </div>
  )
}

// Endpoint Card Component
const EndpointCard = ({ method, endpoint, description, params }: {
  method: string
  endpoint: string
  description: string
  params: Array<{ name: string; required: boolean; description: string }>
}) => {
  const methodColors: Record<string, string> = {
    GET: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    POST: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    PUT: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
    DELETE: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  }

  return (
    <div className="border rounded-lg p-4">
      <div className="flex items-center space-x-3 mb-2">
        <span className={`px-2 py-1 rounded text-xs font-mono font-bold ${methodColors[method]}`}>
          {method}
        </span>
        <code className="text-sm font-mono">{endpoint}</code>
      </div>
      <p className="text-sm text-muted-foreground mb-3">{description}</p>
      <div className="space-y-2">
        {params.map((param) => (
          <div key={param.name} className="flex items-start space-x-2 text-sm">
            <code className="text-xs bg-muted px-2 py-0.5 rounded">{param.name}</code>
            {param.required && (
              <span className="text-xs text-red-500">*</span>
            )}
            <span className="text-muted-foreground text-xs">{param.description}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default DeveloperDocs
