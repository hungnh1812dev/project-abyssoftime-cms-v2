import { ChevronsLeft, ChevronsRight } from 'lucide-react';
import { useSidebar } from './SidebarContext';
import { cn } from '@/lib/utils';

export function SidebarCollapseToggle() {
  const { collapsed, toggle, isMobile } = useSidebar();

  if (isMobile) return null;

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
      className={cn(
        'text-sidebar-muted hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors',
        collapsed && 'justify-center px-0',
      )}>
      {collapsed ? <ChevronsRight className="size-4" /> : <ChevronsLeft className="size-4" />}
      <span className={cn('truncate', collapsed && 'sr-only')}>{collapsed ? 'Expand' : 'Collapse'}</span>
    </button>
  );
}
