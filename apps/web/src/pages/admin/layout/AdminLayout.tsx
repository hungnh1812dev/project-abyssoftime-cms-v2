import { Outlet } from 'react-router-dom';
import { SidebarShell, SidebarProvider } from '@/components/sidebar';
import { TopBar } from './TopBar';

export function AdminLayout() {
  return (
    <SidebarProvider>
      <div className="flex h-screen overflow-hidden">
        <SidebarShell />
        <div className="flex min-w-0 flex-1 flex-col">
          <TopBar />
          <main className="flex-1 overflow-y-auto">
            <Outlet />
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
}
