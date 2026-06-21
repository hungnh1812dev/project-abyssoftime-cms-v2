import { Menu } from 'lucide-react'
import { useBreadcrumbs } from '@/hooks/useBreadcrumbs'
import { useSidebar } from '@/components/sidebar'
import { Button } from '@/components/ui/button'
import { Breadcrumb } from '@/components/ui/breadcrumb'

export function TopBar() {
  const crumbs = useBreadcrumbs()
  const { isMobile, setMobileOpen } = useSidebar()

  return (
    <header className="h-14 border-b border-border flex items-center px-6 shrink-0">
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
    </header>
  )
}
