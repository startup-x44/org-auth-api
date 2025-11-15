import { create } from 'zustand';

const useNotificationStore = create((set, get) => ({
  notifications: [],

  // Add a new notification
  addNotification: (type, message, options = {}) => {
    const id = Date.now() + Math.random();
    const notification = {
      id,
      type,
      message,
      autoClose: options.autoClose !== false,
      duration: options.duration || 5000,
      ...options,
    };

    set(state => ({
      notifications: [...state.notifications, notification],
    }));

    return id;
  },

  // Remove a notification by id
  removeNotification: (id) => {
    set(state => ({
      notifications: state.notifications.filter(notification => notification.id !== id),
    }));
  },

  // Clear all notifications
  clearNotifications: () => {
    set({ notifications: [] });
  },

  // Convenience methods for different notification types
  success: (message, options) => get().addNotification('success', message, options),
  error: (message, options) => get().addNotification('error', message, options),
  warning: (message, options) => get().addNotification('warning', message, options),
  info: (message, options) => get().addNotification('info', message, options),
}));

export default useNotificationStore;