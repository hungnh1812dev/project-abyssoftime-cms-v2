import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const ROLES = [
  { name: 'Super Admin', key: 'super_admin' },
  { name: 'Admin', key: 'admin' },
  { name: 'Editor', key: 'editor' },
  { name: 'Guest', key: 'guest' },
] as const

const PERMISSIONS = [
  'User Management',
  'Access Tokens',
  'Content-Type Mgmt',
  'Content Create/Edit',
  'Content Publish',
  'Content Delete',
  'View Content',
] as const

const MATRIX: Record<string, boolean[]> = {
  super_admin: [true, true, true, true, true, true, true],
  admin:       [true, false, true, true, true, true, true],
  editor:      [false, false, false, true, true, true, true],
  guest:       [false, false, false, false, false, false, true],
}

export function RolesPage() {
  return (
    <div className="space-y-6 p-6">
      <div>
        <h1 className="text-xl font-semibold">Roles & Permissions</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Permission matrix for each role. Roles are fixed and cannot be customized.
        </p>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-40">Role</TableHead>
            {PERMISSIONS.map((perm) => (
              <TableHead key={perm} className="text-center text-xs">
                {perm}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {ROLES.map((role) => (
            <TableRow key={role.key}>
              <TableCell>
                <Badge variant="secondary">{role.name}</Badge>
              </TableCell>
              {MATRIX[role.key].map((has, i) => (
                <TableCell key={PERMISSIONS[i]} className="text-center">
                  {has ? (
                    <span className="text-green-600 font-medium">✓</span>
                  ) : (
                    <span className="text-muted-foreground">—</span>
                  )}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
