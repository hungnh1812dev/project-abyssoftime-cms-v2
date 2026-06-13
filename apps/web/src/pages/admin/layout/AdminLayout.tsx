import { Outlet } from 'react-router-dom'

export function AdminLayout() {
  return (
    <div className="flex min-h-screen">
      <aside className="w-64 border-r p-4">
        <p className="text-muted-foreground text-sm">Sidebar — coming in T3.8</p>
      </aside>
      <main className="flex-1 p-6">
        <Outlet />
      </main>
    </div>
  )
}
