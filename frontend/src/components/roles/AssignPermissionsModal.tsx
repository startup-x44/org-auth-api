import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';


export interface Permission {
    id: string;
    name: string;
    display_name: string;
    description?: string;
    category?: string;
    is_system?: boolean;
}

interface AssignPermissionsModalProps {
    isOpen: boolean;
    roleName: string;
    permissions: Permission[];
    selectedPermissions: string[];
    onPermissionToggle: (permissionId: string, checked: boolean) => void;
    onSave: () => Promise<void>;
    onClose: () => void;
    isLoading: boolean;
}

export function AssignPermissionsModal({
    isOpen,
    roleName,
    permissions,
    selectedPermissions,
    onPermissionToggle,
    onSave,
    onClose,
    isLoading
}: AssignPermissionsModalProps) {
    if (!isOpen) return null;

    // Group permissions by category
    const groupedPermissions = permissions.reduce((acc, perm) => {
        const category = perm.category || 'other';
        if (!acc[category]) {
            acc[category] = [];
        }
        acc[category].push(perm);
        return acc;
    }, {} as Record<string, Permission[]>);

    // Check if all permissions in a category are selected
    const isCategoryFullySelected = (categoryPerms: Permission[]) => {
        return categoryPerms.every(perm => selectedPermissions.includes(perm.id));
    };

    // Check if some (but not all) permissions in a category are selected
    const isCategoryPartiallySelected = (categoryPerms: Permission[]) => {
        const selectedCount = categoryPerms.filter(perm => selectedPermissions.includes(perm.id)).length;
        return selectedCount > 0 && selectedCount < categoryPerms.length;
    };

    // Toggle all permissions in a category
    const handleCategoryToggle = (categoryPerms: Permission[], checked: boolean) => {
        categoryPerms.forEach(perm => {
            onPermissionToggle(perm.id, checked);
        });
    };

    return (
        <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Manage Permissions for "{roleName}"</DialogTitle>
                </DialogHeader>
                <div className="space-y-6">
                    {Object.entries(groupedPermissions).map(([category, categoryPermissions]) => {
                        const allSelected = isCategoryFullySelected(categoryPermissions);
                        const partiallySelected = isCategoryPartiallySelected(categoryPermissions);

                        return (
                            <div key={category} className="border-2 border-primary/30 rounded-lg p-4 bg-primary/5">
                                <div className="flex items-center justify-between mb-3 pb-2 border-b-2 border-primary/20">
                                    <h4 className="font-bold capitalize text-foreground text-lg">
                                        {category} Permissions ({categoryPermissions.length})
                                    </h4>
                                    <div className="flex items-center space-x-2">
                                        <Checkbox
                                            id={`category-${category}`}
                                            checked={allSelected}
                                            onCheckedChange={(checked) => handleCategoryToggle(categoryPermissions, checked === true)}
                                        />
                                        <Label
                                            htmlFor={`category-${category}`}
                                            className="text-sm font-semibold text-foreground cursor-pointer"
                                        >
                                            Select All
                                        </Label>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    {categoryPermissions.map((permission) => (
                                        <div key={permission.id} className="flex items-center space-x-3 p-3 border border-border rounded-lg hover:bg-card bg-card">
                                            <Checkbox
                                                id={permission.id}
                                                checked={selectedPermissions.includes(permission.id)}
                                                onCheckedChange={(checked) => onPermissionToggle(permission.id, checked === true)}
                                            />
                                            <div className="flex-1">
                                                <Label htmlFor={permission.id} className="font-semibold text-foreground cursor-pointer block">
                                                    {permission.display_name}
                                                </Label>
                                                {permission.description && (
                                                    <p className="text-sm text-muted-foreground mt-1">
                                                        {permission.description}
                                                    </p>
                                                )}
                                                <code className="text-xs bg-muted px-2 py-1 rounded text-muted-foreground mt-1 inline-block">
                                                    {permission.name}
                                                </code>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        );
                    })}
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button onClick={onSave} disabled={isLoading}>
                        {isLoading ? 'Saving...' : 'Save Permissions'}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
