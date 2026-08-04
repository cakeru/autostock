import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Droplet, Disc, Wrench, Phone, List, CalendarDays } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { TableCard, Th } from '@/components/ui/table'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { Select } from '@/components/ui/select'
import { useDueForService } from '@/hooks/useVehicleProfile'
import { DueCalendar } from '@/components/vehicle/DueCalendar'
import { formatDate } from '@/utils/date'
import type { DueForServiceItem } from '@/types/vehicleProfile'

const STATUS_STYLE: Record<string, string> = {
  overdue: 'bg-destructive/10 text-destructive',
  due_soon: 'bg-amber-500/10 text-amber-600',
}
const STATUS_LABEL: Record<string, string> = {
  overdue: 'Overdue',
  due_soon: 'Due soon',
}
const STATUS_ORDER: Record<string, number> = { overdue: 0, due_soon: 1 }

export function DueForService() {
  const navigate = useNavigate()
  const [view, setView] = useState<'list' | 'calendar'>('list')
  // The calendar needs upcoming (on-track) dues too, so it asks for a forward
  // horizon; the list stays the overdue + due-soon call sheet.
  const { data, isLoading } = useDueForService(view === 'calendar' ? 120 : undefined)
  const [typeFilter, setTypeFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')

  const items = useMemo(() => {
    const list = (data || []) as DueForServiceItem[]
    return list
      .filter((i) => !typeFilter || i.event_type === typeFilter)
      .filter((i) => !statusFilter || i.status === statusFilter)
      .sort((a, b) => {
        const s = (STATUS_ORDER[a.status] ?? 2) - (STATUS_ORDER[b.status] ?? 2)
        if (s !== 0) return s
        return (a.due_date || '').localeCompare(b.due_date || '')
      })
  }, [data, typeFilter, statusFilter])

  const overdueCount = (data || []).filter((i) => i.status === 'overdue').length
  const dueSoonCount = (data || []).filter((i) => i.status === 'due_soon').length

  return (
    <div className="space-y-6">
      <PageHeader
        title="Due for Service"
        subtitle="Customers whose tires or oil are overdue or coming up, based on their last known mileage and visit history."
        badges={
          <>
            {overdueCount > 0 && (
              <span className="rounded-full bg-destructive/10 px-2.5 py-0.5 text-xs font-medium text-destructive">
                {overdueCount} overdue
              </span>
            )}
            {dueSoonCount > 0 && (
              <span className="rounded-full bg-amber-500/10 px-2.5 py-0.5 text-xs font-medium text-amber-600">
                {dueSoonCount} due soon
              </span>
            )}
          </>
        }
        actions={
          <div className="inline-flex rounded-md border p-0.5">
            <button
              onClick={() => setView('list')}
              className={`inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs font-medium ${view === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
            >
              <List className="h-3.5 w-3.5" /> List
            </button>
            <button
              onClick={() => setView('calendar')}
              className={`inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs font-medium ${view === 'calendar' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
            >
              <CalendarDays className="h-3.5 w-3.5" /> Calendar
            </button>
          </div>
        }
      />

      {isLoading ? (
        <div className="bg-card rounded-lg shadow-sm p-4">
          <TableSkeleton rows={6} cols={6} />
        </div>
      ) : view === 'calendar' ? (
        <DueCalendar items={(data || []) as DueForServiceItem[]} />
      ) : (
        <>
          <div className="hidden md:block">
            <TableCard
              toolbar={
                <>
                  <Select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)} className="w-36">
                    <option value="">All types</option>
                    <option value="oil">Oil</option>
                    <option value="tire">Tires</option>
                    <option value="part">Parts</option>
                  </Select>
                  <Select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="w-36">
                    <option value="">All statuses</option>
                    <option value="overdue">Overdue</option>
                    <option value="due_soon">Due soon</option>
                  </Select>
                </>
              }
            >
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-card sticky top-0 z-10">
                    <Th>Status</Th>
                    <Th>Customer</Th>
                    <Th>Vehicle</Th>
                    <Th>Type</Th>
                    <Th>Last service</Th>
                    <Th>Due</Th>
                  </tr>
                </thead>
                <tbody>
                  {items.length === 0 ? (
                    <tr><td colSpan={6} className="px-4 py-10 text-center text-muted-foreground">Nobody's due right now.</td></tr>
                  ) : (
                    items.map((i) => (
                      <tr
                        key={`${i.vehicle_id}-${i.key}`}
                        className="group border-b last:border-0 hover:bg-muted/50 transition-colors duration-100 cursor-pointer"
                        onClick={() => navigate(`/vehicles/${i.vehicle_id}`)}
                      >
                        <td className="px-4 py-2.5">
                          <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLE[i.status]}`}>
                            {STATUS_LABEL[i.status]}
                          </span>
                        </td>
                        <td className="px-4 py-2.5">
                          <p className="font-medium">{i.customer_name}</p>
                          {i.customer_phone && (
                            <p className="flex items-center gap-1 text-xs text-muted-foreground">
                              <Phone className="h-3 w-3" /> {i.customer_phone}
                            </p>
                          )}
                        </td>
                        <td className="px-4 py-2.5">
                          <p className="font-medium">{i.plate_number}</p>
                          <p className="text-xs text-muted-foreground">{[i.make, i.model].filter(Boolean).join(' ')}</p>
                        </td>
                        <td className="px-4 py-2.5">
                          <span className="flex items-center gap-1.5">
                            {i.event_type === 'oil' ? <Droplet className="h-3.5 w-3.5" /> : i.event_type === 'tire' ? <Disc className="h-3.5 w-3.5" /> : <Wrench className="h-3.5 w-3.5" />}
                            {i.label}
                          </span>
                        </td>
                        <td className="px-4 py-2.5 text-muted-foreground">
                          {i.last_service_at ? formatDate(i.last_service_at) : '—'}
                          {i.last_mileage != null ? ` · ${i.last_mileage.toLocaleString()} ${i.distance_unit === 'mi' ? 'mi' : 'km'}` : ''}
                        </td>
                        <td className="px-4 py-2.5 text-muted-foreground">
                          {i.due_date ? formatDate(i.due_date) : '—'}
                          {i.due_mileage != null ? ` · ${i.due_mileage.toLocaleString()} ${i.distance_unit === 'mi' ? 'mi' : 'km'}` : ''}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </TableCard>
          </div>

          <div className="md:hidden space-y-3">
            <div className="flex gap-2">
              <Select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)} className="w-1/2">
                <option value="">All types</option>
                <option value="oil">Oil</option>
                <option value="tire">Tires</option>
                <option value="part">Parts</option>
              </Select>
              <Select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="w-1/2">
                <option value="">All statuses</option>
                <option value="overdue">Overdue</option>
                <option value="due_soon">Due soon</option>
              </Select>
            </div>
            {items.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-8">Nobody's due right now.</p>
            ) : (
              items.map((i) => (
                <div
                  key={`${i.vehicle_id}-${i.key}`}
                  className="bg-card rounded-lg p-4 shadow-sm cursor-pointer"
                  onClick={() => navigate(`/vehicles/${i.vehicle_id}`)}
                >
                  <div className="flex items-center justify-between">
                    <span className="flex items-center gap-1.5 text-sm font-medium">
                      {i.event_type === 'oil' ? <Droplet className="h-3.5 w-3.5" /> : i.event_type === 'tire' ? <Disc className="h-3.5 w-3.5" /> : <Wrench className="h-3.5 w-3.5" />}
                      {i.plate_number}
                    </span>
                    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLE[i.status]}`}>
                      {STATUS_LABEL[i.status]}
                    </span>
                  </div>
                  <p className="text-sm mt-1">{i.customer_name}{i.customer_phone ? ` · ${i.customer_phone}` : ''}</p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Due {i.due_date ? formatDate(i.due_date) : '—'}
                    {i.due_mileage != null ? ` · ${i.due_mileage.toLocaleString()} ${i.distance_unit === 'mi' ? 'mi' : 'km'}` : ''}
                  </p>
                </div>
              ))
            )}
          </div>
        </>
      )}
    </div>
  )
}
