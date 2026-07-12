import { useState } from 'react'
import { KeyRound, Pencil, Trash2, UserRound } from 'lucide-react'
import { PageHeader } from '@/components/layout/PageHeader'
import { SlideOver } from '@/components/ui/SlideOver'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { TableCard, Th, ActionsTh } from '@/components/ui/table'
import { formatUSD } from '@/utils/currency'
import { EmployeeForm } from '@/components/employee/EmployeeForm'
import {
  useEmployees, useCreateEmployee, useUpdateEmployee, useDeleteEmployee, useCreateEmployeeAccount,
} from '@/hooks/useEmployees'
import type { Employee } from '@/types/employee'

const PAY_LABEL: Record<string, string> = {
  salary: 'Salary', hourly: 'Hourly', commission: 'Commission', hybrid: 'Salary + Commission',
}

export function Employees() {
  const { data: employees, isLoading } = useEmployees()
  const createMutation = useCreateEmployee()
  const updateMutation = useUpdateEmployee()
  const deleteMutation = useDeleteEmployee()

  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Employee | null>(null)
  const [deleting, setDeleting] = useState<Employee | null>(null)
  const [creatingAccountFor, setCreatingAccountFor] = useState<Employee | null>(null)

  const handleSubmit = (data: any) => {
    const onSuccess = () => { setShowForm(false); setEditing(null) }
    if (editing) updateMutation.mutate({ id: editing.id, data }, { onSuccess })
    else createMutation.mutate(data, { onSuccess })
  }

  const list = employees || []

  return (
    <div className="space-y-6">
      <PageHeader
        title="Employees"
        subtitle="HR profiles for pay and job assignment — a login account is optional"
        actions={<Button size="sm" onClick={() => { setEditing(null); setShowForm(true) }}>Add Employee</Button>}
      />

      {isLoading ? (
        <div className="bg-card rounded-lg shadow-sm p-4">
          <TableSkeleton rows={5} cols={5} />
        </div>
      ) : list.length === 0 ? (
        <div className="rounded-lg bg-card p-10 text-center text-sm text-muted-foreground shadow-sm">
          <UserRound className="mx-auto mb-2 h-8 w-8 text-muted-foreground/50" />
          No employees yet
        </div>
      ) : (
        <TableCard>
          <table className="w-full text-sm">
            <thead>
              <tr className="sticky top-0 z-10 border-b bg-card">
                <Th>Name</Th>
                <Th>Position</Th>
                <Th>Account</Th>
                <Th>Pay</Th>
                <ActionsTh />
              </tr>
            </thead>
            <tbody>
              {list.map((e) => (
                <tr key={e.id} className="group border-b last:border-0 transition-colors hover:bg-muted/50">
                  <td className="px-4 py-2.5">
                    <p className="font-medium">{e.name}</p>
                    {e.phone && <p className="text-xs text-muted-foreground">{e.phone}</p>}
                  </td>
                  <td className="px-4 py-2.5 text-muted-foreground">{e.position || '—'}</td>
                  <td className="px-4 py-2.5">
                    {e.username ? (
                      <span className="font-mono text-xs">{e.username}</span>
                    ) : (
                      <span className="text-xs text-muted-foreground">No login</span>
                    )}
                  </td>
                  <td className="px-4 py-2.5">
                    <p>{PAY_LABEL[e.pay_type]}</p>
                    <p className="text-xs text-muted-foreground">
                      {(e.pay_type === 'salary' || e.pay_type === 'hybrid') && `${formatUSD(e.base_salary)}/mo`}
                      {e.pay_type === 'hourly' && `${formatUSD(e.hourly_rate)}/hr`}
                      {(e.pay_type === 'commission' || e.pay_type === 'hybrid') && ` ${e.pay_type === 'hybrid' ? '+ ' : ''}${e.commission_rate}%`}
                    </p>
                  </td>
                  <td className="px-4 py-2.5 text-right">
                    <div className="flex justify-end gap-1 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
                      {!e.user_id && (
                        <Button variant="ghost" size="icon" onClick={() => setCreatingAccountFor(e)} aria-label="Create account" title="Create login account">
                          <KeyRound className="h-4 w-4" />
                        </Button>
                      )}
                      <Button variant="ghost" size="icon" onClick={() => { setEditing(e); setShowForm(true) }} aria-label="Edit" title="Edit">
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={() => setDeleting(e)} aria-label="Delete" title="Deactivate">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </TableCard>
      )}

      <SlideOver
        open={showForm}
        onClose={() => { setShowForm(false); setEditing(null) }}
        title={editing ? 'Edit Employee' : 'Add Employee'}
      >
        <EmployeeForm
          initial={editing || undefined}
          onSubmit={handleSubmit}
          onCancel={() => { setShowForm(false); setEditing(null) }}
          loading={createMutation.isPending || updateMutation.isPending}
        />
      </SlideOver>

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleting && deleteMutation.mutate(deleting.id, { onSuccess: () => setDeleting(null) })}
        title="Deactivate Employee"
        message={`Deactivate ${deleting?.name}? They'll no longer be assignable to new jobs.`}
        destructive
        loading={deleteMutation.isPending}
      />

      {creatingAccountFor && (
        <CreateAccountDialog employee={creatingAccountFor} onClose={() => setCreatingAccountFor(null)} />
      )}
    </div>
  )
}

function CreateAccountDialog({ employee, onClose }: { employee: Employee; onClose: () => void }) {
  const mutation = useCreateEmployeeAccount()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<'admin' | 'staff'>('staff')

  const handleCreate = () => {
    mutation.mutate({ id: employee.id, data: { username, password, role } }, {
      onSuccess: onClose,
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-sm rounded-lg bg-card p-5 shadow-lg">
        <p className="mb-1 text-sm font-semibold">Create Login for {employee.name}</p>
        <p className="mb-3 text-xs text-muted-foreground">
          Their pay info and job history stay exactly as-is — this just adds system access.
        </p>
        <div className="space-y-2">
          <div className="space-y-1">
            <label className="text-xs font-medium">Username *</label>
            <Input value={username} onChange={(e) => setUsername(e.target.value)} />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">Password *</label>
            <Input value={password} onChange={(e) => setPassword(e.target.value)} type="password" placeholder="min 6 characters" />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">Role *</label>
            <Select value={role} onChange={(e) => setRole(e.target.value as 'admin' | 'staff')}>
              <option value="staff">Staff</option>
              <option value="admin">Admin</option>
            </Select>
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-4">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button onClick={handleCreate} disabled={!username || password.length < 6 || mutation.isPending}>
            {mutation.isPending ? 'Creating...' : 'Create Account'}
          </Button>
        </div>
      </div>
    </div>
  )
}
