import { cn } from '@/lib/utils';
import { useSidebar } from './SidebarContext';

interface SidebarSubGroupProps {
  label: string;
  children: React.ReactNode;
}

export function SidebarSubGroup({ label, children }: SidebarSubGroupProps) {
  const { collapsed } = useSidebar();

  return (
    <div className="space-y-0.5">
      <h3 className={cn('text-sidebar-muted px-3 pt-2 pb-1 text-xs font-semibold tracking-wide uppercase', collapsed && 'sr-only')}>{label}</h3>
      <div className={cn('space-y-0.5', !collapsed && 'pl-2')}>{children}</div>
    </div>
  );
}
