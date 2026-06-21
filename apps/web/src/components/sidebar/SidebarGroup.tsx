import { useState, useEffect, type ReactNode } from 'react'
import { ChevronRight, type LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useSidebar } from './SidebarContext'

interface SidebarGroupProps {
  icon: LucideIcon
  label: string
  storageKey: string
  defaultOpen?: boolean
  children: ReactNode
}

export function SidebarGroup({ icon: Icon, label, storageKey, defaultOpen = true, children }: SidebarGroupProps) {
  const { collapsed } = useSidebar()
  const fullKey = `sidebar-group-${storageKey}`

  const [open, setOpen] = useState(() => {
    try {
      const stored = localStorage.getItem(fullKey)
      return stored !== null ? stored === 'true' : defaultOpen
    } catch {
      return defaultOpen
    }
  })

  useEffect(() => {
    try {
      localStorage.setItem(fullKey, String(open))
    } catch { /* noop */ }
  }, [open, fullKey])

  const handleToggle = () => setOpen((prev) => !prev)

  return (
    <div className="space-y-0.5">
      <button
        type="button"
        onClick={handleToggle}
        className={cn(
          'flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm font-semibold text-sidebar-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
          collapsed && 'justify-center px-0',
        )}
      >
        <Icon className="size-4 shrink-0" />
        <span className={cn('flex-1 truncate text-left', collapsed && 'sr-only')}>{label}</span>
        <ChevronRight
          className={cn(
            'size-3.5 shrink-0 transition-transform duration-200',
            open && 'rotate-90',
            collapsed && 'hidden',
          )}
        />
      </button>
      {open && !collapsed && <div className="space-y-0.5 pl-2">{children}</div>}
    </div>
  )
}
