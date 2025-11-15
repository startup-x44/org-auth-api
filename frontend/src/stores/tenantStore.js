import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

const useTenantStore = create(
  persist(
    (set, get) => ({
      // State
      currentTenant: null,
      tenantId: null,
      tenantDomain: null,
      availableTenants: [],
      loading: false,
      error: null,

      // Actions
      setCurrentTenant: (tenant) => set({
        currentTenant: tenant,
        tenantId: tenant?.id || null,
        tenantDomain: tenant?.domain || null
      }),

      setTenantId: (tenantId) => set({ tenantId }),
      setTenantDomain: (domain) => set({ tenantDomain: domain }),
      setAvailableTenants: (tenants) => set({ availableTenants: tenants }),
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),

      // Tenant resolution utilities
      resolveTenantFromSubdomain: () => {
        const hostname = window.location.hostname;
        const parts = hostname.split('.');

        // Handle subdomains like "company.sprout.com"
        if (parts.length > 2) {
          const subdomain = parts[0];
          if (subdomain !== 'www' && subdomain !== 'app') {
            return subdomain;
          }
        }

        return null;
      },

      resolveTenantFromEmail: (email) => {
        if (!email) return null;
        const domain = email.split('@')[1];
        return domain;
      },

      // Initialize tenant context
      initializeTenant: () => {
        const subdomain = get().resolveTenantFromSubdomain();
        if (subdomain) {
          set({ tenantDomain: `${subdomain}.sprout.com` });
        }
      },

      // Clear tenant data
      clearTenant: () => set({
        currentTenant: null,
        tenantId: null,
        tenantDomain: null,
        availableTenants: [],
        error: null,
      }),

      // Clear error
      clearError: () => set({ error: null }),
    }),
    {
      name: 'tenant-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        currentTenant: state.currentTenant,
        tenantId: state.tenantId,
        tenantDomain: state.tenantDomain,
      }),
    }
  )
);

export default useTenantStore;