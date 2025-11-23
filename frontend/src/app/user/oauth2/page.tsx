'use client';

import { useState, useEffect } from 'react';
import { AuthService } from '@/lib/auth';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Lock, Plus, Copy, Trash2, RefreshCw, Eye, EyeOff } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

interface OAuth2App {
    id: string;
    client_id: string;
    client_secret?: string;
    name: string;
    description: string;
    redirect_uris: string[];
    scopes: string[];
    is_confidential: boolean;
    organization_id: string;
    created_at: string;
    updated_at: string;
}

export default function OAuth2Page() {
    const [apps, setApps] = useState<OAuth2App[]>([]);
    const [loading, setLoading] = useState(true);
    const [creating, setCreating] = useState(false);
    const [showCreateForm, setShowCreateForm] = useState(false);
    const { toast } = useToast();

    // Form state
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        is_confidential: true
    });
    
    const [redirectUris, setRedirectUris] = useState<string[]>(['']);
    const [allowedScopes, setAllowedScopes] = useState<string[]>([]);

    // Visibility state for secrets
    const [visibleSecrets, setVisibleSecrets] = useState<Set<string>>(new Set());
    
    // New secret display (shown once after creation/rotation)
    const [newSecret, setNewSecret] = useState<{
        appName: string;
        clientId: string;
        clientSecret: string;
    } | null>(null);
    
    // Edit mode state
    const [editingApp, setEditingApp] = useState<string | null>(null);
    const [editFormData, setEditFormData] = useState<{
        redirect_uris: string[];
        allowed_scopes: string[];
        is_confidential: boolean;
    }>({
        redirect_uris: [],
        allowed_scopes: [],
        is_confidential: true
    });

    useEffect(() => {
        loadOAuth2Apps();
    }, []);

    const loadOAuth2Apps = async () => {
        try {
            setLoading(true);
            const response = await AuthService.apiCall('/api/v1/dev/client-apps');
            
            if (response.ok) {
                const data = await response.json();
                setApps(data.data || data.client_apps || []);
            } else {
                toast({
                    title: "Failed to load OAuth2 apps",
                    description: "Could not retrieve your OAuth2 applications.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to load OAuth2 apps:', error);
            toast({
                title: "Error",
                description: "An unexpected error occurred.",
                variant: "destructive",
            });
        } finally {
            setLoading(false);
        }
    };

    const addRedirectUri = () => {
        setRedirectUris([...redirectUris, '']);
    };

    const removeRedirectUri = (index: number) => {
        if (redirectUris.length > 1) {
            setRedirectUris(redirectUris.filter((_, i) => i !== index));
        }
    };

    const updateRedirectUri = (index: number, value: string) => {
        const updated = [...redirectUris];
        updated[index] = value;
        setRedirectUris(updated);
    };

    const addScope = () => {
        setAllowedScopes([...allowedScopes, '']);
    };

    const removeScope = (index: number) => {
        setAllowedScopes(allowedScopes.filter((_, i) => i !== index));
    };

    const updateScope = (index: number, value: string) => {
        const updated = [...allowedScopes];
        updated[index] = value;
        setAllowedScopes(updated);
    };

    const handleCreateApp = async () => {
        if (!formData.name.trim()) {
            toast({
                title: "Missing information",
                description: "Please provide an application name.",
                variant: "destructive",
            });
            return;
        }

        try {
            setCreating(true);

            const validRedirectUris = redirectUris
                .map(uri => uri.trim())
                .filter(uri => uri.length > 0);

            const validScopes = allowedScopes
                .map(scope => scope.trim())
                .filter(scope => scope.length > 0);

            if (validRedirectUris.length === 0) {
                toast({
                    title: "Missing redirect URIs",
                    description: "Please provide at least one redirect URI.",
                    variant: "destructive",
                });
                return;
            }

            const response = await AuthService.apiCall('/api/v1/dev/client-apps', {
                method: 'POST',
                body: JSON.stringify({
                    name: formData.name.trim(),
                    description: formData.description.trim(),
                    redirect_uris: validRedirectUris,
                    allowed_scopes: validScopes.length > 0 ? validScopes : undefined,
                    is_confidential: formData.is_confidential
                }),
            });

            if (response.ok) {
                const data = await response.json();
                
                // Show the client secret (only shown once!)
                if (data.client_secret && data.data) {
                    setNewSecret({
                        appName: formData.name,
                        clientId: data.data.client_id,
                        clientSecret: data.client_secret
                    });
                }
                
                toast({
                    title: "OAuth2 app created",
                    description: `${formData.name} has been created successfully.`,
                });

                // Reset form
                setFormData({
                    name: '',
                    description: '',
                    is_confidential: true
                });
                setRedirectUris(['']);
                setAllowedScopes([]);
                setShowCreateForm(false);

                // Reload apps
                await loadOAuth2Apps();
            } else {
                const errorData = await response.json();
                toast({
                    title: "Failed to create app",
                    description: errorData.message || "Could not create OAuth2 application.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to create OAuth2 app:', error);
            toast({
                title: "Error",
                description: "An unexpected error occurred.",
                variant: "destructive",
            });
        } finally {
            setCreating(false);
        }
    };

    const handleDeleteApp = async (appId: string, appName: string) => {
        if (!confirm(`Are you sure you want to delete "${appName}"? This action cannot be undone.`)) {
            return;
        }

        try {
            const response = await AuthService.apiCall(`/api/v1/dev/client-apps/${appId}`, {
                method: 'DELETE',
            });

            if (response.ok) {
                toast({
                    title: "App deleted",
                    description: `${appName} has been deleted.`,
                });
                await loadOAuth2Apps();
            } else {
                toast({
                    title: "Failed to delete app",
                    description: "Could not delete the application.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to delete app:', error);
            toast({
                title: "Error",
                description: "An unexpected error occurred.",
                variant: "destructive",
            });
        }
    };

    const handleRotateSecret = async (appId: string, appName: string) => {
        if (!confirm(`Rotate client secret for "${appName}"? The old secret will stop working immediately.`)) {
            return;
        }

        try {
            const response = await AuthService.apiCall(`/api/v1/dev/client-apps/${appId}/rotate-secret`, {
                method: 'POST',
            });

            if (response.ok) {
                const data = await response.json();
                
                // Show the new client secret (only shown once!)
                if (data.client_secret) {
                    const app = apps.find(a => a.id === appId);
                    setNewSecret({
                        appName: appName,
                        clientId: app?.client_id || '',
                        clientSecret: data.client_secret
                    });
                }
                
                toast({
                    title: "Secret rotated",
                    description: `Client secret for ${appName} has been rotated.`,
                });
                await loadOAuth2Apps();
            } else {
                toast({
                    title: "Failed to rotate secret",
                    description: "Could not rotate the client secret.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to rotate secret:', error);
            toast({
                title: "Error",
                description: "An unexpected error occurred.",
                variant: "destructive",
            });
        }
    };

    const copyToClipboard = (text: string, label: string) => {
        navigator.clipboard.writeText(text);
        toast({
            title: "Copied!",
            description: `${label} copied to clipboard.`,
        });
    };

    const toggleSecretVisibility = (appId: string) => {
        setVisibleSecrets(prev => {
            const newSet = new Set(prev);
            if (newSet.has(appId)) {
                newSet.delete(appId);
            } else {
                newSet.add(appId);
            }
            return newSet;
        });
    };

    const startEditingApp = (app: OAuth2App) => {
        setEditingApp(app.id);
        setEditFormData({
            redirect_uris: [...app.redirect_uris],
            allowed_scopes: app.scopes || [],
            is_confidential: app.is_confidential
        });
    };

    const cancelEditing = () => {
        setEditingApp(null);
        setEditFormData({
            redirect_uris: [],
            allowed_scopes: [],
            is_confidential: true
        });
    };

    const addEditRedirectUri = () => {
        setEditFormData(prev => ({
            ...prev,
            redirect_uris: [...prev.redirect_uris, '']
        }));
    };

    const removeEditRedirectUri = (index: number) => {
        if (editFormData.redirect_uris.length > 1) {
            setEditFormData(prev => ({
                ...prev,
                redirect_uris: prev.redirect_uris.filter((_, i) => i !== index)
            }));
        }
    };

    const updateEditRedirectUri = (index: number, value: string) => {
        setEditFormData(prev => ({
            ...prev,
            redirect_uris: prev.redirect_uris.map((uri, i) => i === index ? value : uri)
        }));
    };

    const addEditScope = () => {
        setEditFormData(prev => ({
            ...prev,
            allowed_scopes: [...prev.allowed_scopes, '']
        }));
    };

    const removeEditScope = (index: number) => {
        setEditFormData(prev => ({
            ...prev,
            allowed_scopes: prev.allowed_scopes.filter((_, i) => i !== index)
        }));
    };

    const updateEditScope = (index: number, value: string) => {
        setEditFormData(prev => ({
            ...prev,
            allowed_scopes: prev.allowed_scopes.map((scope, i) => i === index ? value : scope)
        }));
    };

    const handleUpdateApp = async (appId: string, appName: string) => {
        const validRedirectUris = editFormData.redirect_uris
            .map(uri => uri.trim())
            .filter(uri => uri.length > 0);

        const validScopes = editFormData.allowed_scopes
            .map(scope => scope.trim())
            .filter(scope => scope.length > 0);

        if (validRedirectUris.length === 0) {
            toast({
                title: "Missing redirect URIs",
                description: "Please provide at least one redirect URI.",
                variant: "destructive",
            });
            return;
        }

        try {
            const response = await AuthService.apiCall(`/api/v1/dev/client-apps/${appId}`, {
                method: 'PUT',
                body: JSON.stringify({
                    redirect_uris: validRedirectUris,
                    allowed_scopes: validScopes.length > 0 ? validScopes : undefined,
                    is_confidential: editFormData.is_confidential
                }),
            });

            if (response.ok) {
                toast({
                    title: "App updated",
                    description: `${appName} has been updated successfully.`,
                });
                cancelEditing();
                await loadOAuth2Apps();
            } else {
                const errorData = await response.json();
                toast({
                    title: "Failed to update app",
                    description: errorData.message || "Could not update application.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to update app:', error);
            toast({
                title: "Error",
                description: "An unexpected error occurred.",
                variant: "destructive",
            });
        }
    };

    if (loading) {
        return (
            <div className="p-6 max-w-7xl mx-auto">
                <div className="animate-pulse space-y-4">
                    <div className="h-8 bg-muted rounded w-1/3"></div>
                    <div className="h-64 bg-muted rounded"></div>
                </div>
            </div>
        );
    }

    return (
        <div className="p-6 max-w-7xl mx-auto">
            {/* New Secret Alert - Shown once after creation or rotation */}
            {newSecret && (
                <Card className="p-6 mb-6 border-2 border-yellow-500 bg-yellow-50 dark:bg-yellow-950">
                    <div className="flex items-start justify-between mb-4">
                        <div className="flex-1">
                            <h2 className="text-xl font-bold text-foreground flex items-center gap-2">
                                ⚠️ Save Your Client Secret
                            </h2>
                            <p className="text-sm text-muted-foreground mt-1">
                                This secret will only be shown once. Store it securely!
                            </p>
                        </div>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setNewSecret(null)}
                        >
                            ✕
                        </Button>
                    </div>
                    
                    <div className="space-y-3">
                        <div>
                            <label className="block text-xs font-medium text-muted-foreground mb-1">
                                Application Name
                            </label>
                            <div className="text-sm font-semibold text-foreground">
                                {newSecret.appName}
                            </div>
                        </div>
                        
                        <div>
                            <label className="block text-xs font-medium text-muted-foreground mb-1">
                                Client ID
                            </label>
                            <div className="flex gap-2">
                                <Input
                                    value={newSecret.clientId}
                                    readOnly
                                    className="font-mono text-sm bg-white dark:bg-gray-900"
                                />
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => copyToClipboard(newSecret.clientId, 'Client ID')}
                                >
                                    <Copy className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                        
                        <div>
                            <label className="block text-xs font-medium text-muted-foreground mb-1">
                                Client Secret
                            </label>
                            <div className="flex gap-2">
                                <Input
                                    value={newSecret.clientSecret}
                                    readOnly
                                    className="font-mono text-sm bg-white dark:bg-gray-900 font-bold"
                                />
                                <Button
                                    variant="default"
                                    size="sm"
                                    onClick={() => copyToClipboard(newSecret.clientSecret, 'Client Secret')}
                                >
                                    <Copy className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                        
                        <div className="bg-yellow-100 dark:bg-yellow-900 p-3 rounded text-sm">
                            <strong>⚠️ Important:</strong> Copy and save this client secret now. 
                            You won't be able to see it again. If you lose it, you'll need to rotate the secret.
                        </div>
                    </div>
                </Card>
            )}

            {/* Header */}
            <div className="mb-6">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-foreground flex items-center gap-2">
                            <Lock className="h-8 w-8 text-primary" />
                            OAuth2 Applications
                        </h1>
                        <p className="text-muted-foreground mt-1">
                            Manage OAuth2 client applications for secure authentication
                        </p>
                    </div>
                    <Button
                        onClick={() => setShowCreateForm(!showCreateForm)}
                        className="flex items-center gap-2"
                    >
                        <Plus className="h-4 w-4" />
                        Create App
                    </Button>
                </div>
            </div>

            {/* Create Form */}
            {showCreateForm && (
                <Card className="p-6 mb-6">
                    <h2 className="text-xl font-bold text-foreground mb-4">Create OAuth2 Application</h2>
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                Application Name *
                            </label>
                            <Input
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                placeholder="My OAuth2 App"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                Description
                            </label>
                            <Input
                                value={formData.description}
                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                placeholder="Brief description of your application"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                Redirect URIs *
                            </label>
                            <div className="space-y-2">
                                {redirectUris.map((uri, index) => (
                                    <div key={index} className="flex gap-2">
                                        <Input
                                            value={uri}
                                            onChange={(e) => updateRedirectUri(index, e.target.value)}
                                            placeholder="https://example.com/callback"
                                            className="flex-1"
                                        />
                                        {redirectUris.length > 1 && (
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                onClick={() => removeRedirectUri(index)}
                                                className="flex-shrink-0"
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        )}
                                    </div>
                                ))}
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    onClick={addRedirectUri}
                                    className="w-full"
                                >
                                    <Plus className="h-4 w-4 mr-2" />
                                    Add Redirect URI
                                </Button>
                            </div>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                Allowed Scopes (optional)
                            </label>
                            <div className="space-y-2">
                                {allowedScopes.length === 0 ? (
                                    <Button
                                        type="button"
                                        variant="outline"
                                        size="sm"
                                        onClick={addScope}
                                        className="w-full"
                                    >
                                        <Plus className="h-4 w-4 mr-2" />
                                        Add Scope
                                    </Button>
                                ) : (
                                    <>
                                        {allowedScopes.map((scope, index) => (
                                            <div key={index} className="flex gap-2">
                                                <Input
                                                    value={scope}
                                                    onChange={(e) => updateScope(index, e.target.value)}
                                                    placeholder="e.g. email, profile, org.read"
                                                    className="flex-1"
                                                />
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => removeScope(index)}
                                                    className="flex-shrink-0"
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </Button>
                                            </div>
                                        ))}
                                        <Button
                                            type="button"
                                            variant="outline"
                                            size="sm"
                                            onClick={addScope}
                                            className="w-full"
                                        >
                                            <Plus className="h-4 w-4 mr-2" />
                                            Add Scope
                                        </Button>
                                    </>
                                )}
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                                Leave empty to allow all scopes
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-foreground mb-3">
                                Client Type *
                            </label>
                            <div className="space-y-3">
                                {/* Public Client Option */}
                                <div 
                                    onClick={() => setFormData({ ...formData, is_confidential: false })}
                                    className={`border-2 rounded-lg p-4 cursor-pointer transition-all ${
                                        !formData.is_confidential 
                                            ? 'border-primary bg-primary/5' 
                                            : 'border-border hover:border-primary/50'
                                    }`}
                                >
                                    <div className="flex items-start gap-3">
                                        <input
                                            type="radio"
                                            id="public"
                                            checked={!formData.is_confidential}
                                            onChange={() => setFormData({ ...formData, is_confidential: false })}
                                            className="mt-1"
                                        />
                                        <div className="flex-1">
                                            <label htmlFor="public" className="font-semibold text-foreground cursor-pointer flex items-center gap-2">
                                                📱 Public Client
                                                <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full font-normal">
                                                    Recommended for SPAs
                                                </span>
                                            </label>
                                            <p className="text-sm text-muted-foreground mt-1">
                                                For Single Page Applications (React, Vue, Angular), mobile apps, or any client that cannot securely store secrets.
                                            </p>
                                            <ul className="text-xs text-muted-foreground mt-2 space-y-1 ml-4">
                                                <li>✓ Uses PKCE (Proof Key for Code Exchange)</li>
                                                <li>✓ No client secret required</li>
                                                <li>✓ Secure for browser-based apps</li>
                                            </ul>
                                        </div>
                                    </div>
                                </div>

                                {/* Confidential Client Option */}
                                <div 
                                    onClick={() => setFormData({ ...formData, is_confidential: true })}
                                    className={`border-2 rounded-lg p-4 cursor-pointer transition-all ${
                                        formData.is_confidential 
                                            ? 'border-primary bg-primary/5' 
                                            : 'border-border hover:border-primary/50'
                                    }`}
                                >
                                    <div className="flex items-start gap-3">
                                        <input
                                            type="radio"
                                            id="confidential"
                                            checked={formData.is_confidential}
                                            onChange={() => setFormData({ ...formData, is_confidential: true })}
                                            className="mt-1"
                                        />
                                        <div className="flex-1">
                                            <label htmlFor="confidential" className="font-semibold text-foreground cursor-pointer flex items-center gap-2">
                                                🔒 Confidential Client
                                                <span className="text-xs bg-amber-100 text-amber-700 px-2 py-0.5 rounded-full font-normal">
                                                    For server-side
                                                </span>
                                            </label>
                                            <p className="text-sm text-muted-foreground mt-1">
                                                For server-side applications (Node.js, Python, PHP, Java) that can securely store client secrets on the backend.
                                            </p>
                                            <ul className="text-xs text-muted-foreground mt-2 space-y-1 ml-4">
                                                <li>✓ Uses client secret + PKCE</li>
                                                <li>✓ Higher security for server apps</li>
                                                <li>✓ Secret never exposed to browser</li>
                                            </ul>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="flex gap-2">
                            <Button onClick={handleCreateApp} disabled={creating}>
                                {creating ? 'Creating...' : 'Create Application'}
                            </Button>
                            <Button
                                variant="outline"
                                onClick={() => setShowCreateForm(false)}
                            >
                                Cancel
                            </Button>
                        </div>
                    </div>
                </Card>
            )}

            {/* Apps List */}
            <div className="space-y-4">
                {apps.length === 0 ? (
                    <Card className="p-8 text-center text-muted-foreground">
                        <Lock className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                        <p>No OAuth2 applications yet. Create your first app to get started.</p>
                    </Card>
                ) : (
                    apps.map((app) => (
                        <Card key={app.id} className="p-6">
                            <div className="flex items-start justify-between mb-4">
                                <div>
                                    <h3 className="text-xl font-bold text-foreground">{app.name}</h3>
                                    {app.description && (
                                        <p className="text-sm text-muted-foreground mt-1">{app.description}</p>
                                    )}
                                    <p className="text-xs text-muted-foreground mt-2">
                                        Created: {new Date(app.created_at).toLocaleDateString()}
                                    </p>
                                </div>
                                <div className="flex gap-2">
                                    {editingApp === app.id ? (
                                        <>
                                            <Button
                                                variant="default"
                                                size="sm"
                                                onClick={() => handleUpdateApp(app.id, app.name)}
                                            >
                                                Save
                                            </Button>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={cancelEditing}
                                            >
                                                Cancel
                                            </Button>
                                        </>
                                    ) : (
                                        <>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => startEditingApp(app)}
                                            >
                                                Edit
                                            </Button>
                                            {app.is_confidential && (
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => handleRotateSecret(app.id, app.name)}
                                                >
                                                    <RefreshCw className="h-4 w-4" />
                                                </Button>
                                            )}
                                            <Button
                                                variant="destructive"
                                                size="sm"
                                                onClick={() => handleDeleteApp(app.id, app.name)}
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </>
                                    )}
                                </div>
                            </div>

                            <div className="space-y-3">
                                {/* Client ID */}
                                <div>
                                    <label className="block text-xs font-medium text-muted-foreground mb-1">
                                        Client ID
                                    </label>
                                    <div className="flex gap-2">
                                        <Input
                                            value={app.client_id}
                                            readOnly
                                            className="font-mono text-sm"
                                        />
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={() => copyToClipboard(app.client_id, 'Client ID')}
                                        >
                                            <Copy className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>

                                {/* Client Secret */}
                                {app.is_confidential && app.client_secret && (
                                    <div>
                                        <label className="block text-xs font-medium text-muted-foreground mb-1">
                                            Client Secret
                                        </label>
                                        <div className="flex gap-2">
                                            <Input
                                                type={visibleSecrets.has(app.id) ? 'text' : 'password'}
                                                value={app.client_secret}
                                                readOnly
                                                className="font-mono text-sm"
                                            />
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => toggleSecretVisibility(app.id)}
                                            >
                                                {visibleSecrets.has(app.id) ? (
                                                    <EyeOff className="h-4 w-4" />
                                                ) : (
                                                    <Eye className="h-4 w-4" />
                                                )}
                                            </Button>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => copyToClipboard(app.client_secret!, 'Client Secret')}
                                            >
                                                <Copy className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </div>
                                )}

                                {/* Client Type */}
                                <div>
                                    <label className="block text-xs font-medium text-muted-foreground mb-1">
                                        Client Type
                                    </label>
                                    {editingApp === app.id ? (
                                        <div className="space-y-2">
                                            {/* Public Client Option */}
                                            <div 
                                                onClick={() => setEditFormData(prev => ({ ...prev, is_confidential: false }))}
                                                className={`border rounded-lg p-3 cursor-pointer transition-all ${
                                                    !editFormData.is_confidential 
                                                        ? 'border-primary bg-primary/5' 
                                                        : 'border-border hover:border-primary/50'
                                                }`}
                                            >
                                                <div className="flex items-center gap-2">
                                                    <input
                                                        type="radio"
                                                        checked={!editFormData.is_confidential}
                                                        onChange={() => setEditFormData(prev => ({ ...prev, is_confidential: false }))}
                                                        className="cursor-pointer"
                                                    />
                                                    <label className="text-sm font-medium cursor-pointer">
                                                        📱 Public Client (SPA/Mobile - No Secret)
                                                    </label>
                                                </div>
                                            </div>

                                            {/* Confidential Client Option */}
                                            <div 
                                                onClick={() => setEditFormData(prev => ({ ...prev, is_confidential: true }))}
                                                className={`border rounded-lg p-3 cursor-pointer transition-all ${
                                                    editFormData.is_confidential 
                                                        ? 'border-primary bg-primary/5' 
                                                        : 'border-border hover:border-primary/50'
                                                }`}
                                            >
                                                <div className="flex items-center gap-2">
                                                    <input
                                                        type="radio"
                                                        checked={editFormData.is_confidential}
                                                        onChange={() => setEditFormData(prev => ({ ...prev, is_confidential: true }))}
                                                        className="cursor-pointer"
                                                    />
                                                    <label className="text-sm font-medium cursor-pointer">
                                                        🔒 Confidential Client (Server-side - With Secret)
                                                    </label>
                                                </div>
                                            </div>

                                            {editFormData.is_confidential !== app.is_confidential && (
                                                <div className="bg-amber-50 border border-amber-200 rounded p-2 text-xs text-amber-800">
                                                    ⚠️ <strong>Warning:</strong> Changing client type may break existing integrations.
                                                    {!editFormData.is_confidential && app.is_confidential && (
                                                        <span> Client secret will no longer be required.</span>
                                                    )}
                                                    {editFormData.is_confidential && !app.is_confidential && (
                                                        <span> A client secret will be required after this change.</span>
                                                    )}
                                                </div>
                                            )}
                                        </div>
                                    ) : (
                                        <div className="text-sm">
                                            {app.is_confidential ? (
                                                <span className="inline-flex items-center gap-1 bg-amber-100 text-amber-700 px-3 py-1 rounded-full text-xs font-medium">
                                                    🔒 Confidential Client
                                                </span>
                                            ) : (
                                                <span className="inline-flex items-center gap-1 bg-blue-100 text-blue-700 px-3 py-1 rounded-full text-xs font-medium">
                                                    📱 Public Client
                                                </span>
                                            )}
                                        </div>
                                    )}
                                </div>

                                {/* Redirect URIs */}
                                <div>
                                    <label className="block text-xs font-medium text-muted-foreground mb-1">
                                        Redirect URIs
                                    </label>
                                    {editingApp === app.id ? (
                                        <div className="space-y-2">
                                            {editFormData.redirect_uris.map((uri, index) => (
                                                <div key={index} className="flex gap-2">
                                                    <Input
                                                        value={uri}
                                                        onChange={(e) => updateEditRedirectUri(index, e.target.value)}
                                                        placeholder="https://example.com/callback"
                                                        className="flex-1 text-sm"
                                                    />
                                                    {editFormData.redirect_uris.length > 1 && (
                                                        <Button
                                                            type="button"
                                                            variant="outline"
                                                            size="sm"
                                                            onClick={() => removeEditRedirectUri(index)}
                                                        >
                                                            <Trash2 className="h-4 w-4" />
                                                        </Button>
                                                    )}
                                                </div>
                                            ))}
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                onClick={addEditRedirectUri}
                                                className="w-full"
                                            >
                                                <Plus className="h-4 w-4 mr-2" />
                                                Add Redirect URI
                                            </Button>
                                        </div>
                                    ) : (
                                        <div className="space-y-1">
                                            {app.redirect_uris.map((uri, idx) => (
                                                <div key={idx} className="text-sm text-foreground bg-muted px-3 py-2 rounded">
                                                    {uri}
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                </div>

                                {/* Scopes */}
                                <div>
                                    <label className="block text-xs font-medium text-muted-foreground mb-1">
                                        Allowed Scopes
                                    </label>
                                    {editingApp === app.id ? (
                                        <div className="space-y-2">
                                            {editFormData.allowed_scopes.length === 0 ? (
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={addEditScope}
                                                    className="w-full"
                                                >
                                                    <Plus className="h-4 w-4 mr-2" />
                                                    Add Scope
                                                </Button>
                                            ) : (
                                                <>
                                                    {editFormData.allowed_scopes.map((scope, index) => (
                                                        <div key={index} className="flex gap-2">
                                                            <Input
                                                                value={scope}
                                                                onChange={(e) => updateEditScope(index, e.target.value)}
                                                                placeholder="e.g. email, profile, org.read"
                                                                className="flex-1 text-sm"
                                                            />
                                                            <Button
                                                                type="button"
                                                                variant="outline"
                                                                size="sm"
                                                                onClick={() => removeEditScope(index)}
                                                            >
                                                                <Trash2 className="h-4 w-4" />
                                                            </Button>
                                                        </div>
                                                    ))}
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        onClick={addEditScope}
                                                        className="w-full"
                                                    >
                                                        <Plus className="h-4 w-4 mr-2" />
                                                        Add Scope
                                                    </Button>
                                                </>
                                            )}
                                        </div>
                                    ) : (
                                        <div className="flex flex-wrap gap-2">
                                            {app.scopes && app.scopes.length > 0 ? (
                                                app.scopes.map((scope, idx) => (
                                                    <span
                                                        key={idx}
                                                        className="text-xs bg-primary/10 text-primary px-2 py-1 rounded-full"
                                                    >
                                                        {scope}
                                                    </span>
                                                ))
                                            ) : (
                                                <span className="text-sm text-muted-foreground">No scopes configured</span>
                                            )}
                                        </div>
                                    )}
                                </div>
                            </div>
                        </Card>
                    ))
                )}
            </div>
        </div>
    );
}
