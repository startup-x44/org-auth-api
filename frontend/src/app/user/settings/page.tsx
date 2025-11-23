'use client';

import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Settings as SettingsIcon, Bell, Lock, Palette, Globe } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

export default function SettingsPage() {
    const { toast } = useToast();

    const handleSave = (section: string) => {
        toast({
            title: "Settings saved",
            description: `${section} settings have been updated.`,
        });
    };

    return (
        <div className="space-y-6">
            {/* Page Header */}
            <div>
                <h1 className="text-2xl font-bold text-foreground">Settings</h1>
                <p className="text-sm text-muted-foreground mt-1">Manage your preferences and account settings</p>
            </div>

            {/* Notifications Settings */}
            <Card className="p-6 border-none shadow-sm">
                <div className="flex items-center space-x-3 mb-4">
                    <div className="p-2 bg-primary/10 rounded-lg">
                        <Bell className="h-5 w-5 text-primary" />
                    </div>
                    <h3 className="text-lg font-bold text-foreground">Notifications</h3>
                </div>
                <div className="space-y-4">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-3 border-b border-border">
                        <div>
                            <p className="font-medium text-foreground">Email Notifications</p>
                            <p className="text-sm text-muted-foreground">Receive email updates about your account</p>
                        </div>
                        <Switch defaultChecked />
                    </div>
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-3 border-b border-border">
                        <div>
                            <p className="font-medium text-foreground">Organization Updates</p>
                            <p className="text-sm text-muted-foreground">Get notified about organization changes</p>
                        </div>
                        <Switch defaultChecked />
                    </div>
                </div>
                <div className="mt-6 flex justify-end">
                    <Button onClick={() => handleSave('Notification')}>Save Changes</Button>
                </div>
            </Card>

            {/* Security Settings */}
            <Card className="p-6 border-none shadow-sm">
                <div className="flex items-center space-x-3 mb-4">
                    <div className="p-2 bg-destructive/10 rounded-lg">
                        <Lock className="h-5 w-5 text-destructive" />
                    </div>
                    <h3 className="text-lg font-bold text-foreground">Security</h3>
                </div>
                <div className="space-y-4">
                    <div className="py-3 border-b border-border">
                        <p className="font-medium text-foreground mb-2">Password</p>
                        <Button variant="outline" size="sm">Change Password</Button>
                    </div>
                    <div className="py-3 border-b border-border">
                        <p className="font-medium text-foreground mb-2">Two-Factor Authentication</p>
                        <p className="text-sm text-muted-foreground mb-3">Add an extra layer of security to your account</p>
                        <Button variant="outline" size="sm">Enable 2FA</Button>
                    </div>
                    <div className="py-3">
                        <p className="font-medium text-foreground mb-2">Active Sessions</p>
                        <p className="text-sm text-muted-foreground mb-3">Manage your active sessions across devices</p>
                        <Button variant="outline" size="sm">View Sessions</Button>
                    </div>
                </div>
            </Card>

            {/* Appearance Settings */}
            <Card className="p-6 border-none shadow-sm">
                <div className="flex items-center space-x-3 mb-4">
                    <div className="p-2 bg-secondary/10 rounded-lg">
                        <Palette className="h-5 w-5 text-secondary-foreground" />
                    </div>
                    <h3 className="text-lg font-bold text-foreground">Appearance</h3>
                </div>
                <div className="space-y-4">
                    <div className="py-3">
                        <p className="font-medium text-foreground mb-3">Theme</p>
                        <div className="grid grid-cols-3 gap-4">
                            <button className="p-4 border-2 border-primary rounded-lg bg-background hover:bg-muted/50 transition-colors">
                                <div className="text-sm font-medium text-foreground">Light</div>
                            </button>
                            <button className="p-4 border-2 border-border rounded-lg bg-foreground hover:bg-foreground/80 transition-colors">
                                <div className="text-sm font-medium text-background">Dark</div>
                            </button>
                            <button className="p-4 border-2 border-border rounded-lg bg-gradient-to-br from-background to-foreground hover:opacity-90 transition-opacity">
                                <div className="text-sm font-medium text-foreground">Auto</div>
                            </button>
                        </div>
                    </div>
                </div>
                <div className="mt-6 flex justify-end">
                    <Button onClick={() => handleSave('Appearance')}>Save Changes</Button>
                </div>
            </Card>

            {/* Language & Region */}
            <Card className="p-6 border-none shadow-sm">
                <div className="flex items-center space-x-3 mb-4">
                    <div className="p-2 bg-accent/10 rounded-lg">
                        <Globe className="h-5 w-5 text-accent-foreground" />
                    </div>
                    <h3 className="text-lg font-bold text-foreground">Language & Region</h3>
                </div>
                <div className="space-y-4">
                    <div className="py-3">
                        <p className="font-medium text-foreground mb-2">Language</p>
                        <Select defaultValue="en">
                            <SelectTrigger className="w-full">
                                <SelectValue placeholder="Select a language" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="en">English (US)</SelectItem>
                                <SelectItem value="es">Spanish</SelectItem>
                                <SelectItem value="fr">French</SelectItem>
                                <SelectItem value="de">German</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                    <div className="py-3">
                        <p className="font-medium text-foreground mb-2">Timezone</p>
                        <Select defaultValue="utc-8">
                            <SelectTrigger className="w-full">
                                <SelectValue placeholder="Select a timezone" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="utc-8">UTC-8 (Pacific Time)</SelectItem>
                                <SelectItem value="utc-5">UTC-5 (Eastern Time)</SelectItem>
                                <SelectItem value="utc+0">UTC+0 (GMT)</SelectItem>
                                <SelectItem value="utc+8">UTC+8 (Singapore Time)</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
                <div className="mt-6 flex justify-end">
                    <Button onClick={() => handleSave('Language & Region')}>Save Changes</Button>
                </div>
            </Card>
        </div>
    );
}
