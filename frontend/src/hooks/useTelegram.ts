import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { telegramApi } from '@/services/telegram'
import type { TelegramChannel } from '@/types/telegram'

export function useTelegramChannels() {
  return useQuery({
    queryKey: ['telegram-channels'],
    queryFn: () => telegramApi.getChannels(),
  })
}

export function useTelegramRoutes() {
  return useQuery({
    queryKey: ['telegram-routes'],
    queryFn: () => telegramApi.getRoutes(),
  })
}

export function useSaveTelegramChannels() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (channels: TelegramChannel[]) => telegramApi.saveChannels(channels),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['telegram-channels'] }); toast.success('Channels saved') },
  })
}

export function useSaveTelegramRoutes() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (routes: Record<string, string>) => telegramApi.saveRoutes(routes),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['telegram-routes'] }); toast.success('Routing saved') },
  })
}

export function useTestSendTelegram() {
  return useMutation({
    mutationFn: (channelId: string) => telegramApi.testSend(channelId),
    onSuccess: () => toast.success('Test message sent'),
  })
}

export function useTriggerTelegramTopic() {
  return useMutation({
    mutationFn: (topic: string) => telegramApi.trigger(topic),
    onSuccess: () => toast.success('Triggered — check the routed channel shortly'),
  })
}
