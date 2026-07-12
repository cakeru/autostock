import axios from 'axios'

const api = axios.create({
  // Relative by default so the app works on any host/IP: the frontend's nginx
  // proxies /api to the backend over the internal Docker network (no CORS, no
  // hardcoded IP). Override with VITE_API_URL only for split deployments.
  baseURL: import.meta.env.VITE_API_URL || '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api
