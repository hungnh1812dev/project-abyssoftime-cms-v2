import { Outlet } from 'react-router-dom'
import { SidebarShell, SidebarProvider } from '@/components/sidebar'
import { TopBar } from './TopBar'

export function AdminLayout() {
  return (
    <SidebarProvider>
      <div className="flex min-h-screen">
        <SidebarShell />
        <div className="flex flex-col flex-1 min-w-0">
          <TopBar />
          <main className="flex-1 p-6">
            <Outlet />
          </main>
        </div>
      </div>
    </SidebarProvider>
  )
}
