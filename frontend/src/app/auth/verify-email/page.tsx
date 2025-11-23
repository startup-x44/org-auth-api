'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { useToast } from '../../../hooks/use-toast';
import { makeSecureRequest } from '../../../lib/csrf';
import {
  Mail,
  Shield,
  AlertCircle,
  CheckCircle,
  ArrowLeft,
  RefreshCw
} from 'lucide-react';
import Link from 'next/link';

export default function VerifyEmailPage() {
  const [verificationCode, setVerificationCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [resendLoading, setResendLoading] = useState(false);
  const [error, setError] = useState('');
  const [email, setEmail] = useState('');
  const [timeLeft, setTimeLeft] = useState(300); // 5 minutes countdown
  const [alreadyVerified, setAlreadyVerified] = useState(false);
  const [checkingStatus, setCheckingStatus] = useState(true);

  const router = useRouter();
  const searchParams = useSearchParams();
  const { toast } = useToast();

  useEffect(() => {
    // Get email from URL params or localStorage
    const emailParam = searchParams.get('email');
    const storedEmail = localStorage.getItem('verification_email');

    if (emailParam) {
      setEmail(emailParam);
      localStorage.setItem('verification_email', emailParam);
      checkVerificationStatus(emailParam);
    } else if (storedEmail) {
      setEmail(storedEmail);
      checkVerificationStatus(storedEmail);
    } else {
      // No email found, redirect to register
      router.push('/auth/register');
      return;
    }

    // Start countdown timer
    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [searchParams, router]);

  const checkVerificationStatus = async (emailToCheck: string) => {
    try {
      setCheckingStatus(true);
      
      // Try to login to check if already verified
      const response = await fetch('http://localhost:8080/api/v1/auth/verify-email', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: emailToCheck,
          code: '000000', // Dummy code to trigger verification check
        }),
      });

      const data = await response.json();
      
      // If the error indicates already verified
      if (!response.ok && (
        data.message?.includes('already verified') || 
        data.message?.includes('email verified') ||
        data.error_code === 'ALREADY_VERIFIED'
      )) {
        setAlreadyVerified(true);
      }
    } catch (err) {
      console.log('Could not check verification status');
    } finally {
      setCheckingStatus(false);
    }
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!verificationCode.trim()) {
      setError('Verification code is required');
      return;
    }

    if (verificationCode.length !== 6) {
      setError('Verification code must be 6 digits');
      return;
    }

    setLoading(true);
    setError('');

    try {
      console.log('Verifying email with code:', verificationCode);

      const response = await makeSecureRequest('http://localhost:8080/api/v1/auth/verify-email', {
        method: 'POST',
        body: JSON.stringify({
          email,
          code: verificationCode,
        }),
      });

      const data = await response.json();

      if (response.ok && data.success) {
        // Clear stored email
        localStorage.removeItem('verification_email');

        toast({
          title: "Email verified successfully!",
          description: "Your account is now active. You can sign in.",
          variant: "success",
        });

        // Redirect to login
        setTimeout(() => {
          router.push('/auth/login?verified=true');
        }, 1000);
      } else {
        const errorMessage = data.message || 'Verification failed. Please try again.';
        setError(errorMessage);
      }
    } catch (err) {
      console.error('Verification error:', err);
      const errorMessage = err instanceof Error ? err.message : 'Verification failed. Please try again.';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleResendCode = async () => {
    setResendLoading(true);
    setError('');

    try {
      const response = await makeSecureRequest('http://localhost:8080/api/v1/auth/resend-verification', {
        method: 'POST',
        body: JSON.stringify({ email }),
      });

      const data = await response.json();

      if (response.ok && data.success) {
        toast({
          title: "Verification code sent!",
          description: "A new verification code has been sent to your email.",
          variant: "success",
        });

        // Reset timer
        setTimeLeft(300);
      } else {
        setError(data.message || 'Failed to resend code. Please try again.');
      }
    } catch (err) {
      console.error('Resend error:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to resend code. Please try again.';
      setError(errorMessage);
    } finally {
      setResendLoading(false);
    }
  };

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/\D/g, '').slice(0, 6); // Only digits, max 6
    setVerificationCode(value);
    if (error) setError('');
  };

  // Show loading state while checking
  if (checkingStatus) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  // Show already verified state
  if (alreadyVerified) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full">
          <Card className="p-8 shadow-xl border-0 bg-white backdrop-blur">
            <div className="text-center mb-8">
              <div className="flex justify-center mb-4">
                <div className="p-3 bg-green-100 rounded-full">
                  <CheckCircle className="h-8 w-8 text-green-600" />
                </div>
              </div>
              <h1 className="text-2xl font-bold text-foreground mb-2">
                Email Already Verified
              </h1>
              <p className="text-gray-600 text-sm mb-4">
                Your email <span className="font-medium text-green-600">{email}</span> has already been verified.
              </p>
            </div>

            <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
              <div className="flex items-start">
                <CheckCircle className="h-5 w-5 text-green-500 mr-3 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="text-green-800 font-medium text-sm">Account is Active</p>
                  <p className="text-green-700 text-sm mt-1">
                    You can now sign in to your account and access all features.
                  </p>
                </div>
              </div>
            </div>

            <Button
              className="w-full h-12 text-base font-medium"
              onClick={() => {
                localStorage.removeItem('verification_email');
                router.push('/auth/login?verified=true');
              }}
            >
              <span>Go to Login</span>
            </Button>

            <div className="mt-6 text-center">
              <p className="text-xs text-gray-500">
                Secured by NILOAUTH • Email Verification
              </p>
            </div>
          </Card>
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
              <div className="p-3 bg-blue-100 rounded-full">
                <Mail className="h-8 w-8 text-blue-600" />
              </div>
            </div>
            <h1 className="text-2xl font-bold text-foreground mb-2">
              Verify Your Email
            </h1>
            <p className="text-gray-600 text-sm">
              We've sent a 6-digit verification code to
            </p>
            <p className="text-blue-600 font-medium mt-1">{email}</p>
          </div>

          {/* Error Alert */}
          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center">
              <AlertCircle className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
              <p className="text-red-700 text-sm">{error}</p>
            </div>
          )}

          {/* Verification Form */}
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="code">
                Verification Code
              </Label>
              <Input
                id="code"
                name="code"
                type="text"
                value={verificationCode}
                onChange={handleCodeChange}
                placeholder="Enter 6-digit code"
                disabled={loading}
                className="text-center text-2xl font-mono tracking-widest"
                maxLength={6}
              />
              <p className="text-xs text-gray-500 text-center">
                Enter the 6-digit code from your email
              </p>
            </div>

            {/* Timer */}
            {timeLeft > 0 && (
              <div className="text-center">
                <p className="text-sm text-gray-600">
                  Code expires in: <span className="font-mono font-medium text-blue-600">{formatTime(timeLeft)}</span>
                </p>
              </div>
            )}

            {/* Submit Button */}
            <Button
              type="submit"
              disabled={loading || !verificationCode || verificationCode.length !== 6}
              className="w-full h-12 text-base font-medium"
            >
              {loading ? (
                <div className="flex items-center">
                  <LoadingSpinner size="sm" variant="secondary" />
                  <span className="ml-2">Verifying...</span>
                </div>
              ) : (
                <div className="flex items-center justify-center">
                  <CheckCircle className="mr-2 h-5 w-5" />
                  <span>Verify Email</span>
                </div>
              )}
            </Button>
          </form>

          {/* Resend Code */}
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600 mb-3">
              Didn't receive the code?
            </p>
            <Button
              type="button"
              variant="outline"
              onClick={handleResendCode}
              disabled={resendLoading || timeLeft > 240} // Disable for first 1 minute
              className="text-sm"
            >
              {resendLoading ? (
                <div className="flex items-center">
                  <LoadingSpinner size="sm" variant="primary" />
                  <span className="ml-2">Resending...</span>
                </div>
              ) : (
                <div className="flex items-center">
                  <RefreshCw className="mr-2 h-4 w-4" />
                  <span>Resend Code</span>
                </div>
              )}
            </Button>
            {timeLeft > 240 && (
              <p className="text-xs text-gray-400 mt-2">
                You can resend code in {formatTime(timeLeft - 240)}
              </p>
            )}
          </div>

          {/* Back to Register */}
          <div className="mt-8 pt-6 border-t border-gray-200 text-center">
            <Link
              href="/auth/register"
              className="inline-flex items-center text-sm text-gray-600 hover:text-primary"
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Registration
            </Link>
          </div>

          {/* Footer */}
          <div className="mt-6 text-center">
            <p className="text-xs text-gray-500">
              Secured by NILOAUTH • Email Verification
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
}