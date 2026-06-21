import { Menu } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { useBreadcrumbs } from '@/hooks/useBreadcrumbs'
import { useSidebar } from '@/components/sidebar'
import { Button } from '@/components/ui/button'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { Badge } from '@/components/ui/badge'

export function TopBar() {
  const { role, logout } = useAuth()
  const crumbs = useBreadcrumbs()
  const { isMobile, setMobileOpen } = useSidebar()

  return (
    <header className="h-14 border-b border-border flex items-center justify-between px-6 shrink-0">
      <div className="flex items-center gap-3">
        {isMobile && (
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Open menu"
            onClick={() => setMobileOpen(true)}
          >
            <Menu />
          </Button>
        )}
        <Breadcrumb items={crumbs} />
      </div>
      <div className="flex items-center gap-3">
        {role && <Badge variant="secondary" className="capitalize">{role}</Badge>}
        <Button variant="outline" size="sm" onClick={logout}>
          Logout
        </Button>
      </div>
    </header>
  )
}
