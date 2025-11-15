import React, { useState, useEffect, useCallback } from 'react';
import useAdminStore from '../../stores/adminStore';
import { Button, Input, Loading, ConfirmModal } from '../shared';
import useNotificationStore from '../../stores/notificationStore';

const AdminTenants = () => {
  const { tenants, isLoading, fetchTenants, createTenant, updateTenant, deleteTenant } = useAdminStore();
  const { success: showSuccess, error: showError } = useNotificationStore();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingTenant, setEditingTenant] = useState(null);
  const [pagination, setPagination] = useState({
    limit: 10,
    offset: 0,
    total: 0,
  });
  const [formData, setFormData] = useState({
    name: '',
    domain: '',
  });
  const [formErrors, setFormErrors] = useState({});
  const [confirmModal, setConfirmModal] = useState({
    isOpen: false,
    title: '',
    message: '',
    action: null,
    loading: false,
  });

  const loadTenants = useCallback(async () => {
    try {
      await fetchTenants({
        limit: pagination.limit,
        offset: pagination.offset,
      });
    } catch (err) {
      showError('Failed to fetch tenants');
    }
  }, [pagination.limit, pagination.offset, fetchTenants, showError]);

  useEffect(() => {
    loadTenants();
  }, [loadTenants]);

  const validateForm = () => {
    const errors = {};
    if (!formData.name.trim()) {
      errors.name = 'Tenant name is required';
    }
    if (!formData.domain.trim()) {
      errors.domain = 'Domain is required';
    } else if (!/^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(formData.domain)) {
      errors.domain = 'Please enter a valid domain';
    }
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleCreateTenant = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      const result = await createTenant(formData);
      if (result.success) {
        showSuccess('Tenant created successfully');
        setShowCreateForm(false);
        setFormData({ name: '', domain: '' });
        loadTenants();
      } else {
        showError(result.message);
      }
    } catch (err) {
      showError('Failed to create tenant');
    }
  };

  const handleUpdateTenant = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      const result = await updateTenant(editingTenant.id, formData);
      if (result.success) {
        showSuccess('Tenant updated successfully');
        setEditingTenant(null);
        setFormData({ name: '', domain: '' });
        loadTenants();
      } else {
        showError(result.message);
      }
    } catch (err) {
      showError('Failed to update tenant');
    }
  };

  const handleDeleteTenant = async (tenantId) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete Tenant',
      message: 'Are you sure you want to delete this tenant? This action cannot be undone.',
      action: async () => {
        setConfirmModal(prev => ({ ...prev, loading: true }));
        try {
          const result = await deleteTenant(tenantId);
          if (result.success) {
            showSuccess('Tenant deleted successfully');
            loadTenants();
          } else {
            showError(result.message);
          }
        } catch (err) {
          showError('Failed to delete tenant');
        } finally {
          setConfirmModal({ isOpen: false, title: '', message: '', action: null, loading: false });
        }
      },
    });
  };

  const startEdit = (tenant) => {
    setEditingTenant(tenant);
    setFormData({
      name: tenant.name,
      domain: tenant.domain,
    });
    setFormErrors({});
  };

  const cancelEdit = () => {
    setEditingTenant(null);
    setFormData({ name: '', domain: '' });
    setFormErrors({});
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  if (isLoading) {
    return <Loading text="Loading tenants..." />;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Tenant Management</h1>
          <p className="text-gray-600">Manage organizations and their domains</p>
        </div>
        <Button
          onClick={() => setShowCreateForm(true)}
          variant="primary"
        >
          Add Tenant
        </Button>
      </div>

      {/* Create/Edit Form */}
      {(showCreateForm || editingTenant) && (
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h2 className="text-xl font-semibold text-foreground mb-6">
              {editingTenant ? 'Edit Tenant' : 'Create New Tenant'}
            </h2>

            <form onSubmit={editingTenant ? handleUpdateTenant : handleCreateTenant} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Input
                  label="Tenant Name"
                  name="name"
                  type="text"
                  required
                  placeholder="Enter tenant name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  error={formErrors.name}
                />

                <Input
                  label="Domain"
                  name="domain"
                  type="text"
                  required
                  placeholder="example.com"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                  error={formErrors.domain}
                />
              </div>

              <div className="flex justify-end space-x-3">
                <Button
                  type="button"
                  onClick={() => {
                    setShowCreateForm(false);
                    cancelEdit();
                  }}
                  variant="secondary"
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                >
                  {editingTenant ? 'Update Tenant' : 'Create Tenant'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Tenants Table */}
      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Domain
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
              {tenants.map((tenant) => (
                <tr key={tenant.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-foreground">{tenant.name}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-foreground">{tenant.domain}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatDate(tenant.created_at)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div className="flex space-x-2">
                      <Button
                        onClick={() => startEdit(tenant)}
                        variant="secondary"
                        size="sm"
                      >
                        Edit
                      </Button>
                      <Button
                        onClick={() => handleDeleteTenant(tenant.id)}
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

        {tenants.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
            <p className="mt-2 text-sm">No tenants found</p>
          </div>
        )}
      </div>

      {/* Pagination */}
      {tenants.length > 0 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-700">
            Showing {pagination.offset + 1} to {Math.min(pagination.offset + pagination.limit, tenants.length)} of {tenants.length} tenants
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
              disabled={pagination.offset + pagination.limit >= tenants.length}
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

export default AdminTenants;