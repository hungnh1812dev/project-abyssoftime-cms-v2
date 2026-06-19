import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'

export function TopBar() {
  const { role, logout } = useAuth()

  return (
    <header className="h-14 border-b flex items-center justify-between px-6 shrink-0">
      <div />
      <div className="flex items-center gap-3">
        {role && <span className="text-sm text-muted-foreground capitalize">{role}</span>}
        <Button variant="outline" size="sm" onClick={logout}>
          Logout
        </Button>
      </div>
    </header>
  )
}
