import React from 'react';
import { createPortal } from 'react-dom';
import Notification from './Notification';
import useNotificationStore from '../../stores/notificationStore';

const NotificationContainer = () => {
  const { notifications, removeNotification } = useNotificationStore();

  if (notifications.length === 0) {
    return null;
  }

  return createPortal(
    <div className="fixed top-4 right-4 z-50 space-y-2 max-w-sm">
      {notifications.map(notification => (
        <Notification
          key={notification.id}
          type={notification.type}
          message={notification.message}
          onClose={() => removeNotification(notification.id)}
          autoClose={notification.autoClose}
          duration={notification.duration}
        />
      ))}
    </div>,
    document.body
  );
};

export default NotificationContainer;