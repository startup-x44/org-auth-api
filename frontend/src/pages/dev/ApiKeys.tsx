import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  ArrowLeft,
  Plus,
  Key,
  Copy,
  Trash2,
  Eye,
  EyeOff,
  Calendar,
  AlertCircle,
  CheckCircle,
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
import api from '@/lib/axios-instance'

interface APIKey {
  id: string
  key_id: string
  name: string
  description: string
  client_app_id?: string
  scopes: string[]
  expires_at?: string
  revoked: boolean
  last_used_at?: string
  created_at: string
  updated_at: string
}

interface APIKeyCreateResponse extends APIKey {
  secret: string
}

const ApiKeys = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const { isAuthenticated } = useAuthStore()

  const [apiKeys, setApiKeys] = useState<APIKey[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  
  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [secretDialogOpen, setSecretDialogOpen] = useState(false)
  
  const [selectedKey, setSelectedKey] = useState<APIKey | null>(null)
  const [newSecret, setNewSecret] = useState('')
  const [secretVisible, setSecretVisible] = useState(false)
  
  // Form states
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    scopes: '',
    expires_at: '',
  })

  // Redirect if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { replace: true })
    }
  }, [isAuthenticated, navigate])

  // Load API keys
  useEffect(() => {
    if (isAuthenticated) {
      loadAPIKeys()
    }
  }, [isAuthenticated])

  const loadAPIKeys = async () => {
    try {
      setLoading(true)
      setError('')
      const response = await api.get('/dev/api-keys')
      setApiKeys(response.data.data || [])
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Failed to load API keys'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateKey = async () => {
    try {
      const scopes = formData.scopes
        .split(',')
        .map(scope => scope.trim())
        .filter(scope => scope.length > 0)

      const requestData: any = {
        name: formData.name,
        description: formData.description,
        scopes: scopes,
      }

      if (formData.expires_at) {
        requestData.expires_at = new Date(formData.expires_at).toISOString()
      }

      const response = await api.post('/dev/api-keys', requestData)
      const newKey = response.data.data as APIKeyCreateResponse

      toast({
        title: 'API Key Created',
        description: 'Your API key has been created successfully',
      })

      // Show the secret (only time it's visible)
      setNewSecret(newKey.secret)
      setSecretDialogOpen(true)
      setCreateDialogOpen(false)
      
      // Reset form
      setFormData({
        name: '',
        description: '',
        scopes: '',
        expires_at: '',
      })

      // Reload list
      loadAPIKeys()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to create API key',
        variant: 'destructive',
      })
    }
  }

  const handleDeleteKey = async () => {
    if (!selectedKey) return

    try {
      await api.delete(`/dev/api-keys/${selectedKey.id}`)

      toast({
        title: 'API Key Revoked',
        description: 'API key has been revoked successfully',
      })

      setDeleteDialogOpen(false)
      setSelectedKey(null)
      loadAPIKeys()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to revoke API key',
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

  const openDeleteDialog = (apiKey: APIKey) => {
    setSelectedKey(apiKey)
    setDeleteDialogOpen(true)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString()
  }

  const isExpired = (expiresAt?: string) => {
    if (!expiresAt) return false
    return new Date(expiresAt) < new Date()
  }

  if (!isAuthenticated) {
    return null
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-emerald-50 via-teal-50 to-cyan-50 dark:from-slate-900 dark:via-slate-800 dark:to-slate-700">
        <div className="text-center space-y-4">
          <div className="p-4 bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm rounded-xl shadow-lg border border-emerald-200/50 dark:border-slate-600/50">
            <LoadingSpinner size="lg" />
          </div>
          <p className="text-slate-600 dark:text-slate-300 font-medium">Loading API Keys...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-emerald-50 via-teal-50 to-cyan-50 dark:from-slate-900 dark:via-slate-800 dark:to-slate-700">
      {/* Header */}
      <header className="bg-white/80 backdrop-blur-sm dark:bg-slate-800/80 shadow-lg border-b border-emerald-200/50 dark:border-slate-600/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => navigate('/dashboard')}
                className="flex items-center space-x-2 text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white hover:bg-emerald-50 dark:hover:bg-slate-700 transition-all duration-200"
              >
                <ArrowLeft className="h-4 w-4" />
                <span>Back to Dashboard</span>
              </Button>
            </div>
            <div className="flex items-center space-x-4">
              <div className="p-2 bg-gradient-to-r from-emerald-500 to-teal-600 rounded-lg shadow-lg">
                <Key className="h-6 w-6 text-white" />
              </div>
              <div className="text-center">
                <h1 className="text-xl font-bold bg-gradient-to-r from-emerald-600 to-teal-600 bg-clip-text text-transparent">
                  API Keys
                </h1>
                <div className="text-xs text-slate-500 dark:text-slate-400">Developer Tools</div>
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
          <div className="flex flex-col lg:flex-row lg:justify-between lg:items-center gap-6 bg-white/60 dark:bg-slate-800/60 backdrop-blur-sm rounded-xl p-6 border border-emerald-200/30 dark:border-slate-600/30 shadow-lg">
            <div className="space-y-2">
              <h2 className="text-3xl font-bold bg-gradient-to-r from-slate-800 to-slate-600 dark:from-white dark:to-slate-200 bg-clip-text text-transparent">
                API Keys
              </h2>
              <p className="text-slate-600 dark:text-slate-300 text-lg">
                Manage API keys for programmatic access to your account
              </p>
              <div className="flex items-center space-x-2 text-sm text-slate-500 dark:text-slate-400">
                <Globe className="h-4 w-4" />
                <span>Secure programmatic access</span>
              </div>
            </div>
            <Button 
              onClick={() => setCreateDialogOpen(true)} 
              className="flex items-center space-x-2 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 text-white shadow-lg hover:shadow-xl transition-all duration-200 px-6 py-3 text-base font-semibold"
              size="lg"
            >
              <Plus className="h-5 w-5" />
              <span>Create API Key</span>
            </Button>
          </div>

          {/* API Keys List */}
          <div className="grid gap-6">
            {apiKeys.length === 0 ? (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
              >
                <Card className="border-2 border-dashed border-emerald-200 dark:border-slate-600 bg-gradient-to-br from-white to-emerald-50/30 dark:from-slate-800 dark:to-slate-700 shadow-lg">
                  <CardContent className="pt-6">
                    <div className="text-center py-16">
                      <div className="mx-auto w-24 h-24 bg-gradient-to-br from-emerald-100 to-teal-100 dark:from-slate-700 dark:to-slate-600 rounded-full flex items-center justify-center mb-6 shadow-inner">
                        <Key className="h-12 w-12 text-emerald-500 dark:text-emerald-400" />
                      </div>
                      <h3 className="text-xl font-semibold text-slate-800 dark:text-white mb-2">
                        No API keys yet
                      </h3>
                      <p className="text-slate-600 dark:text-slate-300 mb-6">
                        Create your first API key to enable programmatic access to your account
                      </p>
                      <div className="flex flex-col sm:flex-row gap-3 justify-center">
                        <Button
                          onClick={() => setCreateDialogOpen(true)}
                          className="bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 text-white shadow-lg hover:shadow-xl transition-all duration-200"
                        >
                          <Plus className="h-4 w-4 mr-2" />
                          Get Started
                        </Button>
                        <Button variant="outline" className="border-emerald-200 text-emerald-600 hover:bg-emerald-50 dark:border-slate-600 dark:text-emerald-400 dark:hover:bg-slate-700">
                          <Lock className="h-4 w-4 mr-2" />
                          Learn More
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            ) : (
              apiKeys.map((apiKey, index) => (
                <motion.div
                  key={apiKey.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: index * 0.1 }}
                >
                  <Card className="bg-gradient-to-br from-white to-emerald-50/20 dark:from-slate-800 dark:to-slate-700 border border-emerald-200/50 dark:border-slate-600/50 shadow-lg hover:shadow-xl transition-all duration-200 hover:border-emerald-300 dark:hover:border-slate-500">
                    <CardHeader className="pb-4">
                      <div className="flex items-start justify-between">
                        <div className="flex-1 space-y-3">
                          <div className="flex items-center space-x-3">
                            <div className="p-2 bg-gradient-to-r from-emerald-500 to-teal-600 rounded-lg shadow-sm">
                              <Key className="h-5 w-5 text-white" />
                            </div>
                            <div>
                              <CardTitle className="text-xl text-slate-800 dark:text-white flex items-center space-x-2">
                                <span>{apiKey.name}</span>
                                {apiKey.revoked && (
                                  <Badge variant="destructive">Revoked</Badge>
                                )}
                                {!apiKey.revoked && apiKey.expires_at && isExpired(apiKey.expires_at) && (
                                  <Badge variant="destructive">Expired</Badge>
                                )}
                                {!apiKey.revoked && !isExpired(apiKey.expires_at) && (
                                  <Badge className="bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300 border-emerald-200 dark:border-emerald-700">
                                    <Lock className="h-3 w-3 mr-1" />
                                    Active
                                  </Badge>
                                )}
                              </CardTitle>
                              <CardDescription className="text-slate-600 dark:text-slate-300 mt-1">
                                {apiKey.description || 'No description provided'}
                              </CardDescription>
                            </div>
                          </div>
                        </div>
                        <div className="flex space-x-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => openDeleteDialog(apiKey)}
                            disabled={apiKey.revoked}
                            className="text-slate-500 hover:text-red-600 hover:bg-red-50 dark:text-slate-400 dark:hover:text-red-400 dark:hover:bg-red-900/20"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        {/* Key ID */}
                        <div>
                          <Label className="text-xs text-slate-500 dark:text-slate-400 font-medium">Key ID</Label>
                          <div className="flex items-center space-x-2 mt-1">
                            <code className="flex-1 text-sm bg-slate-100 dark:bg-slate-800 px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 font-mono text-slate-800 dark:text-slate-200">
                              {apiKey.key_id}
                            </code>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => copyToClipboard(apiKey.key_id, 'Key ID')}
                              className="shrink-0 border-slate-200 text-slate-600 hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
                            >
                              <Copy className="h-4 w-4" />
                            </Button>
                          </div>
                        </div>

                        {/* Scopes */}
                        {apiKey.scopes && apiKey.scopes.length > 0 && (
                          <div>
                            <Label className="text-xs text-slate-500 dark:text-slate-400 font-medium">Scopes</Label>
                            <div className="mt-1 flex flex-wrap gap-2">
                              {apiKey.scopes.map((scope, scopeIndex) => (
                                <Badge key={scopeIndex} variant="secondary" className="bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300">
                                  {scope}
                                </Badge>
                              ))}
                            </div>
                          </div>
                        )}

                        {/* Expiration */}
                        {apiKey.expires_at && (
                          <div>
                            <Label className="text-xs text-slate-500 dark:text-slate-400 font-medium flex items-center space-x-1">
                              <Calendar className="h-3 w-3" />
                              <span>Expires</span>
                            </Label>
                            <div className="mt-1 text-sm text-slate-700 dark:text-slate-300">
                              {formatDate(apiKey.expires_at)}
                            </div>
                          </div>
                        )}

                        {/* Last Used */}
                        {apiKey.last_used_at && (
                          <div>
                            <Label className="text-xs text-slate-500 dark:text-slate-400 font-medium">Last Used</Label>
                            <div className="mt-1 text-sm text-slate-600 dark:text-slate-400">
                              {formatDate(apiKey.last_used_at)}
                            </div>
                          </div>
                        )}

                        {/* Metadata */}
                        <div className="flex justify-between text-xs text-slate-500 dark:text-slate-400 pt-2 border-t border-slate-200 dark:border-slate-600">
                          <span>Created: {formatDate(apiKey.created_at)}</span>
                          <span>Updated: {formatDate(apiKey.updated_at)}</span>
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
        <DialogContent className="max-w-2xl bg-gradient-to-br from-white to-emerald-50/30 dark:from-slate-800 dark:to-slate-700 border border-emerald-200/50 dark:border-slate-600/50">
          <DialogHeader className="space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-gradient-to-r from-emerald-500 to-teal-600 rounded-lg shadow-lg">
                <Plus className="h-6 w-6 text-white" />
              </div>
              <div>
                <DialogTitle className="text-xl font-bold bg-gradient-to-r from-slate-800 to-slate-600 dark:from-white dark:to-slate-200 bg-clip-text text-transparent">
                  Create API Key
                </DialogTitle>
                <DialogDescription className="text-slate-600 dark:text-slate-300">
                  Create a new API key for secure programmatic access to your account
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="name" className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center space-x-2">
                  <span>API Key Name</span>
                  <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="My API Key"
                  className="border-slate-200 dark:border-slate-600 focus:border-emerald-500 dark:focus:border-emerald-400 bg-white dark:bg-slate-800"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="expires_at" className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center space-x-2">
                  <Calendar className="h-4 w-4" />
                  <span>Expiration Date</span>
                </Label>
                <Input
                  id="expires_at"
                  type="date"
                  value={formData.expires_at}
                  onChange={(e) => setFormData({ ...formData, expires_at: e.target.value })}
                  className="border-slate-200 dark:border-slate-600 focus:border-emerald-500 dark:focus:border-emerald-400 bg-white dark:bg-slate-800"
                />
              </div>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="description" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                Description
              </Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="What this API key will be used for"
                rows={3}
                className="border-slate-200 dark:border-slate-600 focus:border-emerald-500 dark:focus:border-emerald-400 bg-white dark:bg-slate-800 resize-none"
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="scopes" className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center space-x-2">
                <Lock className="h-4 w-4" />
                <span>Scopes</span>
              </Label>
              <Input
                id="scopes"
                value={formData.scopes}
                onChange={(e) => setFormData({ ...formData, scopes: e.target.value })}
                placeholder="profile, email, read:data, write:data"
                className="border-slate-200 dark:border-slate-600 focus:border-emerald-500 dark:focus:border-emerald-400 bg-white dark:bg-slate-800"
              />
              <p className="text-xs text-slate-500 dark:text-slate-400">
                Comma-separated list of permissions this API key can access.
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
              onClick={handleCreateKey} 
              disabled={!formData.name}
              className="bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 text-white shadow-lg hover:shadow-xl transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Plus className="h-4 w-4 mr-2" />
              Create API Key
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-lg bg-gradient-to-br from-white to-red-50/30 dark:from-slate-800 dark:to-red-900/10 border border-red-200/50 dark:border-red-700/50">
          <DialogHeader className="space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-gradient-to-r from-red-500 to-red-600 rounded-lg shadow-lg">
                <AlertCircle className="h-6 w-6 text-white" />
              </div>
              <div>
                <DialogTitle className="text-xl font-bold text-red-600 dark:text-red-400">
                  Revoke API Key
                </DialogTitle>
                <DialogDescription className="text-slate-600 dark:text-slate-300">
                  This action cannot be undone. The API key will be permanently disabled.
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>
          <div className="py-4">
            <p className="text-slate-700 dark:text-slate-300">
              Are you sure you want to revoke <strong>"{selectedKey?.name}"</strong>?
            </p>
          </div>
          <DialogFooter className="space-x-3">
            <Button 
              variant="outline" 
              onClick={() => setDeleteDialogOpen(false)}
              className="border-slate-200 text-slate-600 hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
            >
              Cancel
            </Button>
            <Button 
              variant="destructive" 
              onClick={handleDeleteKey}
              className="bg-gradient-to-r from-red-600 to-red-700 hover:from-red-700 hover:to-red-800"
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Revoke API Key
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
                  API Key Created Successfully
                </DialogTitle>
                <DialogDescription className="text-slate-600 dark:text-slate-300">
                  Save this API key securely. It will not be shown again!
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>
          <div className="space-y-6">
            <Alert className="border-amber-200 bg-amber-50 dark:border-amber-700 dark:bg-amber-900/20">
              <AlertCircle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
              <AlertDescription className="text-amber-800 dark:text-amber-200">
                <strong>Important:</strong> Copy and save this API key in a secure location. 
                You won't be able to see it again after closing this dialog.
              </AlertDescription>
            </Alert>
            <div className="space-y-2">
              <Label className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center space-x-2">
                <Key className="h-4 w-4" />
                <span>API Key</span>
              </Label>
              <div className="flex items-center space-x-2">
                <code className="flex-1 text-sm bg-slate-100 dark:bg-slate-800 px-4 py-3 rounded-lg border border-slate-200 dark:border-slate-600 break-all font-mono text-slate-800 dark:text-slate-200">
                  {secretVisible ? newSecret : newSecret.replace(/./g, '*')}
                </code>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setSecretVisible(!secretVisible)}
                  className="shrink-0 border-slate-200 text-slate-600 hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
                >
                  {secretVisible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => copyToClipboard(newSecret, 'API Key')}
                  className="shrink-0 border-slate-200 text-slate-600 hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
          <DialogFooter className="pt-6 border-t border-slate-200 dark:border-slate-600">
            <Button 
              onClick={() => setSecretDialogOpen(false)}
              className="bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 text-white shadow-lg hover:shadow-xl transition-all duration-200"
            >
              <CheckCircle className="h-4 w-4 mr-2" />
              I've Saved It
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default ApiKeys
    