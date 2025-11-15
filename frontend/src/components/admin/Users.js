import React, { useState, useEffect, useCallback } from 'react';
import useAdminStore from '../../stores/adminStore';
import { Button, Loading, ConfirmModal } from '../shared';
import useNotificationStore from '../../stores/notificationStore';

const AdminUsers = () => {
  const { users, isLoading, fetchUsers, activateUser, deactivateUser, deleteUser } = useAdminStore();
  const { success: showSuccess, error: showError } = useNotificationStore();
  const [pagination, setPagination] = useState({
    limit: 10,
    offset: 0,
    total: 0,
  });
  const [confirmModal, setConfirmModal] = useState({
    isOpen: false,
    title: '',
    message: '',
    action: null,
    loading: false,
  });

  const loadUsers = useCallback(async () => {
    try {
      await fetchUsers({
        limit: pagination.limit,
        offset: pagination.offset,
      });
    } catch (err) {
      showError('Failed to fetch users');
    }
  }, [pagination.limit, pagination.offset, fetchUsers, showError]);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  const handleActivateUser = async (userId) => {
    setConfirmModal({
      isOpen: true,
      title: 'Activate User',
      message: 'Are you sure you want to activate this user?',
      action: async () => {
        setConfirmModal(prev => ({ ...prev, loading: true }));
        try {
          const result = await activateUser(userId);
          if (result.success) {
            showSuccess('User activated successfully');
            loadUsers();
          } else {
            showError(result.message);
          }
        } catch (err) {
          showError('Failed to activate user');
        } finally {
          setConfirmModal({ isOpen: false, title: '', message: '', action: null, loading: false });
        }
      },
    });
  };

  const handleDeactivateUser = async (userId) => {
    setConfirmModal({
      isOpen: true,
      title: 'Deactivate User',
      message: 'Are you sure you want to deactivate this user?',
      action: async () => {
        setConfirmModal(prev => ({ ...prev, loading: true }));
        try {
          const result = await deactivateUser(userId);
          if (result.success) {
            showSuccess('User deactivated successfully');
            loadUsers();
          } else {
            showError(result.message);
          }
        } catch (err) {
          showError('Failed to deactivate user');
        } finally {
          setConfirmModal({ isOpen: false, title: '', message: '', action: null, loading: false });
        }
      },
    });
  };

  const handleDeleteUser = async (userId) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete User',
      message: 'Are you sure you want to delete this user? This action cannot be undone.',
      action: async () => {
        setConfirmModal(prev => ({ ...prev, loading: true }));
        try {
          const result = await deleteUser(userId);
          if (result.success) {
            showSuccess('User deleted successfully');
            loadUsers();
          } else {
            showError(result.message);
          }
        } catch (err) {
          showError('Failed to delete user');
        } finally {
          setConfirmModal({ isOpen: false, title: '', message: '', action: null, loading: false });
        }
      },
    });
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  if (isLoading) {
    return <Loading text="Loading users..." />;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-foreground">User Management</h1>
          <p className="text-gray-600">Manage user accounts and permissions</p>
        </div>
      </div>

      {/* Users Table */}
      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Email
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {users.map((user) => (
                <tr key={user.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-foreground">
                      {user.first_name && user.last_name
                        ? `${user.first_name} ${user.last_name}`
                        : 'N/A'
                      }
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-foreground">{user.email}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="capitalize inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
                      {user.user_type}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      user.is_active
                        ? 'bg-green-100 text-green-800'
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {user.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatDate(user.created_at)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div className="flex space-x-2">
                      {user.is_active ? (
                        <Button
                          onClick={() => handleDeactivateUser(user.id)}
                          variant="secondary"
                          size="sm"
                        >
                          Deactivate
                        </Button>
                      ) : (
                        <Button
                          onClick={() => handleActivateUser(user.id)}
                          variant="secondary"
                          size="sm"
                        >
                          Activate
                        </Button>
                      )}
                      <Button
                        onClick={() => handleDeleteUser(user.id)}
                        variant="danger"
                        size="sm"
                      >
                        Delete
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {users.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
            </svg>
            <p className="mt-2 text-sm">No users found</p>
          </div>
        )}
      </div>

      {/* Pagination */}
      {users.length > 0 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-700">
            Showing {pagination.offset + 1} to {Math.min(pagination.offset + pagination.limit, users.length)} of {users.length} users
          </div>
          <div className="flex space-x-2">
            <Button
              onClick={() => setPagination(prev => ({ ...prev, offset: Math.max(0, prev.offset - prev.limit) }))}
              disabled={pagination.offset === 0}
              variant="secondary"
              size="sm"
            >
              Previous
            </Button>
            <Button
              onClick={() => setPagination(prev => ({ ...prev, offset: prev.offset + prev.limit }))}
              disabled={pagination.offset + pagination.limit >= users.length}
              variant="secondary"
              size="sm"
            >
              Next
            </Button>
          </div>
        </div>
      )}

      {/* Confirmation Modal */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        onClose={() => setConfirmModal({ isOpen: false, title: '', message: '', action: null, loading: false })}
        onConfirm={confirmModal.action}
        title={confirmModal.title}
        message={confirmModal.message}
        loading={confirmModal.loading}
      />
    </div>
  );
};

export default AdminUsers;