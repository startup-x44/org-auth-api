import React, { useState, useEffect } from 'react';
import useAuthStore from '../../stores/authStore';
import { Button, Input } from '../shared';
import useNotificationStore from '../../stores/notificationStore';

const Profile = () => {
  const { user, updateProfile, changePassword } = useAuthStore();
  const { success: showSuccess, error: showError } = useNotificationStore();
  const [activeTab, setActiveTab] = useState('profile');
  const [profileData, setProfileData] = useState({
    first_name: '',
    last_name: '',
    phone: '',
  });
  const [passwordData, setPasswordData] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  });
  const [loading, setLoading] = useState(false);
  const [profileErrors, setProfileErrors] = useState({});
  const [passwordErrors, setPasswordErrors] = useState({});

  useEffect(() => {
    if (user) {
      setProfileData({
        first_name: user.first_name || '',
        last_name: user.last_name || '',
        phone: user.phone || '',
      });
    }
  }, [user]);

  const handleProfileChange = (e) => {
    const { name, value } = e.target;
    setProfileData({
      ...profileData,
      [name]: value,
    });
    // Clear error when user starts typing
    if (profileErrors[name]) {
      setProfileErrors(prev => ({
        ...prev,
        [name]: undefined,
      }));
    }
  };

  const handlePasswordChange = (e) => {
    const { name, value } = e.target;
    setPasswordData({
      ...passwordData,
      [name]: value,
    });
    // Clear error when user starts typing
    if (passwordErrors[name]) {
      setPasswordErrors(prev => ({
        ...prev,
        [name]: undefined,
      }));
    }
  };

  const validateProfile = () => {
    const errors = {};
    if (!profileData.first_name.trim()) {
      errors.first_name = 'First name is required';
    }
    if (!profileData.last_name.trim()) {
      errors.last_name = 'Last name is required';
    }
    setProfileErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const validatePassword = () => {
    const errors = {};
    if (!passwordData.current_password) {
      errors.current_password = 'Current password is required';
    }
    if (!passwordData.new_password) {
      errors.new_password = 'New password is required';
    } else if (passwordData.new_password.length < 8) {
      errors.new_password = 'Password must be at least 8 characters';
    }
    if (!passwordData.confirm_password) {
      errors.confirm_password = 'Please confirm your new password';
    } else if (passwordData.new_password !== passwordData.confirm_password) {
      errors.confirm_password = 'Passwords do not match';
    }
    setPasswordErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleProfileSubmit = async (e) => {
    e.preventDefault();

    if (!validateProfile()) {
      return;
    }

    setLoading(true);

    try {
      const result = await updateProfile(profileData);

      if (result.success) {
        showSuccess('Profile updated successfully!');
      } else {
        showError(result.message);
      }
    } catch (err) {
      showError('Failed to update profile. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordSubmit = async (e) => {
    e.preventDefault();

    if (!validatePassword()) {
      return;
    }

    setLoading(true);

    try {
      const result = await changePassword({
        current_password: passwordData.current_password,
        new_password: passwordData.new_password,
      });

      if (result.success) {
        showSuccess('Password changed successfully!');
        setPasswordData({
          current_password: '',
          new_password: '',
          confirm_password: '',
        });
      } else {
        showError(result.message);
      }
    } catch (err) {
      showError('Failed to change password. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div>
  <h1 className="text-2xl font-bold text-foreground">Profile Settings</h1>
        <p className="text-gray-600">Manage your account settings and preferences.</p>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('profile')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'profile'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Profile Information
          </button>
          <button
            onClick={() => setActiveTab('password')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'password'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Change Password
          </button>
        </nav>
      </div>

      {/* Profile Tab */}
      {activeTab === 'profile' && (
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h2 className="text-xl font-semibold text-foreground mb-6">Profile Information</h2>

            <form onSubmit={handleProfileSubmit} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Input
                  label="First Name"
                  name="first_name"
                  type="text"
                  required
                  placeholder="Enter your first name"
                  value={profileData.first_name}
                  onChange={handleProfileChange}
                  error={profileErrors.first_name}
                />

                <Input
                  label="Last Name"
                  name="last_name"
                  type="text"
                  required
                  placeholder="Enter your last name"
                  value={profileData.last_name}
                  onChange={handleProfileChange}
                  error={profileErrors.last_name}
                />
              </div>

              <Input
                label="Phone Number"
                name="phone"
                type="tel"
                placeholder="Enter your phone number"
                value={profileData.phone}
                onChange={handleProfileChange}
                error={profileErrors.phone}
              />

              <div className="flex justify-end">
                <Button
                  type="submit"
                  variant="primary"
                  loading={loading}
                >
                  Update Profile
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Password Tab */}
      {activeTab === 'password' && (
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h2 className="text-xl font-semibold text-foreground mb-6">Change Password</h2>

            <form onSubmit={handlePasswordSubmit} className="space-y-6">
              <Input
                label="Current Password"
                name="current_password"
                type="password"
                required
                placeholder="Enter your current password"
                value={passwordData.current_password}
                onChange={handlePasswordChange}
                error={passwordErrors.current_password}
              />

              <Input
                label="New Password"
                name="new_password"
                type="password"
                required
                placeholder="Enter your new password"
                value={passwordData.new_password}
                onChange={handlePasswordChange}
                error={passwordErrors.new_password}
              />

              <Input
                label="Confirm New Password"
                name="confirm_password"
                type="password"
                required
                placeholder="Confirm your new password"
                value={passwordData.confirm_password}
                onChange={handlePasswordChange}
                error={passwordErrors.confirm_password}
              />

              <div className="text-sm text-gray-500">
                Password must be at least 8 characters with uppercase, lowercase, number, and special character.
              </div>

              <div className="flex justify-end">
                <Button
                  type="submit"
                  variant="primary"
                  loading={loading}
                >
                  Change Password
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default Profile;