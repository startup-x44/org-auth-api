'use client'

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../contexts/auth-context';
import { Button } from '../components/ui/button';
import { Card } from '../components/ui/card';
import { LoadingSpinner } from '../components/ui/loading-spinner';
import { Shield, Users, Lock, Zap, CheckCircle, ArrowRight } from 'lucide-react';

export default function HomePage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && user) {
      const redirectPath = user.is_superadmin 
        ? '/superadmin/dashboard' 
        : '/user/dashboard';
      router.push(redirectPath);
    }
  }, [user, loading, router]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="text-center">
          <LoadingSpinner size="lg" variant="primary" />
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (user) {
    return null; // Will redirect via useEffect
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Header */}
      <nav className="bg-white/95 backdrop-blur shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <div className="flex items-center">
                <div className="p-2 bg-primary/10 rounded-lg mr-3">
                  <Shield className="h-6 w-6 text-primary" />
                </div>
                <div>
                  <h1 className="text-xl font-bold text-gray-900">NILOAUTH</h1>
                  <p className="text-xs text-gray-500">Production Ready Authentication</p>
                </div>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <Button
                variant="outline"
                onClick={() => router.push('/auth/login')}
                className="flex items-center"
              >
                Sign In
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="text-center mb-16">
          <div className="flex justify-center mb-8">
            <div className="p-4 bg-primary/10 rounded-full">
              <Shield className="h-16 w-16 text-primary" />
            </div>
          </div>
          <h1 className="text-4xl md:text-6xl font-bold text-foreground mb-6">
            Welcome to <span className="text-primary">NILOAUTH</span>
          </h1>
          <p className="text-xl text-gray-600 mb-8 max-w-2xl mx-auto">
            A modern, secure, and production-ready authentication service built with cutting-edge technology.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button
              size="lg"
              onClick={() => router.push('/auth/login')}
              className="text-lg px-8 py-4"
            >
              Get Started
              <ArrowRight className="ml-2 h-5 w-5" />
            </Button>
            <Button
              variant="outline"
              size="lg"
              className="text-lg px-8 py-4"
            >
              Learn More
            </Button>
          </div>
        </div>

        {/* Features */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-16">
          <Card className="p-8 text-center hover:shadow-lg transition-shadow">
            <div className="flex justify-center mb-4">
              <div className="p-3 bg-success/20 rounded-lg">
                <Lock className="h-8 w-8 text-success" />
              </div>
            </div>
            <h3 className="text-xl font-semibold text-foreground mb-4">Secure by Design</h3>
            <p className="text-gray-600">
              Built with enterprise-grade security features including JWT tokens, role-based access control, and comprehensive audit logging.
            </p>
          </Card>

          <Card className="p-8 text-center hover:shadow-lg transition-shadow">
            <div className="flex justify-center mb-4">
              <div className="p-3 bg-primary/10 rounded-lg">
                <Users className="h-8 w-8 text-primary" />
              </div>
            </div>
            <h3 className="text-xl font-semibold text-foreground mb-4">Multi-Organization</h3>
            <p className="text-gray-600">
              Support for multiple organizations with granular permissions and role management across different tenants.
            </p>
          </Card>

          <Card className="p-8 text-center hover:shadow-lg transition-shadow">
            <div className="flex justify-center mb-4">
              <div className="p-3 bg-orange-100 rounded-lg">
                <Zap className="h-8 w-8 text-orange-600" />
              </div>
            </div>
            <h3 className="text-xl font-semibold text-foreground mb-4">Modern Stack</h3>
            <p className="text-gray-600">
              Built with Next.js, TypeScript, Go backend, PostgreSQL, and modern UI components for optimal performance.
            </p>
          </Card>
        </div>

        {/* Tech Stack */}
        <Card className="p-8 mb-16">
          <h2 className="text-2xl font-bold text-foreground mb-6 text-center">
            Built with Modern Technology
          </h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
            <div>
              <div className="text-2xl font-bold text-primary mb-2">Next.js 16</div>
              <div className="text-sm text-gray-600">React Framework</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-success mb-2">Go</div>
              <div className="text-sm text-gray-600">Backend API</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-primary mb-2">PostgreSQL</div>
              <div className="text-sm text-gray-600">Database</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-orange-600 mb-2">TypeScript</div>
              <div className="text-sm text-gray-600">Type Safety</div>
            </div>
          </div>
        </Card>

        {/* Demo Accounts */}
        <Card className="p-8 bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200">
          <div className="text-center mb-6">
            <h2 className="text-2xl font-bold text-foreground mb-2">Try Demo Accounts</h2>
            <p className="text-gray-600">Experience the full functionality with our test accounts</p>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-2xl mx-auto">
            <div className="bg-white rounded-lg p-6 text-center">
              <div className="flex justify-center mb-3">
                <Shield className="h-8 w-8 text-primary" />
              </div>
              <h3 className="font-semibold text-foreground mb-2">SuperAdmin Access</h3>
              <p className="text-sm text-gray-600 mb-4">Full system administration capabilities</p>
              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
                  <span>User Management</span>
                </div>
                <div className="flex items-center justify-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
                  <span>System Settings</span>
                </div>
                <div className="flex items-center justify-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
                  <span>Organization Control</span>
                </div>
              </div>
            </div>
            
            <div className="bg-white rounded-lg p-6 text-center">
              <div className="flex justify-center mb-3">
                <Users className="h-8 w-8 text-success" />
              </div>
              <h3 className="font-semibold text-foreground mb-2">Regular User</h3>
              <p className="text-sm text-gray-600 mb-4">Standard user dashboard and features</p>
              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
                  <span>Profile Management</span>
                </div>
                <div className="flex items-center justify-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
                  <span>Organization Access</span>
                </div>
                <div className="flex items-center justify-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
                  <span>Activity Tracking</span>
                </div>
              </div>
            </div>
          </div>
          
          <div className="text-center mt-8">
            <Button
              size="lg"
              onClick={() => router.push('/auth/login')}
              className="text-lg px-8 py-4"
            >
              Try Demo Accounts
              <ArrowRight className="ml-2 h-5 w-5" />
            </Button>
          </div>
        </Card>
      </div>

      {/* Footer */}
      <footer className="bg-white border-t mt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="flex justify-center items-center mb-4">
              <div className="p-2 bg-primary/10 rounded-lg mr-3">
                <Shield className="h-5 w-5 text-primary" />
              </div>
              <span className="text-lg font-bold text-foreground">NILOAUTH</span>
            </div>
            <p className="text-gray-600">
              Production Ready Authentication Service
            </p>
            <p className="text-xs text-gray-500 mt-2">
              Built with ❤️ using modern web technologies
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}
