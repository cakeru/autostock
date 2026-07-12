import { useEffect, useRef, type ReactNode } from 'react'
import { Button } from '@/components/ui/button'

interface ConfirmDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  message: string
  confirmLabel?: string
  loading?: boolean
  destructive?: boolean
  children?: ReactNode
}

export function ConfirmDialog({ open, onClose, onConfirm, title, message, confirmLabel, loading, destructive, children }: ConfirmDialogProps) {
  const cancelRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (open) {
      setTimeout(() => cancelRef.current?.focus(), 50)
    }
  }, [open])

  useEffect(() => {
    if (!open) return
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [open, onClose])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onClose}
      role="alertdialog"
      aria-modal="true"
      aria-label={title}
    >
      <div
        className="bg-card rounded-lg p-5 shadow-lg w-full max-w-sm mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <p className="text-sm font-semibold mb-1">{title}</p>
        <p className="text-sm text-muted-foreground mb-3">{message}</p>
        {children}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" ref={cancelRef} onClick={onClose}>Cancel</Button>
          <Button
            variant={destructive ? 'destructive' : 'default'}
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? 'Processing...' : confirmLabel || (destructive ? 'Delete' : 'Confirm')}
          </Button>
        </div>
      </div>
    </div>
  )
}
