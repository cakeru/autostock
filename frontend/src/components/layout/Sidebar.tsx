import { SearchBar } from './SearchBar'
import { BrandMark } from './BrandMark'
import { useState, useMemo } from 'react'
import { Link, useLocation } from 'react-router-dom'
import {
  LayoutDashboard, Package, Users, Wrench, FileText,
  Settings, ChevronLeft, LogOut, UserCog, ShoppingCart, BarChart3, Truck, ClipboardList, ShoppingBag, UserRound, BellRing, QrCode
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/contexts/AuthContext'
import { useSettings } from '@/hooks/useSettings'

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()
  const { logout, hasPermission } = useAuth()
  const { data: settings } = useSettings()
  const scanOn = !!settings?.feature_batch_scan && hasPermission('install:scan')

  // Grouped by how often each area is used: the daily transaction loop first,
  // then reference data, then owner-only insights and admin.
  const navGroups = useMemo(() => [
    { heading: null, items: [
      { icon: LayoutDashboard, label: 'Dashboard', path: '/' },
    ] },
    { heading: 'Operations', items: [
      { icon: Wrench, label: 'Service Jobs', path: '/service-jobs' },
      { icon: FileText, label: 'Invoices', path: '/invoices' },
      { icon: BellRing, label: 'Due for Service', path: '/due-for-service' },
      ...(scanOn ? [{ icon: QrCode, label: 'Scan Install', path: '/scan' }] : []),
    ] },
    { heading: 'Records', items: [
      { icon: Package, label: 'Inventory', path: '/inventory' },
      { icon: ClipboardList, label: 'Stocktakes', path: '/stocktakes' },
      { icon: Users, label: 'Customers', path: '/customers' },
      { icon: Truck, label: 'Suppliers', path: '/suppliers' },
      { icon: ShoppingBag, label: 'Purchase Orders', path: '/purchase-orders' },
    ] },
    ...(hasPermission('report:view') ? [{ heading: 'Insights', items: [
      { icon: BarChart3, label: 'Analytics', path: '/analytics' },
    ] }] : []),
    { heading: 'Admin', items: [
      { icon: Settings, label: 'Settings', path: '/settings' },
      ...(hasPermission('user:view') ? [{ icon: UserCog, label: 'Users', path: '/users' }] : []),
      ...(hasPermission('user:view') ? [{ icon: UserRound, label: 'Employees', path: '/employees' }] : []),
    ] },
  ], [hasPermission, scanOn])

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/'
    return location.pathname.startsWith(path)
  }

  return (
    <aside
      className={cn(
        'fixed left-0 top-0 h-full z-30 flex flex-col text-white transition-all duration-200 print:hidden',
        collapsed ? 'w-[64px]' : 'w-[240px]'
      )}
      style={{ background: 'linear-gradient(180deg, var(--color-brand-navy) 0%, var(--color-brand-navy-deep) 100%)' }}
    >
      {/* Brand header */}
      <div className={cn('flex h-16 items-center gap-2.5 px-4', collapsed && 'justify-center px-0')}>
        <BrandMark className="h-9 w-9" />
        {!collapsed && (
          <div className="min-w-0 leading-tight">
            <p className="truncate text-sm font-bold tracking-wide">K&amp;S Wheel-Tyre</p>
            <p className="truncate text-[10px] uppercase tracking-[0.2em] text-white/45">Inventory</p>
          </div>
        )}
        <button
          onClick={() => setCollapsed(!collapsed)}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          className={cn(
            'grid h-7 w-7 place-items-center rounded-md text-white/60 transition-colors hover:bg-white/10 hover:text-white',
            collapsed ? 'absolute -right-3 top-5 h-6 w-6 bg-[var(--color-brand-navy)] shadow-md' : 'ml-auto'
          )}
        >
          <ChevronLeft className={cn('h-4 w-4 transition-transform', collapsed && 'rotate-180')} />
        </button>
      </div>

      <div className={cn('px-3 pb-2', collapsed && 'px-2')}>
        <Link
          to="/sale"
          title={collapsed ? 'New Sale' : undefined}
          className={cn(
            'flex items-center justify-center gap-2 rounded-lg bg-[var(--color-brand-gold)] font-semibold text-[var(--color-brand-navy-deep)] shadow-sm transition-opacity hover:opacity-90',
            collapsed ? 'h-10 w-full' : 'px-3 py-2.5 text-sm'
          )}
        >
          <ShoppingCart className="h-4 w-4 flex-shrink-0" />
          {!collapsed && <span>New Sale</span>}
        </Link>
      </div>

      {!collapsed && <div className="px-3 pb-2"><SearchBar dark /></div>}

      <nav className="flex flex-1 flex-col gap-0.5 overflow-y-auto p-3">
        {navGroups.map((group, gi) => (
          <div key={group.heading ?? `g${gi}`} className={cn(gi > 0 && 'mt-3')}>
            {group.heading && (
              collapsed
                ? <div className="mx-auto my-1 h-px w-6 bg-white/10" />
                : <p className="px-3 pb-1 text-[10px] font-semibold uppercase tracking-[0.15em] text-white/35">{group.heading}</p>
            )}
            <div className="flex flex-col gap-0.5">
              {group.items.map((item) => {
                const Icon = item.icon
                const active = isActive(item.path)
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    title={collapsed ? item.label : undefined}
                    className={cn(
                      'group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition-colors',
                      collapsed && 'justify-center px-0',
                      active
                        ? 'bg-white/10 font-medium text-white'
                        : 'text-white/65 hover:bg-white/5 hover:text-white'
                    )}
                  >
                    {active && (
                      <span className="absolute left-0 top-1/2 h-5 w-1 -translate-y-1/2 rounded-r-full bg-[var(--color-brand-gold)]" />
                    )}
                    <Icon className={cn('h-[18px] w-[18px] flex-shrink-0', active && 'text-[var(--color-brand-gold)]')} />
                    {!collapsed && <span>{item.label}</span>}
                  </Link>
                )
              })}
            </div>
          </div>
        ))}
      </nav>

      <div className="border-t border-white/10 p-3">
        <button
          onClick={logout}
          title={collapsed ? 'Logout' : undefined}
          className={cn(
            'flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm text-white/65 transition-colors hover:bg-white/5 hover:text-white',
            collapsed && 'justify-center px-0'
          )}
        >
          <LogOut className="h-[18px] w-[18px] flex-shrink-0" />
          {!collapsed && <span>Logout</span>}
        </button>
      </div>
    </aside>
  )
}
