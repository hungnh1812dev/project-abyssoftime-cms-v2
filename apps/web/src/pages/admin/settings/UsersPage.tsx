import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useAuth } from '@/context/AuthContext';
import { roleLevel } from '@/lib/roles';
import { useUserList, useUpdateUserRole, useDeleteUser } from '@/hooks/useUsers';
import { useInviteList, useCreateInvite, useRevokeInvite } from '@/hooks/useInvites';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

const ALL_ROLES = ['super_admin', 'admin', 'editor', 'guest'] as const;

function rolesBelow(currentRole: string | null): string[] {
  const level = roleLevel(currentRole);
  return ALL_ROLES.filter((role) => roleLevel(role) < level);
}

interface InviteFields {
  email: string;
}

export function UsersPage() {
  const { role: myRole, userId } = useAuth();
  const [page, setPage] = useState(1);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [inviteRole, setInviteRole] = useState('');
  const [inviteLink, setInviteLink] = useState<string | null>(null);

  const { data: usersData, isLoading } = useUserList(page);
  const { data: invitesData } = useInviteList();
  const updateRole = useUpdateUserRole();
  const deleteUser = useDeleteUser();
  const createInvite = useCreateInvite();
  const revokeInvite = useRevokeInvite();

  const users = usersData?.items ?? [];
  const total = usersData?.total ?? 0;
  const invites = invitesData?.items ?? [];
  const hasNext = page * 20 < total;
  const hasPrev = page > 1;
  const availableRoles = rolesBelow(myRole);

  const inviteForm = useForm<InviteFields>();

  function handleInvite(data: InviteFields) {
    if (!inviteRole) return;
    createInvite.mutate(
      { email: data.email, role: inviteRole },
      {
        onSuccess: (response) => {
          setInviteLink(`${window.location.origin}/invite/${response.token}`);
          inviteForm.reset();
          setInviteRole('');
        },
      },
    );
  }

  function copyLink() {
    if (inviteLink) {
      navigator.clipboard.writeText(inviteLink);
    }
  }

  const now = new Date();

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Users</h1>
        <Dialog
          open={inviteOpen}
          onOpenChange={(open: boolean) => {
            setInviteOpen(open);
            if (!open) setInviteLink(null);
          }}>
          <DialogTrigger render={<Button size="sm" />}>Invite User</DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{inviteLink ? 'Invite Created' : 'Invite User'}</DialogTitle>
            </DialogHeader>
            {inviteLink ? (
              <div className="space-y-3">
                <p className="text-muted-foreground text-sm">Share this link with the invitee. It expires in 7 days and can only be used once.</p>
                <div className="flex gap-2">
                  <Input readOnly value={inviteLink} className="font-mono text-xs" />
                  <Button size="sm" variant="outline" onClick={copyLink}>
                    Copy
                  </Button>
                </div>
              </div>
            ) : (
              <form onSubmit={inviteForm.handleSubmit(handleInvite)} className="space-y-4">
                <div className="space-y-1">
                  <Label htmlFor="invite-email">Email</Label>
                  <Input id="invite-email" type="email" {...inviteForm.register('email', { required: 'Email is required' })} />
                </div>
                <div className="space-y-1">
                  <Label>Role</Label>
                  <Select value={inviteRole} onValueChange={(value: string | null) => setInviteRole(value ?? '')}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent>
                      {availableRoles.map((role) => (
                        <SelectItem key={role} value={role}>
                          {role}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <Button type="submit" className="w-full" disabled={createInvite.isPending || !inviteRole}>
                  {createInvite.isPending ? 'Creating…' : 'Send Invite'}
                </Button>
              </form>
            )}
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.map((user) => {
              const isMe = user.id === userId;
              const canManage = !isMe && roleLevel(myRole) > roleLevel(user.role);
              return (
                <TableRow key={user.id} className={isMe ? 'bg-accent/30' : undefined}>
                  <TableCell>
                    {user.email}
                    {isMe && <span className="text-muted-foreground ml-2 text-xs">(you)</span>}
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{user.role}</Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    {canManage && (
                      <div className="flex justify-end gap-2">
                        <Select
                          onValueChange={(role: string | null) => {
                            if (role) updateRole.mutate({ id: user.id, role });
                          }}>
                          <SelectTrigger className="h-8 w-32 text-xs">
                            <SelectValue placeholder="Change role" />
                          </SelectTrigger>
                          <SelectContent>
                            {availableRoles.map((role) => (
                              <SelectItem key={role} value={role}>
                                {role}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => {
                            if (confirm(`Delete user ${user.email}?`)) {
                              deleteUser.mutate(user.id);
                            }
                          }}>
                          Delete
                        </Button>
                      </div>
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      )}

      <div className="flex items-center justify-between">
        <span className="text-muted-foreground text-sm">
          {total} user{total !== 1 ? 's' : ''}
        </span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => setPage((currentPage) => currentPage - 1)} disabled={!hasPrev}>
            Prev
          </Button>
          <Button variant="outline" size="sm" onClick={() => setPage((currentPage) => currentPage + 1)} disabled={!hasNext}>
            Next
          </Button>
        </div>
      </div>

      {invites.length > 0 && (
        <div className="space-y-3">
          <h2 className="text-lg font-semibold">Pending Invites</h2>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {invites.map((invite) => {
                const expired = new Date(invite.expiresAt) < now;
                return (
                  <TableRow key={invite.id} className={expired ? 'opacity-50' : undefined}>
                    <TableCell>{invite.email}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{invite.role}</Badge>
                    </TableCell>
                    <TableCell>{expired ? <Badge variant="destructive">Expired</Badge> : new Date(invite.expiresAt).toLocaleDateString()}</TableCell>
                    <TableCell className="text-right">
                      <Button variant="outline" size="sm" onClick={() => revokeInvite.mutate(invite.id)}>
                        Revoke
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
