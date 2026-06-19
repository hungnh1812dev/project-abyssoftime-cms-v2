import { NavLink } from 'react-router-dom'
import { useContentTypes } from '@/hooks/useContentTypes'
import { useAuth } from '@/context/AuthContext'
import { roleLevel } from '@/lib/roles'
import type { ContentTypeSummary } from '@/types/cms'

function navLinkClass({ isActive }: { isActive: boolean }) {
  return `block px-3 py-2 rounded-md text-sm transition-colors ${
    isActive
      ? 'bg-accent text-accent-foreground font-medium'
      : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
  }`
}

function NavGroup({ title, items }: { title: string; items: ContentTypeSummary[] }) {
  if (items.length === 0) {
    return null
  }
  return (
    <div className="space-y-0.5">
      <h2 className="px-3 pt-3 pb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        {title}
      </h2>
      {items.map((ct) => (
        <NavLink
          key={ct.ID}
          to={`/admin/content-type/${ct.Kind === 'single' ? 'single-type' : 'collection-type'}/${ct.Slug}`}
          className={navLinkClass}
        >
          {ct.Name}
        </NavLink>
      ))}
    </div>
  )
}

export function Sidebar() {
  const { data: contentTypes } = useContentTypes()
  const { role } = useAuth()
  const singleTypes = (contentTypes ?? []).filter((ct) => ct.Kind === 'single')
  const collectionTypes = (contentTypes ?? []).filter((ct) => ct.Kind === 'collection')

  const isSuperAdmin = role === 'super_admin'
  const isAdminOrAbove = roleLevel(role) >= roleLevel('admin')

  return (
    <aside className="w-64 border-r flex flex-col shrink-0">
      <div className="px-4 py-3 border-b">
        <span className="font-semibold text-sm">Abyssoftime CMS</span>
      </div>
      <nav className="flex-1 p-2 space-y-1">
        <NavGroup title="Single Types" items={singleTypes} />
        <NavGroup title="Collection Types" items={collectionTypes} />
        <div className="space-y-0.5">
          <h2 className="px-3 pt-3 pb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Settings
          </h2>
          <NavLink to="/admin/settings/media" className={navLinkClass}>
            Media Library
          </NavLink>
          {isAdminOrAbove && (
            <NavLink to="/admin/settings/users" className={navLinkClass}>
              Users
            </NavLink>
          )}
          {isSuperAdmin && (
            <NavLink to="/admin/settings/access-tokens" className={navLinkClass}>
              Access Tokens
            </NavLink>
          )}
          {isSuperAdmin && (
            <NavLink to="/admin/settings/roles" className={navLinkClass}>
              Roles
            </NavLink>
          )}
        </div>
      </nav>
    </aside>
  )
}
