import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useState, useEffect } from 'react';

interface CreatePermissionModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (name: string, displayName: string, category: string) => Promise<void>;
    isLoading: boolean;
}

export function CreatePermissionModal({ isOpen, onClose, onSubmit, isLoading }: CreatePermissionModalProps) {
    const [name, setName] = useState('');
    const [displayName, setDisplayName] = useState('');
    const [category, setCategory] = useState('');

    // Auto-generate display name from permission name
    useEffect(() => {
        if (name && !displayName) {
            const parts = name.split(':');
            if (parts.length === 2) {
                const generated = parts.map(p => p.charAt(0).toUpperCase() + p.slice(1).replace(/_/g, ' ')).join(' ');
                setDisplayName(generated);
            }
        }
    }, [name]);

    const handleSubmit = async () => {
        await onSubmit(name, displayName, category);
        setName('');
        setDisplayName('');
        setCategory('');
    };

    const handleClose = () => {
        setName('');
        setDisplayName('');
        setCategory('');
        onClose();
    };

    if (!isOpen) return null;

    return (
        <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Create New Permission</DialogTitle>
                </DialogHeader>
                <div className="space-y-4 py-4">
                    <div className="space-y-2">
                        <Label className="text-sm font-medium text-foreground">
                            Permission Name *
                        </Label>
                        <Input
                            type="text"
                            placeholder="action:resource (e.g., view:profile, edit:content)"
                            value={name}
                            onChange={(e) => setName(e.target.value.toLowerCase())}
                        />
                        <p className="text-xs text-muted-foreground">
                            Format: <code className="bg-muted px-1 rounded">action:resource</code> (lowercase, underscores allowed)
                        </p>
                    </div>
                    <div className="space-y-2">
                        <Label className="text-sm font-medium text-foreground">
                            Display Name *
                        </Label>
                        <Input
                            type="text"
                            placeholder="e.g., View Profile, Edit Content"
                            value={displayName}
                            onChange={(e) => setDisplayName(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground">Human-readable name shown in the UI</p>
                    </div>
                    <div className="space-y-2">
                        <Label className="text-sm font-medium text-foreground">
                            Category *
                        </Label>
                        <Input
                            type="text"
                            placeholder="e.g., users, content, settings"
                            value={category}
                            onChange={(e) => setCategory(e.target.value.toLowerCase())}
                        />
                        <p className="text-xs text-muted-foreground">Group related permissions together</p>
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={handleClose}>
                        Cancel
                    </Button>
                    <Button onClick={handleSubmit} disabled={isLoading || !name.trim() || !displayName.trim() || !category.trim()}>
                        {isLoading ? 'Creating...' : 'Create Permission'}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
