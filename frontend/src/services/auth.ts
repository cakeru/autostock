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

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: User
}

export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const res = await api.post('/auth/login', data)
    return res.data
  },

  getMe: async (): Promise<User> => {
    const res = await api.get('/auth/me')
    return res.data.data
  },

  changePassword: async (currentPassword: string, newPassword: string) => {
    await api.put('/auth/password', {
      current_password: currentPassword,
      new_password: newPassword,
    })
  },
}
