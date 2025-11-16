import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  ArrowLeft,
  Plus,
  Key,
  Copy,
  Trash2,
  Edit,
  RotateCw,
  AlertCircle,
  CheckCircle,
  Shield,
  Globe,
  Lock,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import LoadingSpinner from '@/components/ui/loading-spinner'
import { useToast } from '@/hooks/use-toast'
import useAuthStore from '@/store/auth'
import { clientAppAPI } from '@/lib/api'

interface ClientApp {
  id: string
  client_id: string
  name: string
  description: string
  redirect_uris: string[]
  allowed_scopes: string[]
  created_at: string
  updated_at: string
}

const OAuthClientApps = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const { isSuperadmin } = useAuthStore()

  const [clientApps, setClientApps] = useState<ClientApp[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  
  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [secretDialogOpen, setSecretDialogOpen] = useState(false)
  
  const [selectedApp, setSelectedApp] = useState<ClientApp | null>(null)
  const [newSecret, setNewSecret] = useState('')
  
  // Form states
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    redirect_uris: '',
    allowed_scopes: '',
  })

  // Redirect if not superadmin
  useEffect(() => {
    if (!isSuperadmin) {
      navigate('/dashboard', { replace: true })
    }
  }, [isSuperadmin, navigate])

  // Load client apps
  useEffect(() => {
    if (isSuperadmin) {
      loadClientApps()
    }
  }, [isSuperadmin])

  const loadClientApps = async () => {
    try {
      setLoading(true)
      setError('')
      const response = await clientAppAPI.listClientApps()
      setClientApps(response.data || [])
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to load client apps'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateApp = async () => {
    try {
      const redirectUris = formData.redirect_uris
        .split('\n')
        .map(uri => uri.trim())
        .filter(uri => uri.length > 0)

      const allowedScopes = formData.allowed_scopes
        .split(',')
        .map(scope => scope.trim())
        .filter(scope => scope.length > 0)

      const response = await clientAppAPI.createClientApp({
        name: formData.name,
        description: formData.description,
        redirect_uris: redirectUris,
        allowed_scopes: allowedScopes,
      })

      toast({
        title: 'Client App Created',
        description: 'OAuth client application created successfully',
      })

      // Show the client secret (from top-level response)
      setNewSecret((response as any).client_secret || '')
      setSecretDialogOpen(true)
      setCreateDialogOpen(false)
      
      // Reset form
      setFormData({
        name: '',
        description: '',
        redirect_uris: '',
        allowed_scopes: '',
      })

      // Reload list
      loadClientApps()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || 'Failed to create client app',
        variant: 'destructive',
      })
    }
  }

  const handleUpdateApp = async () => {
    if (!selectedApp) return

    try {
      const redirectUris = formData.redirect_uris
        .split('\n')
        .map(uri => uri.trim())
        .filter(uri => uri.length > 0)

      const allowedScopes = formData.allowed_scopes
        .split(',')
        .map(scope => scope.trim())
        .filter(scope => scope.length > 0)

      await clientAppAPI.updateClientApp(selectedApp.id, {
        name: formData.name,
        description: formData.description,
        redirect_uris: redirectUris,
        allowed_scopes: allowedScopes,
      })

      toast({
        title: 'Client App Updated',
        description: 'OAuth client application updated successfully',
      })

      setEditDialogOpen(false)
      setSelectedApp(null)
      loadClientApps()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || 'Failed to update client app',
        variant: 'destructive',
      })
    }
  }

  const handleDeleteApp = async () => {
    if (!selectedApp) return

    try {
      await clientAppAPI.deleteClientApp(selectedApp.id)

      toast({
        title: 'Client App Deleted',
        description: 'OAuth client application deleted successfully',
      })

      setDeleteDialogOpen(false)
      setSelectedApp(null)
      loadClientApps()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || 'Failed to delete client app',
        variant: 'destructive',
      })
    }
  }

  const handleRotateSecret = async (app: ClientApp) => {
    try {
      const response = await clientAppAPI.rotateSecret(app.id)
      setNewSecret((response as any).client_secret || response.data?.client_secret || '')
      setSecretDialogOpen(true)

      toast({
        title: 'Secret Rotated',
        description: 'Client secret rotated successfully',
      })
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || 'Failed to rotate secret',
        variant: 'destructive',
      })
    }
  }

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast({
      title: 'Copied!',
      description: `${label} copied to clipboard`,
    })
  }

  const openEditDialog = (app: ClientApp) => {
    setSelectedApp(app)
    setFormData({
      name: app.name,
      description: app.description,
      redirect_uris: app.redirect_uris.join('\n'),
      allowed_scopes: app.allowed_scopes.join(', '),
    })
    setEditDialogOpen(true)
  }

  const openDeleteDialog = (app: ClientApp) => {
    setSelectedApp(app)
    setDeleteDialogOpen(true)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString()
  }

  if (!isSuperadmin) {
    return null
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 dark:from-slate-900 dark:via-slate-800 dark:to-slate-700">
        <div className="text-center space-y-4">
          <div className="p-4 bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm rounded-xl shadow-lg border border-blue-200/50 dark:border-slate-600/50">
            <LoadingSpinner size="lg" />
          </div>
          <p className="text-slate-600 dark:text-slate-300 font-medium">Loading OAuth Client Applications...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 dark:from-slate-900 dark:via-slate-800 dark:to-slate-700">
      {/* Header */}
      <header className="bg-white/80 backdrop-blur-sm dark:bg-slate-800/80 shadow-lg border-b border-blue-200/50 dark:border-slate-600/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => navigate('/admin')}
                className="flex items-center space-x-2 text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white hover:bg-blue-50 dark:hover:bg-slate-700 transition-all duration-200"
              >
                <ArrowLeft className="h-4 w-4" />
                <span>Back to Admin</span>
              </Button>
            </div>
            <div className="flex items-center space-x-4">
              <div className="p-2 bg-gradient-to-r from-blue-500 to-indigo-600 rounded-lg shadow-lg">
                <Shield className="h-6 w-6 text-white" />
              </div>
              <div className="text-center">
                <h1 className="text-xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
                  OAuth Client Applications
                </h1>
                <div className="text-xs text-slate-500 dark:text-slate-400">Security & Integration</div>
              </div>
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
          {/* Error Message */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Header Actions */}
          <div className="flex flex-col lg:flex-row lg:justify-between lg:items-center gap-6 bg-white/60 dark:bg-slate-800/60 backdrop-blur-sm rounded-xl p-6 border border-blue-200/30 dark:border-slate-600/30 shadow-lg">
            <div className="space-y-2">
              <h2 className="text-3xl font-bold bg-gradient-to-r from-slate-800 to-slate-600 dark:from-white dark:to-slate-200 bg-clip-text text-transparent">
                Client Applications
              </h2>
              <p className="text-slate-600 dark:text-slate-300 text-lg">
                Manage OAuth2 client applications for third-party integrations
              </p>
              <div className="flex items-center space-x-2 text-sm text-slate-500 dark:text-slate-400">
                <Globe className="h-4 w-4" />
                <span>Secure authentication & authorization</span>
              </div>
            </div>
            <Button 
              onClick={() => setCreateDialogOpen(true)} 
              className="flex items-center space-x-2 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white shadow-lg hover:shadow-xl transition-all duration-200 px-6 py-3 text-base font-semibold"
              size="lg"
            >
              <Plus className="h-5 w-5" />
              <span>Create Client App</span>
            </Button>
          </div>

          {/* Client Apps List */}
          <div className="grid gap-6">
            {clientApps.length === 0 ? (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
              >
                <Card className="border-2 border-dashed border-blue-200 dark:border-slate-600 bg-gradient-to-br from-white to-blue-50/30 dark:from-slate-800 dark:to-slate-700 shadow-lg">
                  <CardContent className="pt-6">
                    <div className="text-center py-16">
                      <div className="mx-auto w-24 h-24 bg-gradient-to-br from-blue-100 to-indigo-100 dark:from-slate-700 dark:to-slate-600 rounded-full flex items-center justify-center mb-6 shadow-inner">
                        <Key className="h-12 w-12 text-blue-500 dark:text-blue-400" />
                      </div>
                      <h3 className="text-xl font-semibold text-slate-800 dark:text-white mb-2">
                        No client applications yet
                      </h3>
                      <p className="text-slate-600 dark:text-slate-300 mb-6">
                        Create your first OAuth client app to enable secure third-party integrations
                      </p>
                      <div className="flex flex-col sm:flex-row gap-3 justify-center">
                        <Button
                          onClick={() => setCreateDialogOpen(true)}
                          className="bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white shadow-lg hover:shadow-xl transition-all duration-200"
                        >
                          <Plus className="h-4 w-4 mr-2" />
                          Get Started
                        </Button>
                        <Button variant="outline" className="border-blue-200 text-blue-600 hover:bg-blue-50 dark:border-slate-600 dark:text-blue-400 dark:hover:bg-slate-700">
                          <Globe className="h-4 w-4 mr-2" />
                          Learn More
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            ) : (
              clientApps.map((app, index) => (
                <motion.div
                  key={app.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: index * 0.1 }}
                >
                  <Card className="bg-gradient-to-br from-white to-blue-50/20 dark:from-slate-800 dark:to-slate-700 border border-blue-200/50 dark:border-slate-600/50 shadow-lg hover:shadow-xl transition-all duration-200 hover:border-blue-300 dark:hover:border-slate-500">
                    <CardHeader className="pb-4">
                      <div className="flex items-start justify-between">
                        <div className="flex-1 space-y-3">
                          <div className="flex items-center space-x-3">
                            <div className="p-2 bg-gradient-to-r from-blue-500 to-indigo-600 rounded-lg shadow-sm">
                              <Key className="h-5 w-5 text-white" />
                            </div>
                            <div>
                              <CardTitle className="text-xl text-slate-800 dark:text-white flex items-center space-x-2">
                                <span>{app.name}</span>
                                <Badge variant="secondary" className="bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300 border-blue-200 dark:border-blue-700">
                                  <Lock className="h-3 w-3 mr-1" />
                                  OAuth2.1
                                </Badge>
                              </CardTitle>
                              <CardDescription className="text-slate-600 dark:text-slate-300 mt-1">
                                {app.description || 'No description provided'}
                              </CardDescription>
                            </div>
                          </div>
                        </div>
                        <div className="flex space-x-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => openEditDialog(app)}
                            className="text-slate-500 hover:text-blue-600 hover:bg-blue-50 dark:text-slate-400 dark:hover:text-blue-400 dark:hover:bg-slate-700"
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRotateSecret(app)}
                            className="text-slate-500 hover:text-indigo-600 hover:bg-indigo-50 dark:text-slate-400 dark:hover:text-indigo-400 dark:hover:bg-slate-700"
                          >
                            <RotateCw className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => openDeleteDialog(app)}
                            className="text-slate-500 hover:text-red-600 hover:bg-red-50 dark:text-slate-400 dark:hover:text-red-400 dark:hover:bg-red-900/20"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      {/* Client ID */}
                      <div>
                        <Label className="text-xs text-muted-foreground">Client ID</Label>
                        <div className="flex items-center space-x-2 mt-1">
                          <code className="flex-1 text-sm bg-muted px-3 py-2 rounded">
                            {app.client_id}
                          </code>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => copyToClipboard(app.client_id, 'Client ID')}
                          >
                            <Copy className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>

                      {/* Redirect URIs */}
                      <div>
                        <Label className="text-xs text-muted-foreground flex items-center space-x-1">
                          <Globe className="h-3 w-3" />
                          <span>Redirect URIs</span>
                        </Label>
                        <div className="mt-1 space-y-1">
                          {app.redirect_uris.map((uri, index) => (
                            <div key={index} className="text-sm bg-muted px-3 py-1 rounded">
                              {uri}
                            </div>
                          ))}
                        </div>
                      </div>

                      {/* Allowed Scopes */}
                      {app.allowed_scopes && app.allowed_scopes.length > 0 && (
                        <div>
                          <Label className="text-xs text-muted-foreground flex items-center space-x-1">
                            <Lock className="h-3 w-3" />
                            <span>Allowed Scopes</span>
                          </Label>
                          <div className="mt-1 flex flex-wrap gap-2">
                            {app.allowed_scopes.map((scope, index) => (
                              <Badge key={index} variant="secondary">
                                {scope}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Metadata */}
                      <div className="flex justify-between text-xs text-muted-foreground pt-2 border-t">
                        <span>Created: {formatDate(app.created_at)}</span>
                        <span>Updated: {formatDate(app.updated_at)}</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
                </motion.div>
              ))
            )}
          </div>
        </motion.div>
      </main>

      {/* Create Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent className="max-w-2xl bg-gradient-to-br from-white to-blue-50/30 dark:from-slate-800 dark:to-slate-700 border border-blue-200/50 dark:border-slate-600/50">
          <DialogHeader className="space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-gradient-to-r from-blue-500 to-indigo-600 rounded-lg shadow-lg">
                <Plus className="h-6 w-6 text-white" />
              </div>
              <div>
                <DialogTitle className="text-xl font-bold bg-gradient-to-r from-slate-800 to-slate-600 dark:from-white dark:to-slate-200 bg-clip-text text-transparent">
                  Create OAuth Client Application
                </DialogTitle>
                <DialogDescription className="text-slate-600 dark:text-slate-300">
                  Create a new OAuth2.1 client application for secure third-party integrations
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="name" className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center space-x-2">
                  <span>Application Name</span>
                  <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="My Application"
                  className="border-slate-200 dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-white dark:bg-slate-800"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                  Description
                </Label>
                <Input
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Application description"
                  className="border-slate-200 dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-white dark:bg-slate-800"
                />
              </div>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="redirect_uris" className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center space-x-2">
                <Globe className="h-4 w-4" />
                <span>Redirect URIs</span>
                <span className="text-red-500">*</span>
              </Label>
              <Textarea
                id="redirect_uris"
                value={formData.redirect_uris}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setFormData({ ...formData, redirect_uris: e.target.value })}
                placeholder="https://example.com/callback&#10;https://app.example.com/auth/callback"
                rows={4}
                className="border-slate-200 dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-white dark:bg-slate-800 resize-none"
              />
              <p className="text-xs text-slate-500 dark:text-slate-400">
                Enter one URI per line. These are the allowed callback URLs after authentication.
              </p>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="allowed_scopes" className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center space-x-2">
                <Lock className="h-4 w-4" />
                <span>Allowed Scopes</span>
              </Label>
              <Input
                id="allowed_scopes"
                value={formData.allowed_scopes}
                onChange={(e) => setFormData({ ...formData, allowed_scopes: e.target.value })}
                placeholder="profile, email, org:read, org:write"
                className="border-slate-200 dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 bg-white dark:bg-slate-800"
              />
              <p className="text-xs text-slate-500 dark:text-slate-400">
                Comma-separated list of OAuth scopes this application can request.
              </p>
            </div>
          </div>
          <DialogFooter className="space-x-3 pt-6 border-t border-slate-200 dark:border-slate-600">
            <Button 
              variant="outline" 
              onClick={() => setCreateDialogOpen(false)}
              className="border-slate-200 text-slate-600 hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
            >
              Cancel
            </Button>
            <Button 
              onClick={handleCreateApp} 
              disabled={!formData.name || !formData.redirect_uris}
              className="bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white shadow-lg hover:shadow-xl transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Plus className="h-4 w-4 mr-2" />
              Create Application
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit OAuth Client Application</DialogTitle>
            <DialogDescription>
              Update the client application configuration
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="edit-name">Application Name *</Label>
              <Input
                id="edit-name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="edit-description">Description</Label>
              <Input
                id="edit-description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="edit-redirect_uris">Redirect URIs * (one per line)</Label>
              <Textarea
                id="edit-redirect_uris"
                value={formData.redirect_uris}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setFormData({ ...formData, redirect_uris: e.target.value })}
                rows={4}
              />
            </div>
            <div>
              <Label htmlFor="edit-allowed_scopes">Allowed Scopes (comma-separated)</Label>
              <Input
                id="edit-allowed_scopes"
                value={formData.allowed_scopes}
                onChange={(e) => setFormData({ ...formData, allowed_scopes: e.target.value })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateApp}>Update Application</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Client Application</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{selectedApp?.name}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteApp}>
              Delete Application
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Secret Display Dialog */}
      <Dialog open={secretDialogOpen} onOpenChange={setSecretDialogOpen}>
        <DialogContent className="max-w-lg bg-gradient-to-br from-white to-green-50/30 dark:from-slate-800 dark:to-green-900/10 border border-green-200/50 dark:border-green-700/50">
          <DialogHeader className="space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-gradient-to-r from-green-500 to-emerald-600 rounded-lg shadow-lg">
                <CheckCircle className="h-6 w-6 text-white" />
              </div>
              <div>
                <DialogTitle className="text-xl font-bold bg-gradient-to-r from-green-600 to-emerald-600 bg-clip-text text-transparent">
                  Client Secret Generated
                </DialogTitle>
                <DialogDescription className="text-slate-600 dark:text-slate-300">
                  Save this secret securely. It will not be shown again!
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>
          <div className="space-y-6">
            <Alert className="border-amber-200 bg-amber-50 dark:border-amber-700 dark:bg-amber-900/20">
              <AlertCircle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
              <AlertDescription className="text-amber-800 dark:text-amber-200">
                <strong>Important:</strong> Copy and save this client secret in a secure location. 
                You won't be able to see it again after closing this dialog.
              </AlertDescription>
            </Alert>
            <div className="space-y-2">
              <Label className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center space-x-2">
                <Key className="h-4 w-4" />
                <span>Client Secret</span>
              </Label>
              <div className="flex items-center space-x-2">
                <code className="flex-1 text-sm bg-slate-100 dark:bg-slate-800 px-4 py-3 rounded-lg border border-slate-200 dark:border-slate-600 break-all font-mono text-slate-800 dark:text-slate-200">
                  {newSecret}
                </code>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => copyToClipboard(newSecret, 'Client Secret')}
                  className="shrink-0 border-slate-200 text-slate-600 hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setSecretDialogOpen(false)}>I've Saved It</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default OAuthClientApps
