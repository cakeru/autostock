import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { MobileNav } from './MobileNav'

export function MainLayout() {
  return (
    <div className="min-h-screen bg-background">
      <div className="hidden md:block">
        <Sidebar />
      </div>
      <main className="md:ml-[240px] print:ml-0 print:p-0">
        <div className="max-w-6xl mx-auto p-4 md:p-6 pb-20 md:pb-8 print:p-0">
          <Outlet />
        </div>
      </main>
      <MobileNav />
    </div>
  )
}
