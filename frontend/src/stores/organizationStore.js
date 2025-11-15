import { create } from 'zustand';
import { organizationAPI } from '../services/api';

const useOrganizationStore = create((set, get) => ({
  // State
  organizations: [],
  currentOrganization: null,
  members: [],
  invitations: [],
  loading: false,
  error: null,

  // Actions
  setOrganizations: (organizations) => set({ organizations }),
  setCurrentOrganization: (organization) => set({ currentOrganization: organization }),
  setMembers: (members) => set({ members }),
  setInvitations: (invitations) => set({ invitations }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),

  // Organization actions
  fetchOrganizations: async () => {
    try {
      set({ loading: true, error: null });
      const response = await organizationAPI.listOrganizations();
      set({
        organizations: response.data.data,
        loading: false,
        error: null,
      });
      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to fetch organizations';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  createOrganization: async (orgData) => {
    try {
      set({ loading: true, error: null });
      const response = await organizationAPI.createOrganization(orgData);
      const newOrg = response.data.data;

      // Add to organizations list
      const { organizations } = get();
      set({
        organizations: [...organizations, newOrg],
        loading: false,
        error: null,
      });

      return { success: true, data: newOrg };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to create organization';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  fetchOrganization: async (orgId) => {
    try {
      set({ loading: true, error: null });
      const response = await organizationAPI.getOrganization(orgId);
      set({
        currentOrganization: response.data.data,
        loading: false,
        error: null,
      });
      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to fetch organization';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  updateOrganization: async (orgId, orgData) => {
    try {
      set({ loading: true, error: null });
      const response = await organizationAPI.updateOrganization(orgId, orgData);
      const updatedOrg = response.data.data;

      // Update in organizations list
      const { organizations } = get();
      const updatedOrganizations = organizations.map(org =>
        org.id === orgId ? updatedOrg : org
      );

      set({
        organizations: updatedOrganizations,
        currentOrganization: updatedOrg,
        loading: false,
        error: null,
      });

      return { success: true, data: updatedOrg };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to update organization';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  deleteOrganization: async (orgId) => {
    try {
      set({ loading: true, error: null });
      await organizationAPI.deleteOrganization(orgId);

      // Remove from organizations list
      const { organizations } = get();
      const filteredOrganizations = organizations.filter(org => org.id !== orgId);

      set({
        organizations: filteredOrganizations,
        currentOrganization: null,
        loading: false,
        error: null,
      });

      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to delete organization';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  // Member actions
  fetchMembers: async (orgId) => {
    try {
      set({ loading: true, error: null });
      const response = await organizationAPI.listMembers(orgId);
      set({
        members: response.data.data,
        loading: false,
        error: null,
      });
      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to fetch members';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  inviteUser: async (orgId, email, role) => {
    try {
      set({ loading: true, error: null });
      await organizationAPI.inviteUser(orgId, { email, role });
      set({ loading: false, error: null });

      // Refresh invitations
      get().fetchInvitations(orgId);

      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to invite user';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  updateMemberRole: async (orgId, userId, role) => {
    try {
      set({ loading: true, error: null });
      await organizationAPI.updateMember(orgId, userId, { role });

      // Refresh members
      get().fetchMembers(orgId);

      set({ loading: false, error: null });
      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to update member role';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  removeMember: async (orgId, userId) => {
    try {
      set({ loading: true, error: null });
      await organizationAPI.removeMember(orgId, userId);

      // Refresh members
      get().fetchMembers(orgId);

      set({ loading: false, error: null });
      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to remove member';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  // Invitation actions
  fetchInvitations: async (orgId) => {
    try {
      set({ loading: true, error: null });
      const response = await organizationAPI.listInvitations(orgId);
      set({
        invitations: response.data.data,
        loading: false,
        error: null,
      });
      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to fetch invitations';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  cancelInvitation: async (orgId, invitationId) => {
    try {
      set({ loading: true, error: null });
      await organizationAPI.cancelInvitation(orgId, invitationId);

      // Refresh invitations
      get().fetchInvitations(orgId);

      set({ loading: false, error: null });
      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to cancel invitation';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  acceptInvitation: async (token) => {
    try {
      set({ loading: true, error: null });
      const response = await organizationAPI.acceptInvitation(token);
      set({ loading: false, error: null });

      // Refresh organizations
      get().fetchOrganizations();

      return { success: true, data: response.data.data };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to accept invitation';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  getInvitationDetails: async (token) => {
    try {
      set({ loading: true, error: null });
      const response = await organizationAPI.getInvitationDetails(token);
      set({ loading: false, error: null });
      return { success: true, data: response.data.data };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to get invitation details';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Reset state
  reset: () => set({
    organizations: [],
    currentOrganization: null,
    members: [],
    invitations: [],
    loading: false,
    error: null,
  }),
}));

export default useOrganizationStore;