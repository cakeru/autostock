import { PageHeader } from '@/components/layout/PageHeader'
import { useState, useEffect, useCallback } from 'react'
import { useSettings, useUpdateSetting, useBatchUpdateSettings, useUpdateExchangeRate } from '@/hooks/useSettings'
import { useIntervalSettings, useUpdateIntervalSettings } from '@/hooks/useVehicleProfile'
import type { PartRule } from '@/types/vehicleProfile'
import { distanceUnit, unitLabel } from '@/utils/units'
import { Trash2, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { PACKAGES, LABOR_PRESETS, FEE_PRESETS, DEFAULT_PAYMENT_METHODS, parseArraySetting } from '@/lib/packages'
import type { SalePackage, Preset } from '@/lib/packages'
import { TelegramSettings } from '@/components/settings/TelegramSettings'

function useSavedTimeout() {
  const [saved, setSaved] = useState(false)
  const trigger = useCallback(() => {
    setSaved(true)
    setTimeout(() => setSaved(false), 3000)
  }, [])
  return { saved, trigger }
}

export function Settings() {
  const { data: settings, isLoading } = useSettings()
  const updateExchangeRate = useUpdateExchangeRate()
  const [exchangeRate, setExchangeRate] = useState('4050')
  const exchangeSaved = useSavedTimeout()

  const [taxEnabled, setTaxEnabled] = useState('false')
  const [taxRate, setTaxRate] = useState('0')
  const taxSave = useBatchUpdateSettings()
  const taxSaved = useSavedTimeout()

  const [invoicePrefix, setInvoicePrefix] = useState('INV')
  const invoiceSave = useUpdateSetting()
  const invoiceSaved = useSavedTimeout()

  const [lowStock, setLowStock] = useState('5')
  const lowStockSave = useUpdateSetting()
  const lowStockSaved = useSavedTimeout()

  const [batchScan, setBatchScan] = useState(false)
  const batchScanSave = useUpdateSetting()

  const [distanceUnitValue, setDistanceUnitValue] = useState('km')
  const distanceUnitSave = useUpdateSetting()
  const distanceUnitSaved = useSavedTimeout()

  const [shopName, setShopName] = useState('K&S Wheel-Tyre')
  const [shopAddress, setShopAddress] = useState('')
  const [shopPhone, setShopPhone] = useState('')
  const [shopEmail, setShopEmail] = useState('')
  const shopSave = useBatchUpdateSettings()
  const shopSaved = useSavedTimeout()

  useEffect(() => {
    if (settings) {
      setExchangeRate(settings.exchange_rate_usd_khr.toString())
      setTaxEnabled(settings.tax_enabled.toString())
      setTaxRate(settings.tax_rate_percent.toString())
      setInvoicePrefix(settings.invoice_prefix)
      setLowStock(settings.low_stock_threshold.toString())
      setBatchScan(!!settings.feature_batch_scan)
      setDistanceUnitValue(settings.distance_unit === 'mi' ? 'mi' : 'km')
      setShopName(settings.shop_name || 'K&S Wheel-Tyre')
      setShopAddress(settings.shop_address || '')
      setShopPhone(settings.shop_phone || '')
      setShopEmail(settings.shop_email || '')
    }
  }, [settings])

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading settings...</p>

  return (
    <div className="space-y-6">
      <PageHeader title="Settings" />

      <Section title="Shop Details" description="Business name and contact info shown on printed invoices and receipts">
        <div className="space-y-3 max-w-md">
          <div className="space-y-1.5">
            <label className="text-xs text-muted-foreground">Shop Name</label>
            <Input value={shopName} onChange={(e) => setShopName(e.target.value)} placeholder="K&S Wheel-Tyre" />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs text-muted-foreground">Address</label>
            <Input value={shopAddress} onChange={(e) => setShopAddress(e.target.value)} placeholder="Street, City, Cambodia" />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <div className="space-y-1.5">
              <label className="text-xs text-muted-foreground">Phone</label>
              <Input value={shopPhone} onChange={(e) => setShopPhone(e.target.value)} placeholder="+855 12 345 678" />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs text-muted-foreground">Email</label>
              <Input value={shopEmail} onChange={(e) => setShopEmail(e.target.value)} placeholder="shop@example.com" />
            </div>
          </div>
          <SaveButton
            onClick={() => {
              shopSave.mutate([
                { key: 'shop_name', value: shopName },
                { key: 'shop_address', value: shopAddress },
                { key: 'shop_phone', value: shopPhone },
                { key: 'shop_email', value: shopEmail },
              ], { onSuccess: shopSaved.trigger })
            }}
            loading={shopSave.isPending}
            saved={shopSaved.saved}
            error={shopSave.isError}
          />
        </div>
      </Section>

      <Section title="Exchange Rate" description="USD to KHR conversion rate">
        <div className="flex items-center gap-2">
          <div className="text-xs text-muted-foreground">1 USD =</div>
          <Input
            value={exchangeRate}
            onChange={(e) => setExchangeRate(e.target.value)}
            type="number"
            className="w-32"
          />
          <div className="text-xs text-muted-foreground">KHR</div>
          <SaveButton
            onClick={() => {
              updateExchangeRate.mutate(parseFloat(exchangeRate), { onSuccess: exchangeSaved.trigger })
            }}
            loading={updateExchangeRate.isPending}
            saved={exchangeSaved.saved}
            error={updateExchangeRate.isError}
          />
        </div>
      </Section>

      <Section title="Tax Settings" description="Configure tax calculation">
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <label className="text-sm">Enable Tax</label>
            <Select value={taxEnabled} onChange={(e) => setTaxEnabled(e.target.value)} className="w-24">
              <option value="true">Enabled</option>
              <option value="false">Disabled</option>
            </Select>
          </div>
          <div className="flex items-center gap-2">
            <Input
              value={taxRate}
              onChange={(e) => setTaxRate(e.target.value)}
              type="number"
              className="w-24"
              disabled={taxEnabled !== 'true'}
            />
            <span className="text-xs text-muted-foreground">%</span>
          </div>
          <SaveButton
            onClick={() => {
              taxSave.mutate([
                { key: 'tax_enabled', value: taxEnabled },
                { key: 'tax_rate_percent', value: taxRate },
              ], { onSuccess: taxSaved.trigger })
            }}
            loading={taxSave.isPending}
            saved={taxSaved.saved}
            error={taxSave.isError}
          />
        </div>
      </Section>

      <Section title="Distance Units" description="The unit used for odometer readings and service intervals (sale, invoices, vehicle profiles)">
        <div className="flex items-center gap-2">
          <Select value={distanceUnitValue} onChange={(e) => setDistanceUnitValue(e.target.value)} className="w-36">
            <option value="km">Kilometers (km)</option>
            <option value="mi">Miles (mi)</option>
          </Select>
          <SaveButton
            onClick={() => {
              distanceUnitSave.mutate({ key: 'distance_unit', value: distanceUnitValue }, { onSuccess: distanceUnitSaved.trigger })
            }}
            loading={distanceUnitSave.isPending}
            saved={distanceUnitSaved.saved}
            error={distanceUnitSave.isError}
          />
        </div>
      </Section>

      <Section title="Invoice Settings" description="Invoice number prefix">
        <div className="flex items-center gap-2">
          <Input
            value={invoicePrefix}
            onChange={(e) => setInvoicePrefix(e.target.value)}
            className="w-32"
          />
          <span className="text-xs text-muted-foreground">-YYYY-NNNN</span>
          <SaveButton
            onClick={() => {
              invoiceSave.mutate({ key: 'invoice_prefix', value: invoicePrefix }, { onSuccess: invoiceSaved.trigger })
            }}
            loading={invoiceSave.isPending}
            saved={invoiceSaved.saved}
            error={invoiceSave.isError}
          />
        </div>
      </Section>

      <Section title="Inventory Alerts" description="Low stock notification threshold">
        <div className="flex items-center gap-2">
          <Input
            value={lowStock}
            onChange={(e) => setLowStock(e.target.value)}
            type="number"
            className="w-24"
          />
          <span className="text-xs text-muted-foreground">Alert when stock is below this value</span>
          <SaveButton
            onClick={() => {
              lowStockSave.mutate({ key: 'low_stock_threshold', value: lowStock }, { onSuccess: lowStockSaved.trigger })
            }}
            loading={lowStockSave.isPending}
            saved={lowStockSaved.saved}
            error={lowStockSave.isError}
          />
        </div>
      </Section>

      <Section title="Batch Scan (Traceability)" description="Optional: mechanics scan a batch QR before fitting a part, so you know the exact batch on each car">
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={batchScan}
            onChange={(e) => {
              const on = e.target.checked
              setBatchScan(on)
              batchScanSave.mutate({ key: 'feature_batch_scan', value: on ? 'true' : 'false' })
            }}
            className="accent-primary"
          />
          Enable batch-scan install tracking
          <span className="text-xs text-muted-foreground">— adds a Scan Install screen and QR label printing. Never changes the sale or stock counts.</span>
        </label>
      </Section>

      <ServiceIntervalsEditor />

      <PackagesEditor initial={settings?.sale_packages} />

      <PresetsEditor
        title="Labor Presets"
        description="Quick-add labor buttons on the New Sale screen. Prices stay editable in the cart."
        settingKey="labor_presets"
        initial={settings?.labor_presets}
        defaults={LABOR_PRESETS}
        addLabel="Add labor"
      />

      <PresetsEditor
        title="Fee Presets"
        description="Quick-add fee buttons on the New Sale screen (disposal, shop supplies, etc.)."
        settingKey="fee_presets"
        initial={settings?.fee_presets}
        defaults={FEE_PRESETS}
        addLabel="Add fee"
      />

      <PaymentMethodsEditor initial={settings?.payment_methods} />

      <TelegramSettings />
    </div>
  )
}

function PackagesEditor({ initial }: { initial?: string }) {
  const save = useUpdateSetting()
  const saved = useSavedTimeout()
  const [pkgs, setPkgs] = useState<SalePackage[]>([])

  useEffect(() => {
    if (initial) {
      try {
        const p = JSON.parse(initial)
        if (Array.isArray(p)) { setPkgs(p); return }
      } catch { /* ignore */ }
    }
    setPkgs(PACKAGES)
  }, [initial])

  const update = (i: number, patch: Partial<SalePackage>) =>
    setPkgs((ps) => ps.map((p, idx) => (idx === i ? { ...p, ...patch } : p)))
  const updateAddon = (pi: number, ai: number, patch: any) =>
    setPkgs((ps) => ps.map((p, idx) => idx === pi ? { ...p, addons: p.addons.map((a, aj) => aj === ai ? { ...a, ...patch } : a) } : p))
  const addAddon = (pi: number) =>
    setPkgs((ps) => ps.map((p, idx) => idx === pi ? { ...p, addons: [...p.addons, { description: '', item_type: 'labor', unit_price_usd: 0, per_tire: true }] } : p))
  const removeAddon = (pi: number, ai: number) =>
    setPkgs((ps) => ps.map((p, idx) => idx === pi ? { ...p, addons: p.addons.filter((_, aj) => aj !== ai) } : p))
  const addPkg = () =>
    setPkgs((ps) => [...ps, { id: `pkg-${Date.now()}`, name: 'New package', blurb: '', tireType: 'passenger', addons: [] }])
  const removePkg = (i: number) => setPkgs((ps) => ps.filter((_, idx) => idx !== i))

  return (
    <Section title="Sale Packages" description="Bundles shown on the New Sale screen. Each expands into editable lines; the tire is chosen at the sale.">
      <div className="space-y-4">
        {pkgs.map((p, pi) => (
          <div key={p.id} className="rounded-md border p-3">
            <div className="flex flex-wrap items-center gap-2">
              <Input value={p.name} onChange={(e) => update(pi, { name: e.target.value })} placeholder="Package name" className="w-48" />
              <Select value={p.tireType || ''} onChange={(e) => update(pi, { tireType: e.target.value || undefined })} className="w-36">
                <option value="">Any tire</option>
                <option value="passenger">Passenger</option>
                <option value="truck">Truck</option>
                <option value="suv">SUV</option>
                <option value="motorcycle">Motorcycle</option>
              </Select>
              <Input value={p.blurb} onChange={(e) => update(pi, { blurb: e.target.value })} placeholder="Short description" className="flex-1 min-w-[10rem]" />
              <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={() => removePkg(pi)} aria-label="Remove package"><Trash2 className="h-4 w-4" /></Button>
            </div>
            <div className="mt-2 space-y-1.5">
              {p.addons.map((a, ai) => (
                <div key={ai} className="flex flex-wrap items-center gap-2">
                  <Input value={a.description} onChange={(e) => updateAddon(pi, ai, { description: e.target.value })} placeholder="Add-on (e.g. Mounting)" className="w-44" />
                  <Select value={a.item_type} onChange={(e) => updateAddon(pi, ai, { item_type: e.target.value })} className="w-28">
                    <option value="labor">Labor</option>
                    <option value="fee">Fee</option>
                  </Select>
                  <div className="flex items-center rounded-md border px-2">
                    <span className="text-xs text-muted-foreground">$</span>
                    <input type="number" step="0.01" min="0" value={a.unit_price_usd}
                      onChange={(e) => updateAddon(pi, ai, { unit_price_usd: parseFloat(e.target.value) || 0 })}
                      className="w-16 bg-transparent py-1.5 text-sm tabular-nums focus:outline-none" />
                  </div>
                  <label className="flex items-center gap-1 text-xs text-muted-foreground">
                    <input type="checkbox" checked={a.per_tire} onChange={(e) => updateAddon(pi, ai, { per_tire: e.target.checked })} className="accent-primary" />
                    per tire
                  </label>
                  <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={() => removeAddon(pi, ai)} aria-label="Remove add-on"><Trash2 className="h-3.5 w-3.5" /></Button>
                </div>
              ))}
              <Button variant="ghost" size="sm" className="gap-1" onClick={() => addAddon(pi)}><Plus className="h-3.5 w-3.5" /> Add-on</Button>
            </div>
          </div>
        ))}
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" className="gap-1" onClick={addPkg}><Plus className="h-4 w-4" /> Add package</Button>
          <SaveButton
            onClick={() => save.mutate({ key: 'sale_packages', value: JSON.stringify(pkgs) }, { onSuccess: saved.trigger })}
            loading={save.isPending}
            saved={saved.saved}
            error={save.isError}
          />
        </div>
      </div>
    </Section>
  )
}

function NumBox({ value, onChange, unit, min = '0' }: { value: string; onChange: (v: string) => void; unit: string; min?: string }) {
  return (
    <div className="flex items-center rounded-md border px-2">
      <input type="number" min={min} value={value} onChange={(e) => onChange(e.target.value)}
        className="w-full bg-transparent py-1.5 text-sm tabular-nums focus:outline-none" />
      <span className="text-xs text-muted-foreground whitespace-nowrap">{unit}</span>
    </div>
  )
}

function ServiceIntervalsEditor() {
  const { data: settings } = useSettings()
  const unit = distanceUnit(settings)
  const { data: intervals, isLoading } = useIntervalSettings()
  const save = useUpdateIntervalSettings()
  const saved = useSavedTimeout()

  const [oilKm, setOilKm] = useState('5000')
  const [oilDays, setOilDays] = useState('90')
  const [tireLifeKm, setTireLifeKm] = useState('40000')
  const [tireDays, setTireDays] = useState('')
  const [dueSoonDays, setDueSoonDays] = useState('14')
  const [fallbackKmPerDay, setFallbackKmPerDay] = useState('30')
  const [partRules, setPartRules] = useState<PartRule[]>([])

  useEffect(() => {
    if (intervals) {
      setOilKm(intervals.oil_interval_km.toString())
      setOilDays(intervals.oil_interval_days.toString())
      setTireLifeKm(intervals.tire_life_km.toString())
      setTireDays(intervals.tire_interval_days ? intervals.tire_interval_days.toString() : '')
      setDueSoonDays(intervals.due_soon_days.toString())
      setFallbackKmPerDay(intervals.fallback_km_per_day.toString())
      setPartRules(intervals.part_rules || [])
    }
  }, [intervals])

  if (isLoading) return null

  const updateRule = (i: number, patch: Partial<PartRule>) =>
    setPartRules((rs) => rs.map((r, idx) => (idx === i ? { ...r, ...patch } : r)))
  const removeRule = (i: number) => setPartRules((rs) => rs.filter((_, idx) => idx !== i))
  const addRule = () => setPartRules((rs) => [...rs, { part_key: '', label: '', km: null, days: null }])

  return (
    <Section
      title="Service Reminder Intervals"
      description="How the shop flags customers as due. Tires are judged by each tire's own rated life (set on the product); here you set the oil and per-part intervals."
    >
      <div className="space-y-5 max-w-xl">
        <div>
          <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Oil change — due at whichever comes first</p>
          <div className="grid grid-cols-2 gap-2">
            <NumBox value={oilKm} onChange={setOilKm} unit={unit} />
            <NumBox value={oilDays} onChange={setOilDays} unit="days" />
          </div>
        </div>

        <div>
          <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Tires — due at whichever comes first</p>
          <div className="grid grid-cols-2 gap-2">
            <NumBox value={tireLifeKm} onChange={setTireLifeKm} unit={`${unit} default life`} />
            <NumBox value={tireDays} onChange={setTireDays} unit="days (optional)" />
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            {unit} life is a fallback used when a sold tire has no rated life set on its product. Days is optional — set it to also flag tires by age (a firm calendar date); leave blank to judge by wear-to-{unit} only.
          </p>
        </div>

        <div>
          <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Parts</p>
          <p className="mb-2 text-xs text-muted-foreground">
            Each part is due at whichever limit is hit first, measured from the last time it was logged as replaced. Leave a box blank to skip that limit.
          </p>
          <div className="space-y-2">
            <div className="grid grid-cols-[1fr_5rem_5rem_2rem] items-center gap-2 text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
              <span>Part key</span><span>{unit}</span><span>days</span><span />
            </div>
            {partRules.map((r, i) => (
              <div key={i} className="grid grid-cols-[1fr_5rem_5rem_2rem] items-center gap-2">
                <Input value={r.part_key} onChange={(e) => updateRule(i, { part_key: e.target.value })} placeholder="brakes_front" className="font-mono text-xs" />
                <Input value={r.km?.toString() ?? ''} onChange={(e) => updateRule(i, { km: e.target.value ? parseInt(e.target.value.replace(/[^\d]/g, '')) : null })} inputMode="numeric" placeholder="—" />
                <Input value={r.days?.toString() ?? ''} onChange={(e) => updateRule(i, { days: e.target.value ? parseInt(e.target.value.replace(/[^\d]/g, '')) : null })} inputMode="numeric" placeholder="—" />
                <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={() => removeRule(i)} aria-label="Remove"><Trash2 className="h-4 w-4" /></Button>
              </div>
            ))}
          </div>
          <Button variant="outline" size="sm" className="mt-2 gap-1" onClick={addRule}><Plus className="h-4 w-4" /> Add part</Button>
          <p className="mt-1.5 text-[11px] text-muted-foreground">
            Part keys match the inspection diagram: brakes_front, brakes_rear, battery, lights, wipers, suspension, air_filter, coolant, belts.
          </p>
        </div>

        <div>
          <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Estimation</p>
          <div className="grid grid-cols-2 gap-2">
            <NumBox value={fallbackKmPerDay} onChange={setFallbackKmPerDay} unit={`${unit}/day fallback`} min="1" />
            <NumBox value={dueSoonDays} onChange={setDueSoonDays} unit={'"due soon" window (days)'} />
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            The {unit}/day fallback estimates today's odometer for a vehicle with too little visit history to project from.
          </p>
        </div>

        <SaveButton
          onClick={() => {
            save.mutate({
              oil_interval_km: parseInt(oilKm) || 0,
              oil_interval_days: parseInt(oilDays) || 0,
              tire_life_km: parseInt(tireLifeKm) || 0,
              tire_interval_days: parseInt(tireDays) || 0,
              fallback_km_per_day: parseFloat(fallbackKmPerDay) || 0,
              due_soon_days: parseInt(dueSoonDays) || 0,
              part_rules: partRules.filter((r) => r.part_key.trim()),
            }, { onSuccess: saved.trigger })
          }}
          loading={save.isPending}
          saved={saved.saved}
          error={save.isError}
        />
      </div>
    </Section>
  )
}

function PresetsEditor({ title, description, settingKey, initial, defaults, addLabel }: {
  title: string
  description: string
  settingKey: string
  initial?: string
  defaults: Preset[]
  addLabel: string
}) {
  const save = useUpdateSetting()
  const saved = useSavedTimeout()
  const [rows, setRows] = useState<Preset[]>([])

  useEffect(() => { setRows(parseArraySetting<Preset>(initial, defaults)) }, [initial, defaults])

  const update = (i: number, patch: Partial<Preset>) =>
    setRows((rs) => rs.map((r, idx) => (idx === i ? { ...r, ...patch } : r)))
  const remove = (i: number) => setRows((rs) => rs.filter((_, idx) => idx !== i))
  const add = () => setRows((rs) => [...rs, { description: '', unit_price_usd: 0 }])

  return (
    <Section title={title} description={description}>
      <div className="space-y-2">
        {rows.map((r, i) => (
          <div key={i} className="flex flex-wrap items-center gap-2">
            <Input value={r.description} onChange={(e) => update(i, { description: e.target.value })} placeholder="Description" className="w-56" />
            <div className="flex items-center rounded-md border px-2">
              <span className="text-xs text-muted-foreground">$</span>
              <input type="number" step="0.01" min="0" value={r.unit_price_usd}
                onChange={(e) => update(i, { unit_price_usd: parseFloat(e.target.value) || 0 })}
                className="w-20 bg-transparent py-1.5 text-sm tabular-nums focus:outline-none" />
            </div>
            <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={() => remove(i)} aria-label="Remove"><Trash2 className="h-4 w-4" /></Button>
          </div>
        ))}
        <div className="flex items-center gap-2 pt-1">
          <Button variant="outline" size="sm" className="gap-1" onClick={add}><Plus className="h-4 w-4" /> {addLabel}</Button>
          <SaveButton
            onClick={() => save.mutate({ key: settingKey, value: JSON.stringify(rows.filter((r) => r.description.trim())) }, { onSuccess: saved.trigger })}
            loading={save.isPending}
            saved={saved.saved}
            error={save.isError}
          />
        </div>
      </div>
    </Section>
  )
}

function PaymentMethodsEditor({ initial }: { initial?: string }) {
  const save = useUpdateSetting()
  const saved = useSavedTimeout()
  const [methods, setMethods] = useState<string[]>([])

  useEffect(() => { setMethods(parseArraySetting<string>(initial, DEFAULT_PAYMENT_METHODS)) }, [initial])

  const update = (i: number, value: string) => setMethods((ms) => ms.map((m, idx) => (idx === i ? value : m)))
  const remove = (i: number) => setMethods((ms) => ms.filter((_, idx) => idx !== i))
  const add = () => setMethods((ms) => [...ms, ''])

  return (
    <Section title="Payment Methods" description="Options shown when taking payment on a sale or invoice (e.g. Cash, ABA, Wing).">
      <div className="space-y-2">
        {methods.map((m, i) => (
          <div key={i} className="flex items-center gap-2">
            <Input value={m} onChange={(e) => update(i, e.target.value)} placeholder="Method name" className="w-56" />
            <Button variant="ghost" size="icon" className="hover:text-destructive" onClick={() => remove(i)} aria-label="Remove"><Trash2 className="h-4 w-4" /></Button>
          </div>
        ))}
        <div className="flex items-center gap-2 pt-1">
          <Button variant="outline" size="sm" className="gap-1" onClick={add}><Plus className="h-4 w-4" /> Add method</Button>
          <SaveButton
            onClick={() => {
              const cleaned = methods.map((m) => m.trim()).filter(Boolean)
              save.mutate({ key: 'payment_methods', value: JSON.stringify(cleaned.length ? cleaned : DEFAULT_PAYMENT_METHODS) }, { onSuccess: saved.trigger })
            }}
            loading={save.isPending}
            saved={saved.saved}
            error={save.isError}
          />
        </div>
      </div>
    </Section>
  )
}

function Section({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return (
    <div className="bg-card rounded-lg p-5 shadow-sm">
      <div className="mb-2">
        <h2 className="text-sm font-medium">{title}</h2>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      {children}
    </div>
  )
}

function SaveButton({ onClick, loading, saved, error }: {
  onClick: () => void
  loading: boolean
  saved: boolean
  error: boolean
}) {
  return (
    <div className="flex items-center gap-2">
      <Button size="sm" onClick={onClick} disabled={loading}>
        {loading ? 'Saving...' : 'Save'}
      </Button>
      {saved && <span className="text-xs text-primary">Saved!</span>}
      {error && <span className="text-xs text-destructive">Error</span>}
    </div>
  )
}
