import { create } from 'zustand';
import { adminAPI } from '../services/api';

const useAdminStore = create((set, get) => ({
  // State
  users: [],
  totalUsers: 0,
  nextCursor: null,
  loading: false,
  error: null,

  // Actions
  setUsers: (users) => set({ users }),
  setTotalUsers: (total) => set({ totalUsers: total }),
  setNextCursor: (cursor) => set({ nextCursor: cursor }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),

  // User management actions
  fetchUsers: async (params = {}) => {
    try {
      set({ loading: true, error: null });
      const response = await adminAPI.listUsers(params);
      const { users, total, next_cursor } = response.data.data;

      set({
        users,
        totalUsers: total,
        nextCursor: next_cursor,
        loading: false,
        error: null,
      });

      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to fetch users';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  activateUser: async (userId) => {
    try {
      set({ loading: true, error: null });
      await adminAPI.activateUser(userId);

      // Update user in list
      const { users } = get();
      const updatedUsers = users.map(user =>
        user.id === userId ? { ...user, is_active: true } : user
      );

      set({
        users: updatedUsers,
        loading: false,
        error: null,
      });

      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to activate user';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  deactivateUser: async (userId) => {
    try {
      set({ loading: true, error: null });
      await adminAPI.deactivateUser(userId);

      // Update user in list
      const { users } = get();
      const updatedUsers = users.map(user =>
        user.id === userId ? { ...user, is_active: false } : user
      );

      set({
        users: updatedUsers,
        loading: false,
        error: null,
      });

      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to deactivate user';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  deleteUser: async (userId) => {
    try {
      set({ loading: true, error: null });
      await adminAPI.deleteUser(userId);

      // Remove user from list
      const { users } = get();
      const filteredUsers = users.filter(user => user.id !== userId);

      set({
        users: filteredUsers,
        loading: false,
        error: null,
      });

      return { success: true };
    } catch (error) {
      const errorMessage = error.response?.data?.message || 'Failed to delete user';
      set({ loading: false, error: errorMessage });
      return { success: false, message: errorMessage };
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Reset state
  reset: () => set({
    users: [],
    totalUsers: 0,
    nextCursor: null,
    loading: false,
    error: null,
  }),
}));

export default useAdminStore;