import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import type { CreateVehicleRequest } from '@/types/customer'

export interface VehicleFormProps {
  initial?: Partial<CreateVehicleRequest>
  onSubmit: (data: CreateVehicleRequest) => void
  onCancel: () => void
  loading?: boolean
  defaultUnit?: string
}

export function VehicleForm({ initial, onSubmit, onCancel, loading, defaultUnit = 'km' }: VehicleFormProps) {
  const [plateNumber, setPlateNumber] = useState(initial?.plate_number || '')
  const [make, setMake] = useState(initial?.make || '')
  const [model, setModel] = useState(initial?.model || '')
  const [year, setYear] = useState(initial?.year?.toString() || '')
  const [color, setColor] = useState(initial?.color || '')
  const [bodyType, setBodyType] = useState(initial?.body_type || '')
  const [distanceUnit, setDistanceUnit] = useState(initial?.distance_unit || defaultUnit)
  const [notes, setNotes] = useState(initial?.notes || '')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({
      plate_number: plateNumber,
      make,
      model,
      year: year ? parseInt(year) : undefined,
      color,
      body_type: bodyType || undefined,
      distance_unit: distanceUnit,
      notes,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-2">
      <div className="space-y-1">
        <Label>Plate Number *</Label>
        <Input value={plateNumber} onChange={(e) => setPlateNumber(e.target.value)} required />
      </div>
      <div className="grid grid-cols-2 gap-2">
        <div className="space-y-1">
          <Label>Make</Label>
          <Input value={make} onChange={(e) => setMake(e.target.value)} />
        </div>
        <div className="space-y-1">
          <Label>Model</Label>
          <Input value={model} onChange={(e) => setModel(e.target.value)} />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-2">
        <div className="space-y-1">
          <Label>Year</Label>
          <Input value={year} onChange={(e) => setYear(e.target.value)} type="number" min="1900" max="2100" />
        </div>
        <div className="space-y-1">
          <Label>Odometer unit</Label>
          <Select value={distanceUnit} onChange={(e) => setDistanceUnit(e.target.value)}>
            <option value="km">Kilometers (km)</option>
            <option value="mi">Miles (mi)</option>
          </Select>
          <p className="text-[11px] text-muted-foreground">Imported cars often read miles only — pick what this car's odometer shows.</p>
        </div>
      </div>
      <div className="space-y-1">
        <Label>Color</Label>
        <Input value={color} onChange={(e) => setColor(e.target.value)} />
      </div>
      <div className="space-y-1">
        <Label>Body type</Label>
        <Select value={bodyType} onChange={(e) => setBodyType(e.target.value)}>
          <option value="">— unspecified —</option>
          <option value="sedan">Sedan / hatchback</option>
          <option value="suv">SUV / van</option>
          <option value="pickup">Pickup</option>
          <option value="motorcycle">Motorcycle</option>
        </Select>
        <p className="text-xs text-muted-foreground">Sets which top-down car shape the profile draws.</p>
      </div>
      <div className="space-y-1">
        <Label>Notes</Label>
        <Input value={notes} onChange={(e) => setNotes(e.target.value)} />
      </div>
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onCancel}>Cancel</Button>
        <Button type="submit" disabled={loading}>{loading ? 'Saving...' : 'Save'}</Button>
      </div>
    </form>
  )
}
