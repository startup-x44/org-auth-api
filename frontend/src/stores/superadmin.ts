import { create } from 'zustand'
import { User } from '../lib/auth'

interface SuperAdminState {
  sidebarOpen: boolean
  currentUser: User | null
  users: User[]
  organizations: any[]
  loading: boolean
  error: string | null
  
  // Actions
  setSidebarOpen: (open: boolean) => void
  setCurrentUser: (user: User | null) => void
  setUsers: (users: User[]) => void
  setOrganizations: (organizations: any[]) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
}

export const useSuperAdminStore = create<SuperAdminState>((set) => ({
  sidebarOpen: true,
  currentUser: null,
  users: [],
  organizations: [],
  loading: false,
  error: null,
  
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  setCurrentUser: (user) => set({ currentUser: user }),
  setUsers: (users) => set({ users }),
  setOrganizations: (organizations) => set({ organizations }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  clearError: () => set({ error: null }),
}))