import { NavLink } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useSidebar } from './SidebarContext'

interface SidebarItemProps {
  to: string
  children: React.ReactNode
}

export function SidebarItem({ to, children }: SidebarItemProps) {
  const { collapsed, setMobileOpen, isMobile } = useSidebar()

  return (
    <NavLink
      to={to}
      onClick={() => isMobile && setMobileOpen(false)}
      className={({ isActive }) =>
        cn(
          'flex items-center gap-2 rounded-md px-3 py-1.5 text-sm transition-colors',
          isActive
            ? 'bg-sidebar-accent text-sidebar-primary font-medium border-l-2 border-sidebar-primary'
            : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
          collapsed && 'justify-center px-0',
        )
      }
    >
      <span className={cn('truncate', collapsed && 'sr-only')}>{children}</span>
    </NavLink>
  )
}
