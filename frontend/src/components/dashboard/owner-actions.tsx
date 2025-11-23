import { Card } from '../ui/card';
import { Button } from '../ui/button';
import { Crown, UserPlus, Plus, Key } from 'lucide-react';

interface OwnerActionsProps {
    ownedOrganizationsCount: number;
    onInvite: () => void;
    onCreateOrg: () => void;
    onManageRoles: () => void;
}

export function OwnerActions({
    ownedOrganizationsCount,
    onInvite,
    onCreateOrg,
    onManageRoles,
}: OwnerActionsProps) {
    return (
        <Card className="p-6 border-none shadow-sm bg-accent/10">
            <div className="absolute top-0 right-0 -mt-4 -mr-4 w-24 h-24 bg-accent rounded-full opacity-10 blur-2xl"></div>
            <div className="absolute bottom-0 left-0 -mb-4 -ml-4 w-32 h-32 bg-accent rounded-full opacity-10 blur-2xl"></div>

            <div className="relative z-10">
                <div className="flex items-center mb-6">
                    <div className="p-3 bg-accent rounded-xl mr-4 shadow-sm">
                        <Crown className="h-6 w-6 text-accent-foreground" />
                    </div>
                    <div>
                        <h3 className="text-lg font-bold text-foreground">Owner Actions</h3>
                        <p className="text-sm text-muted-foreground font-medium">Manage your organizations</p>
                    </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <Button
                        className="flex flex-col items-center justify-center py-6 h-auto bg-card/60 hover:bg-card border-2 border-accent/30 hover:border-accent/50 text-foreground shadow-sm hover:shadow-md transition-all duration-200 group"
                        variant="outline"
                        onClick={onInvite}
                    >
                        <div className="p-2 bg-primary/10 rounded-full mb-2 group-hover:scale-110 transition-transform duration-200">
                            <UserPlus className="h-6 w-6 text-primary" />
                        </div>
                        <span className="font-semibold">Invite Members</span>
                        <span className="text-xs text-muted-foreground mt-1">Add to team</span>
                    </Button>

                    <Button
                        className="flex flex-col items-center justify-center py-6 h-auto bg-card/60 hover:bg-card border-2 border-success/20 hover:border-success/40 text-foreground shadow-sm hover:shadow-md transition-all duration-200 group"
                        variant="outline"
                        onClick={onCreateOrg}
                    >
                        <div className="p-2 bg-success/10 rounded-full mb-2 group-hover:scale-110 transition-transform duration-200">
                            <Plus className="h-6 w-6 text-success" />
                        </div>
                        <span className="font-semibold">New Organization</span>
                        <span className="text-xs text-muted-foreground mt-1">Start fresh</span>
                    </Button>

                    <Button
                        className="flex flex-col items-center justify-center py-6 h-auto bg-card/60 hover:bg-card border-2 border-secondary/20 hover:border-secondary/30 text-foreground shadow-sm hover:shadow-md transition-all duration-200 group"
                        variant="outline"
                        onClick={onManageRoles}
                    >
                        <div className="p-2 bg-secondary/10 rounded-full mb-2 group-hover:scale-110 transition-transform duration-200">
                            <Key className="h-6 w-6 text-secondary-foreground" />
                        </div>
                        <span className="font-semibold">Manage Roles</span>
                        <span className="text-xs text-muted-foreground mt-1">Permissions</span>
                    </Button>
                </div>

                <div className="mt-6 flex items-center justify-center">
                    <div className="inline-flex items-center px-4 py-2 bg-accent/20 rounded-full border border-accent/30 backdrop-blur-sm">
                        <Crown className="h-4 w-4 text-accent-foreground mr-2" />
                        <p className="text-sm text-accent-foreground font-medium">
                            You own <strong>{ownedOrganizationsCount}</strong> organization{ownedOrganizationsCount !== 1 ? 's' : ''}
                        </p>
                    </div>
                </div>
            </div>
        </Card>
    );
}
