export type PayType = 'salary' | 'hourly' | 'commission' | 'hybrid'

export interface Employee {
  id: number
  user_id?: number
  username?: string
  name: string
  position?: string
  phone?: string
  email?: string
  pay_type: PayType
  base_salary: number
  hourly_rate: number
  commission_rate: number
  hire_date?: string
  notes?: string
  is_active: boolean
  created_at: string
}

export interface CreateEmployeeRequest {
  name: string
  position?: string
  phone?: string
  email?: string
  pay_type?: PayType
  base_salary?: number
  hourly_rate?: number
  commission_rate?: number
  hire_date?: string
  notes?: string
}

export interface UpdateEmployeeRequest {
  name?: string
  position?: string
  phone?: string
  email?: string
  pay_type?: PayType
  base_salary?: number
  hourly_rate?: number
  commission_rate?: number
  hire_date?: string
  notes?: string
}

export interface CreateAccountRequest {
  username: string
  password: string
  role: 'admin' | 'staff'
  permissions?: string[]
}
