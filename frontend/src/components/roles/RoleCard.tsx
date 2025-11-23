import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Shield, Edit, Settings } from 'lucide-react';

export interface Role {
    id: string;
    name: string;
    display_name: string;
    description?: string;
    is_default?: boolean;
    is_system?: boolean;
    permissions?: any[];
}

interface RoleCardProps {
    role: Role;
    onEdit?: (role: Role) => void;
    onManagePermissions: (role: Role) => void;
}

export function RoleCard({ role, onEdit, onManagePermissions }: RoleCardProps) {
    return (
        <Card className="p-4 hover:shadow-md transition-shadow">
            <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
                <div className="flex items-start space-x-3 flex-1 w-full">
                    <div className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center flex-shrink-0">
                        <Shield className="h-5 w-5 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                            <h3 className="text-lg font-semibold text-foreground">{role.display_name}</h3>
                            {role.is_default && (
                                <span className="px-2 py-0.5 text-xs font-medium bg-success/10 text-success rounded">
                                    Default
                                </span>
                            )}
                        </div>
                        {role.description && (
                            <p className="text-sm text-muted-foreground mt-1">{role.description}</p>
                        )}
                        <p className="text-xs text-muted-foreground mt-1 font-mono">{role.name}</p>
                        {role.permissions && role.permissions.length > 0 && (
                            <div className="flex flex-wrap gap-1 mt-2">
                                {role.permissions.slice(0, 3).map((perm: any, index: number) => (
                                    <span key={perm.id || `perm-${index}`} className="px-2 py-0.5 text-xs bg-secondary/10 text-secondary-foreground rounded">
                                        {perm.display_name || perm.name}
                                    </span>
                                ))}
                                {role.permissions.length > 3 && (
                                    <span className="px-2 py-0.5 text-xs bg-muted text-muted-foreground rounded">
                                        +{role.permissions.length - 3} more
                                    </span>
                                )}
                            </div>
                        )}
                    </div>
                </div>
                <div className="flex items-center justify-end space-x-2 w-full sm:w-auto">
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onManagePermissions(role)}
                        className="flex-1 sm:flex-none"
                    >
                        <Settings className="h-4 w-4 mr-1" />
                        Permissions
                    </Button>
                    {onEdit && (
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onEdit(role)}
                        >
                            <Edit className="h-4 w-4" />
                        </Button>
                    )}
                </div>
            </div>
        </Card>
    );
}
