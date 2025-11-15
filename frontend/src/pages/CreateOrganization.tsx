import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Building2, Sparkles, ArrowLeft, Check, AlertCircle } from 'lucide-react';
import useAuthStore from '../store/auth';
import { Card, CardContent, CardDescription, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import LoadingSpinner from '@/components/ui/loading-spinner';

interface FormData {
  name: string;
  slug: string;
}

const CreateOrganization: React.FC = () => {
  const { createOrganization, user } = useAuthStore();
  const navigate = useNavigate();
  const [formData, setFormData] = useState<FormData>({
    name: '',
    slug: '',
  });
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>('');
  const [slugManuallyEdited, setSlugManuallyEdited] = useState<boolean>(false);

  // Check if user is authenticated
  React.useEffect(() => {
    console.log('CreateOrganization - Current user:', user);
    
    // If no user in Zustand, try to get from localStorage
    if (!user) {
      const userGlobal = localStorage.getItem('user_global');
      console.log('User from localStorage:', userGlobal);
      
      if (!userGlobal) {
        // No user found - redirect to login
        console.error('No user found, redirecting to login');
        navigate('/login', { replace: true });
      }
    }
  }, [user, navigate]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    
    if (name === 'slug') {
      setSlugManuallyEdited(true);
    }

    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    // Auto-generate slug from name only if slug hasn't been manually edited
    if (name === 'name' && !slugManuallyEdited) {
      const slug = value
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');
      setFormData((prev) => ({
        ...prev,
        slug,
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    const result = await createOrganization(formData.name, formData.slug);

    if (result.success) {
      navigate('/dashboard');
    } else {
      setError(result.message || 'Failed to create organization');
      setLoading(false);
    }
  };

  const handleCancel = () => {
    navigate('/choose-organization');
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 via-purple-50 to-indigo-100 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900 p-4">
      {/* Background Pattern */}
      <div className="absolute inset-0 bg-grid-slate-200/50 [mask-image:linear-gradient(0deg,white,rgba(255,255,255,0.8))] dark:bg-grid-slate-700/30" />

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-2xl relative z-10"
      >
        <Card className="shadow-2xl border border-slate-200/50 dark:border-slate-700/50 bg-white/95 dark:bg-slate-800/95 backdrop-blur-md overflow-hidden">
          {/* Header with Gradient */}
          <div className="bg-gradient-to-r from-purple-600 via-blue-600 to-indigo-600 p-8 text-white">
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ delay: 0.2, type: "spring" }}
              className="mx-auto w-20 h-20 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center shadow-lg mb-4"
            >
              <Sparkles className="h-10 w-10 text-white" />
            </motion.div>
            <CardTitle className="text-3xl font-bold text-center text-white">
              Create Your Workspace
            </CardTitle>
            <CardDescription className="text-center text-purple-100 mt-2">
              Set up your organization and start collaborating
            </CardDescription>
          </div>

          <CardContent className="p-8">
            <form onSubmit={handleSubmit} className="space-y-6">
              {error && (
                <motion.div
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                >
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                </motion.div>
              )}

              <motion.div
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.3 }}
                className="space-y-2"
              >
                <Label htmlFor="name" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                  Workspace Name *
                </Label>
                <div className="relative">
                  <Building2 className="absolute left-3 top-3 h-5 w-5 text-slate-400" />
                  <Input
                    id="name"
                    name="name"
                    type="text"
                    required
                    value={formData.name}
                    onChange={handleChange}
                    className="pl-10 h-12 text-lg border-slate-300 dark:border-slate-600 focus:border-purple-500 dark:focus:border-purple-400 bg-white dark:bg-slate-800"
                    placeholder="Acme Corporation"
                    disabled={loading}
                  />
                </div>
                <p className="text-sm text-slate-500 dark:text-slate-400">
                  This is how your workspace will appear to members
                </p>
              </motion.div>

              <motion.div
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.4 }}
                className="space-y-2"
              >
                <Label htmlFor="slug" className="text-base font-semibold text-slate-700 dark:text-slate-300">
                  Workspace URL *
                </Label>
                <div className="relative">
                  <span className="absolute left-3 top-3 text-slate-500 dark:text-slate-400 text-lg">
                    @
                  </span>
                  <Input
                    id="slug"
                    name="slug"
                    type="text"
                    required
                    value={formData.slug}
                    onChange={handleChange}
                    pattern="[a-z0-9\-]+"
                    className="pl-8 h-12 text-lg font-mono border-slate-300 dark:border-slate-600 focus:border-purple-500 dark:focus:border-purple-400 bg-white dark:bg-slate-800"
                    placeholder="acme-corporation"
                    disabled={loading}
                  />
                  {formData.slug && /^[a-z0-9-]+$/.test(formData.slug) && (
                    <Check className="absolute right-3 top-3 h-5 w-5 text-green-500" />
                  )}
                </div>
                <p className="text-sm text-slate-500 dark:text-slate-400">
                  Lowercase letters, numbers, and hyphens only. Cannot be changed later.
                </p>
              </motion.div>

              {/* Features List */}
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.5 }}
                className="bg-gradient-to-br from-purple-50 to-blue-50 dark:from-slate-700/50 dark:to-slate-700/30 rounded-lg p-6 space-y-3"
              >
                <h4 className="font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-2">
                  <Sparkles className="h-4 w-4 text-purple-600" />
                  What you'll get
                </h4>
                <ul className="space-y-2 text-sm text-slate-600 dark:text-slate-300">
                  {[
                    'Full admin control as workspace owner',
                    'Invite team members and manage roles',
                    'Secure authentication and data isolation',
                    'Customizable workspace settings'
                  ].map((feature, index) => (
                    <motion.li
                      key={index}
                      initial={{ opacity: 0, x: -10 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.6 + index * 0.1 }}
                      className="flex items-start gap-2"
                    >
                      <Check className="h-4 w-4 text-green-600 mt-0.5 flex-shrink-0" />
                      <span>{feature}</span>
                    </motion.li>
                  ))}
                </ul>
              </motion.div>

              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.7 }}
                className="flex gap-3 pt-4"
              >
                <Button
                  type="button"
                  onClick={handleCancel}
                  variant="outline"
                  className="flex-1 h-12"
                  disabled={loading}
                >
                  <ArrowLeft className="h-4 w-4 mr-2" />
                  Back
                </Button>
                <Button
                  type="submit"
                  disabled={loading || !formData.name || !formData.slug}
                  className="flex-1 h-12 bg-gradient-to-r from-purple-600 via-blue-600 to-indigo-600 hover:from-purple-700 hover:via-blue-700 hover:to-indigo-700 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200"
                >
                  {loading ? (
                    <>
                      <LoadingSpinner size="sm" className="mr-2" />
                      Creating Workspace...
                    </>
                  ) : (
                    <>
                      <Sparkles className="h-4 w-4 mr-2" />
                      Create Workspace
                    </>
                  )}
                </Button>
              </motion.div>
            </form>
          </CardContent>
        </Card>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.8 }}
          className="mt-6 text-center"
        >
          <p className="text-sm text-slate-600 dark:text-slate-400">
            By creating a workspace, you agree to our{' '}
            <button
              type="button"
              onClick={() => window.open('/terms', '_blank')}
              className="text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300 font-medium underline bg-transparent border-0 cursor-pointer"
            >
              Terms of Service
            </button>
          </p>
        </motion.div>
      </motion.div>
    </div>
  );
};

export default CreateOrganization;
