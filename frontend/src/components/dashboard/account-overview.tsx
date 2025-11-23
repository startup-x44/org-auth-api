import { Card } from '../ui/card';
import { Button } from '../ui/button';
import { Mail, Phone, MapPin, Shield, Activity, Calendar, Clock, CheckCircle, XCircle, User } from 'lucide-react';

interface AccountOverviewProps {
    user: {
        email: string;
        phone?: string;
        address?: string;
        global_role: string;
        is_active: boolean;
        created_at: string;
        last_login?: string;
    };
    profile: {
        email: string;
        phone?: string;
        address?: string;
        global_role: string;
        is_active: boolean;
        created_at: string;
        last_login?: string;
    } | null;
    onEditProfile: () => void;
}

export function AccountOverview({ user, profile, onEditProfile }: AccountOverviewProps) {
    const displayData = profile || user;

    return (
        <Card className="p-6 shadow-md border-none bg-card/80 backdrop-blur-sm h-full">
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-3">
                    <div className="p-2 bg-secondary/10 rounded-lg">
                        <User className="h-5 w-5 text-secondary-foreground" />
                    </div>
                    <h3 className="text-lg font-bold text-foreground">Account Information</h3>
                </div>
                <Button variant="outline" size="sm" onClick={onEditProfile}>
                    Edit Profile
                </Button>
            </div>

            <div className="space-y-5">
                <div className="flex items-center p-3 rounded-lg hover:bg-muted/50 transition-colors">
                    <div className="p-2 bg-primary/10 rounded-full mr-4">
                        <Mail className="h-5 w-5 text-primary" />
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Email Address</p>
                        <p className="text-sm font-medium text-foreground">{displayData.email}</p>
                    </div>
                </div>

                {displayData.phone && (
                    <div className="flex items-center p-3 rounded-lg hover:bg-muted/50 transition-colors">
                        <div className="p-2 bg-success/10 rounded-full mr-4">
                            <Phone className="h-5 w-5 text-success" />
                        </div>
                        <div>
                            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Phone Number</p>
                            <p className="text-sm font-medium text-foreground">{displayData.phone}</p>
                        </div>
                    </div>
                )}

                {displayData.address && (
                    <div className="flex items-center p-3 rounded-lg hover:bg-muted/50 transition-colors">
                        <div className="p-2 bg-secondary/10 rounded-full mr-4">
                            <MapPin className="h-5 w-5 text-secondary-foreground" />
                        </div>
                        <div>
                            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Address</p>
                            <p className="text-sm font-medium text-foreground">{displayData.address}</p>
                        </div>
                    </div>
                )}

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-4 border-t border-border">
                    <div className="flex items-center">
                        <Shield className="h-4 w-4 text-muted-foreground mr-2" />
                        <div>
                            <p className="text-xs text-muted-foreground">Role</p>
                            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary mt-0.5">
                                {displayData.global_role}
                            </span>
                        </div>
                    </div>

                    <div className="flex items-center">
                        <Activity className="h-4 w-4 text-muted-foreground mr-2" />
                        <div>
                            <p className="text-xs text-muted-foreground">Status</p>
                            <div className="flex items-center mt-0.5">
                                {displayData.is_active ? (
                                    <>
                                        <CheckCircle className="h-3 w-3 text-success mr-1" />
                                        <span className="text-xs font-medium text-success">Active</span>
                                    </>
                                ) : (
                                    <>
                                        <XCircle className="h-3 w-3 text-destructive mr-1" />
                                        <span className="text-xs font-medium text-destructive">Inactive</span>
                                    </>
                                )}
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center">
                        <Calendar className="h-4 w-4 text-muted-foreground mr-2" />
                        <div>
                            <p className="text-xs text-muted-foreground">Joined</p>
                            <p className="text-xs font-medium text-foreground mt-0.5">
                                {new Date(displayData.created_at).toLocaleDateString('en-US', {
                                    month: 'short',
                                    year: 'numeric'
                                })}
                            </p>
                        </div>
                    </div>

                    {displayData.last_login && (
                        <div className="flex items-center">
                            <Clock className="h-4 w-4 text-muted-foreground mr-2" />
                            <div>
                                <p className="text-xs text-muted-foreground">Last Login</p>
                                <p className="text-xs font-medium text-foreground mt-0.5">
                                    {new Date(displayData.last_login).toLocaleDateString('en-US', {
                                        month: 'short',
                                        day: 'numeric'
                                    })}
                                </p>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </Card>
    );
}
