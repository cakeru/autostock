import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import type { CreateEmployeeRequest, Employee, PayType } from '@/types/employee'

export interface EmployeeFormProps {
  initial?: Employee
  onSubmit: (data: CreateEmployeeRequest) => void
  onCancel: () => void
  loading?: boolean
}

export function EmployeeForm({ initial, onSubmit, onCancel, loading }: EmployeeFormProps) {
  const [name, setName] = useState(initial?.name || '')
  const [position, setPosition] = useState(initial?.position || '')
  const [phone, setPhone] = useState(initial?.phone || '')
  const [email, setEmail] = useState(initial?.email || '')
  const [payType, setPayType] = useState<PayType>(initial?.pay_type || 'salary')
  const [baseSalary, setBaseSalary] = useState(initial?.base_salary?.toString() || '0')
  const [hourlyRate, setHourlyRate] = useState(initial?.hourly_rate?.toString() || '0')
  const [commissionRate, setCommissionRate] = useState(initial?.commission_rate?.toString() || '0')
  const [hireDate, setHireDate] = useState(initial?.hire_date || '')
  const [notes, setNotes] = useState(initial?.notes || '')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({
      name, position, phone, email,
      pay_type: payType,
      base_salary: parseFloat(baseSalary) || 0,
      hourly_rate: parseFloat(hourlyRate) || 0,
      commission_rate: parseFloat(commissionRate) || 0,
      hire_date: hireDate || undefined,
      notes,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label>Name</Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div className="space-y-1.5">
          <Label>Position</Label>
          <Input value={position} onChange={(e) => setPosition(e.target.value)} placeholder="Technician, Cashier..." />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label>Phone</Label>
          <Input value={phone} onChange={(e) => setPhone(e.target.value)} />
        </div>
        <div className="space-y-1.5">
          <Label>Email</Label>
          <Input value={email} onChange={(e) => setEmail(e.target.value)} type="email" />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label>Hire Date</Label>
        <Input value={hireDate} onChange={(e) => setHireDate(e.target.value)} type="date" />
      </div>

      <div className="bg-muted/50 rounded-md p-4 space-y-3">
        <Label className="text-sm font-semibold">Pay</Label>
        <div className="space-y-1.5">
          <Label className="text-xs">Type</Label>
          <Select value={payType} onChange={(e) => setPayType(e.target.value as PayType)}>
            <option value="salary">Salary — fixed monthly</option>
            <option value="hourly">Hourly — rate × hours logged on jobs</option>
            <option value="commission">Commission — % of labor revenue</option>
            <option value="hybrid">Hybrid — base salary + commission</option>
          </Select>
        </div>
        <div className="grid grid-cols-2 gap-4">
          {(payType === 'salary' || payType === 'hybrid') && (
            <div className="space-y-1.5">
              <Label className="text-xs">Base Salary (USD/month)</Label>
              <Input value={baseSalary} onChange={(e) => setBaseSalary(e.target.value)} type="number" step="0.01" min="0" />
            </div>
          )}
          {payType === 'hourly' && (
            <div className="space-y-1.5">
              <Label className="text-xs">Hourly Rate (USD)</Label>
              <Input value={hourlyRate} onChange={(e) => setHourlyRate(e.target.value)} type="number" step="0.01" min="0" />
            </div>
          )}
          {(payType === 'commission' || payType === 'hybrid') && (
            <div className="space-y-1.5">
              <Label className="text-xs">Commission Rate (%)</Label>
              <Input value={commissionRate} onChange={(e) => setCommissionRate(e.target.value)} type="number" step="0.1" min="0" max="100" />
            </div>
          )}
        </div>
        <p className="text-[10px] text-muted-foreground">
          Commission is calculated on labor revenue from completed jobs assigned to this employee.
        </p>
      </div>

      <div className="space-y-1.5">
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
