import { NavLink } from 'react-router-dom'
import { useContentTypes } from '@/hooks/useContentTypes'

export function Sidebar() {
  const { data: contentTypes } = useContentTypes()

  return (
    <aside className="w-64 border-r flex flex-col shrink-0">
      <div className="px-4 py-3 border-b">
        <span className="font-semibold text-sm">Abyssoftime CMS</span>
      </div>
      <nav className="flex-1 p-2 space-y-0.5">
        {(contentTypes ?? []).map((ct) => (
          <NavLink
            key={ct.ID}
            to={`/admin/content-types/${ct.Slug}`}
            className={({ isActive }) =>
              `block px-3 py-2 rounded-md text-sm transition-colors ${
                isActive
                  ? 'bg-accent text-accent-foreground font-medium'
                  : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
              }`
            }
          >
            {ct.Name}
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}
