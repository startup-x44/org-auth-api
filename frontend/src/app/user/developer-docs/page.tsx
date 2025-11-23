'use client';

import { useState } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { BookOpen, Lock, Key, Copy, Check, Code, AlertTriangle } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

export default function DeveloperDocsPage() {
    const [copiedSection, setCopiedSection] = useState<string | null>(null);
    const { toast } = useToast();

    const copyToClipboard = (text: string, section: string) => {
        navigator.clipboard.writeText(text);
        setCopiedSection(section);
        setTimeout(() => setCopiedSection(null), 2000);
        toast({
            title: "Copied!",
            description: "Code snippet copied to clipboard.",
        });
    };

    return (
        <div className="p-6 max-w-7xl mx-auto">
            {/* Header */}
            <div className="mb-8">
                <h1 className="text-3xl font-bold text-foreground flex items-center gap-2">
                    <BookOpen className="h-8 w-8 text-primary" />
                    Developer Documentation
                </h1>
                <p className="text-muted-foreground mt-2">
                    Learn how to integrate OAuth2 and API Keys into your applications
                </p>
            </div>

            {/* Table of Contents */}
            <Card className="p-6 mb-8 bg-muted/30">
                <h2 className="text-xl font-bold text-foreground mb-4">Quick Links</h2>
                <div className="grid md:grid-cols-2 gap-4">
                    <a href="#oauth2" className="flex items-center gap-2 text-primary hover:underline">
                        <Lock className="h-4 w-4" />
                        OAuth2 Integration
                    </a>
                    <a href="#api-keys" className="flex items-center gap-2 text-primary hover:underline">
                        <Key className="h-4 w-4" />
                        API Key Integration
                    </a>
                </div>
            </Card>

            {/* OAuth2 Integration Section */}
            <div id="oauth2" className="mb-12">
                <Card className="p-8">
                    <div className="flex items-center gap-3 mb-6">
                        <Lock className="h-8 w-8 text-primary" />
                        <h2 className="text-2xl font-bold text-foreground">OAuth2 Integration</h2>
                    </div>

                    <div className="space-y-6">
                        {/* Overview */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-2">Overview</h3>
                            <p className="text-muted-foreground">
                                OAuth2 allows users to authorize your application to access their data without sharing their password.
                                This is ideal for third-party integrations and multi-user applications.
                            </p>
                        </div>

                        {/* Step 1: Create OAuth2 App */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Step 1: Create an OAuth2 Application</h3>
                            <ol className="list-decimal list-inside space-y-2 text-muted-foreground ml-4">
                                <li>Navigate to <a href="/user/oauth2" className="text-primary hover:underline">OAuth2 Apps</a></li>
                                <li>Click "Create OAuth2 App"</li>
                                <li>Enter your application name and redirect URIs</li>
                                <li>Configure the required scopes (permissions)</li>
                                <li>Save your Client ID and Client Secret securely</li>
                            </ol>
                        </div>

                        {/* Client Types */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Client Types</h3>
                            <div className="grid md:grid-cols-2 gap-4 mb-4">
                                <Card className="p-4 border-primary/30 bg-primary/5">
                                    <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                                        📱 Public Client
                                    </h4>
                                    <p className="text-sm text-muted-foreground mb-2">
                                        For SPAs, mobile apps, or any client that cannot securely store secrets.
                                    </p>
                                    <ul className="text-xs text-muted-foreground space-y-1 ml-4 list-disc">
                                        <li>Uses PKCE (required)</li>
                                        <li>No client secret needed</li>
                                        <li>Best for React, Vue, Angular apps</li>
                                    </ul>
                                </Card>
                                <Card className="p-4 border-primary/30 bg-primary/5">
                                    <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                                        🔒 Confidential Client
                                    </h4>
                                    <p className="text-sm text-muted-foreground mb-2">
                                        For server-side apps that can securely store secrets on the backend.
                                    </p>
                                    <ul className="text-xs text-muted-foreground space-y-1 ml-4 list-disc">
                                        <li>Uses PKCE + client secret</li>
                                        <li>Client secret required</li>
                                        <li>Best for Node.js, Python, PHP, Java</li>
                                    </ul>
                                </Card>
                            </div>
                        </div>

                        {/* Step 2: PKCE Setup */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Step 2: Generate PKCE Parameters</h3>
                            <p className="text-muted-foreground mb-4">
                                PKCE (Proof Key for Code Exchange) is <strong>required</strong> for all OAuth2 flows for security.
                            </p>
                            
                            <div className="relative bg-slate-950 rounded-lg p-4 mb-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                    onClick={() => copyToClipboard(
                                        `// Generate code verifier (random string)
const codeVerifier = base64URLEncode(crypto.getRandomValues(new Uint8Array(32)));

// Generate code challenge (SHA-256 hash of verifier)
const encoder = new TextEncoder();
const data = encoder.encode(codeVerifier);
const hash = await crypto.subtle.digest('SHA-256', data);
const codeChallenge = base64URLEncode(new Uint8Array(hash));

// Store code verifier for later use
sessionStorage.setItem('pkce_code_verifier', codeVerifier);`,
                                        'pkce-gen'
                                    )}
                                >
                                    {copiedSection === 'pkce-gen' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                </Button>
                                <pre className="text-sm text-white overflow-x-auto">
{`// Generate code verifier (random string)
const codeVerifier = base64URLEncode(crypto.getRandomValues(new Uint8Array(32)));

// Generate code challenge (SHA-256 hash of verifier)
const encoder = new TextEncoder();
const data = encoder.encode(codeVerifier);
const hash = await crypto.subtle.digest('SHA-256', data);
const codeChallenge = base64URLEncode(new Uint8Array(hash));

// Store code verifier for later use
sessionStorage.setItem('pkce_code_verifier', codeVerifier);`}
                                </pre>
                            </div>
                        </div>

                        {/* Step 3: Authorization Flow */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Step 3: Authorization Request</h3>
                            <p className="text-muted-foreground mb-4">Redirect users to the authorization endpoint with PKCE:</p>
                            
                            <div className="relative bg-slate-950 rounded-lg p-4 mb-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                    onClick={() => copyToClipboard(
                                        `http://localhost:8080/api/v1/oauth/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=YOUR_REDIRECT_URI&response_type=code&scope=read:profile read:organization&state=RANDOM_STATE&code_challenge=CODE_CHALLENGE&code_challenge_method=S256`,
                                        'auth-url'
                                    )}
                                >
                                    {copiedSection === 'auth-url' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                </Button>
                                <pre className="text-sm text-white overflow-x-auto">
{`http://localhost:8080/api/v1/oauth/authorize?
  client_id=YOUR_CLIENT_ID&
  redirect_uri=YOUR_REDIRECT_URI&
  response_type=code&
  scope=read:profile read:organization&
  state=RANDOM_STATE&
  code_challenge=CODE_CHALLENGE&
  code_challenge_method=S256`}
                                </pre>
                            </div>

                            <Card className="bg-primary/10 border-primary/30 p-4 mb-4">
                                <p className="text-sm font-semibold text-foreground mb-3">
                                    Required Parameters:
                                </p>
                                <div className="space-y-2">
                                    <div className="flex items-start gap-2">
                                        <code className="bg-primary/20 px-2 py-0.5 rounded text-sm font-mono text-foreground whitespace-nowrap">client_id</code>
                                        <span className="text-sm text-muted-foreground">- Your OAuth2 Client ID</span>
                                    </div>
                                    <div className="flex items-start gap-2">
                                        <code className="bg-primary/20 px-2 py-0.5 rounded text-sm font-mono text-foreground whitespace-nowrap">redirect_uri</code>
                                        <span className="text-sm text-muted-foreground">- One of your registered redirect URIs (must match exactly)</span>
                                    </div>
                                    <div className="flex items-start gap-2">
                                        <code className="bg-primary/20 px-2 py-0.5 rounded text-sm font-mono text-foreground whitespace-nowrap">response_type</code>
                                        <span className="text-sm text-muted-foreground">- Must be "code"</span>
                                    </div>
                                    <div className="flex items-start gap-2">
                                        <code className="bg-primary/20 px-2 py-0.5 rounded text-sm font-mono text-foreground whitespace-nowrap">code_challenge</code>
                                        <span className="text-sm text-muted-foreground">- SHA-256 hash of code_verifier (PKCE)</span>
                                    </div>
                                    <div className="flex items-start gap-2">
                                        <code className="bg-primary/20 px-2 py-0.5 rounded text-sm font-mono text-foreground whitespace-nowrap">code_challenge_method</code>
                                        <span className="text-sm text-muted-foreground">- Must be "S256"</span>
                                    </div>
                                    <div className="flex items-start gap-2">
                                        <code className="bg-primary/20 px-2 py-0.5 rounded text-sm font-mono text-foreground whitespace-nowrap">state</code>
                                        <span className="text-sm text-muted-foreground">- Random string for CSRF protection (optional but recommended)</span>
                                    </div>
                                    <div className="flex items-start gap-2">
                                        <code className="bg-primary/20 px-2 py-0.5 rounded text-sm font-mono text-foreground whitespace-nowrap">scope</code>
                                        <span className="text-sm text-muted-foreground">- Space-separated list of permissions (optional)</span>
                                    </div>
                                </div>
                            </Card>
                        </div>

                        {/* Step 4: Exchange Code for Token */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Step 4: Exchange Authorization Code for Access Token</h3>
                            <p className="text-muted-foreground mb-4">After the user authorizes, exchange the code for an access token:</p>
                            
                            <p className="text-sm font-medium text-foreground mb-2">For Public Clients (SPAs - NO client_secret):</p>
                            <div className="relative bg-slate-950 rounded-lg p-4 mb-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                    onClick={() => copyToClipboard(
                                        `curl -X POST http://localhost:8080/api/v1/oauth/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "client_id=YOUR_CLIENT_ID" \\
  -d "redirect_uri=YOUR_REDIRECT_URI" \\
  -d "code_verifier=YOUR_CODE_VERIFIER"`,
                                        'token-exchange-public'
                                    )}
                                >
                                    {copiedSection === 'token-exchange-public' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                </Button>
                                <pre className="text-sm text-white overflow-x-auto">
{`curl -X POST http://localhost:8080/api/v1/oauth/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "client_id=YOUR_CLIENT_ID" \\
  -d "redirect_uri=YOUR_REDIRECT_URI" \\
  -d "code_verifier=YOUR_CODE_VERIFIER"`}
                                </pre>
                            </div>

                            <p className="text-sm font-medium text-foreground mb-2">For Confidential Clients (Server-side - WITH client_secret):</p>
                            <div className="relative bg-slate-950 rounded-lg p-4 mb-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                    onClick={() => copyToClipboard(
                                        `curl -X POST http://localhost:8080/api/v1/oauth/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "client_id=YOUR_CLIENT_ID" \\
  -d "client_secret=YOUR_CLIENT_SECRET" \\
  -d "redirect_uri=YOUR_REDIRECT_URI" \\
  -d "code_verifier=YOUR_CODE_VERIFIER"`,
                                        'token-exchange-confidential'
                                    )}
                                >
                                    {copiedSection === 'token-exchange-confidential' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                </Button>
                                <pre className="text-sm text-white overflow-x-auto">
{`curl -X POST http://localhost:8080/api/v1/oauth/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "client_id=YOUR_CLIENT_ID" \\
  -d "client_secret=YOUR_CLIENT_SECRET" \\
  -d "redirect_uri=YOUR_REDIRECT_URI" \\
  -d "code_verifier=YOUR_CODE_VERIFIER"`}
                                </pre>
                            </div>

                            <p className="text-muted-foreground mb-2">Response:</p>
                            <div className="relative bg-slate-950 rounded-lg p-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                    onClick={() => copyToClipboard(
                                        `{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "def50200..."
}`,
                                        'token-response'
                                    )}
                                >
                                    {copiedSection === 'token-response' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                </Button>
                                <pre className="text-sm text-white overflow-x-auto">
{`{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "def50200..."
}`}
                                </pre>
                            </div>
                        </div>

                        {/* Step 5: Use Access Token */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Step 5: Make Authenticated Requests</h3>
                            <p className="text-muted-foreground mb-4">Include the access token in your API requests:</p>
                            
                            <div className="relative bg-slate-950 rounded-lg p-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                    onClick={() => copyToClipboard(
                                        `curl -X GET https://api.yourdomain.com/v1/user/profile \\
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"`,
                                        'api-request'
                                    )}
                                >
                                    {copiedSection === 'api-request' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                </Button>
                                <pre className="text-sm text-white overflow-x-auto">
{`curl -X GET https://api.yourdomain.com/v1/user/profile \\
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"`}
                                </pre>
                            </div>
                        </div>

                        {/* Refresh Token */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Step 6: Refresh Expired Tokens</h3>
                            <p className="text-muted-foreground mb-4">When the access token expires (1 hour), use the refresh token:</p>
                            
                            <div className="relative bg-slate-950 rounded-lg p-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                    onClick={() => copyToClipboard(
                                        `curl -X POST http://localhost:8080/api/v1/oauth/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=refresh_token" \\
  -d "refresh_token=YOUR_REFRESH_TOKEN" \\
  -d "client_id=YOUR_CLIENT_ID"`,
                                        'refresh-token'
                                    )}
                                >
                                    {copiedSection === 'refresh-token' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                </Button>
                                <pre className="text-sm text-white overflow-x-auto">
{`curl -X POST http://localhost:8080/api/v1/oauth/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=refresh_token" \\
  -d "refresh_token=YOUR_REFRESH_TOKEN" \\
  -d "client_id=YOUR_CLIENT_ID"`}
                                </pre>
                            </div>
                            <p className="text-xs text-muted-foreground mt-2">
                                Note: Client secret is not required for refresh token grant, even for confidential clients.
                            </p>
                        </div>
                    </div>
                </Card>
            </div>

            {/* API Keys Integration Section */}
            <div id="api-keys">
                <Card className="p-8">
                    <div className="flex items-center gap-3 mb-6">
                        <Key className="h-8 w-8 text-primary" />
                        <h2 className="text-2xl font-bold text-foreground">API Key Integration</h2>
                    </div>

                    <div className="space-y-6">
                        {/* Overview */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-2">Overview</h3>
                            <p className="text-muted-foreground">
                                API Keys provide a simple way to authenticate server-to-server requests without user interaction.
                                Best for backend services, scripts, and automated workflows.
                            </p>
                        </div>

                        {/* Step 1: Create API Key */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Step 1: Create an API Key</h3>
                            <ol className="list-decimal list-inside space-y-2 text-muted-foreground ml-4">
                                <li>Navigate to <a href="/user/api-keys" className="text-primary hover:underline">API Keys</a></li>
                                <li>Click "Create API Key"</li>
                                <li>Enter a descriptive name</li>
                                <li>Set expiration (optional, recommended for security)</li>
                                <li>Copy and save the secret immediately - it won't be shown again</li>
                            </ol>
                        </div>

                        {/* Step 2: Use API Key */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Step 2: Authenticate Your Requests</h3>
                            <p className="text-muted-foreground mb-4">Include the API key in the Authorization header:</p>
                            
                            <div className="relative bg-slate-950 rounded-lg p-4 mb-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                    onClick={() => copyToClipboard(
                                        `curl -X GET https://api.yourdomain.com/v1/organizations \\
  -H "Authorization: Bearer YOUR_API_KEY"`,
                                        'apikey-request'
                                    )}
                                >
                                    {copiedSection === 'apikey-request' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                </Button>
                                <pre className="text-sm text-white overflow-x-auto">
{`curl -X GET https://api.yourdomain.com/v1/organizations \\
  -H "Authorization: Bearer YOUR_API_KEY"`}
                                </pre>
                            </div>

                            <Card className="bg-destructive/10 border-destructive/30 p-4">
                                <p className="text-sm font-semibold text-foreground mb-3 flex items-center gap-2">
                                    <AlertTriangle className="h-4 w-4 text-destructive" />
                                    Security Best Practices
                                </p>
                                <ul className="text-sm text-muted-foreground space-y-2 ml-4 list-disc">
                                    <li>Never commit API keys to version control</li>
                                    <li>Store keys in environment variables or secure vaults</li>
                                    <li>Set expiration dates for keys</li>
                                    <li>Rotate keys regularly</li>
                                    <li>Revoke unused or compromised keys immediately</li>
                                    <li>Use different keys for different environments (dev, staging, prod)</li>
                                </ul>
                            </Card>
                        </div>

                        {/* Code Examples */}
                        <div>
                            <h3 className="text-lg font-semibold text-foreground mb-3">Code Examples</h3>
                            
                            {/* JavaScript/Node.js */}
                            <div className="mb-4">
                                <p className="text-sm font-medium text-foreground mb-2">JavaScript / Node.js</p>
                                <div className="relative bg-slate-950 rounded-lg p-4">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                        onClick={() => copyToClipboard(
                                            `const API_KEY = process.env.BLOCKSURE_API_KEY;

const response = await fetch('https://api.yourdomain.com/v1/organizations', {
  headers: {
    'Authorization': \`Bearer \${API_KEY}\`,
    'Content-Type': 'application/json'
  }
});

const data = await response.json();
console.log(data);`,
                                            'js-example'
                                        )}
                                    >
                                        {copiedSection === 'js-example' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                    </Button>
                                    <pre className="text-sm text-white overflow-x-auto">
{`const API_KEY = process.env.BLOCKSURE_API_KEY;

const response = await fetch('https://api.yourdomain.com/v1/organizations', {
  headers: {
    'Authorization': \`Bearer \${API_KEY}\`,
    'Content-Type': 'application/json'
  }
});

const data = await response.json();
console.log(data);`}
                                    </pre>
                                </div>
                            </div>

                            {/* Python */}
                            <div className="mb-4">
                                <p className="text-sm font-medium text-foreground mb-2">Python</p>
                                <div className="relative bg-slate-950 rounded-lg p-4">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                        onClick={() => copyToClipboard(
                                            `import os
import requests

API_KEY = os.environ.get('BLOCKSURE_API_KEY')

response = requests.get(
    'https://api.yourdomain.com/v1/organizations',
    headers={
        'Authorization': f'Bearer {API_KEY}',
        'Content-Type': 'application/json'
    }
)

data = response.json()
print(data)`,
                                            'python-example'
                                        )}
                                    >
                                        {copiedSection === 'python-example' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                    </Button>
                                    <pre className="text-sm text-white overflow-x-auto">
{`import os
import requests

API_KEY = os.environ.get('BLOCKSURE_API_KEY')

response = requests.get(
    'https://api.yourdomain.com/v1/organizations',
    headers={
        'Authorization': f'Bearer {API_KEY}',
        'Content-Type': 'application/json'
    }
)

data = response.json()
print(data)`}
                                    </pre>
                                </div>
                            </div>

                            {/* Go */}
                            <div>
                                <p className="text-sm font-medium text-foreground mb-2">Go</p>
                                <div className="relative bg-slate-950 rounded-lg p-4">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        className="absolute top-2 right-2 text-white hover:bg-slate-800"
                                        onClick={() => copyToClipboard(
                                            `package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
)

func main() {
    apiKey := os.Getenv("BLOCKSURE_API_KEY")
    
    req, _ := http.NewRequest("GET", "https://api.yourdomain.com/v1/organizations", nil)
    req.Header.Set("Authorization", "Bearer " + apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}`,
                                            'go-example'
                                        )}
                                    >
                                        {copiedSection === 'go-example' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                    </Button>
                                    <pre className="text-sm text-white overflow-x-auto">
{`package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
)

func main() {
    apiKey := os.Getenv("BLOCKSURE_API_KEY")
    
    req, _ := http.NewRequest("GET", "https://api.yourdomain.com/v1/organizations", nil)
    req.Header.Set("Authorization", "Bearer " + apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}`}
                                    </pre>
                                </div>
                            </div>
                        </div>
                    </div>
                </Card>
            </div>

            {/* Comparison Table */}
            <Card className="p-8 mt-12">
                <h2 className="text-2xl font-bold text-foreground mb-6">OAuth2 vs API Keys</h2>
                <div className="overflow-x-auto">
                    <table className="w-full text-left">
                        <thead>
                            <tr className="border-b-2 border-border">
                                <th className="pb-3 pr-4 text-foreground font-semibold text-sm">Feature</th>
                                <th className="pb-3 px-4 text-foreground font-semibold text-sm">OAuth2</th>
                                <th className="pb-3 pl-4 text-foreground font-semibold text-sm">API Keys</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr className="border-b border-border">
                                <td className="py-4 pr-4 font-medium text-foreground text-sm">Use Case</td>
                                <td className="py-4 px-4 text-muted-foreground text-sm">User-facing apps, third-party integrations</td>
                                <td className="py-4 pl-4 text-muted-foreground text-sm">Server-to-server, automated scripts</td>
                            </tr>
                            <tr className="border-b border-border">
                                <td className="py-4 pr-4 font-medium text-foreground text-sm">User Interaction</td>
                                <td className="py-4 px-4 text-muted-foreground text-sm">Required (authorization flow)</td>
                                <td className="py-4 pl-4 text-muted-foreground text-sm">Not required</td>
                            </tr>
                            <tr className="border-b border-border">
                                <td className="py-4 pr-4 font-medium text-foreground text-sm">Token Expiration</td>
                                <td className="py-4 px-4 text-muted-foreground text-sm">Short-lived (1 hour default)</td>
                                <td className="py-4 pl-4 text-muted-foreground text-sm">Long-lived or never expires</td>
                            </tr>
                            <tr className="border-b border-border">
                                <td className="py-4 pr-4 font-medium text-foreground text-sm">Refresh Capability</td>
                                <td className="py-4 px-4 text-muted-foreground text-sm">Yes (refresh tokens)</td>
                                <td className="py-4 pl-4 text-muted-foreground text-sm">No (create new key)</td>
                            </tr>
                            <tr className="border-b border-border">
                                <td className="py-4 pr-4 font-medium text-foreground text-sm">Granular Scopes</td>
                                <td className="py-4 px-4 text-muted-foreground text-sm">Yes</td>
                                <td className="py-4 pl-4 text-muted-foreground text-sm">Limited</td>
                            </tr>
                            <tr>
                                <td className="py-4 pr-4 font-medium text-foreground text-sm">Best For</td>
                                <td className="py-4 px-4 text-muted-foreground text-sm">Web/mobile apps, delegated access</td>
                                <td className="py-4 pl-4 text-muted-foreground text-sm">Backend services, CLI tools, cron jobs</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </Card>

            {/* Support Section */}
            <Card className="p-8 mt-8 bg-primary/5 border-primary/20">
                <h2 className="text-xl font-bold text-foreground mb-3">Need Help?</h2>
                <p className="text-muted-foreground mb-4">
                    If you encounter any issues or have questions about integrating with our platform:
                </p>
                <div className="flex gap-4">
                    <Button variant="outline">
                        <Code className="h-4 w-4 mr-2" />
                        View API Reference
                    </Button>
                    <Button variant="outline">
                        Contact Support
                    </Button>
                </div>
            </Card>
        </div>
    );
}
