import { Link } from 'react-router-dom'
import { Hexagon } from 'lucide-react'
import { useSidebar } from './SidebarContext'

export function SidebarBrand() {
  const { collapsed } = useSidebar()

  return (
    <div className="border-b border-sidebar-border px-3 py-3">
      <Link to="/admin" className="flex items-center gap-2 text-sidebar-foreground">
        <Hexagon className="size-6 shrink-0 text-sidebar-primary" />
        <span
          className="font-semibold text-sm whitespace-nowrap overflow-hidden transition-[opacity,width] duration-200"
          style={{ opacity: collapsed ? 0 : 1, width: collapsed ? 0 : 'auto' }}
        >
          AbyssOfTime CMS
        </span>
      </Link>
    </div>
  )
}
