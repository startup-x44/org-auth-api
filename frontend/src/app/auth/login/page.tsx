'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../../contexts/auth-context';
import { AuthService } from '../../../lib/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { useToast } from '../../../hooks/use-toast';
import { Eye, EyeOff, Mail, Lock, Shield, AlertCircle, CheckCircle } from 'lucide-react';
import { Alert, AlertDescription } from '../../../components/ui/alert';
import Link from 'next/link';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const [validationErrors, setValidationErrors] = useState<{ email?: string; password?: string }>({});

  const router = useRouter();
  const { login, user } = useAuth();
  const { toast } = useToast();

  // Redirect if already logged in
  useEffect(() => {
    if (user) {
      if (user.is_superadmin) {
        router.push('/superadmin/dashboard');
      } else {
        // For regular users, check if they have organization context
        const token = AuthService.getAccessToken();
        const payload = token ? AuthService.decodeToken(token) : null;

        if (payload && payload.organization_id) {
          // User has organization context, go to dashboard
          router.push('/user/dashboard');
        } else {
          // User needs to select organization
          router.push('/auth/select-organization');
        }
      }
    }
  }, [user, router]);

  // Debug error state
  useEffect(() => {
    console.log('Current error state:', error);
  }, [error]);

  // Track component lifecycle
  useEffect(() => {
    console.log('LoginPage component mounted');
    return () => {
      console.log('LoginPage component unmounted');
    };
  }, []);

  // Show messages based on URL parameters
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);

    if (urlParams.get('verified') === 'true') {
      toast({
        title: "Email verified successfully!",
        description: "You can now sign in to your account.",
        variant: "success",
      });
    } else if (urlParams.get('reason') === 'unauthorized') {
      toast({
        title: "Session expired",
        description: "Please sign in to continue.",
        variant: "destructive",
      });
    }

    // Clean up URL
    if (urlParams.get('verified') || urlParams.get('reason')) {
      window.history.replaceState({}, '', '/auth/login');
    }
  }, [toast]);

  const validateForm = () => {
    const errors: { email?: string; password?: string } = {};

    if (!email) {
      errors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      errors.email = 'Please enter a valid email address';
    }

    if (!password) {
      errors.password = 'Password is required';
    } else if (password.length < 6) {
      errors.password = 'Password must be at least 6 characters';
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setLoading(true);
    setError('');
    console.log('Cleared error, starting login...');

    try {
      console.log('Attempting login with:', email);
      const result = await login(email, password);
      console.log('Login result:', result);

      if (result.success) {
        console.log('LOGIN SUCCESS FLOW REACHED');
        console.log('Organizations from login:', result.organizations);
        console.log('Needs org selection:', result.needsOrgSelection);

        // Check if user is superadmin
        const token = AuthService.getAccessToken();
        const isSuperAdmin = token ? AuthService.isSuperAdmin(token) : false;
        console.log('Is SuperAdmin:', isSuperAdmin);

        if (isSuperAdmin) {
          // Show success toast for superadmin
          toast({
            title: "Login successful!",
            description: "Welcome back to NILOAUTH Admin",
            variant: "success",
          });

          console.log('Redirecting to superadmin dashboard');
          setTimeout(() => {
            router.push('/superadmin/dashboard');
          }, 500);
        } else if (result.needsOrgSelection || (result.organizations && result.organizations.length === 0)) {
          // User needs to select/create organization
          toast({
            title: "Login successful!",
            description: "Please create or select an organization to continue",
            variant: "success",
          });

          console.log('Redirecting to organization selection');
          console.log('User data to store:', result.user);
          console.log('Organizations to store:', result.organizations);

          // Store organizations and user data for the selection page
          if (result.organizations) {
            sessionStorage.setItem('user_organizations', JSON.stringify(result.organizations));
          }
          // Store user data so we can get the user ID - use result.user directly
          if (result.user) {
            sessionStorage.setItem('pending_user', JSON.stringify(result.user));
            console.log('Stored user data in session storage');
          } else {
            console.error('No user data available to store');
          }

          setTimeout(() => {
            router.push('/auth/select-organization');
          }, 500);
        } else {
          // User has organization context, go to dashboard
          toast({
            title: "Login successful!",
            description: "Welcome back to NILOAUTH",
            variant: "success",
          });

          console.log('Redirecting to user dashboard');
          setTimeout(() => {
            router.push('/user/dashboard');
          }, 500);
        }
      }
    } catch (err) {
      console.error('CATCH BLOCK REACHED - Login error:', err);
      const errorMessage = err instanceof Error ? err.message : 'Login failed. Please try again.';
      console.log('Setting error message:', errorMessage);
      setError(errorMessage);
      console.log('Error state should be set now');
    } finally {
      setLoading(false);
    }
  };

  const fillTestCredentials = (type: 'superadmin' | 'user') => {
    if (type === 'superadmin') {
      setEmail('superadmin@platform.com');
      setPassword('Admin123!');
    } else {
      setEmail('user@example.com');
      setPassword('Admin123!');
    }
    setValidationErrors({});
  };

  if (user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="text-center">
          <LoadingSpinner size="lg" variant="primary" />
          <p className="mt-4 text-gray-600">Redirecting to dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        <Card className="p-8 shadow-xl border-0 bg-white backdrop-blur">
          {/* Header */}
          <div className="text-center mb-8">
            <div className="flex justify-center mb-4">
              <div className="p-3 bg-primary/10 rounded-full">
                <Shield className="h-8 w-8 text-primary" />
              </div>
            </div>
            <h1 className="text-2xl font-bold text-foreground mb-2">
              Welcome to NILOAUTH
            </h1>
            <p className="text-gray-600">
              Sign in to access your account
            </p>
          </div>

          {/* Error Alert */}
          {error && (
            <div className="mb-6">
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  {error}
                </AlertDescription>
              </Alert>
            </div>
          )}

          {/* Login Form */}
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Email Input */}
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium text-gray-700">
                Email Address
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                <Input
                  id="email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={(e) => {
                    setEmail(e.target.value);
                    if (validationErrors.email) {
                      setValidationErrors(prev => {
                        const { email, ...rest } = prev;
                        return rest;
                      });
                    }
                  }}
                  placeholder="Enter your email"
                  disabled={loading}
                  className={`pl-10 ${validationErrors.email ? 'border-red-300 focus:border-red-500 focus:ring-red-500' : ''}`}
                />
              </div>
              {validationErrors.email && (
                <p className="text-sm text-red-600 flex items-center">
                  <AlertCircle className="h-4 w-4 mr-1" />
                  {validationErrors.email}
                </p>
              )}
            </div>

            {/* Password Input */}
            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium text-gray-700">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                <Input
                  id="password"
                  name="password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => {
                    setPassword(e.target.value);
                    if (validationErrors.password) {
                      setValidationErrors(prev => {
                        const { password, ...rest } = prev;
                        return rest;
                      });
                    }
                  }}
                  placeholder="Enter your password"
                  disabled={loading}
                  className={`pl-10 pr-10 ${validationErrors.password ? 'border-red-300 focus:border-red-500 focus:ring-red-500' : ''}`}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                </button>
              </div>
              {validationErrors.password && (
                <p className="text-sm text-red-600 flex items-center">
                  <AlertCircle className="h-4 w-4 mr-1" />
                  {validationErrors.password}
                </p>
              )}
            </div>

            {/* Remember Me */}
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="remember-me"
                  checked={rememberMe}
                  onCheckedChange={(checked) => setRememberMe(checked as boolean)}
                />
                <Label htmlFor="remember-me" className="text-sm text-gray-700 font-normal cursor-pointer">
                  Remember me
                </Label>
              </div>
              <button
                type="button"
                className="text-sm text-primary hover:text-primary/80"
              >
                Forgot password?
              </button>
            </div>

            {/* Submit Button */}
            <Button
              type="submit"
              disabled={loading}
              className="w-full h-12 text-base font-medium"
            >
              {loading ? (
                <div className="flex items-center">
                  <LoadingSpinner size="sm" variant="secondary" />
                  <span className="ml-2">Signing in...</span>
                </div>
              ) : (
                <div className="flex items-center justify-center">
                  <span>Sign in</span>
                  <CheckCircle className="ml-2 h-5 w-5" />
                </div>
              )}
            </Button>
          </form>

          {/* Test Accounts */}
          <div className="mt-8 pt-6 border-t border-gray-200">
            <div className="text-center mb-4">
              <p className="text-sm font-medium text-gray-700">Test Accounts</p>
              <p className="text-xs text-gray-500 mt-1">Click to auto-fill credentials</p>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={() => fillTestCredentials('superadmin')}
                disabled={loading}
                className="h-auto p-3 flex flex-col items-center space-y-2"
              >
                <Shield className="h-5 w-5 text-primary" />
                <div className="text-center">
                  <div className="text-xs font-medium">SuperAdmin</div>
                  <div className="text-xs text-gray-500">Full Access</div>
                </div>
              </Button>

              <Button
                type="button"
                variant="outline"
                onClick={() => fillTestCredentials('user')}
                disabled={loading}
                className="h-auto p-3 flex flex-col items-center space-y-2"
              >
                <div className="h-5 w-5 bg-success/20 rounded-full flex items-center justify-center">
                  <div className="h-2 w-2 bg-success rounded-full"></div>
                </div>
                <div className="text-center">
                  <div className="text-xs font-medium">Regular User</div>
                  <div className="text-xs text-gray-500">Standard Access</div>
                </div>
              </Button>
            </div>
          </div>

          {/* Sign Up Link */}
          <div className="mt-8 pt-6 border-t border-gray-200 text-center">
            <p className="text-sm text-gray-600">
              Don't have an account?{' '}
              <Link
                href="/auth/register"
                className="text-primary hover:text-primary/80 font-medium"
              >
                Create one here
              </Link>
            </p>
          </div>

          {/* Footer */}
          <div className="mt-6 text-center">
            <p className="text-xs text-gray-500">
              Secured by NILOAUTH • Production Ready Authentication
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
}