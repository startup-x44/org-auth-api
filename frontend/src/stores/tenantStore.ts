import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

import type { Organization, User } from '@/types/auth'
import { apiClient } from '@/api/client'
import { secureStorage } from '@/utils/storage'
import { logger } from '@/utils/logger'

// Tenant-specific types
export interface TenantMember {
  id: string
  user: User
  roles: string[]
  status: 'active' | 'inactive' | 'invited' | 'suspended'
  invitedBy?: string
  invitedAt?: string
  joinedAt?: string
  lastSeenAt?: string
}

export interface TenantInvitation {
  id: string
  email: string
  roles: string[]
  invitedBy: User
  invitedAt: string
  expiresAt: string
  status: 'pending' | 'accepted' | 'expired' | 'revoked'
}

export interface TenantUsage {
  users: {
    current: number
    limit: number
  }
  storage: {
    current: number
    limit: number
  }
  apiCalls: {
    current: number
    limit: number
    resetDate: string
  }
}

export interface TenantBilling {
  subscription: {
    plan: string
    status: 'active' | 'canceled' | 'past_due' | 'trial'
    currentPeriodStart: string
    currentPeriodEnd: string
    cancelAtPeriodEnd: boolean
  }
  nextInvoice?: {
    amount: number
    currency: string
    dueDate: string
  }
  paymentMethod?: {
    type: string
    last4: string
    expiresAt: string
  }
}

export interface TenantAuditLog {
  id: string
  userId: string
  userName: string
  action: string
  resource: string
  resourceId?: string
  metadata?: Record<string, any>
  ip: string
  userAgent: string
  timestamp: string
}

// Store state
export interface TenantState {
  // Current tenant context
  currentTenant: Organization | null
  availableTenants: Organization[]
  
  // Tenant management
  members: TenantMember[]
  invitations: TenantInvitation[]
  
  // Tenant analytics
  usage: TenantUsage | null
  billing: TenantBilling | null
  auditLogs: TenantAuditLog[]
  
  // UI state
  isLoading: boolean
  isLoadingMembers: boolean
  isLoadingUsage: boolean
  error: string | null
  
  // Permissions for current user in current tenant
  currentUserPermissions: string[]
  currentUserRoles: string[]
}

// Store actions
export interface TenantActions {
  // Tenant switching
  switchTenant: (tenantId: string) => Promise<void>
  refreshTenants: () => Promise<void>
  
  // Member management
  loadMembers: () => Promise<void>
  inviteMember: (email: string, roles: string[]) => Promise<void>
  removeMember: (memberId: string) => Promise<void>
  updateMemberRoles: (memberId: string, roles: string[]) => Promise<void>
  resendInvitation: (invitationId: string) => Promise<void>
  revokeInvitation: (invitationId: string) => Promise<void>
  
  // Tenant settings
  updateTenantSettings: (settings: Partial<Organization['settings']>) => Promise<void>
  
  // Usage and billing
  loadUsage: () => Promise<void>
  loadBilling: () => Promise<void>
  loadAuditLogs: (filters?: any) => Promise<void>
  
  // Permissions
  refreshPermissions: () => Promise<void>
  hasPermission: (permission: string) => boolean
  hasRole: (role: string) => boolean
  
  // Utility
  initialize: (user?: User) => Promise<void>
  clearError: () => void
  setError: (error: string) => void
}

export type TenantStore = TenantState & TenantActions

// Initial state
const initialState: TenantState = {
  currentTenant: null,
  availableTenants: [],
  members: [],
  invitations: [],
  usage: null,
  billing: null,
  auditLogs: [],
  isLoading: false,
  isLoadingMembers: false,
  isLoadingUsage: false,
  error: null,
  currentUserPermissions: [],
  currentUserRoles: [],
}

// Store implementation
export const useTenantStore = create<TenantStore>()(
  devtools(
    persist(
      immer((set, get) => ({
        ...initialState,

        switchTenant: async (tenantId: string) => {
          try {
            set((state) => {
              state.isLoading = true
              state.error = null
            })

            logger.info('Switching tenant', { tenantId })

            // Find the tenant in available tenants
            const tenant = get().availableTenants.find(t => t.id === tenantId)
            if (!tenant) {
              throw new Error('Tenant not found')
            }

            // Call API to switch tenant context
            await apiClient.post('/v1/auth/select-organization', { organizationId: tenantId })

            // Update secure storage
            secureStorage.setCurrentOrganization(tenant)

            // Update state
            set((state) => {
              state.currentTenant = tenant
              state.isLoading = false
              // Reset tenant-specific data
              state.members = []
              state.invitations = []
              state.usage = null
              state.billing = null
              state.auditLogs = []
              state.currentUserPermissions = []
              state.currentUserRoles = []
            })

            // Refresh permissions for new tenant
            await get().refreshPermissions()

            logger.info('Tenant switch successful', { 
              tenantId, 
              tenantName: tenant.name 
            })
          } catch (error: any) {
            logger.error('Tenant switch failed', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'Failed to switch tenant'
              state.isLoading = false
            })
            
            throw error
          }
        },

        refreshTenants: async () => {
          try {
            logger.debug('Refreshing available tenants')

            const response = await apiClient.get<{ organizations: Organization[] }>('/organizations/me')
            
            set((state) => {
              state.availableTenants = response.data.organizations
            })

            logger.debug('Tenants refreshed', { 
              count: response.data.organizations.length 
            })
          } catch (error: any) {
            logger.error('Failed to refresh tenants', error)
            throw error
          }
        },

        loadMembers: async () => {
          try {
            set((state) => {
              state.isLoadingMembers = true
              state.error = null
            })

            const response = await apiClient.get<{ 
              members: TenantMember[]
              invitations: TenantInvitation[]
            }>('/organizations/current/members')
            
            set((state) => {
              state.members = response.data.members
              state.invitations = response.data.invitations
              state.isLoadingMembers = false
            })

            logger.debug('Members loaded', { 
              memberCount: response.data.members.length,
              invitationCount: response.data.invitations.length 
            })
          } catch (error: any) {
            logger.error('Failed to load members', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'Failed to load members'
              state.isLoadingMembers = false
            })
            
            throw error
          }
        },

        inviteMember: async (email: string, roles: string[]) => {
          try {
            logger.info('Inviting member', { email, roles })

            const response = await apiClient.post<{ invitation: TenantInvitation }>('/organizations/current/invitations', {
              email,
              roles,
            })

            set((state) => {
              state.invitations.push(response.data.invitation)
            })

            logger.info('Member invitation sent', { email })
          } catch (error: any) {
            logger.error('Failed to invite member', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'Failed to invite member'
            })
            
            throw error
          }
        },

        removeMember: async (memberId: string) => {
          try {
            logger.info('Removing member', { memberId })

            await apiClient.delete(`/organizations/current/members/${memberId}`)

            set((state) => {
              state.members = state.members.filter((m: TenantMember) => m.id !== memberId)
            })

            logger.info('Member removed successfully', { memberId })
          } catch (error: any) {
            logger.error('Failed to remove member', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'Failed to remove member'
            })
            
            throw error
          }
        },

        updateMemberRoles: async (memberId: string, roles: string[]) => {
          try {
            logger.info('Updating member roles', { memberId, roles })

            const response = await apiClient.patch<{ member: TenantMember }>(`/organizations/current/members/${memberId}`, {
              roles,
            })

            set((state) => {
              const index = state.members.findIndex((m: TenantMember) => m.id === memberId)
              if (index !== -1) {
                state.members[index] = response.data.member
              }
            })

            logger.info('Member roles updated successfully', { memberId })
          } catch (error: any) {
            logger.error('Failed to update member roles', error)
            
            set((state) => {
              state.error = error.response?.data?.message || 'Failed to update member roles'
            })
            
            throw error
          }
        },

        resendInvitation: async (invitationId: string) => {
          try {
            await apiClient.post(`/organizations/current/invitations/${invitationId}/resend`)
            logger.info('Invitation resent', { invitationId })
          } catch (error: any) {
            logger.error('Failed to resend invitation', error)
            throw error
          }
        },

        revokeInvitation: async (invitationId: string) => {
          try {
            await apiClient.delete(`/organizations/current/invitations/${invitationId}`)
            
            set((state) => {
              state.invitations = state.invitations.filter((i: TenantInvitation) => i.id !== invitationId)
            })

            logger.info('Invitation revoked', { invitationId })
          } catch (error: any) {
            logger.error('Failed to revoke invitation', error)
            throw error
          }
        },

        updateTenantSettings: async (settings: Partial<Organization['settings']>) => {
          try {
            const response = await apiClient.patch<{ organization: Organization }>('/organizations/current/settings', settings)
            
            set((state) => {
              if (state.currentTenant) {
                state.currentTenant.settings = response.data.organization.settings
              }
            })

            logger.info('Tenant settings updated')
          } catch (error: any) {
            logger.error('Failed to update tenant settings', error)
            throw error
          }
        },

        loadUsage: async () => {
          try {
            set((state) => {
              state.isLoadingUsage = true
            })

            const response = await apiClient.get<{ usage: TenantUsage }>('/organizations/current/usage')
            
            set((state) => {
              state.usage = response.data.usage
              state.isLoadingUsage = false
            })
          } catch (error: any) {
            logger.error('Failed to load usage', error)
            
            set((state) => {
              state.isLoadingUsage = false
            })
          }
        },

        loadBilling: async () => {
          try {
            const response = await apiClient.get<{ billing: TenantBilling }>('/organizations/current/billing')
            
            set((state) => {
              state.billing = response.data.billing
            })
          } catch (error: any) {
            logger.error('Failed to load billing', error)
          }
        },

        loadAuditLogs: async (filters?: any) => {
          try {
            const response = await apiClient.get<{ logs: TenantAuditLog[] }>('/organizations/current/audit-logs', {
              params: filters,
            })
            
            set((state) => {
              state.auditLogs = response.data.logs
            })
          } catch (error: any) {
            logger.error('Failed to load audit logs', error)
          }
        },

        refreshPermissions: async () => {
          try {
            const response = await apiClient.get<{ 
              permissions: string[]
              roles: string[]
            }>('/organizations/current/permissions/me')
            
            set((state) => {
              state.currentUserPermissions = response.data.permissions
              state.currentUserRoles = response.data.roles
            })
          } catch (error: any) {
            logger.error('Failed to refresh permissions', error)
          }
        },

        hasPermission: (permission: string) => {
          return get().currentUserPermissions.includes(permission)
        },

        hasRole: (role: string) => {
          return get().currentUserRoles.includes(role)
        },

        initialize: async (user?: User) => {
          // Prevent multiple simultaneous initializations
          const state = get()
          if (state.isLoading) {
            return
          }

          try {
            set((state) => {
              state.isLoading = true
            })

            // Check if user is superadmin - they don't need organization data
            if (user?.is_superadmin) {
              logger.info('Superadmin detected - skipping tenant initialization')
              set((state) => {
                state.isLoading = false
              })
              return
            }

            // Get current organization from secure storage
            const currentOrg = secureStorage.getCurrentOrganization()
            
            if (currentOrg) {
              set((state) => {
                state.currentTenant = currentOrg
              })
            }

            // Refresh available tenants
            await get().refreshTenants()

            // If we have a current tenant, refresh its permissions
            if (currentOrg) {
              await get().refreshPermissions()
            }

            set((state) => {
              state.isLoading = false
            })

            logger.info('Tenant store initialized')
          } catch (error: any) {
            logger.error('Failed to initialize tenant store', error)
            set((state) => {
              state.isLoading = false
            })
          }
        },

        clearError: () => {
          set((state) => {
            state.error = null
          })
        },

        setError: (error: string) => {
          set((state) => {
            state.error = error
          })
        },
      })),
      {
        name: 'tenant-store',
        partialize: (state) => ({
          currentTenant: state.currentTenant,
        }),
      }
    ),
    {
      name: 'tenant-store',
    }
  )
)

// Helper hooks
export const useTenant = () => {
  const store = useTenantStore()
  return {
    currentTenant: store.currentTenant,
    availableTenants: store.availableTenants,
    isLoading: store.isLoading,
    error: store.error,
    switchTenant: store.switchTenant,
    refreshTenants: store.refreshTenants,
    clearError: store.clearError,
  }
}

export const useCurrentTenant = () => useTenantStore((state) => state.currentTenant)
export const useTenantMembers = () => useTenantStore((state) => ({ 
  members: state.members, 
  invitations: state.invitations,
  isLoading: state.isLoadingMembers 
}))
export const useTenantPermissions = () => useTenantStore((state) => ({
  permissions: state.currentUserPermissions,
  roles: state.currentUserRoles,
  hasPermission: state.hasPermission,
  hasRole: state.hasRole,
}))

export default useTenantStore