import React, { useState, useEffect } from 'react';
import { adminAPI } from '../../services/api';

const AdminTenants = () => {
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);
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
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    fetchTenants();
  }, [pagination.limit, pagination.offset]);

  const fetchTenants = async () => {
    try {
      setLoading(true);
      const response = await adminAPI.listTenants({
        limit: pagination.limit,
        offset: pagination.offset,
      });

      setTenants(response.data.data.tenants);
      setPagination(prev => ({
        ...prev,
        total: response.data.data.total,
      }));
    } catch (err) {
      setError('Failed to fetch tenants');
      console.error('Error fetching tenants:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateTenant = async (e) => {
    e.preventDefault();
    try {
      await adminAPI.createTenant(formData);
      setMessage('Tenant created successfully');
      setShowCreateForm(false);
      setFormData({ name: '', domain: '' });
      fetchTenants();
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to create tenant');
      console.error('Error creating tenant:', err);
    }
  };

  const handleUpdateTenant = async (e) => {
    e.preventDefault();
    try {
      await adminAPI.updateTenant(editingTenant.id, formData);
      setMessage('Tenant updated successfully');
      setEditingTenant(null);
      setFormData({ name: '', domain: '' });
      fetchTenants();
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to update tenant');
      console.error('Error updating tenant:', err);
    }
  };

  const handleDeleteTenant = async (tenantId) => {
    if (!window.confirm('Are you sure you want to delete this tenant? This action cannot be undone.')) {
      return;
    }

    try {
      await adminAPI.deleteTenant(tenantId);
      setMessage('Tenant deleted successfully');
      fetchTenants();
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to delete tenant');
      console.error('Error deleting tenant:', err);
    }
  };

  const startEdit = (tenant) => {
    setEditingTenant(tenant);
    setFormData({
      name: tenant.name,
      domain: tenant.domain,
    });
  };

  const cancelEdit = () => {
    setEditingTenant(null);
    setFormData({ name: '', domain: '' });
  };

  const clearMessages = () => {
    setMessage('');
    setError('');
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Tenant Management</h1>
          <p className="text-gray-600">Manage organizations and their domains</p>
        </div>
        <button
          onClick={() => setShowCreateForm(true)}
          className="btn btn-primary"
        >
          Add Tenant
        </button>
      </div>

      {/* Messages */}
      {message && (
        <div className="bg-success bg-opacity-10 border border-success border-opacity-20 text-success px-4 py-3 rounded-md flex justify-between items-center">
          <span>{message}</span>
          <button onClick={clearMessages} className="text-success hover:text-success-dark">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      )}

      {error && (
        <div className="bg-error bg-opacity-10 border border-error border-opacity-20 text-error px-4 py-3 rounded-md flex justify-between items-center">
          <span>{error}</span>
          <button onClick={clearMessages} className="text-error hover:text-error-dark">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      )}

      {/* Create/Edit Form */}
      {(showCreateForm || editingTenant) && (
        <div className="card">
          <h2 className="text-xl font-semibold text-gray-900 mb-6">
            {editingTenant ? 'Edit Tenant' : 'Create New Tenant'}
          </h2>

          <form onSubmit={editingTenant ? handleUpdateTenant : handleCreateTenant} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                  Tenant Name
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  className="input mt-1"
                  placeholder="Enter tenant name"
                />
              </div>

              <div>
                <label htmlFor="domain" className="block text-sm font-medium text-gray-700">
                  Domain
                </label>
                <input
                  type="text"
                  id="domain"
                  name="domain"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                  required
                  className="input mt-1"
                  placeholder="example.com"
                />
                <p className="mt-1 text-sm text-gray-500">
                  Domain must be a valid domain name
                </p>
              </div>
            </div>

            <div className="flex justify-end space-x-3">
              <button
                type="button"
                onClick={() => {
                  setShowCreateForm(false);
                  cancelEdit();
                }}
                className="btn btn-secondary"
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn btn-primary"
              >
                {editingTenant ? 'Update Tenant' : 'Create Tenant'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Tenants Table */}
      <div className="card">
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Domain</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {tenants.map((tenant) => (
                <tr key={tenant.id}>
                  <td className="font-medium text-gray-900">{tenant.name}</td>
                  <td>{tenant.domain}</td>
                  <td>{formatDate(tenant.created_at)}</td>
                  <td>
                    <div className="flex space-x-2">
                      <button
                        onClick={() => startEdit(tenant)}
                        className="text-primary-600 hover:text-primary-800 text-sm font-medium"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteTenant(tenant.id)}
                        className="text-error hover:text-error-dark text-sm font-medium"
                      >
                        Delete
                      </button>
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
      {pagination.total > pagination.limit && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-700">
            Showing {pagination.offset + 1} to {Math.min(pagination.offset + pagination.limit, pagination.total)} of {pagination.total} tenants
          </div>
          <div className="flex space-x-2">
            <button
              onClick={() => setPagination(prev => ({ ...prev, offset: Math.max(0, prev.offset - prev.limit) }))}
              disabled={pagination.offset === 0}
              className="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <button
              onClick={() => setPagination(prev => ({ ...prev, offset: prev.offset + prev.limit }))}
              disabled={pagination.offset + pagination.limit >= pagination.total}
              className="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminTenants;