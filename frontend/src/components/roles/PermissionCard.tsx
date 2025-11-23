import { Card } from '@/components/ui/card';
import { Key } from 'lucide-react';

export interface Permission {
    id: string;
    name: string;
    display_name: string;
    description?: string;
    category?: string;
    is_system?: boolean;
}

interface PermissionCardProps {
    permission: Permission;
}

export function PermissionCard({ permission }: PermissionCardProps) {
    return (
        <Card className="p-4 hover:shadow-md transition-shadow">
            <div className="flex items-start space-x-3">
                <div className="w-10 h-10 bg-secondary/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Key className="h-5 w-5 text-secondary-foreground" />
                </div>
                <div className="flex-1 min-w-0">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                        <h3 className="text-base font-semibold text-foreground break-all">{permission.display_name}</h3>
                        {permission.category && (
                            <span className="px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary rounded capitalize flex-shrink-0">
                                {permission.category}
                            </span>
                        )}
                    </div>
                    {permission.description && (
                        <p className="text-sm text-muted-foreground mt-1">{permission.description}</p>
                    )}
                    <code className="text-xs text-muted-foreground mt-1 block font-mono bg-muted px-2 py-1 rounded inline-block">
                        {permission.name}
                    </code>
                </div>
            </div>
        </Card>
    );
}
