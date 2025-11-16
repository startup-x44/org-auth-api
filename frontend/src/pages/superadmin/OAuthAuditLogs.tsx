import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import useAuthStore from '@/store/auth'
import { oauthAuditAPI, clientAppAPI } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ArrowLeft, Activity, Key, Users, TrendingUp, Filter } from 'lucide-react'

interface AuditLog {
  id: string
  client_id: string
  client_name: string
  user_id: string
  user_email: string
  organization_id?: string
  scope: string // singular, not scopes
  status: string // 'used' | 'expired' | 'active'
  created_at: string
}

interface TokenGrant {
  id: string
  client_id: string
  client_name: string
  user_id: string
  user_email: string
  organization_id?: string
  scope: string // singular, not scopes
  grant_type: string
  is_active: boolean
  created_at: string
  expires_at: string
}

interface AuditStats {
  total_authorizations: number
  authorizations_today: number
  active_tokens: number
  total_tokens: number
  unique_users: number
  unique_clients: number
}

export default function OAuthAuditLogs() {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [activeTab, setActiveTab] = useState<'authorizations' | 'tokens' | 'stats'>('authorizations')
  
  // Authorizations state
  const [authorizations, setAuthorizations] = useState<AuditLog[]>([])
  const [authLoading, setAuthLoading] = useState(true)
  const [authFilters, setAuthFilters] = useState({ client_id: '', user_id: '' })

  // Tokens state
  const [tokens, setTokens] = useState<TokenGrant[]>([])
  const [tokenLoading, setTokenLoading] = useState(false)
  const [tokenFilters, setTokenFilters] = useState({ client_id: '', user_id: '' })

  // Stats state
  const [stats, setStats] = useState<AuditStats | null>(null)
  const [statsLoading, setStatsLoading] = useState(false)

  // Client apps for filter dropdown
  const [clientApps, setClientApps] = useState<any[]>([])

  // Redirect if not superadmin
  useEffect(() => {
    if (user?.user_type !== 'superadmin') {
      navigate('/dashboard')
    }
  }, [user, navigate])

  // Load client apps for filters
  useEffect(() => {
    clientAppAPI.listClientApps({ limit: 100 })
      .then(res => {
        if (res.success) {
          setClientApps(res.data.client_apps || [])
        }
      })
      .catch(err => console.error('Failed to load client apps:', err))
  }, [])

  // Load authorizations
  useEffect(() => {
    if (activeTab === 'authorizations') {
      loadAuthorizations()
    }
  }, [activeTab, authFilters])

  // Load tokens
  useEffect(() => {
    if (activeTab === 'tokens') {
      loadTokens()
    }
  }, [activeTab, tokenFilters])

  // Load stats
  useEffect(() => {
    if (activeTab === 'stats') {
      loadStats()
    }
  }, [activeTab])

  const loadAuthorizations = async () => {
    setAuthLoading(true)
    try {
      const params: any = { limit: 50 }
      if (authFilters.client_id) params.client_id = authFilters.client_id
      if (authFilters.user_id) params.user_id = authFilters.user_id

      const res = await oauthAuditAPI.listAuthorizations(params)
      if (res.success) {
        // Backend returns logs directly in data array
        setAuthorizations(res.data || [])
      }
    } catch (err) {
      console.error('Failed to load authorizations:', err)
      setAuthorizations([])
    } finally {
      setAuthLoading(false)
    }
  }

  const loadTokens = async () => {
    setTokenLoading(true)
    try {
      const params: any = { limit: 50 }
      if (tokenFilters.client_id) params.client_id = tokenFilters.client_id
      if (tokenFilters.user_id) params.user_id = tokenFilters.user_id

      const res = await oauthAuditAPI.listTokens(params)
      if (res.success) {
        // Backend returns tokens directly in data array
        setTokens(res.data || [])
      }
    } catch (err) {
      console.error('Failed to load tokens:', err)
      setTokens([])
    } finally {
      setTokenLoading(false)
    }
  }

  const loadStats = async () => {
    setStatsLoading(true)
    try {
      const res = await oauthAuditAPI.getStats()
      if (res.success) {
        setStats(res.data)
      }
    } catch (err) {
      console.error('Failed to load stats:', err)
    } finally {
      setStatsLoading(false)
    }
  }

  const formatDate = (date: string) => {
    return new Date(date).toLocaleString()
  }

  const getStatusBadge = (status: string) => {
    const styles = {
      used: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      expired: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
      active: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    }
    return (
      <span className={`px-2 py-1 rounded text-xs font-medium ${styles[status as keyof typeof styles] || styles.active}`}>
        {status.toUpperCase()}
      </span>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <Button
          variant="ghost"
          onClick={() => navigate('/admin')}
          className="mb-4"
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Admin
        </Button>
        <h1 className="text-3xl font-bold">OAuth Audit Logs</h1>
        <p className="text-muted-foreground mt-2">
          Monitor OAuth authorization flows, active tokens, and security events
        </p>
      </div>

      {/* Tabs */}
      <div className="border-b mb-6">
        <div className="flex space-x-4">
          <button
            onClick={() => setActiveTab('authorizations')}
            className={`pb-4 px-2 border-b-2 transition-colors ${
              activeTab === 'authorizations'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Activity className="inline-block mr-2 h-4 w-4" />
            Authorizations
          </button>
          <button
            onClick={() => setActiveTab('tokens')}
            className={`pb-4 px-2 border-b-2 transition-colors ${
              activeTab === 'tokens'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Key className="inline-block mr-2 h-4 w-4" />
            Active Tokens
          </button>
          <button
            onClick={() => setActiveTab('stats')}
            className={`pb-4 px-2 border-b-2 transition-colors ${
              activeTab === 'stats'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <TrendingUp className="inline-block mr-2 h-4 w-4" />
            Statistics
          </button>
        </div>
      </div>

      {/* Authorizations Tab */}
      {activeTab === 'authorizations' && (
        <div>
          {/* Filters */}
          <div className="bg-card rounded-lg border p-4 mb-6">
            <div className="flex items-center mb-3">
              <Filter className="mr-2 h-4 w-4 text-muted-foreground" />
              <h3 className="font-medium">Filters</h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">Client App</label>
                <select
                  value={authFilters.client_id}
                  onChange={(e) => setAuthFilters({ ...authFilters, client_id: e.target.value })}
                  className="w-full px-3 py-2 border rounded-md bg-background"
                >
                  <option value="">All Apps</option>
                  {clientApps.map(app => (
                    <option key={app.id} value={app.id}>{app.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">User ID</label>
                <Input
                  placeholder="Filter by user ID..."
                  value={authFilters.user_id}
                  onChange={(e) => setAuthFilters({ ...authFilters, user_id: e.target.value })}
                />
              </div>
            </div>
          </div>

          {/* Authorizations Table */}
          {authLoading ? (
            <div className="text-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
              <p className="mt-4 text-muted-foreground">Loading authorizations...</p>
            </div>
          ) : authorizations.length === 0 ? (
            <div className="text-center py-12">
              <Activity className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No authorization logs found</p>
            </div>
          ) : (
            <div className="bg-card rounded-lg border overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-muted">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-medium">Client App</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">User</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Scopes</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Status</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Created</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Used At</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {authorizations.map((log) => (
                      <tr key={log.id} className="hover:bg-muted/50">
                        <td className="px-4 py-3 text-sm font-medium">{log.client_name}</td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">{log.user_email}</td>
                        <td className="px-4 py-3 text-sm">
                          <code className="text-xs bg-muted px-2 py-1 rounded">{log.scope}</code>
                        </td>
                        <td className="px-4 py-3">{getStatusBadge(log.status)}</td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">{formatDate(log.created_at)}</td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">
                          {log.status === 'used' ? 'Used' : '-'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tokens Tab */}
      {activeTab === 'tokens' && (
        <div>
          {/* Filters */}
          <div className="bg-card rounded-lg border p-4 mb-6">
            <div className="flex items-center mb-3">
              <Filter className="mr-2 h-4 w-4 text-muted-foreground" />
              <h3 className="font-medium">Filters</h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">Client App</label>
                <select
                  value={tokenFilters.client_id}
                  onChange={(e) => setTokenFilters({ ...tokenFilters, client_id: e.target.value })}
                  className="w-full px-3 py-2 border rounded-md bg-background"
                >
                  <option value="">All Apps</option>
                  {clientApps.map(app => (
                    <option key={app.id} value={app.id}>{app.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">User ID</label>
                <Input
                  placeholder="Filter by user ID..."
                  value={tokenFilters.user_id}
                  onChange={(e) => setTokenFilters({ ...tokenFilters, user_id: e.target.value })}
                />
              </div>
            </div>
          </div>

          {/* Tokens Table */}
          {tokenLoading ? (
            <div className="text-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
              <p className="mt-4 text-muted-foreground">Loading tokens...</p>
            </div>
          ) : tokens.length === 0 ? (
            <div className="text-center py-12">
              <Key className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No active tokens found</p>
            </div>
          ) : (
            <div className="bg-card rounded-lg border overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-muted">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-medium">Client App</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">User</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Scopes</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Status</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Created</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Last Used</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">Expires</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {tokens.map((token) => (
                      <tr key={token.id} className="hover:bg-muted/50">
                        <td className="px-4 py-3 text-sm font-medium">{token.client_name}</td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">{token.user_email}</td>
                        <td className="px-4 py-3 text-sm">
                          <code className="text-xs bg-muted px-2 py-1 rounded">{token.scope}</code>
                        </td>
                        <td className="px-4 py-3">
                          {token.is_active ? (
                            <span className="px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                              ACTIVE
                            </span>
                          ) : (
                            <span className="px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300">
                              INACTIVE
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">{formatDate(token.created_at)}</td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">
                          {token.grant_type === 'refresh_token' ? 'Refresh Token' : 'Never'}
                        </td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">{formatDate(token.expires_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Statistics Tab */}
      {activeTab === 'stats' && (
        <div>
          {statsLoading ? (
            <div className="text-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
              <p className="mt-4 text-muted-foreground">Loading statistics...</p>
            </div>
          ) : stats ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {/* Total Authorizations */}
              <div className="bg-card rounded-lg border p-6">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium text-muted-foreground">Total Authorizations</h3>
                  <Activity className="h-4 w-4 text-muted-foreground" />
                </div>
                <p className="text-3xl font-bold">{stats.total_authorizations}</p>
                <p className="text-sm text-muted-foreground mt-2">
                  {stats.authorizations_today} today
                </p>
              </div>

              {/* Active Tokens */}
              <div className="bg-card rounded-lg border p-6">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium text-muted-foreground">Active Tokens</h3>
                  <Key className="h-4 w-4 text-muted-foreground" />
                </div>
                <p className="text-3xl font-bold">{stats.active_tokens}</p>
                <p className="text-sm text-muted-foreground mt-2">
                  of {stats.total_tokens} total
                </p>
              </div>

              {/* Unique Users */}
              <div className="bg-card rounded-lg border p-6">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium text-muted-foreground">Unique Users</h3>
                  <Users className="h-4 w-4 text-muted-foreground" />
                </div>
                <p className="text-3xl font-bold">{stats.unique_users}</p>
                <p className="text-sm text-muted-foreground mt-2">
                  across {stats.unique_clients} client apps
                </p>
              </div>
            </div>
          ) : (
            <div className="text-center py-12">
              <TrendingUp className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">Failed to load statistics</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
