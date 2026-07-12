import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import type { CreateCustomerRequest } from '@/types/customer'

export interface CustomerFormProps {
  initial?: Partial<CreateCustomerRequest>
  onSubmit: (data: CreateCustomerRequest) => void
  onCancel: () => void
  loading?: boolean
}

export function CustomerForm({ initial, onSubmit, onCancel, loading }: CustomerFormProps) {
  const [name, setName] = useState(initial?.name || '')
  const [phone, setPhone] = useState(initial?.phone || '')
  const [email, setEmail] = useState(initial?.email || '')
  const [address, setAddress] = useState(initial?.address || '')
  const [notes, setNotes] = useState(initial?.notes || '')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({ name, phone, email, address, notes })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-2">
      <div className="space-y-1">
        <Label>Name *</Label>
        <Input value={name} onChange={(e) => setName(e.target.value)} required />
      </div>
      <div className="grid grid-cols-2 gap-2">
        <div className="space-y-1">
          <Label>Phone</Label>
          <Input value={phone} onChange={(e) => setPhone(e.target.value)} />
        </div>
        <div className="space-y-1">
          <Label>Email</Label>
          <Input value={email} onChange={(e) => setEmail(e.target.value)} type="email" />
        </div>
      </div>
      <div className="space-y-1">
        <Label>Address</Label>
        <Input value={address} onChange={(e) => setAddress(e.target.value)} />
      </div>
      <div className="space-y-1">
        <Label>Notes</Label>
        <Textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={2} />
      </div>
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onCancel}>Cancel</Button>
        <Button type="submit" disabled={loading}>{loading ? 'Saving...' : 'Save'}</Button>
      </div>
    </form>
  )
}
