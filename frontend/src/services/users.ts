import api from './api'

export interface User {
  id: number
  username: string
  email?: string
  full_name?: string
  role: string
  permissions: string[]
  branch_id: number
  is_active: boolean
}

export interface CreateUserRequest {
  username: string
  password: string
  email?: string
  full_name?: string
  role: string
  permissions: string[]
}

export interface UpdateUserRequest {
  email?: string
  full_name?: string
  permissions?: string[]
  is_active?: boolean
}

export const usersApi = {
  list: async (): Promise<User[]> => {
    const res = await api.get('/users')
    return res.data.data
  },

  create: async (data: CreateUserRequest): Promise<User> => {
    const res = await api.post('/users', data)
    return res.data.data
  },

  update: async (id: number, data: UpdateUserRequest): Promise<User> => {
    const res = await api.put(`/users/${id}`, data)
    return res.data.data
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/users/${id}`)
  },
}
