import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import {
  Package, ShoppingCart, Wrench, FileText, MoreHorizontal,
  LayoutDashboard, Users, BarChart3, Settings, UserCog, X, ClipboardList, ShoppingBag, UserRound, BellRing,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/contexts/AuthContext'

// The 4 highest-frequency floor tasks live in the bar; everything else is one
// tap away under "More" (progressive disclosure — see nav best practices).
const coreItems = [
  { icon: ShoppingCart, label: 'Sell', path: '/sale' },
  { icon: Package, label: 'Inventory', path: '/inventory' },
  { icon: Wrench, label: 'Jobs', path: '/service-jobs' },
  { icon: FileText, label: 'Invoices', path: '/invoices' },
]

export function MobileNav() {
  const location = useLocation()
  const { hasPermission } = useAuth()
  const [moreOpen, setMoreOpen] = useState(false)

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/'
    return location.pathname.startsWith(path)
  }

  const moreItems = [
    { icon: LayoutDashboard, label: 'Dashboard', path: '/' },
    { icon: Users, label: 'Customers', path: '/customers' },
    { icon: BellRing, label: 'Due for Service', path: '/due-for-service' },
    { icon: ClipboardList, label: 'Stocktakes', path: '/stocktakes' },
    { icon: ShoppingBag, label: 'Purchase Orders', path: '/purchase-orders' },
    ...(hasPermission('report:view') ? [{ icon: BarChart3, label: 'Analytics', path: '/analytics' }] : []),
    { icon: Settings, label: 'Settings', path: '/settings' },
    ...(hasPermission('user:view') ? [{ icon: UserCog, label: 'Users', path: '/users' }] : []),
    ...(hasPermission('user:view') ? [{ icon: UserRound, label: 'Employees', path: '/employees' }] : []),
  ]
  const moreActive = moreItems.some((i) => isActive(i.path))

  return (
    <>
      {moreOpen && (
        <div className="fixed inset-0 z-40 bg-black/40 md:hidden" onClick={() => setMoreOpen(false)}>
          <div
            className="absolute inset-x-0 bottom-0 rounded-t-2xl bg-card p-2 pb-[calc(env(safe-area-inset-bottom,0px)+0.5rem)] shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between px-3 py-2">
              <p className="text-sm font-semibold">More</p>
              <button onClick={() => setMoreOpen(false)} aria-label="Close" className="text-muted-foreground hover:text-foreground">
                <X className="h-5 w-5" />
              </button>
            </div>
            <div className="grid grid-cols-3 gap-2 p-1">
              {moreItems.map((item) => {
                const Icon = item.icon
                const active = isActive(item.path)
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    onClick={() => setMoreOpen(false)}
                    className={cn(
                      'flex flex-col items-center gap-1.5 rounded-xl border p-4 text-xs font-medium',
                      active ? 'border-primary/40 bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted/50'
                    )}
                  >
                    <Icon className="h-6 w-6" />
                    <span>{item.label}</span>
                  </Link>
                )
              })}
            </div>
          </div>
        </div>
      )}

      <nav className="fixed bottom-0 left-0 right-0 z-30 border-t bg-card md:hidden print:hidden" style={{ paddingBottom: 'env(safe-area-inset-bottom, 0px)' }}>
        <div className="flex h-14 items-center justify-around px-2">
          {coreItems.map((item) => {
            const Icon = item.icon
            const active = isActive(item.path)
            return (
              <Link
                key={item.path}
                to={item.path}
                className={cn('flex flex-col items-center gap-0.5 px-2 py-1 text-xs min-w-0', active ? 'text-primary' : 'text-muted-foreground')}
              >
                <Icon className="h-5 w-5" />
                <span className="truncate">{item.label}</span>
              </Link>
            )
          })}
          <button
            onClick={() => setMoreOpen(true)}
            className={cn('flex flex-col items-center gap-0.5 px-2 py-1 text-xs min-w-0', moreActive || moreOpen ? 'text-primary' : 'text-muted-foreground')}
          >
            <MoreHorizontal className="h-5 w-5" />
            <span className="truncate">More</span>
          </button>
        </div>
      </nav>
    </>
  )
}
