import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ClipboardList } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import { TableCard, Th } from '@/components/ui/table'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { useStocktakes, useCreateStocktake } from '@/hooks/useStocktakes'
import type { StocktakeListItem } from '@/types/stocktake'

function StatusPill({ status }: { status: StocktakeListItem['status'] }) {
  const styles = {
    draft: 'bg-accent/10 text-accent',
    completed: 'bg-primary/10 text-primary',
    cancelled: 'bg-muted text-muted-foreground',
  }
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize ${styles[status]}`}>
      {status}
    </span>
  )
}

export function Stocktakes() {
  const navigate = useNavigate()
  const { data: stocktakes, isLoading } = useStocktakes()
  const createMutation = useCreateStocktake()
  const [showCreate, setShowCreate] = useState(false)

  return (
    <div className="space-y-6">
      <PageHeader
        title="Stocktakes"
        subtitle="Physical inventory counts — expected vs. counted, with variances applied as real stock adjustments"
        actions={<Button size="sm" onClick={() => setShowCreate(true)}>New Stocktake</Button>}
      />

      {isLoading ? (
        <div className="bg-card rounded-lg shadow-sm p-4">
          <TableSkeleton rows={5} cols={5} />
        </div>
      ) : !stocktakes || stocktakes.length === 0 ? (
        <div className="rounded-lg bg-card p-10 text-center text-sm text-muted-foreground shadow-sm">
          <ClipboardList className="mx-auto mb-2 h-8 w-8 text-muted-foreground/50" />
          No stocktakes yet
        </div>
      ) : (
        <TableCard>
          <table className="w-full text-sm">
            <thead>
              <tr className="sticky top-0 z-10 border-b bg-card">
                <Th>Status</Th>
                <Th>Notes</Th>
                <Th className="text-right">Items</Th>
                <Th className="text-right">Counted</Th>
                <Th className="text-right">Variances</Th>
                <Th>Created by</Th>
                <Th>Date</Th>
              </tr>
            </thead>
            <tbody>
              {stocktakes.map((st) => (
                <tr
                  key={st.id}
                  onClick={() => navigate(`/stocktakes/${st.id}`)}
                  className="cursor-pointer border-b last:border-0 transition-colors hover:bg-muted/50"
                >
                  <td className="px-4 py-2.5"><StatusPill status={st.status} /></td>
                  <td className="px-4 py-2.5 text-muted-foreground">{st.notes || '—'}</td>
                  <td className="px-4 py-2.5 text-right tabular-nums">{st.item_count}</td>
                  <td className="px-4 py-2.5 text-right tabular-nums">{st.counted_count} / {st.item_count}</td>
                  <td className="px-4 py-2.5 text-right tabular-nums">
                    {st.variance_count > 0
                      ? <span className="font-medium text-destructive">{st.variance_count}</span>
                      : <span className="text-muted-foreground">0</span>}
                  </td>
                  <td className="px-4 py-2.5 text-muted-foreground">{st.created_by_name || '—'}</td>
                  <td className="px-4 py-2.5 text-muted-foreground">{new Date(st.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </TableCard>
      )}

      {showCreate && (
        <CreateStocktakeDialog
          onClose={() => setShowCreate(false)}
          onCreate={(data) => createMutation.mutate(data, {
            onSuccess: (st) => { setShowCreate(false); navigate(`/stocktakes/${st.id}`) },
          })}
          loading={createMutation.isPending}
        />
      )}
    </div>
  )
}

function CreateStocktakeDialog({ onClose, onCreate, loading }: {
  onClose: () => void
  onCreate: (data: { notes?: string; type?: string; category?: string }) => void
  loading: boolean
}) {
  const [notes, setNotes] = useState('')
  const [type, setType] = useState('')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-sm rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">New Stocktake</p>
        <p className="mb-3 text-xs text-muted-foreground">
          Start a blank count sheet, or pre-load it with every product of one type at its current on-hand quantity.
        </p>
        <div className="space-y-3">
          <div className="space-y-1.5">
            <Label>Notes</Label>
            <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="e.g. Monthly tire count" />
          </div>
          <div className="space-y-1.5">
            <Label>Pre-load by type (optional)</Label>
            <Select value={type} onChange={(e) => setType(e.target.value)}>
              <option value="">Start blank — add products one at a time</option>
              <option value="tire">All tires</option>
              <option value="part">All parts</option>
              <option value="labor">All labor</option>
              <option value="consumable">All consumables</option>
            </Select>
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-4">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="button" onClick={() => onCreate({ notes: notes || undefined, type: type || undefined })} disabled={loading}>
            {loading ? 'Creating...' : 'Create'}
          </Button>
        </div>
      </div>
    </div>
  )
}
