import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useAuth } from '@/context/AuthContext'
import { roleLevel } from '@/lib/roles'
import { useUserList, useUpdateUserRole, useDeleteUser } from '@/hooks/useUsers'
import { useInviteList, useCreateInvite, useRevokeInvite } from '@/hooks/useInvites'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const ALL_ROLES = ['super_admin', 'admin', 'editor', 'guest'] as const

function rolesBelow(currentRole: string | null): string[] {
  const level = roleLevel(currentRole)
  return ALL_ROLES.filter((r) => roleLevel(r) < level)
}

interface InviteFields {
  email: string
}

export function UsersPage() {
  const { role: myRole, userId } = useAuth()
  const [page, setPage] = useState(1)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [inviteRole, setInviteRole] = useState('')
  const [inviteLink, setInviteLink] = useState<string | null>(null)

  const { data: usersData, isLoading } = useUserList(page)
  const { data: invitesData } = useInviteList()
  const updateRole = useUpdateUserRole()
  const deleteUser = useDeleteUser()
  const createInvite = useCreateInvite()
  const revokeInvite = useRevokeInvite()

  const users = usersData?.items ?? []
  const total = usersData?.total ?? 0
  const invites = invitesData?.items ?? []
  const hasNext = page * 20 < total
  const hasPrev = page > 1
  const availableRoles = rolesBelow(myRole)

  const inviteForm = useForm<InviteFields>()

  function handleInvite(data: InviteFields) {
    if (!inviteRole) return
    createInvite.mutate(
      { email: data.email, role: inviteRole },
      {
        onSuccess: (res) => {
          setInviteLink(`${window.location.origin}/invite/${res.token}`)
          inviteForm.reset()
          setInviteRole('')
        },
      },
    )
  }

  function copyLink() {
    if (inviteLink) {
      navigator.clipboard.writeText(inviteLink)
    }
  }

  const now = new Date()

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Users</h1>
        <Dialog open={inviteOpen} onOpenChange={(open: boolean) => { setInviteOpen(open); if (!open) setInviteLink(null) }}>
          <DialogTrigger render={<Button size="sm" />}>
            Invite User
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{inviteLink ? 'Invite Created' : 'Invite User'}</DialogTitle>
            </DialogHeader>
            {inviteLink ? (
              <div className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Share this link with the invitee. It expires in 7 days and can only be used once.
                </p>
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
                  <Input
                    id="invite-email"
                    type="email"
                    {...inviteForm.register('email', { required: 'Email is required' })}
                  />
                </div>
                <div className="space-y-1">
                  <Label>Role</Label>
                  <Select value={inviteRole} onValueChange={(v: string | null) => setInviteRole(v ?? '')}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent>
                      {availableRoles.map((r) => (
                        <SelectItem key={r} value={r}>
                          {r}
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
            {users.map((u) => {
              const isMe = u.id === userId
              const canManage = !isMe && roleLevel(myRole) > roleLevel(u.role)
              return (
                <TableRow key={u.id} className={isMe ? 'bg-accent/30' : undefined}>
                  <TableCell>{u.email}{isMe && <span className="text-xs text-muted-foreground ml-2">(you)</span>}</TableCell>
                  <TableCell><Badge variant="secondary">{u.role}</Badge></TableCell>
                  <TableCell className="text-right">
                    {canManage && (
                      <div className="flex justify-end gap-2">
                        <Select
                          onValueChange={(role: string | null) => { if (role) updateRole.mutate({ id: u.id, role }) }}
                        >
                          <SelectTrigger className="w-32 h-8 text-xs">
                            <SelectValue placeholder="Change role" />
                          </SelectTrigger>
                          <SelectContent>
                            {availableRoles.map((r) => (
                              <SelectItem key={r} value={r}>
                                {r}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => {
                            if (confirm(`Delete user ${u.email}?`)) {
                              deleteUser.mutate(u.id)
                            }
                          }}
                        >
                          Delete
                        </Button>
                      </div>
                    )}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      )}

      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">{total} user{total !== 1 ? 's' : ''}</span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => setPage((p) => p - 1)} disabled={!hasPrev}>
            Prev
          </Button>
          <Button variant="outline" size="sm" onClick={() => setPage((p) => p + 1)} disabled={!hasNext}>
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
              {invites.map((inv) => {
                const expired = new Date(inv.expiresAt) < now
                return (
                  <TableRow key={inv.id} className={expired ? 'opacity-50' : undefined}>
                    <TableCell>{inv.email}</TableCell>
                    <TableCell><Badge variant="secondary">{inv.role}</Badge></TableCell>
                    <TableCell>
                      {expired ? (
                        <Badge variant="destructive">Expired</Badge>
                      ) : (
                        new Date(inv.expiresAt).toLocaleDateString()
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => revokeInvite.mutate(inv.id)}
                      >
                        Revoke
                      </Button>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
