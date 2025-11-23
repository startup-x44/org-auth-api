'use client';

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { AuthService } from '../../lib/auth';
import { Card } from '../../components/ui/card';
import { Button } from '../../components/ui/button';
import { Input } from '../../components/ui/input';
import { Label } from '../../components/ui/label';
import { LoadingSpinner } from '../../components/ui/loading-spinner';
import { useToast } from '../../hooks/use-toast';
import { Mail, Building2, UserPlus, CheckCircle2, Eye, EyeOff, Check, X } from 'lucide-react';

function AcceptInvitationContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const { toast } = useToast();

    const [loading, setLoading] = useState(true);
    const [processing, setProcessing] = useState(false);
    const [userExists, setUserExists] = useState(false);
    const [invitationData, setInvitationData] = useState<any>(null);
    const [invitationError, setInvitationError] = useState<string | null>(null);

    // Registration form state
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [firstName, setFirstName] = useState('');
    const [lastName, setLastName] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

    const token = searchParams.get('token');
    const emailParam = searchParams.get('email');

    // Password validation
    const passwordValidation = {
        minLength: password.length >= 8,
        hasUpperCase: /[A-Z]/.test(password),
        hasLowerCase: /[a-z]/.test(password),
        hasNumber: /[0-9]/.test(password),
        hasSpecialChar: /[!@#$%^&*(),.?":{}|<>]/.test(password),
    };

    const isPasswordValid = Object.values(passwordValidation).every(Boolean);
    const passwordsMatch = password === confirmPassword && confirmPassword.length > 0;

    useEffect(() => {
        if (token && emailParam) {
            setEmail(decodeURIComponent(emailParam));
            checkInvitation();
        } else {
            toast({
                title: "Invalid invitation link",
                description: "The invitation link is missing required parameters.",
                variant: "destructive",
            });
            router.push('/auth/login');
        }
    }, [token, emailParam]);

    const checkInvitation = async () => {
        try {
            setLoading(true);
            
            // Try to get invitation details (public endpoint, no auth required)
            const response = await fetch(`http://localhost:8080/api/v1/invitations/${token}`);

            if (response.ok) {
                const data = await response.json();
                setInvitationData(data.data || {});
                setInvitationError(null);
            } else {
                const error = await response.json();
                // Invitation not found or expired
                setInvitationError(error.message || 'Invitation not found or expired');
            }

            setLoading(false);

        } catch (error) {
            console.error('Error checking invitation:', error);
            setInvitationError('Unable to verify invitation. Please try again later.');
            setLoading(false);
        }
    };

    const handleLoginAndAccept = async () => {
        try {
            setProcessing(true);

            // Step 1: Login
            const loginResponse = await AuthService.apiCall('/api/v1/auth/login', {
                method: 'POST',
                body: JSON.stringify({ email, password }),
            });

            if (!loginResponse.ok) {
                const error = await loginResponse.json();
                throw new Error(error.message || 'Login failed');
            }

            const loginData = await loginResponse.json();
            
            // Store auth token
            localStorage.setItem('auth_token', loginData.data.token.access_token);
            localStorage.setItem('refresh_token', loginData.data.token.refresh_token);

            // Step 2: Accept invitation
            const acceptResponse = await AuthService.apiCall(`/api/v1/invitations/${token}/accept`, {
                method: 'POST',
            });

            if (!acceptResponse.ok) {
                const error = await acceptResponse.json();
                throw new Error(error.message || 'Failed to accept invitation');
            }

            toast({
                title: "Success!",
                description: "Invitation accepted. Redirecting...",
            });

            // Redirect to select organization or dashboard
            setTimeout(() => {
                router.push('/select-organization');
            }, 1500);

        } catch (error) {
            console.error('Error:', error);
            toast({
                title: "Failed to accept invitation",
                description: error instanceof Error ? error.message : "Please try again.",
                variant: "destructive",
            });
        } finally {
            setProcessing(false);
        }
    };

    const handleRegisterAndAccept = async () => {
        if (!isPasswordValid) {
            toast({
                title: "Password requirements not met",
                description: "Please ensure your password meets all the requirements.",
                variant: "destructive",
            });
            return;
        }

        if (!passwordsMatch) {
            toast({
                title: "Passwords don't match",
                description: "Please make sure both passwords are the same.",
                variant: "destructive",
            });
            return;
        }

        try {
            setProcessing(true);

            // Step 1: Register with invitation token
            const registerResponse = await AuthService.apiCall('/api/v1/auth/register', {
                method: 'POST',
                body: JSON.stringify({
                    email,
                    password,
                    confirm_password: confirmPassword,
                    first_name: firstName,
                    last_name: lastName,
                    invitation_token: token,
                }),
            });

            if (!registerResponse.ok) {
                const error = await registerResponse.json();
                
                // Check if user exists but not verified
                if (error.error_code === 'USER_EXISTS' || error.message?.includes('already registered')) {
                    toast({
                        title: "Account already exists",
                        description: "Redirecting to email verification...",
                        variant: "default",
                    });
                    
                    // Store email for OTP verification
                    localStorage.setItem('verification_email', email);
                    
                    // Redirect to OTP verification
                    setTimeout(() => {
                        router.push(`/auth/verify-email?email=${encodeURIComponent(email)}`);
                    }, 1500);
                    return;
                }
                
                throw new Error(error.message || 'Registration failed');
            }

            const registerData = await registerResponse.json();

            // Store email for OTP verification
            localStorage.setItem('verification_email', email);

            toast({
                title: "Success!",
                description: "Account created! Please verify your email with the OTP code sent.",
            });

            // Redirect to OTP verification
            setTimeout(() => {
                router.push(`/auth/verify-email?email=${encodeURIComponent(email)}`);
            }, 1500);

        } catch (error) {
            console.error('Error:', error);
            toast({
                title: "Failed to accept invitation",
                description: error instanceof Error ? error.message : "Please try again.",
                variant: "destructive",
            });
        } finally {
            setProcessing(false);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-background">
                <LoadingSpinner size="lg" />
            </div>
        );
    }

    // Show error state if invitation is invalid
    if (invitationError) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-background p-4">
                <Card className="w-full max-w-md p-8 border-none shadow-lg">
                    <div className="text-center mb-6">
                        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-destructive/10 mb-4">
                            <X className="h-8 w-8 text-destructive" />
                        </div>
                        <h1 className="text-2xl font-bold text-foreground mb-2">
                            Invitation Not Found
                        </h1>
                        <p className="text-sm text-muted-foreground mb-4">
                            {invitationError}
                        </p>
                        <div className="space-y-2 text-left bg-muted/50 p-4 rounded-lg mb-6">
                            <p className="text-sm text-muted-foreground">This could mean:</p>
                            <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
                                <li>The invitation link has expired</li>
                                <li>The invitation has already been used</li>
                                <li>The invitation was cancelled by the organization</li>
                                <li>The link is invalid or incomplete</li>
                            </ul>
                        </div>
                        <div className="space-y-3">
                            <Button
                                className="w-full"
                                onClick={() => router.push('/auth/login')}
                            >
                                Go to Login
                            </Button>
                            <Button
                                variant="outline"
                                className="w-full"
                                onClick={() => router.push('/auth/register')}
                            >
                                Create New Account
                            </Button>
                        </div>
                        <p className="text-xs text-muted-foreground mt-6">
                            Need help? Contact the organization that sent you the invitation.
                        </p>
                    </div>
                </Card>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex items-center justify-center bg-background p-4">
            <Card className="w-full max-w-md p-8 border-none shadow-lg">
                <div className="text-center mb-6">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
                        <UserPlus className="h-8 w-8 text-primary" />
                    </div>
                    <h1 className="text-2xl font-bold text-foreground mb-2">
                        Organization Invitation
                    </h1>
                    <p className="text-sm text-muted-foreground">
                        You've been invited to join {invitationData?.organizationName || 'an organization'}
                    </p>
                    {invitationData?.roleName && (
                        <p className="text-xs text-muted-foreground mt-1">
                            Role: <span className="font-medium">{invitationData.roleName}</span>
                        </p>
                    )}
                </div>

                <div className="mb-6 p-4 bg-muted/50 rounded-lg flex items-center space-x-3">
                    <Mail className="h-5 w-5 text-muted-foreground flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                        <p className="text-sm text-muted-foreground">Invited email</p>
                        <p className="text-sm font-medium text-foreground truncate">{email}</p>
                    </div>
                </div>

                <div className="space-y-6">
                    {/* Tab selection */}
                    <div className="flex space-x-2 p-1 bg-muted rounded-lg">
                        <button
                            className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
                                !userExists
                                    ? 'bg-background text-foreground shadow-sm'
                                    : 'text-muted-foreground hover:text-foreground'
                            }`}
                            onClick={() => setUserExists(false)}
                        >
                            Create Account
                        </button>
                        <button
                            className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
                                userExists
                                    ? 'bg-background text-foreground shadow-sm'
                                    : 'text-muted-foreground hover:text-foreground'
                            }`}
                            onClick={() => setUserExists(true)}
                        >
                            Login
                        </button>
                    </div>

                    {userExists ? (
                        // Login form
                        <div className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="password">Password</Label>
                                <div className="relative">
                                    <Input
                                        id="password"
                                        type={showPassword ? "text" : "password"}
                                        placeholder="Enter your password"
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                        disabled={processing}
                                        className="pr-10"
                                    />
                                    <button
                                        type="button"
                                        onClick={() => setShowPassword(!showPassword)}
                                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                    >
                                        {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                    </button>
                                </div>
                            </div>

                            <Button
                                className="w-full"
                                onClick={handleLoginAndAccept}
                                disabled={processing || !password}
                            >
                                {processing ? (
                                    <div className="flex items-center">
                                        <LoadingSpinner size="sm" variant="secondary" />
                                        <span className="ml-2">Processing...</span>
                                    </div>
                                ) : (
                                    <div className="flex items-center justify-center">
                                        <CheckCircle2 className="h-4 w-4 mr-2" />
                                        Login & Accept Invitation
                                    </div>
                                )}
                            </Button>
                        </div>
                    ) : (
                        // Registration form
                        <div className="space-y-4">
                        <div className="text-center mb-4">
                            <p className="text-sm text-muted-foreground">
                                Create your account to accept the invitation.
                            </p>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="firstName">First Name</Label>
                                <Input
                                    id="firstName"
                                    type="text"
                                    placeholder="John"
                                    value={firstName}
                                    onChange={(e) => setFirstName(e.target.value)}
                                    disabled={processing}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="lastName">Last Name</Label>
                                <Input
                                    id="lastName"
                                    type="text"
                                    placeholder="Doe"
                                    value={lastName}
                                    onChange={(e) => setLastName(e.target.value)}
                                    disabled={processing}
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="newPassword">Password</Label>
                            <div className="relative">
                                <Input
                                    id="newPassword"
                                    type={showPassword ? "text" : "password"}
                                    placeholder="Create a strong password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    disabled={processing}
                                    className="pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                >
                                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </button>
                            </div>
                            
                            {/* Password strength indicator */}
                            {password.length > 0 && (
                                <div className="space-y-2 mt-2 p-3 bg-muted/50 rounded-md">
                                    <p className="text-xs font-medium text-muted-foreground mb-2">Password must contain:</p>
                                    <div className="space-y-1">
                                        <div className="flex items-center text-xs">
                                            {passwordValidation.minLength ? (
                                                <Check className="h-3 w-3 mr-2 text-success" />
                                            ) : (
                                                <X className="h-3 w-3 mr-2 text-muted-foreground" />
                                            )}
                                            <span className={passwordValidation.minLength ? 'text-success' : 'text-muted-foreground'}>
                                                At least 8 characters
                                            </span>
                                        </div>
                                        <div className="flex items-center text-xs">
                                            {passwordValidation.hasUpperCase ? (
                                                <Check className="h-3 w-3 mr-2 text-success" />
                                            ) : (
                                                <X className="h-3 w-3 mr-2 text-muted-foreground" />
                                            )}
                                            <span className={passwordValidation.hasUpperCase ? 'text-success' : 'text-muted-foreground'}>
                                                One uppercase letter
                                            </span>
                                        </div>
                                        <div className="flex items-center text-xs">
                                            {passwordValidation.hasLowerCase ? (
                                                <Check className="h-3 w-3 mr-2 text-success" />
                                            ) : (
                                                <X className="h-3 w-3 mr-2 text-muted-foreground" />
                                            )}
                                            <span className={passwordValidation.hasLowerCase ? 'text-success' : 'text-muted-foreground'}>
                                                One lowercase letter
                                            </span>
                                        </div>
                                        <div className="flex items-center text-xs">
                                            {passwordValidation.hasNumber ? (
                                                <Check className="h-3 w-3 mr-2 text-success" />
                                            ) : (
                                                <X className="h-3 w-3 mr-2 text-muted-foreground" />
                                            )}
                                            <span className={passwordValidation.hasNumber ? 'text-success' : 'text-muted-foreground'}>
                                                One number
                                            </span>
                                        </div>
                                        <div className="flex items-center text-xs">
                                            {passwordValidation.hasSpecialChar ? (
                                                <Check className="h-3 w-3 mr-2 text-success" />
                                            ) : (
                                                <X className="h-3 w-3 mr-2 text-muted-foreground" />
                                            )}
                                            <span className={passwordValidation.hasSpecialChar ? 'text-success' : 'text-muted-foreground'}>
                                                One special character (!@#$%^&*...)
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="confirmPassword">Confirm Password</Label>
                            <div className="relative">
                                <Input
                                    id="confirmPassword"
                                    type={showConfirmPassword ? "text" : "password"}
                                    placeholder="Re-enter your password"
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    disabled={processing}
                                    className="pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                >
                                    {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </button>
                            </div>
                            {confirmPassword.length > 0 && (
                                <div className="flex items-center text-xs mt-1">
                                    {passwordsMatch ? (
                                        <>
                                            <Check className="h-3 w-3 mr-1 text-success" />
                                            <span className="text-success">Passwords match</span>
                                        </>
                                    ) : (
                                        <>
                                            <X className="h-3 w-3 mr-1 text-destructive" />
                                            <span className="text-destructive">Passwords do not match</span>
                                        </>
                                    )}
                                </div>
                            )}
                        </div>

                        <Button
                            className="w-full"
                            onClick={handleRegisterAndAccept}
                            disabled={processing || !firstName || !lastName || !isPasswordValid || !passwordsMatch}
                        >
                            {processing ? (
                                <div className="flex items-center">
                                    <LoadingSpinner size="sm" variant="secondary" />
                                    <span className="ml-2">Creating account...</span>
                                </div>
                            ) : (
                                <div className="flex items-center justify-center">
                                    <CheckCircle2 className="h-4 w-4 mr-2" />
                                    Create Account & Accept
                                </div>
                            )}
                        </Button>

                        <div className="text-center">
                            <p className="text-xs text-muted-foreground">
                                Already have an account?{' '}
                                <Button
                                    variant="link"
                                    className="text-xs p-0 h-auto"
                                    onClick={() => router.push('/login')}
                                >
                                    Login here
                                </Button>
                            </p>
                        </div>
                    </div>
                )}
                </div>
            </Card>
        </div>
    );
}

export default function AcceptInvitationPage() {
    return (
        <Suspense fallback={
            <div className="min-h-screen flex items-center justify-center bg-background">
                <LoadingSpinner size="lg" />
            </div>
        }>
            <AcceptInvitationContent />
        </Suspense>
    );
}
