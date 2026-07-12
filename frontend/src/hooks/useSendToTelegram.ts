import { useState } from 'react'
import { toast } from 'sonner'
import { captureElementToPdf } from '@/utils/capturePdf'
import { telegramApi } from '@/services/telegram'

// Captures a DOM element to a PDF and sends it to Telegram's "documents"
// channel. Shared by the invoice and vehicle-report "Send to Telegram" buttons.
export function useSendToTelegram() {
  const [sending, setSending] = useState(false)

  const send = async (el: HTMLElement, filename: string, caption: string) => {
    setSending(true)
    const t = toast.loading('Preparing PDF…')
    try {
      const blob = await captureElementToPdf(el)
      toast.loading('Sending to Telegram…', { id: t })
      await telegramApi.sendDocument(blob, filename, caption)
      toast.success('Sent to Telegram — open Telegram to forward it', { id: t })
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      toast.error(msg || 'Could not send to Telegram', { id: t })
    } finally {
      setSending(false)
    }
  }

  return { send, sending }
}
