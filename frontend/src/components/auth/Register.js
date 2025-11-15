import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import useAuthStore from '../../stores/authStore';
import useTenantStore from '../../stores/tenantStore';
import { Button, Input, Loading } from '../shared';
import useNotificationStore from '../../stores/notificationStore';

const Register = () => {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirm_password: '',
    user_type: 'student',
    first_name: '',
    last_name: '',
    phone: '',
    tenant_id: '',
  });
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState({});

  const { register } = useAuthStore();
  const { tenantId, resolveTenantFromEmail, setTenantId } = useTenantStore();
  const { error: showError, success: showSuccess } = useNotificationStore();
  const navigate = useNavigate();

  // Initialize tenant from subdomain or set default
  useEffect(() => {
    const subdomainTenant = tenantId;
    if (subdomainTenant) {
      setFormData(prev => ({ ...prev, tenant_id: subdomainTenant }));
    }
  }, [tenantId]);

  const validateForm = () => {
    const newErrors = {};

    if (!formData.email) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    }

    if (!formData.confirm_password) {
      newErrors.confirm_password = 'Please confirm your password';
    } else if (formData.password !== formData.confirm_password) {
      newErrors.confirm_password = 'Passwords do not match';
    }

    if (!formData.first_name) {
      newErrors.first_name = 'First name is required';
    }

    if (!formData.last_name) {
      newErrors.last_name = 'Last name is required';
    }

    if (!formData.tenant_id) {
      newErrors.tenant_id = 'Tenant is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });

    // Auto-resolve tenant from email domain
    if (name === 'email' && value.includes('@')) {
      const emailTenant = resolveTenantFromEmail(value);
      if (emailTenant && !formData.tenant_id) {
        setFormData(prev => ({ ...prev, tenant_id: emailTenant }));
      }
    }

    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({
        ...prev,
        [name]: undefined,
      }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setLoading(true);

    try {
      const result = await register(formData);

      if (result.success) {
        // Set tenant ID in store after successful registration
        setTenantId(formData.tenant_id);
        showSuccess('Account created successfully! Please check your email for verification.');
        navigate('/login');
      } else {
        showError(result.message);
      }
    } catch (err) {
      const errorMessage = 'An unexpected error occurred. Please try again.';
      showError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <Loading fullScreen text="Creating your account..." />;
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-foreground">
            Create your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Or{' '}
            <Link
              to="/login"
              className="font-medium text-primary-600 hover:text-primary-500"
            >
              sign in to existing account
            </Link>
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <Input
                label="First Name"
                name="first_name"
                type="text"
                required
                placeholder="First name"
                value={formData.first_name}
                onChange={handleChange}
                error={errors.first_name}
              />
              <Input
                label="Last Name"
                name="last_name"
                type="text"
                required
                placeholder="Last name"
                value={formData.last_name}
                onChange={handleChange}
                error={errors.last_name}
              />
            </div>

            <Input
              label="Email Address"
              name="email"
              type="email"
              required
              placeholder="Enter your email"
              value={formData.email}
              onChange={handleChange}
              error={errors.email}
            />

            <div>
              <label htmlFor="user_type" className="block text-sm font-medium text-gray-700">
                User Type
              </label>
              <select
                id="user_type"
                name="user_type"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                value={formData.user_type}
                onChange={handleChange}
              >
                <option value="student">Student</option>
                <option value="rto">RTO</option>
                <option value="issuer">Issuer</option>
                <option value="validator">Validator</option>
                <option value="badger">Badger</option>
                <option value="non_partner">Non Partner</option>
                <option value="partner">Partner</option>
              </select>
            </div>

            <Input
              label="Phone Number"
              name="phone"
              type="tel"
              placeholder="Phone number (optional)"
              value={formData.phone}
              onChange={handleChange}
              error={errors.phone}
            />

            <Input
              label="Password"
              name="password"
              type="password"
              required
              placeholder="Create a password"
              value={formData.password}
              onChange={handleChange}
              error={errors.password}
            />

            <Input
              label="Confirm Password"
              name="confirm_password"
              type="password"
              required
              placeholder="Confirm your password"
              value={formData.confirm_password}
              onChange={handleChange}
              error={errors.confirm_password}
            />
          </div>

          <div>
            <Button
              type="submit"
              variant="primary"
              className="w-full"
              loading={loading}
            >
              Create account
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default Register;