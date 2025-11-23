'use client';

import { useAuth } from '../../../contexts/auth-context';
import { useState, useEffect } from 'react';
import { AuthService } from '../../../lib/auth';
import { Card } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Label } from '../../../components/ui/label';
import { User, Mail, Phone, MapPin, Calendar, Save } from 'lucide-react';
import { useToast } from '../../../hooks/use-toast';

interface UserProfile {
    id: string;
    email: string;
    first_name?: string;
    last_name?: string;
    phone?: string;
    address?: string;
    global_role: string;
    is_superadmin: boolean;
    is_active: boolean;
    created_at: string;
    updated_at: string;
    last_login?: string;
}

export default function ProfilePage() {
    const { user } = useAuth();
    const { toast } = useToast();
    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        loadProfile();
    }, []);

    const loadProfile = async () => {
        try {
            setLoading(true);
            const response = await AuthService.apiCall('/api/v1/user/profile');
            if (response.ok) {
                const data = await response.json();
                setProfile(data.data || data.user || data);
            }
        } catch (error) {
            console.error('Failed to load profile:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        // TODO: Implement profile update API call
        setTimeout(() => {
            setSaving(false);
            toast({
                title: "Profile updated",
                description: "Your profile has been updated successfully.",
            });
        }, 1000);
    };

    if (loading) {
        return (
            <div className="space-y-6">
                <div className="h-8 w-48 bg-muted rounded animate-pulse"></div>
                <div className="h-96 bg-muted rounded-lg animate-pulse"></div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Page Header */}
            <div>
                <h1 className="text-2xl font-bold text-foreground">Profile</h1>
                <p className="text-sm text-muted-foreground mt-1">Manage your account information</p>
            </div>

            {/* Profile Card */}
            <Card className="p-6 border-none shadow-sm">
                <div className="flex flex-col sm:flex-row items-center sm:items-start space-y-4 sm:space-y-0 sm:space-x-4 mb-6 pb-6 border-b border-border">
                    <div className="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center flex-shrink-0">
                        <User className="h-10 w-10 text-primary" />
                    </div>
                    <div className="text-center sm:text-left">
                        <h2 className="text-xl font-bold text-foreground">
                            {profile?.first_name} {profile?.last_name}
                        </h2>
                        <p className="text-sm text-muted-foreground">{profile?.email}</p>
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary mt-2">
                            {profile?.global_role}
                        </span>
                    </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <Label className="block text-sm font-medium text-foreground mb-2">
                            <Mail className="h-4 w-4 inline mr-1" />
                            Email Address
                        </Label>
                        <Input
                            type="email"
                            value={profile?.email || ''}
                            disabled
                            className="bg-muted"
                        />
                        <p className="text-xs text-muted-foreground mt-1">Email cannot be changed</p>
                    </div>

                    <div>
                        <Label className="block text-sm font-medium text-foreground mb-2">
                            <Phone className="h-4 w-4 inline mr-1" />
                            Phone Number
                        </Label>
                        <Input
                            type="tel"
                            value={profile?.phone || ''}
                            onChange={(e) => setProfile({ ...profile!, phone: e.target.value })}
                            placeholder="Enter phone number"
                        />
                    </div>

                    <div className="md:col-span-2">
                        <Label className="block text-sm font-medium text-foreground mb-2">
                            <MapPin className="h-4 w-4 inline mr-1" />
                            Address
                        </Label>
                        <Input
                            type="text"
                            value={profile?.address || ''}
                            onChange={(e) => setProfile({ ...profile!, address: e.target.value })}
                            placeholder="Enter address"
                        />
                    </div>

                    <div>
                        <Label className="block text-sm font-medium text-foreground mb-2">
                            <Calendar className="h-4 w-4 inline mr-1" />
                            Member Since
                        </Label>
                        <Input
                            type="text"
                            value={new Date(profile?.created_at || '').toLocaleDateString('en-US', {
                                year: 'numeric',
                                month: 'long',
                                day: 'numeric'
                            })}
                            disabled
                            className="bg-muted"
                        />
                    </div>

                    <div>
                        <Label className="block text-sm font-medium text-foreground mb-2">
                            Last Login
                        </Label>
                        <Input
                            type="text"
                            value={profile?.last_login ? new Date(profile.last_login).toLocaleString() : 'Never'}
                            disabled
                            className="bg-muted"
                        />
                    </div>
                </div>

                <div className="mt-6 pt-6 border-t border-border flex justify-end">
                    <Button onClick={handleSave} disabled={saving}>
                        <Save className="h-4 w-4 mr-2" />
                        {saving ? 'Saving...' : 'Save Changes'}
                    </Button>
                </div>
            </Card >
        </div >
    );
}
