'use client';

import { useState, useEffect } from 'react';
import { AuthService } from '@/lib/auth';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Key, Plus, Copy, Trash2, Eye, EyeOff, AlertTriangle } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

interface APIKey {
    id: string;
    key_id: string;
    name: string;
    description?: string;
    scopes: string[];
    revoked: boolean;
    last_used_at?: string;
    expires_at?: string;
    created_at: string;
    updated_at: string;
}

export default function APIKeysPage() {
    const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
    const [loading, setLoading] = useState(true);
    const [creating, setCreating] = useState(false);
    const [showCreateForm, setShowCreateForm] = useState(false);
    const [newlyCreatedKey, setNewlyCreatedKey] = useState<string | null>(null);
    const { toast } = useToast();

    // Form state
    const [formData, setFormData] = useState({
        name: '',
        expires_in_days: 90
    });

    // Visibility state for new key
    const [showNewKey, setShowNewKey] = useState(false);

    useEffect(() => {
        loadAPIKeys();
    }, []);

    const loadAPIKeys = async () => {
        try {
            setLoading(true);
            const response = await AuthService.apiCall('/api/v1/dev/api-keys');
            
            if (response.ok) {
                const data = await response.json();
                setApiKeys(data.data || data.api_keys || []);
            } else {
                toast({
                    title: "Failed to load API keys",
                    description: "Could not retrieve your API keys.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to load API keys:', error);
            toast({
                title: "Error",
                description: "An unexpected error occurred.",
                variant: "destructive",
            });
        } finally {
            setLoading(false);
        }
    };

    const handleCreateKey = async () => {
        if (!formData.name.trim()) {
            toast({
                title: "Missing information",
                description: "Please provide a name for the API key.",
                variant: "destructive",
            });
            return;
        }

        try {
            setCreating(true);

            const response = await AuthService.apiCall('/api/v1/dev/api-keys', {
                method: 'POST',
                body: JSON.stringify({
                    name: formData.name.trim(),
                    expires_in_days: formData.expires_in_days
                }),
            });

            if (response.ok) {
                const data = await response.json();
                // Backend returns { success: true, data: { secret: "key", id: "...", ... } }
                const newKey = data.data?.secret;
                
                if (newKey) {
                    setNewlyCreatedKey(newKey);
                    setShowNewKey(true); // Show the key by default
                } else {
                    console.error('No secret in response:', data);
                }

                toast({
                    title: "API key created",
                    description: `${formData.name} has been created. Save it now - you won't see it again!`,
                });

                // Reset form
                setFormData({
                    name: '',
                    expires_in_days: 90
                });
                setShowCreateForm(false);

                // Reload keys
                await loadAPIKeys();
            } else {
                const errorData = await response.json();
                toast({
                    title: "Failed to create API key",
                    description: errorData.message || "Could not create API key.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to create API key:', error);
            toast({
                title: "Error",
                description: "An unexpected error occurred.",
                variant: "destructive",
            });
        } finally {
            setCreating(false);
        }
    };

    const handleDeleteKey = async (keyId: string, keyName: string) => {
        if (!confirm(`Are you sure you want to revoke "${keyName}"? This action cannot be undone and will break any integrations using this key.`)) {
            return;
        }

        try {
            const response = await AuthService.apiCall(`/api/v1/dev/api-keys/${keyId}`, {
                method: 'DELETE',
            });

            if (response.ok) {
                toast({
                    title: "API key revoked",
                    description: `${keyName} has been revoked.`,
                });
                await loadAPIKeys();
            } else {
                toast({
                    title: "Failed to revoke key",
                    description: "Could not revoke the API key.",
                    variant: "destructive",
                });
            }
        } catch (error) {
            console.error('Failed to delete key:', error);
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

    const dismissNewKey = () => {
        setNewlyCreatedKey(null);
        setShowNewKey(false);
    };

    const isExpired = (expiresAt?: string) => {
        if (!expiresAt) return false;
        return new Date(expiresAt) < new Date();
    };

    const getExpiryText = (expiresAt?: string) => {
        if (!expiresAt) return 'Never expires';
        const date = new Date(expiresAt);
        const now = new Date();
        if (date < now) return 'Expired';
        
        const days = Math.ceil((date.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
        if (days === 0) return 'Expires today';
        if (days === 1) return 'Expires tomorrow';
        return `Expires in ${days} days`;
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
            {/* Header */}
            <div className="mb-6">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-foreground flex items-center gap-2">
                            <Key className="h-8 w-8 text-primary" />
                            API Keys
                        </h1>
                        <p className="text-muted-foreground mt-1">
                            Manage API keys for programmatic access to your organization
                        </p>
                    </div>
                    <Button
                        onClick={() => setShowCreateForm(!showCreateForm)}
                        className="flex items-center gap-2"
                    >
                        <Plus className="h-4 w-4" />
                        Create API Key
                    </Button>
                </div>
            </div>

            {/* Newly Created Key Alert */}
            {newlyCreatedKey && (
                <Card className="p-6 mb-6 border-primary bg-primary/5">
                    <div className="flex items-start gap-3">
                        <AlertTriangle className="h-5 w-5 text-primary flex-shrink-0 mt-0.5" />
                        <div className="flex-1">
                            <h3 className="font-bold text-foreground mb-2">Save Your API Key</h3>
                            <p className="text-sm text-muted-foreground mb-3">
                                This is the only time you'll see this key. Store it securely - we don't store the full key.
                            </p>
                            <div className="flex gap-2">
                                <Input
                                    type={showNewKey ? 'text' : 'password'}
                                    value={newlyCreatedKey}
                                    readOnly
                                    className="font-mono text-sm bg-background"
                                />
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => setShowNewKey(!showNewKey)}
                                >
                                    {showNewKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </Button>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => copyToClipboard(newlyCreatedKey, 'API Key')}
                                >
                                    <Copy className="h-4 w-4" />
                                </Button>
                            </div>
                            <Button
                                variant="ghost"
                                size="sm"
                                onClick={dismissNewKey}
                                className="mt-3"
                            >
                                I've saved my key
                            </Button>
                        </div>
                    </div>
                </Card>
            )}

            {/* Create Form */}
            {showCreateForm && (
                <Card className="p-6 mb-6">
                    <h2 className="text-xl font-bold text-foreground mb-4">Create API Key</h2>
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                Key Name *
                            </label>
                            <Input
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                placeholder="Production API Key"
                            />
                            <p className="text-xs text-muted-foreground mt-1">
                                A descriptive name to help you identify this key
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                Expires in (days)
                            </label>
                            <Input
                                type="number"
                                value={formData.expires_in_days}
                                onChange={(e) => setFormData({ ...formData, expires_in_days: parseInt(e.target.value) || 90 })}
                                min={1}
                                max={365}
                            />
                            <p className="text-xs text-muted-foreground mt-1">
                                Set to 0 for no expiration (not recommended)
                            </p>
                        </div>

                        <div className="flex gap-2">
                            <Button onClick={handleCreateKey} disabled={creating}>
                                {creating ? 'Creating...' : 'Create API Key'}
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

            {/* API Keys List */}
            <div className="space-y-4">
                {apiKeys.length === 0 ? (
                    <Card className="p-8 text-center text-muted-foreground">
                        <Key className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                        <p>No API keys yet. Create your first key to get started.</p>
                    </Card>
                ) : (
                    apiKeys.map((key) => (
                        <Card key={key.id} className={`p-6 ${key.revoked || isExpired(key.expires_at) ? 'border-destructive bg-destructive/5 opacity-60' : ''}`}>
                            <div className="flex items-start justify-between mb-4">
                                <div className="flex-1">
                                    <div className="flex items-center gap-2">
                                        <h3 className="text-xl font-bold text-foreground">{key.name}</h3>
                                        {key.revoked && (
                                            <span className="text-xs font-semibold px-2 py-1 rounded bg-destructive text-destructive-foreground">
                                                REVOKED
                                            </span>
                                        )}
                                        {isExpired(key.expires_at) && !key.revoked && (
                                            <span className="text-xs font-semibold px-2 py-1 rounded bg-destructive text-destructive-foreground">
                                                EXPIRED
                                            </span>
                                        )}
                                    </div>
                                    <div className="flex items-center gap-4 mt-2">
                                        <span className="text-xs text-muted-foreground">
                                            Created: {new Date(key.created_at).toLocaleDateString()}
                                        </span>
                                        {key.last_used_at && (
                                            <span className="text-xs text-muted-foreground">
                                                Last used: {new Date(key.last_used_at).toLocaleDateString()}
                                            </span>
                                        )}
                                        {!key.revoked && (
                                            <span className={`text-xs font-medium ${isExpired(key.expires_at) ? 'text-destructive' : 'text-muted-foreground'}`}>
                                                {getExpiryText(key.expires_at)}
                                            </span>
                                        )}
                                    </div>
                                </div>
                                {!key.revoked && (
                                    <Button
                                        variant="destructive"
                                        size="sm"
                                        onClick={() => handleDeleteKey(key.key_id, key.name)}
                                    >
                                        <Trash2 className="h-4 w-4" />
                                    </Button>
                                )}
                            </div>

                            <div>
                                <label className="block text-xs font-medium text-muted-foreground mb-1">
                                    Key ID (Public)
                                </label>
                                <div className="flex gap-2">
                                    <div className="flex-1 font-mono text-sm text-foreground bg-muted px-3 py-2 rounded">
                                        {key.key_id}
                                    </div>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => copyToClipboard(key.key_id, 'Key ID')}
                                    >
                                        <Copy className="h-4 w-4" />
                                    </Button>
                                </div>
                                <p className="text-xs text-muted-foreground mt-1">
                                    This is your public key identifier. Use this to identify the key in API requests.
                                </p>
                            </div>
                        </Card>
                    ))
                )}
            </div>
        </div>
    );
}
