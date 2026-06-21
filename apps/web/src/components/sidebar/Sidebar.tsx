import { FileText, LogOut, Settings } from 'lucide-react'
import { useContentTypes } from '@/hooks/useContentTypes'
import { useAuth } from '@/hooks/useAuth'
import { roleLevel } from '@/lib/roles'
import { cn } from '@/lib/utils'
import { useSidebar } from './SidebarContext'
import { SidebarBrand } from './SidebarBrand'
import { SidebarGroup } from './SidebarGroup'
import { SidebarSubGroup } from './SidebarSubGroup'
import { SidebarItem } from './SidebarItem'
import { SidebarCollapseToggle } from './SidebarCollapseToggle'

export function Sidebar() {
  const { collapsed } = useSidebar()
  const { data: contentTypes } = useContentTypes()
  const { role, logout } = useAuth()

  const singleTypes = (contentTypes ?? []).filter((ct) => ct.Kind === 'single')
  const collectionTypes = (contentTypes ?? []).filter((ct) => ct.Kind === 'collection')

  const isSuperAdmin = role === 'super_admin'
  const isAdminOrAbove = roleLevel(role) >= roleLevel('admin')

  return (
    <aside
      role="complementary"
      className={cn(
        'flex flex-col shrink-0 bg-sidebar border-r border-sidebar-border transition-[width] duration-200 ease-in-out overflow-hidden',
        collapsed ? 'w-16' : 'w-64',
      )}
    >
      <SidebarBrand />

      <nav className="flex-1 overflow-y-auto p-2 space-y-1">
        <SidebarGroup icon={FileText} label="Content Manager" storageKey="content-manager" defaultOpen>
          {singleTypes.length > 0 && (
            <SidebarSubGroup label="Single Types">
              {singleTypes.map((ct) => (
                <SidebarItem key={ct.ID} to={`/admin/content-type/single-type/${ct.Slug}`}>
                  {ct.Name}
                </SidebarItem>
              ))}
            </SidebarSubGroup>
          )}
          {collectionTypes.length > 0 && (
            <SidebarSubGroup label="Collection Types">
              {collectionTypes.map((ct) => (
                <SidebarItem key={ct.ID} to={`/admin/content-type/collection-type/${ct.Slug}`}>
                  {ct.Name}
                </SidebarItem>
              ))}
            </SidebarSubGroup>
          )}
        </SidebarGroup>

        <SidebarGroup icon={Settings} label="Settings" storageKey="settings" defaultOpen>
          <SidebarItem to="/admin/settings/media">Media Library</SidebarItem>
          {isAdminOrAbove && (
            <SidebarItem to="/admin/settings/users">Users</SidebarItem>
          )}
          {isSuperAdmin && (
            <SidebarItem to="/admin/settings/access-tokens">Access Tokens</SidebarItem>
          )}
          {isSuperAdmin && (
            <SidebarItem to="/admin/settings/roles">Roles</SidebarItem>
          )}
        </SidebarGroup>
      </nav>

      <div className="border-t border-sidebar-border p-2 space-y-1">
        {role && !collapsed && (
          <span className="block px-3 py-1 text-xs text-sidebar-muted capitalize">{role}</span>
        )}
        <button
          type="button"
          onClick={logout}
          className={cn(
            'flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm text-sidebar-muted transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
            collapsed && 'justify-center px-0',
          )}
        >
          <LogOut className="size-4 shrink-0" />
          <span className={cn('truncate', collapsed && 'sr-only')}>Logout</span>
        </button>
        <SidebarCollapseToggle />
      </div>
    </aside>
  )
}
