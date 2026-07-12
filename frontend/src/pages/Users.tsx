import { Pencil, Trash2 } from 'lucide-react'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { PageHeader } from '@/components/layout/PageHeader'
import { useState } from 'react'
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from '@/hooks/useUsers'
import { useAuth } from '@/contexts/AuthContext'
import { PERMISSION_GROUPS, DEFAULT_STAFF_PERMISSIONS } from '@/constants/permissions'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { TableCard, TableFooter, Th, ActionsTh } from '@/components/ui/table'

export function Users() {
  const { user: currentUser } = useAuth()
  const { data: users, isLoading } = useUsers()
  const createMutation = useCreateUser()
  const updateMutation = useUpdateUser()
  const deleteMutation = useDeleteUser()

  const [showForm, setShowForm] = useState(false)
  const [editingUser, setEditingUser] = useState<any>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [deleteError, setDeleteError] = useState('')

  const [formUsername, setFormUsername] = useState('')
  const [formPassword, setFormPassword] = useState('')
  const [formFullName, setFormFullName] = useState('')
  const [formEmail, setFormEmail] = useState('')
  const [formRole, setFormRole] = useState('staff')
  const [formPermissions, setFormPermissions] = useState<string[]>(DEFAULT_STAFF_PERMISSIONS)
  const [formActive, setFormActive] = useState(true)

  const userList = users || []
  const adminCount = userList.filter((u: any) => u.role === 'admin' && u.is_active).length

  const resetForm = () => {
    setFormUsername('')
    setFormPassword('')
    setFormFullName('')
    setFormEmail('')
    setFormRole('staff')
    setFormPermissions(DEFAULT_STAFF_PERMISSIONS)
    setFormActive(true)
    setEditingUser(null)
  }

  const openCreate = () => {
    resetForm()
    setShowForm(true)
  }

  const openEdit = (u: any) => {
    setEditingUser(u)
    setFormUsername(u.username)
    setFormPassword('')
    setFormFullName(u.full_name || '')
    setFormEmail(u.email || '')
    setFormRole(u.role)
    setFormPermissions(u.permissions || [])
    setFormActive(u.is_active)
    setShowForm(true)
  }

  const togglePermission = (perm: string) => {
    setFormPermissions((prev) =>
      prev.includes(perm) ? prev.filter((p) => p !== perm) : [...prev, perm]
    )
  }

  const handleSubmit = () => {
    if (!formUsername) return
    if (editingUser) {
      updateMutation.mutate({
        id: editingUser.id,
        data: {
          full_name: formFullName || undefined,
          email: formEmail || undefined,
          permissions: formRole === 'admin' ? undefined : formPermissions,
          is_active: formActive,
        },
      }, { onSuccess: () => { setShowForm(false); resetForm() } })
    } else {
      if (!formPassword || formPassword.length < 6) return
      createMutation.mutate({
        username: formUsername,
        password: formPassword,
        full_name: formFullName || undefined,
        email: formEmail || undefined,
        role: formRole,
        permissions: formRole === 'admin' ? [] : formPermissions,
      }, { onSuccess: () => { setShowForm(false); resetForm() } })
    }
  }

  const handleDelete = () => {
    if (!deletingId) return
    const userToDelete = userList.find((u: any) => u.id === deletingId)
    if (!userToDelete) return

    if (userToDelete.role === 'admin' && adminCount <= 1) {
      setDeleteError('Cannot delete the last admin user')
      return
    }

    deleteMutation.mutate(deletingId, {
      onSuccess: () => { setDeletingId(null); setDeleteError('') },
    })
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Users"
        actions={
          <Button onClick={openCreate} size="sm">Add User</Button>
        }
      />

      {isLoading ? (
        <div className="bg-card rounded-lg shadow-sm p-4">
          <TableSkeleton rows={5} cols={6} />
        </div>
      ) : (
        <>
        <div className="hidden md:block">
        <TableCard
          footer={<TableFooter total={userList.length} page={1} totalPages={1} onPage={() => {}} noun="users" />}
        >
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-card sticky top-0 z-10">
                <Th>User</Th>
                <Th>Email</Th>
                <Th>Role</Th>
                <Th>Status</Th>
                <ActionsTh />
              </tr>
            </thead>
            <tbody>
              {userList.length === 0 ? (
                <tr><td colSpan={5} className="px-4 py-10 text-center text-muted-foreground">No users found</td></tr>
              ) : (
                userList.map((u: any) => {
                  const isLastAdmin = u.role === 'admin' && adminCount <= 1
                  return (
                    <tr key={u.id} className="group border-b last:border-0 hover:bg-muted/50 transition-colors duration-100">
                      <td className="px-4 py-2.5">
                        <p className="font-medium">{u.username}</p>
                        {u.full_name && <p className="text-xs text-muted-foreground">{u.full_name}</p>}
                      </td>
                      <td className="px-4 py-2.5 text-muted-foreground">{u.email || '—'}</td>
                      <td className="px-4 py-2.5">
                        <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                          u.role === 'admin' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'
                        }`}>{u.role}</span>
                      </td>
                      <td className="px-4 py-2.5">
                        <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                          u.is_active ? 'bg-primary/10 text-primary' : 'bg-destructive/10 text-destructive'
                        }`}>{u.is_active ? 'Active' : 'Inactive'}</span>
                      </td>
                      <td className="px-4 py-2.5 text-right">
                        <div className="flex justify-end gap-1 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
                          <Button variant="ghost" size="icon" onClick={() => openEdit(u)} aria-label="Edit" title="Edit user">
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            disabled={isLastAdmin}
                            className="hover:text-destructive"
                            onClick={() => { setDeletingId(u.id); setDeleteError('') }}
                            aria-label="Delete"
                            title="Delete user"
                          ><Trash2 className="h-4 w-4" /></Button>
                        </div>
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </TableCard>
        </div>

        {/* Mobile cards — actions are always visible (no hover on touch) */}
        <div className="md:hidden space-y-3">
          {userList.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-8">No users found</p>
          ) : (
            userList.map((u: any) => {
              const isLastAdmin = u.role === 'admin' && adminCount <= 1
              return (
                <div key={u.id} className="bg-card rounded-lg p-4 shadow-sm">
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="font-medium text-sm truncate">{u.username}</p>
                      {u.full_name && <p className="text-xs text-muted-foreground truncate">{u.full_name}</p>}
                      {u.email && <p className="text-xs text-muted-foreground truncate">{u.email}</p>}
                    </div>
                    <div className="flex flex-shrink-0 gap-1 text-muted-foreground">
                      <Button variant="ghost" size="icon" onClick={() => openEdit(u)} aria-label="Edit"><Pencil className="h-4 w-4" /></Button>
                      <Button variant="ghost" size="icon" disabled={isLastAdmin} className="hover:text-destructive" onClick={() => { setDeletingId(u.id); setDeleteError('') }} aria-label="Delete"><Trash2 className="h-4 w-4" /></Button>
                    </div>
                  </div>
                  <div className="mt-2 flex gap-1.5">
                    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${u.role === 'admin' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'}`}>{u.role}</span>
                    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${u.is_active ? 'bg-primary/10 text-primary' : 'bg-destructive/10 text-destructive'}`}>{u.is_active ? 'Active' : 'Inactive'}</span>
                  </div>
                </div>
              )
            })
          )}
        </div>
        </>
      )}

      {/* Create/Edit Modal */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-10 bg-black/50 overflow-auto">
          <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-lg mx-4 mt-10 mb-10">
            <h2 className="text-sm font-semibold mb-3">{editingUser ? 'Edit User' : 'Add User'}</h2>

            <div className="space-y-3">
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Username *</label>
                <Input value={formUsername} onChange={(e) => setFormUsername(e.target.value)} disabled={!!editingUser} />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium">{editingUser ? 'New Password (leave blank to keep)' : 'Password *'}</label>
                <Input value={formPassword} onChange={(e) => setFormPassword(e.target.value)} type="password" />
                {!editingUser && formPassword.length > 0 && formPassword.length < 6 &&
                  <p className="text-xs text-destructive">Minimum 6 characters</p>}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Full Name</label>
                  <Input value={formFullName} onChange={(e) => setFormFullName(e.target.value)} />
                </div>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Email</label>
                  <Input value={formEmail} onChange={(e) => setFormEmail(e.target.value)} type="email" />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Role</label>
                  <Select value={formRole} onChange={(e) => {
                    setFormRole(e.target.value)
                    if (e.target.value === 'staff') {
                      setFormPermissions(DEFAULT_STAFF_PERMISSIONS)
                    }
                  }}>
                    <option value="admin">Admin</option>
                    <option value="staff">Staff</option>
                  </Select>
                </div>
                <div className="space-y-1.5 flex items-end">
                  <label className="flex items-center gap-2 text-sm cursor-pointer pb-2">
                    <input
                      type="checkbox"
                      checked={formActive}
                      onChange={(e) => setFormActive(e.target.checked)}
                      className="rounded border-input"
                    />
                    Active
                  </label>
                </div>
              </div>

              {/* Permissions (only for staff) */}
              {formRole === 'staff' && (
                <div className="border rounded-md p-4 space-y-3">
                  <label className="text-sm font-semibold font-medium">Permissions</label>
                  {PERMISSION_GROUPS.map((group) => (
                    <div key={group.group}>
                      <p className="text-xs font-medium text-muted-foreground mb-1">{group.group}</p>
                      <div className="grid grid-cols-2 gap-1">
                        {group.permissions.map((perm) => (
                          <label key={perm.key} className="flex items-center gap-1.5 text-xs cursor-pointer py-0.5">
                            <input
                              type="checkbox"
                              checked={formPermissions.includes(perm.key)}
                              onChange={() => togglePermission(perm.key)}
                              className="rounded border-input"
                            />
                            {perm.label}
                          </label>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}

              <div className="flex justify-end gap-2 pt-2">
                <Button variant="ghost" onClick={() => { setShowForm(false); resetForm() }}>Cancel</Button>
                <Button onClick={handleSubmit} disabled={createMutation.isPending || updateMutation.isPending}>
                  {createMutation.isPending || updateMutation.isPending ? 'Saving...' : editingUser ? 'Update User' : 'Create User'}
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation */}
      {deletingId && (() => {
        const user = users?.find(u => u.id === deletingId)
        return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-card rounded-lg p-5 shadow-lg w-full max-w-sm mx-4">
            {deleteError ? (
              <p className="text-sm text-destructive mb-3">{deleteError}</p>
            ) : (
              <p className="text-sm mb-3">Delete <strong>{user?.username || 'this user'}</strong>?</p>
            )}
            <div className="flex justify-end gap-2">
              <Button variant="ghost" onClick={() => { setDeletingId(null); setDeleteError('') }}>Cancel</Button>
              {!deleteError && (
                <Button variant="destructive" onClick={handleDelete} disabled={deleteMutation.isPending}>
                  {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
                </Button>
              )}
            </div>
          </div>
        </div>
        )
      })()}
    </div>
  )
}
