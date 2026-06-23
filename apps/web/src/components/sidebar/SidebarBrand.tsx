import { Link } from 'react-router-dom';
import { Hexagon } from 'lucide-react';
import { useSidebar } from './SidebarContext';

export function SidebarBrand() {
  const { collapsed } = useSidebar();

  return (
    <div className="border-sidebar-border border-b px-3 py-3">
      <Link to="/admin" className="text-sidebar-foreground flex items-center gap-2">
        <Hexagon className="text-sidebar-primary size-6 shrink-0" />
        <span
          className="overflow-hidden text-sm font-semibold whitespace-nowrap transition-[opacity,width] duration-200"
          style={{ opacity: collapsed ? 0 : 1, width: collapsed ? 0 : 'auto' }}>
          AbyssOfTime CMS
        </span>
      </Link>
    </div>
  );
}
