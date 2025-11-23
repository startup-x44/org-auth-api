'use client';

import { useState } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Key, Send, CheckCircle, Copy, AlertCircle, Code } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

/**
 * API Key Demo
 * 
 * This demonstrates how to use API Keys for server-to-server authentication.
 * API Keys are simpler than OAuth2 and ideal for:
 * - Backend services
 * - CLI tools
 * - Automated scripts
 * - Cron jobs
 * - Server-to-server communication
 */

interface ApiResponse {
    success: boolean;
    data?: any;
    message?: string;
    error?: string;
}

export default function ApiKeyDemo() {
    const [apiKey, setApiKey] = useState('');
    const [endpoint, setEndpoint] = useState('/api/v1/user/profile');
    const [method, setMethod] = useState<'GET' | 'POST' | 'PUT' | 'DELETE'>('GET');
    const [requestBody, setRequestBody] = useState('');
    const [response, setResponse] = useState<ApiResponse | null>(null);
    const [loading, setLoading] = useState(false);
    const [responseTime, setResponseTime] = useState<number | null>(null);
    const { toast } = useToast();

    const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    // Example endpoints
    const exampleEndpoints = [
        { method: 'GET', path: '/api/v1/user/profile', description: 'Get your user profile' },
        { method: 'GET', path: '/api/v1/organizations', description: 'List organizations' },
        { method: 'GET', path: '/api/v1/dev/api-keys', description: 'List your API keys' },
        { method: 'GET', path: '/api/v1/dev/client-apps', description: 'List OAuth2 apps' },
    ];

    const makeApiRequest = async () => {
        if (!apiKey.trim()) {
            toast({
                title: "Missing API Key",
                description: "Please enter your API key",
                variant: "destructive"
            });
            return;
        }

        if (!endpoint.trim()) {
            toast({
                title: "Missing Endpoint",
                description: "Please enter an API endpoint",
                variant: "destructive"
            });
            return;
        }

        setLoading(true);
        setResponse(null);
        setResponseTime(null);

        const startTime = performance.now();

        try {
            const options: RequestInit = {
                method,
                headers: {
                    'Authorization': `Bearer ${apiKey}`,
                    'Content-Type': 'application/json'
                }
            };

            // Add request body for POST/PUT
            if ((method === 'POST' || method === 'PUT') && requestBody.trim()) {
                try {
                    JSON.parse(requestBody); // Validate JSON
                    options.body = requestBody;
                } catch (e) {
                    throw new Error('Invalid JSON in request body');
                }
            }

            const res = await fetch(`${API_BASE}${endpoint}`, options);
            const endTime = performance.now();
            setResponseTime(Math.round(endTime - startTime));

            const data = await res.json();

            if (res.ok) {
                setResponse({
                    success: true,
                    data: data.data || data,
                    message: data.message
                });
                toast({
                    title: "Request Successful",
                    description: `API returned ${res.status} ${res.statusText}`,
                });
            } else {
                setResponse({
                    success: false,
                    error: data.message || data.error || 'Request failed',
                    data: data
                });
                toast({
                    title: "Request Failed",
                    description: data.message || data.error || 'API request failed',
                    variant: "destructive"
                });
            }
        } catch (error: any) {
            const endTime = performance.now();
            setResponseTime(Math.round(endTime - startTime));
            
            setResponse({
                success: false,
                error: error.message || 'Network error'
            });
            toast({
                title: "Error",
                description: error.message || "Failed to make API request",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    const copyToClipboard = (text: string, label: string) => {
        navigator.clipboard.writeText(text);
        toast({
            title: "Copied!",
            description: `${label} copied to clipboard.`,
        });
    };

    const generateCurlCommand = () => {
        let curl = `curl -X ${method} ${API_BASE}${endpoint} \\\n  -H "Authorization: Bearer YOUR_API_KEY" \\\n  -H "Content-Type: application/json"`;
        
        if ((method === 'POST' || method === 'PUT') && requestBody.trim()) {
            curl += ` \\\n  -d '${requestBody}'`;
        }
        
        return curl;
    };

    const setExampleEndpoint = (exampleMethod: string, examplePath: string) => {
        setMethod(exampleMethod as 'GET' | 'POST' | 'PUT' | 'DELETE');
        setEndpoint(examplePath);
        setRequestBody('');
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-primary/5 via-background to-primary/10 p-6">
            <div className="max-w-7xl mx-auto space-y-6">
                {/* Header */}
                <div className="text-center space-y-3">
                    <div className="flex items-center justify-center gap-3">
                        <div className="p-3 bg-primary/10 rounded-xl">
                            <Key className="h-8 w-8 text-primary" />
                        </div>
                    </div>
                    <h1 className="text-4xl font-bold bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
                        API Key Demo
                    </h1>
                    <p className="text-muted-foreground max-w-2xl mx-auto">
                        Test your API keys and explore the API. Perfect for backend services, scripts, and automation.
                    </p>
                </div>

                {/* Info Cards */}
                <div className="grid md:grid-cols-3 gap-4">
                    <Card className="p-4 border-primary/30 bg-primary/5">
                        <div className="flex items-center gap-2 mb-2">
                            <div className="p-2 bg-primary/10 rounded-lg">
                                <Key className="h-4 w-4 text-primary" />
                            </div>
                            <h3 className="font-semibold text-foreground">Simple Auth</h3>
                        </div>
                        <p className="text-sm text-muted-foreground">
                            No OAuth flow required. Just include your API key in the Authorization header.
                        </p>
                    </Card>
                    
                    <Card className="p-4 border-primary/30 bg-primary/5">
                        <div className="flex items-center gap-2 mb-2">
                            <div className="p-2 bg-primary/10 rounded-lg">
                                <CheckCircle className="h-4 w-4 text-primary" />
                            </div>
                            <h3 className="font-semibold text-foreground">Long-lived</h3>
                        </div>
                        <p className="text-sm text-muted-foreground">
                            API keys don't expire unless you set an expiration date or rotate them.
                        </p>
                    </Card>
                    
                    <Card className="p-4 border-primary/30 bg-primary/5">
                        <div className="flex items-center gap-2 mb-2">
                            <div className="p-2 bg-primary/10 rounded-lg">
                                <Code className="h-4 w-4 text-primary" />
                            </div>
                            <h3 className="font-semibold text-foreground">Server-to-Server</h3>
                        </div>
                        <p className="text-sm text-muted-foreground">
                            Ideal for backend services, cron jobs, CLI tools, and automated scripts.
                        </p>
                    </Card>
                </div>

                {/* Main Content */}
                <div className="grid lg:grid-cols-2 gap-6">
                    {/* Request Panel */}
                    <Card className="p-6">
                        <h2 className="text-2xl font-bold text-foreground mb-6 flex items-center gap-2">
                            <Send className="h-6 w-6 text-primary" />
                            API Request
                        </h2>

                        <div className="space-y-4">
                            {/* API Key Input */}
                            <div>
                                <label className="block text-sm font-medium text-foreground mb-2">
                                    API Secret Key *
                                </label>
                                <Input
                                    type="password"
                                    value={apiKey}
                                    onChange={(e) => setApiKey(e.target.value)}
                                    placeholder="bsk_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
                                    className="font-mono"
                                />
                                <Card className="bg-amber-500/10 border-amber-500/30 p-3 mt-2">
                                    <div className="flex items-start gap-2">
                                        <AlertCircle className="h-4 w-4 text-amber-600 mt-0.5 flex-shrink-0" />
                                        <div className="text-xs text-foreground">
                                            <strong>Use the SECRET key</strong> (shown only once during creation), NOT the public Key ID.
                                            <br />
                                            Get your API key from <a href="/user/api-keys" className="text-primary hover:underline font-semibold">API Keys page</a>
                                        </div>
                                    </div>
                                </Card>
                            </div>

                            {/* HTTP Method */}
                            <div>
                                <label className="block text-sm font-medium text-foreground mb-2">
                                    HTTP Method
                                </label>
                                <div className="flex gap-2">
                                    {(['GET', 'POST', 'PUT', 'DELETE'] as const).map((m) => (
                                        <Button
                                            key={m}
                                            variant={method === m ? 'default' : 'outline'}
                                            size="sm"
                                            onClick={() => setMethod(m)}
                                            className="flex-1"
                                        >
                                            {m}
                                        </Button>
                                    ))}
                                </div>
                            </div>

                            {/* Endpoint Input */}
                            <div>
                                <label className="block text-sm font-medium text-foreground mb-2">
                                    API Endpoint *
                                </label>
                                <Input
                                    value={endpoint}
                                    onChange={(e) => setEndpoint(e.target.value)}
                                    placeholder="/api/v1/user/profile"
                                    className="font-mono"
                                />
                            </div>

                            {/* Request Body (for POST/PUT) */}
                            {(method === 'POST' || method === 'PUT') && (
                                <div>
                                    <label className="block text-sm font-medium text-foreground mb-2">
                                        Request Body (JSON)
                                    </label>
                                    <textarea
                                        value={requestBody}
                                        onChange={(e) => setRequestBody(e.target.value)}
                                        placeholder='{"key": "value"}'
                                        className="w-full h-24 px-3 py-2 text-sm font-mono border rounded-md bg-background"
                                    />
                                </div>
                            )}

                            {/* Example Endpoints */}
                            <div>
                                <label className="block text-sm font-medium text-foreground mb-2">
                                    Quick Examples
                                </label>
                                <div className="space-y-2">
                                    {exampleEndpoints.map((example, idx) => (
                                        <button
                                            key={idx}
                                            onClick={() => setExampleEndpoint(example.method, example.path)}
                                            className="w-full text-left p-3 border rounded-lg hover:border-primary/50 hover:bg-primary/5 transition-all"
                                        >
                                            <div className="flex items-center gap-2 mb-1">
                                                <Badge variant="outline" className="text-xs">
                                                    {example.method}
                                                </Badge>
                                                <code className="text-xs font-mono text-foreground">
                                                    {example.path}
                                                </code>
                                            </div>
                                            <p className="text-xs text-muted-foreground">
                                                {example.description}
                                            </p>
                                        </button>
                                    ))}
                                </div>
                            </div>

                            {/* Send Button */}
                            <Button
                                onClick={makeApiRequest}
                                disabled={loading || !apiKey || !endpoint}
                                className="w-full"
                                size="lg"
                            >
                                <Send className="h-4 w-4 mr-2" />
                                {loading ? 'Sending...' : 'Send Request'}
                            </Button>

                            {/* cURL Command */}
                            <div>
                                <div className="flex items-center justify-between mb-2">
                                    <label className="text-sm font-medium text-foreground">
                                        cURL Command
                                    </label>
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => copyToClipboard(generateCurlCommand(), 'cURL command')}
                                    >
                                        <Copy className="h-4 w-4" />
                                    </Button>
                                </div>
                                <pre className="text-xs bg-slate-950 text-white p-3 rounded-lg overflow-x-auto">
                                    {generateCurlCommand()}
                                </pre>
                            </div>
                        </div>
                    </Card>

                    {/* Response Panel */}
                    <Card className="p-6">
                        <h2 className="text-2xl font-bold text-foreground mb-6 flex items-center gap-2">
                            {response?.success ? (
                                <CheckCircle className="h-6 w-6 text-green-600" />
                            ) : response ? (
                                <AlertCircle className="h-6 w-6 text-red-600" />
                            ) : (
                                <Code className="h-6 w-6 text-primary" />
                            )}
                            Response
                        </h2>

                        {!response ? (
                            <div className="text-center py-12 text-muted-foreground">
                                <Code className="h-12 w-12 mx-auto mb-4 opacity-50" />
                                <p>Send a request to see the response</p>
                            </div>
                        ) : (
                            <div className="space-y-4">
                                {/* Status Badge */}
                                <div className="flex items-center gap-2">
                                    <Badge
                                        variant={response.success ? 'default' : 'destructive'}
                                        className="text-sm"
                                    >
                                        {response.success ? 'Success' : 'Error'}
                                    </Badge>
                                    {responseTime && (
                                        <Badge variant="outline" className="text-sm">
                                            {responseTime}ms
                                        </Badge>
                                    )}
                                </div>

                                {/* Error Message */}
                                {response.error && (
                                    <Card className="p-4 bg-destructive/10 border-destructive/30">
                                        <p className="text-sm font-semibold text-destructive mb-1">
                                            Error
                                        </p>
                                        <p className="text-sm text-foreground">
                                            {response.error}
                                        </p>
                                    </Card>
                                )}

                                {/* Success Message */}
                                {response.success && response.message && (
                                    <Card className="p-4 bg-green-500/10 border-green-500/30">
                                        <p className="text-sm text-foreground">
                                            {response.message}
                                        </p>
                                    </Card>
                                )}

                                {/* Response Data */}
                                <div>
                                    <div className="flex items-center justify-between mb-2">
                                        <label className="text-sm font-medium text-foreground">
                                            Response Body
                                        </label>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => copyToClipboard(JSON.stringify(response.data, null, 2), 'Response')}
                                        >
                                            <Copy className="h-4 w-4" />
                                        </Button>
                                    </div>
                                    <pre className="text-xs bg-slate-950 text-white p-4 rounded-lg overflow-x-auto max-h-96">
                                        {JSON.stringify(response.data || response, null, 2)}
                                    </pre>
                                </div>
                            </div>
                        )}
                    </Card>
                </div>

                {/* Code Examples Section */}
                <Card className="p-6">
                    <h2 className="text-2xl font-bold text-foreground mb-4">Integration Examples</h2>
                    <div className="grid md:grid-cols-2 gap-4">
                        {/* JavaScript Example */}
                        <div>
                            <div className="flex items-center justify-between mb-2">
                                <p className="text-sm font-medium text-foreground">JavaScript / Node.js</p>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => copyToClipboard(
                                        `const response = await fetch('${API_BASE}${endpoint}', {\n  method: '${method}',\n  headers: {\n    'Authorization': 'Bearer YOUR_API_KEY',\n    'Content-Type': 'application/json'\n  }\n});\n\nconst data = await response.json();`,
                                        'JavaScript code'
                                    )}
                                >
                                    <Copy className="h-4 w-4" />
                                </Button>
                            </div>
                            <pre className="text-xs bg-slate-950 text-white p-3 rounded-lg overflow-x-auto">
{`const response = await fetch('${API_BASE}${endpoint}', {
  method: '${method}',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  }
});

const data = await response.json();`}
                            </pre>
                        </div>

                        {/* Python Example */}
                        <div>
                            <div className="flex items-center justify-between mb-2">
                                <p className="text-sm font-medium text-foreground">Python</p>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => copyToClipboard(
                                        `import requests\n\nresponse = requests.${method.toLowerCase()}(\n    '${API_BASE}${endpoint}',\n    headers={\n        'Authorization': 'Bearer YOUR_API_KEY',\n        'Content-Type': 'application/json'\n    }\n)\n\ndata = response.json()`,
                                        'Python code'
                                    )}
                                >
                                    <Copy className="h-4 w-4" />
                                </Button>
                            </div>
                            <pre className="text-xs bg-slate-950 text-white p-3 rounded-lg overflow-x-auto">
{`import requests

response = requests.${method.toLowerCase()}(
    '${API_BASE}${endpoint}',
    headers={
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
)

data = response.json()`}
                            </pre>
                        </div>
                    </div>
                </Card>

                {/* Security Warning */}
                <Card className="p-6 bg-destructive/5 border-destructive/20">
                    <div className="flex items-start gap-3">
                        <AlertCircle className="h-5 w-5 text-destructive mt-0.5" />
                        <div>
                            <h3 className="font-semibold text-foreground mb-2">Security Best Practices</h3>
                            <ul className="text-sm text-muted-foreground space-y-1 list-disc ml-4">
                                <li>Never commit API keys to version control</li>
                                <li>Store keys in environment variables or secure vaults</li>
                                <li>Set expiration dates for API keys</li>
                                <li>Rotate keys regularly and revoke unused ones</li>
                                <li>Use different keys for different environments (dev, staging, prod)</li>
                                <li>Monitor API key usage and set up alerts for suspicious activity</li>
                            </ul>
                        </div>
                    </div>
                </Card>
            </div>
        </div>
    );
}
