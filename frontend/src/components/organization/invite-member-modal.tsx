'use client';

import { useState } from 'react';
import { AuthService } from '../../lib/auth';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Card } from '../ui/card';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import { LoadingSpinner } from '../ui/loading-spinner';
import { useToast } from '../../hooks/use-toast';
import {
  X,
  Mail,
  UserPlus,
  Send,
  Plus,
  Trash2
} from 'lucide-react';

interface InviteMemberModalProps {
  isOpen: boolean;
  onClose: () => void;
  organizationId: string;
  organizationName: string;
}

interface InviteData {
  email: string;
  role: string;
}

export function InviteMemberModal({ isOpen, onClose, organizationId, organizationName }: InviteMemberModalProps) {
  const [invites, setInvites] = useState<InviteData[]>([{ email: '', role: 'member' }]);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  if (!isOpen) return null;

  const addInviteField = () => {
    setInvites([...invites, { email: '', role: 'member' }]);
  };

  const removeInviteField = (index: number) => {
    if (invites.length > 1) {
      setInvites(invites.filter((_, i) => i !== index));
    }
  };

  const updateInvite = (index: number, field: keyof InviteData, value: string) => {
    const updated = invites.map((invite, i) =>
      i === index ? { ...invite, [field]: value } : invite
    );
    setInvites(updated);
  };

  const handleSendInvites = async () => {
    const validInvites = invites.filter(invite => invite.email.trim() !== '');

    if (validInvites.length === 0) {
      toast({
        title: "No valid invites",
        description: "Please enter at least one email address.",
        variant: "destructive",
      });
      return;
    }

    setLoading(true);
    try {
      const response = await AuthService.apiCall(`/api/v1/organizations/${organizationId}/members`, {
        method: 'POST',
        body: JSON.stringify({ invites: validInvites }),
      });

      if (response.ok) {
        toast({
          title: "Invites sent!",
          description: `Successfully sent ${validInvites.length} invitation${validInvites.length !== 1 ? 's' : ''} to ${organizationName}.`,
          variant: "success",
        });
        onClose();
        setInvites([{ email: '', role: 'member' }]);
      } else {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to send invites');
      }
    } catch (error) {
      console.error('Invite error:', error);
      toast({
        title: "Failed to send invites",
        description: error instanceof Error ? error.message : "Please try again later.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-hidden p-0 gap-0">
        <DialogHeader className="p-6 border-b">
          <div className="flex items-center">
            <div className="p-2 bg-primary/10 rounded-lg mr-3">
              <UserPlus className="h-5 w-5 text-primary" />
            </div>
            <div>
              <DialogTitle className="text-lg font-semibold text-foreground">Invite Members</DialogTitle>
              <p className="text-sm text-muted-foreground">Add people to {organizationName}</p>
            </div>
          </div>
        </DialogHeader>

        <div className="p-6 max-h-96 overflow-y-auto">
          <div className="space-y-4">
            {invites.map((invite, index) => (
              <div key={index} className="flex items-center space-x-3 p-4 border border-border rounded-lg">
                <div className="flex-1">
                  <div className="flex items-center space-x-3">
                    <Mail className="h-4 w-4 text-muted-foreground" />
                    <Input
                      type="email"
                      placeholder="colleague@company.com"
                      value={invite.email}
                      onChange={(e) => updateInvite(index, 'email', e.target.value)}
                      className="flex-1"
                    />
                  </div>
                </div>

                <Select
                  value={invite.role}
                  onValueChange={(value) => updateInvite(index, 'role', value)}
                >
                  <SelectTrigger className="w-[120px]">
                    <SelectValue placeholder="Select role" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="member">Member</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                    <SelectItem value="manager">Manager</SelectItem>
                  </SelectContent>
                </Select>

                {invites.length > 1 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => removeInviteField(index)}
                    className="text-red-600 hover:text-red-700"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </div>
            ))}
          </div>

          <Button
            variant="outline"
            onClick={addInviteField}
            className="mt-4 w-full"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Another Invite
          </Button>
        </div>

        <DialogFooter className="flex items-center justify-between p-6 border-t bg-muted/50 sm:justify-between">
          <div className="text-sm text-muted-foreground">
            {invites.filter(invite => invite.email.trim() !== '').length} invite{invites.filter(invite => invite.email.trim() !== '').length !== 1 ? 's' : ''} ready to send
          </div>
          <div className="flex items-center space-x-3">
            <Button variant="outline" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button onClick={handleSendInvites} disabled={loading}>
              {loading ? (
                <div className="flex items-center">
                  <LoadingSpinner size="sm" variant="secondary" />
                  <span className="ml-2">Sending...</span>
                </div>
              ) : (
                <div className="flex items-center">
                  <Send className="h-4 w-4 mr-2" />
                  <span>Send Invites</span>
                </div>
              )}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}