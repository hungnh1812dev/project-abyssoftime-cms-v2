import { FileText, LogOut, Settings } from 'lucide-react';
import { useContentTypes } from '@/hooks/useContentTypes';
import { useAuth } from '@/hooks/useAuth';
import { roleLevel } from '@/lib/roles';
import { cn } from '@/lib/utils';
import { useSidebar } from './SidebarContext';
import { SidebarBrand } from './SidebarBrand';
import { SidebarGroup } from './SidebarGroup';
import { SidebarSubGroup } from './SidebarSubGroup';
import { SidebarItem } from './SidebarItem';
import { SidebarCollapseToggle } from './SidebarCollapseToggle';

export function Sidebar() {
  const { collapsed } = useSidebar();
  const { data: contentTypes } = useContentTypes();
  const { role, logout } = useAuth();

  const singleTypes = (contentTypes ?? []).filter((contentType) => contentType.Kind === 'single');
  const collectionTypes = (contentTypes ?? []).filter((contentType) => contentType.Kind === 'collection');

  const isSuperAdmin = role === 'super_admin';
  const isAdminOrAbove = roleLevel(role) >= roleLevel('admin');

  return (
    <aside
      role="complementary"
      className={cn('bg-sidebar border-sidebar-border flex shrink-0 flex-col overflow-hidden border-r transition-[width] duration-200 ease-in-out', collapsed ? 'w-16' : 'w-64')}>
      <SidebarBrand />

      <nav className="flex-1 space-y-1 overflow-y-auto p-2">
        <SidebarGroup icon={FileText} label="Content Manager" storageKey="content-manager" defaultOpen>
          {singleTypes.length > 0 && (
            <SidebarSubGroup label="Single Types">
              {singleTypes.map((contentType) => (
                <SidebarItem key={contentType.ID} to={`/admin/content-type/single-type/${contentType.Slug}`}>
                  {contentType.Name}
                </SidebarItem>
              ))}
            </SidebarSubGroup>
          )}
          {collectionTypes.length > 0 && (
            <SidebarSubGroup label="Collection Types">
              {collectionTypes.map((contentType) => (
                <SidebarItem key={contentType.ID} to={`/admin/content-type/collection-type/${contentType.Slug}`}>
                  {contentType.Name}
                </SidebarItem>
              ))}
            </SidebarSubGroup>
          )}
        </SidebarGroup>

        <SidebarGroup icon={Settings} label="Settings" storageKey="settings" defaultOpen>
          <SidebarItem to="/admin/settings/media">Media Library</SidebarItem>
          {isAdminOrAbove && <SidebarItem to="/admin/settings/users">Users</SidebarItem>}
          {isSuperAdmin && <SidebarItem to="/admin/settings/access-tokens">Access Tokens</SidebarItem>}
          {isSuperAdmin && <SidebarItem to="/admin/settings/roles">Roles</SidebarItem>}
          {isSuperAdmin && <SidebarItem to="/admin/settings/internationalize">Internationalize</SidebarItem>}
        </SidebarGroup>
      </nav>

      <div className="border-sidebar-border space-y-1 border-t p-2">
        {role && !collapsed && <span className="text-sidebar-muted block px-3 py-1 text-xs capitalize">{role}</span>}
        <button
          type="button"
          onClick={logout}
          className={cn(
            'text-sidebar-muted hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors',
            collapsed && 'justify-center px-0',
          )}>
          <LogOut className="size-4 shrink-0" />
          <span className={cn('truncate', collapsed && 'sr-only')}>Logout</span>
        </button>
        <SidebarCollapseToggle />
      </div>
    </aside>
  );
}
