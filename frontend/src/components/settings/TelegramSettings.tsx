import { useEffect, useState } from 'react'
import { Plus, Trash2, Send, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import {
  useTelegramChannels, useTelegramRoutes, useSaveTelegramChannels, useSaveTelegramRoutes,
  useTestSendTelegram, useTriggerTelegramTopic,
} from '@/hooks/useTelegram'
import { TELEGRAM_TOPICS } from '@/types/telegram'
import type { TelegramChannel } from '@/types/telegram'

function newChannelID() {
  return `c${Math.random().toString(36).slice(2, 9)}`
}

export function TelegramSettings() {
  const { data: savedChannels, isLoading: channelsLoading } = useTelegramChannels()
  const { data: savedRoutes, isLoading: routesLoading } = useTelegramRoutes()
  const saveChannels = useSaveTelegramChannels()
  const saveRoutes = useSaveTelegramRoutes()
  const testSend = useTestSendTelegram()
  const trigger = useTriggerTelegramTopic()

  const [channels, setChannels] = useState<TelegramChannel[]>([])
  const [routes, setRoutes] = useState<Record<string, string>>({})

  useEffect(() => {
    if (!savedChannels) return
    // Always present at least one (blank) channel row so the bot token /
    // chat ID fields are visible immediately — first-time setup shouldn't
    // require discovering the "Add channel" button.
    setChannels(savedChannels.length > 0 ? savedChannels
      : [{ id: newChannelID(), label: '', bot_token: '', chat_id: '' }])
  }, [savedChannels])
  useEffect(() => { if (savedRoutes) setRoutes(savedRoutes) }, [savedRoutes])

  const updateChannel = (id: string, patch: Partial<TelegramChannel>) => {
    setChannels((cs) => cs.map((c) => (c.id === id ? { ...c, ...patch } : c)))
  }
  const addChannel = () => {
    setChannels((cs) => [...cs, { id: newChannelID(), label: '', bot_token: '', chat_id: '' }])
  }
  const removeChannel = (id: string) => {
    setChannels((cs) => cs.filter((c) => c.id !== id))
    setRoutes((rs) => {
      const next = { ...rs }
      for (const topic of Object.keys(next)) if (next[topic] === id) next[topic] = ''
      return next
    })
  }

  // Untouched blank rows (like the auto-added starter row) don't count against
  // validity and are dropped on save; only partially-filled rows block saving.
  const isBlank = (c: TelegramChannel) => !c.label.trim() && !c.bot_token.trim() && !c.chat_id.trim()
  const filledChannels = channels.filter((c) => !isBlank(c))
  const channelsValid = filledChannels.every((c) => c.label.trim() && c.bot_token.trim() && c.chat_id.trim())

  if (channelsLoading || routesLoading) {
    return <p className="text-sm text-muted-foreground">Loading Telegram settings...</p>
  }

  return (
    <div className="bg-card rounded-lg p-5 shadow-sm space-y-5">
      <div>
        <h2 className="text-sm font-medium">Telegram Notifications</h2>
        <p className="text-xs text-muted-foreground">
          Configure one or more bot + chat destinations ("channels"), then route each type of notification to whichever
          channel you want — the same group for everything, or a different one per topic. A channel's chat can be a
          group or a personal DM; Telegram treats them the same way.
        </p>
      </div>

      <div className="space-y-3">
        <Label className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Channels</Label>
        <p className="text-xs text-muted-foreground">
          Create a bot by messaging <span className="font-mono">@BotFather</span> on Telegram — it replies with the
          bot token. For a group chat, add the bot to the group; for a personal chat, send the bot any message first.
          Then use the "send test" button to confirm it works.
        </p>
        {channels.map((c) => (
          <div key={c.id} className="flex flex-wrap items-end gap-2 rounded-md border p-3">
            <div className="min-w-[120px] flex-1 space-y-1">
              <label className="text-xs text-muted-foreground">Label</label>
              <Input value={c.label} onChange={(e) => updateChannel(c.id, { label: e.target.value })} placeholder="Shop Floor" />
            </div>
            <div className="min-w-[160px] flex-1 space-y-1">
              <label className="text-xs text-muted-foreground">Bot token (from @BotFather)</label>
              <Input value={c.bot_token} onChange={(e) => updateChannel(c.id, { bot_token: e.target.value })} type="password" placeholder="123456:ABC-DEF..." />
            </div>
            <div className="min-w-[140px] flex-1 space-y-1">
              <label className="text-xs text-muted-foreground">Chat ID (group or personal)</label>
              <Input value={c.chat_id} onChange={(e) => updateChannel(c.id, { chat_id: e.target.value })} placeholder="-100123456789" />
            </div>
            <Button
              type="button" variant="outline" size="sm"
              onClick={() => testSend.mutate(c.id)}
              disabled={!c.bot_token.trim() || !c.chat_id.trim() || testSend.isPending}
              title="Send a test message to verify this channel"
            >
              <Send className="h-3.5 w-3.5" />
            </Button>
            <Button type="button" variant="ghost" size="icon" onClick={() => removeChannel(c.id)} aria-label="Remove channel">
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))}
        <div className="flex items-center gap-2">
          <Button type="button" variant="outline" size="sm" onClick={addChannel} className="gap-1">
            <Plus className="h-3.5 w-3.5" /> Add channel
          </Button>
          <Button
            type="button" size="sm"
            onClick={() => saveChannels.mutate(filledChannels)}
            disabled={!channelsValid || saveChannels.isPending}
          >
            {saveChannels.isPending ? 'Saving...' : 'Save channels'}
          </Button>
          {!channelsValid && (
            <span className="text-xs text-destructive">Every channel needs a label, bot token, and chat ID</span>
          )}
        </div>
      </div>

      <div className="space-y-3">
        <Label className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Routing</Label>
        <div className="space-y-2">
          {TELEGRAM_TOPICS.map((t) => (
            <div key={t.value} className="flex flex-wrap items-center gap-2 rounded-md border p-3">
              <div className="min-w-[180px] flex-1">
                <p className="text-sm font-medium">{t.label}</p>
                <p className="text-xs text-muted-foreground">{t.description}</p>
              </div>
              <Select
                value={routes[t.value] || ''}
                onChange={(e) => setRoutes((rs) => ({ ...rs, [t.value]: e.target.value }))}
                className="w-48"
              >
                <option value="">— off —</option>
                {channels.filter((c) => c.label.trim()).map((c) => (
                  <option key={c.id} value={c.id}>{c.label}</option>
                ))}
              </Select>
              {t.scheduled && (
                <Button
                  type="button" variant="outline" size="sm"
                  onClick={() => trigger.mutate(t.value)}
                  disabled={trigger.isPending}
                  title="Send this one now, without waiting for its schedule"
                  className="gap-1"
                >
                  <Zap className="h-3.5 w-3.5" /> Send now
                </Button>
              )}
            </div>
          ))}
        </div>
        <Button type="button" size="sm" onClick={() => saveRoutes.mutate(routes)} disabled={saveRoutes.isPending}>
          {saveRoutes.isPending ? 'Saving...' : 'Save routing'}
        </Button>
      </div>
    </div>
  )
}
