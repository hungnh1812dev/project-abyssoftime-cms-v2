import { useState } from 'react'
import { useAccessTokenList, useCreateAccessToken, useDeleteAccessToken } from '@/hooks/useAccessTokens'
import { useContentTypes } from '@/hooks/useContentTypes'
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

const EXPIRY_OPTIONS = [
  { label: '7 days', value: '168h' },
  { label: '30 days', value: '720h' },
  { label: '90 days', value: '2160h' },
  { label: '1 year', value: '8760h' },
  { label: 'No expiration', value: '' },
] as const

function formatScope(scope: string): string {
  if (scope === 'documents:read') return 'All Documents'
  if (scope.startsWith('documents:read:')) return scope.replace('documents:read:', '')
  if (scope === 'media:read') return 'Media'
  if (scope === 'content-types:read') return 'Content Types'
  return scope
}

export function AccessTokensPage() {
  const [page, setPage] = useState(1)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [tokenName, setTokenName] = useState('')
  const [selectedScopes, setSelectedScopes] = useState<string[]>([])
  const [expiresIn, setExpiresIn] = useState('')
  const [createdToken, setCreatedToken] = useState<string | null>(null)

  const { data, isLoading } = useAccessTokenList(page)
  const createToken = useCreateAccessToken()
  const deleteToken = useDeleteAccessToken()
  const { data: contentTypes } = useContentTypes()

  const tokens = data?.items ?? []
  const total = data?.total ?? 0
  const hasNext = page * 20 < total
  const hasPrev = page > 1
  const ctList = contentTypes ?? []

  const hasDocumentsAll = selectedScopes.includes('documents:read')

  function toggleScope(scope: string) {
    setSelectedScopes((previousScopes) => {
      if (scope === 'documents:read') {
        if (previousScopes.includes('documents:read')) {
          return previousScopes.filter((item) => item !== 'documents:read')
        }
        return [...previousScopes.filter((item) => !item.startsWith('documents:read')), 'documents:read']
      }

      if (scope.startsWith('documents:read:')) {
        const without = previousScopes.filter((item) => item !== scope && item !== 'documents:read')
        if (previousScopes.includes(scope)) {
          return without
        }
        return [...without, scope]
      }

      return previousScopes.includes(scope) ? previousScopes.filter((item) => item !== scope) : [...previousScopes, scope]
    })
  }

  function handleCreate() {
    if (!tokenName || selectedScopes.length === 0) return
    createToken.mutate(
      { name: tokenName, scopes: selectedScopes, expiresIn: expiresIn || undefined },
      {
        onSuccess: (response) => {
          setCreatedToken(response.token)
          setTokenName('')
          setSelectedScopes([])
          setExpiresIn('')
        },
      },
    )
  }

  function copyToken() {
    if (createdToken) {
      navigator.clipboard.writeText(createdToken)
    }
  }

  function closeDialog() {
    setDialogOpen(false)
    setCreatedToken(null)
  }

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Access Tokens</h1>
        <Dialog open={dialogOpen} onOpenChange={(open: boolean) => { if (!open) closeDialog(); else setDialogOpen(true) }}>
          <DialogTrigger render={<Button size="sm" />}>
            Create new token
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{createdToken ? 'Token Created' : 'Create Access Token'}</DialogTitle>
            </DialogHeader>
            {createdToken ? (
              <div className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Copy this token now. It will not be shown again.
                </p>
                <div className="rounded-md border bg-muted p-3">
                  <code className="text-xs break-all">{createdToken}</code>
                </div>
                <Button size="sm" onClick={copyToken} className="w-full">
                  Copy Token
                </Button>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="space-y-1">
                  <Label htmlFor="token-name">Name</Label>
                  <Input
                    id="token-name"
                    value={tokenName}
                    onChange={(event) => setTokenName(event.target.value)}
                    placeholder="e.g. Frontend production"
                  />
                </div>
                <div className="space-y-3">
                  <Label>Scopes</Label>

                  <div className="space-y-2">
                    <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Documents</div>
                    <div className="rounded-md border p-3 space-y-2">
                      <label className="flex items-center gap-2 text-sm cursor-pointer font-medium">
                        <input
                          type="checkbox"
                          checked={hasDocumentsAll}
                          onChange={() => toggleScope('documents:read')}
                          className="rounded"
                        />
                        All content types
                      </label>
                      {!hasDocumentsAll && ctList.length > 0 && (
                        <div className="ml-5 space-y-1 border-l pl-3">
                          {ctList.map((ct) => {
                            const scope = `documents:read:${ct.Slug}`
                            return (
                              <label key={ct.ID} className="flex items-center gap-2 text-sm cursor-pointer">
                                <input
                                  type="checkbox"
                                  checked={selectedScopes.includes(scope)}
                                  onChange={() => toggleScope(scope)}
                                  className="rounded"
                                />
                                <span>{ct.Name}</span>
                                <span className="text-xs text-muted-foreground">({ct.Kind})</span>
                              </label>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="space-y-1">
                    <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Other</div>
                    <div className="rounded-md border p-3 space-y-2">
                      <label className="flex items-center gap-2 text-sm cursor-pointer">
                        <input
                          type="checkbox"
                          checked={selectedScopes.includes('media:read')}
                          onChange={() => toggleScope('media:read')}
                          className="rounded"
                        />
                        Media assets
                      </label>
                      <label className="flex items-center gap-2 text-sm cursor-pointer">
                        <input
                          type="checkbox"
                          checked={selectedScopes.includes('content-types:read')}
                          onChange={() => toggleScope('content-types:read')}
                          className="rounded"
                        />
                        Content type definitions
                      </label>
                    </div>
                  </div>
                </div>
                <div className="space-y-1">
                  <Label>Expiration</Label>
                  <Select value={expiresIn} onValueChange={(value: string | null) => setExpiresIn(value ?? '')}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select expiration" />
                    </SelectTrigger>
                    <SelectContent>
                      {EXPIRY_OPTIONS.map((option) => (
                        <SelectItem key={option.value || 'none'} value={option.value || 'none'}>
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <Button
                  className="w-full"
                  onClick={handleCreate}
                  disabled={createToken.isPending || !tokenName || selectedScopes.length === 0}
                >
                  {createToken.isPending ? 'Creating…' : 'Create Token'}
                </Button>
              </div>
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
              <TableHead>Name</TableHead>
              <TableHead>Token</TableHead>
              <TableHead>Scopes</TableHead>
              <TableHead>Expires</TableHead>
              <TableHead>Last Used</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tokens.map((token) => (
              <TableRow key={token.id}>
                <TableCell className="font-medium">{token.name}</TableCell>
                <TableCell>
                  <code className="text-xs text-muted-foreground">{token.prefix}••••••</code>
                </TableCell>
                <TableCell>
                  <div className="flex gap-1 flex-wrap">
                    {token.scopes.map((scope) => (
                      <Badge key={scope} variant="secondary" className="text-xs">
                        {formatScope(scope)}
                      </Badge>
                    ))}
                  </div>
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {token.expiresAt ? new Date(token.expiresAt).toLocaleDateString() : 'Never'}
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {token.lastUsedAt ? new Date(token.lastUsedAt).toLocaleDateString() : 'Never'}
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => {
                      if (confirm(`Delete token "${token.name}"?`)) {
                        deleteToken.mutate(token.id)
                      }
                    }}
                  >
                    Delete
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">{total} token{total !== 1 ? 's' : ''}</span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => setPage((currentPage) => currentPage - 1)} disabled={!hasPrev}>
            Prev
          </Button>
          <Button variant="outline" size="sm" onClick={() => setPage((currentPage) => currentPage + 1)} disabled={!hasNext}>
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
