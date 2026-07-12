import api from './api'
import type { TelegramChannel } from '@/types/telegram'

export const telegramApi = {
  getChannels: async (): Promise<TelegramChannel[]> => {
    const res = await api.get('/telegram/channels')
    return res.data.data.channels
  },

  saveChannels: async (channels: TelegramChannel[]): Promise<void> => {
    await api.put('/telegram/channels', { channels })
  },

  getRoutes: async (): Promise<Record<string, string>> => {
    const res = await api.get('/telegram/routes')
    return res.data.data.routes || {}
  },

  saveRoutes: async (routes: Record<string, string>): Promise<void> => {
    await api.put('/telegram/routes', { routes })
  },

  testSend: async (channelId: string): Promise<void> => {
    await api.post('/telegram/test-send', { channel_id: channelId })
  },

  trigger: async (topic: string): Promise<void> => {
    await api.post('/telegram/trigger', { topic })
  },

  sendDocument: async (file: Blob, filename: string, caption: string): Promise<void> => {
    const form = new FormData()
    form.append('file', file, filename)
    if (caption) form.append('caption', caption)
    await api.post('/telegram/send-document', form, { headers: { 'Content-Type': 'multipart/form-data' } })
  },
}
