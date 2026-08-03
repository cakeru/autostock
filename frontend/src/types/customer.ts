export interface Customer {
  id: number
  name: string
  customer_type: 'garage' | 'retail' | 'company'
  phone?: string
  email?: string
  address?: string
  notes?: string
  customer_since?: string
  is_active: boolean
  vehicle_count: number
  vehicle_plates?: string
  total_spent?: number
  last_visit?: string
  created_at: string
  updated_at: string
}

export interface CustomerDetail extends Customer {
  vehicles: Vehicle[]
  visit_count?: number
  outstanding?: number
}

export interface Vehicle {
  id: number
  customer_id: number
  plate_number: string
  distance_unit?: string
  make?: string
  model?: string
  year?: number
  vin?: string
  color?: string
  body_type?: string
  notes?: string
  created_at: string
  updated_at: string
}

export interface ActivityItem {
  type: 'job' | 'invoice'
  id: number
  ref: string
  date?: string
  title: string
  status: string
  amount: number
  outstanding: number
  plate?: string
}

export interface CustomerListParams {
  name_like?: string
  phone?: string
  email?: string
  search?: string
  page?: number
  per_page?: number
}

export interface CreateCustomerRequest {
  name: string
  customer_type?: string
  phone?: string
  email?: string
  address?: string
  notes?: string
}

export interface UpdateCustomerRequest {
  name?: string
  customer_type?: string
  phone?: string
  email?: string
  address?: string
  notes?: string
}

export interface CreateVehicleRequest {
  plate_number: string
  distance_unit?: string
  make?: string
  model?: string
  year?: number
  vin?: string
  color?: string
  body_type?: string
  notes?: string
}

export interface UpdateVehicleRequest {
  plate_number?: string
  distance_unit?: string
  make?: string
  model?: string
  year?: number
  vin?: string
  color?: string
  body_type?: string
  notes?: string
}
