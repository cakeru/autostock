import { NavLink, Outlet } from 'react-router-dom'
import { cn } from '@/lib/utils'

const TABS = [
  { to: '/analytics/sales', label: 'Sales' },
  { to: '/analytics/inventory', label: 'Inventory' },
  { to: '/analytics/customers', label: 'Customers' },
  { to: '/analytics/receivables', label: 'Receivables' },
  { to: '/analytics/pnl', label: 'Profit & Loss' },
  { to: '/analytics/technicians', label: 'Technicians' },
  { to: '/analytics/audit', label: 'Audit log' },
]

export function AnalyticsLayout() {
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center gap-1 border-b">
        {TABS.map((t) => (
          <NavLink
            key={t.to}
            to={t.to}
            className={({ isActive }) => cn(
              'relative px-3 py-2 text-sm font-medium transition-colors',
              isActive ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {({ isActive }) => (
              <>
                {t.label}
                {isActive && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-[var(--color-brand-gold)]" />}
              </>
            )}
          </NavLink>
        ))}
      </div>
      <Outlet />
    </div>
  )
}
