import type { InvoiceDetail } from '@/types/invoice'
import type { Settings } from '@/services/settings'
import { BrandMark } from '@/components/layout/BrandMark'

const usd = (n: number) => `$${n.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
const khr = (n: number) => `${Math.round(n).toLocaleString()}៛`

export type PrintFormat = 'thermal' | 'a4' | 'classic'

interface Shop {
  name: string
  address?: string
  phone?: string
  email?: string
}

function shopFrom(settings?: Settings): Shop {
  return {
    name: settings?.shop_name || 'K&S Wheel-Tyre',
    address: settings?.shop_address,
    phone: settings?.shop_phone,
    email: settings?.shop_email,
  }
}

const issuedDate = (invoice: InvoiceDetail) =>
  invoice.issued_at ? new Date(invoice.issued_at) : new Date(invoice.created_at)

/** Chooses which printable layout to mount. Rendered off-screen and revealed only by print media. */
export function PrintReceipt({
  invoice,
  settings,
  format,
  capture,
}: {
  invoice: InvoiceDetail
  settings?: Settings
  format: PrintFormat
  // capture: render as a normal on-screen block (not print-only) so it can be
  // screenshotted to PDF for "Send to Telegram".
  capture?: boolean
}) {
  const shop = shopFrom(settings)
  if (format === 'a4') return <InvoiceA4 invoice={invoice} shop={shop} capture={capture} />
  if (format === 'classic') return <ClassicInvoice invoice={invoice} />
  return <ThermalReceipt invoice={invoice} shop={shop} />
}

/* ------------------------------------------------------------------ */
/* 80mm thermal receipt                                                */
/* ------------------------------------------------------------------ */

function ThermalReceipt({ invoice, shop }: { invoice: InvoiceDetail; shop: Shop }) {
  return (
    <div className="hidden print:block mx-auto w-[72mm] text-black text-[12px] leading-snug font-mono">
      <style>{`@page { size: 80mm auto; margin: 4mm; }`}</style>

      <div className="mb-3 text-center">
        <p className="text-[15px] font-bold uppercase tracking-wide">{shop.name}</p>
        {shop.address && <p>{shop.address}</p>}
        {shop.phone && <p>{shop.phone}</p>}
      </div>

      <div className="mb-2 border-y border-dashed border-black py-1">
        <div className="flex justify-between"><span>Receipt</span><span>{invoice.invoice_number}</span></div>
        <div className="flex justify-between"><span>Date</span><span>{issuedDate(invoice).toLocaleString()}</span></div>
        {invoice.customer_name && <div className="flex justify-between"><span>Customer</span><span>{invoice.customer_name}</span></div>}
        {invoice.plate_number && <div className="flex justify-between"><span>Vehicle</span><span>{invoice.plate_number}</span></div>}
        {invoice.mileage != null && <div className="flex justify-between"><span>Odometer</span><span>{invoice.mileage.toLocaleString()} km</span></div>}
        {invoice.job_number && <div className="flex justify-between"><span>Job</span><span>{invoice.job_number}</span></div>}
      </div>

      <table className="mb-2 w-full">
        <tbody>
          {invoice.items.map((item) => (
            <tr key={item.id} className="align-top">
              <td className="pr-1">
                {item.description}
                <br />
                <span>{item.quantity} x {usd(item.unit_price_usd)}</span>
              </td>
              <td className="whitespace-nowrap text-right">{usd(item.total_usd)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="border-t border-dashed border-black pt-1">
        <div className="flex justify-between"><span>Subtotal</span><span>{usd(invoice.subtotal)}</span></div>
        {invoice.discount > 0 && <div className="flex justify-between"><span>Discount</span><span>-{usd(invoice.discount)}</span></div>}
        {invoice.tax_amount > 0 && <div className="flex justify-between"><span>Tax ({invoice.tax_rate}%)</span><span>{usd(invoice.tax_amount)}</span></div>}
        <div className="mt-1 flex justify-between text-[14px] font-bold"><span>TOTAL USD</span><span>{usd(invoice.total_usd)}</span></div>
        <div className="flex justify-between"><span>TOTAL KHR</span><span>{khr(invoice.total_khr)}</span></div>
        <div className="mt-1 flex justify-between">
          <span>Payment</span>
          <span>{invoice.payment_status}{invoice.payment_method ? ` (${invoice.payment_method})` : ''}</span>
        </div>
        {invoice.paid_amount > 0 && <div className="flex justify-between"><span>Paid</span><span>{usd(invoice.paid_amount)}</span></div>}
      </div>

      {invoice.status === 'voided' && (
        <p className="mt-2 text-center text-[14px] font-bold">*** VOIDED ***</p>
      )}

      <p className="mt-3 text-center">Thank you! សូមអរគុណ!</p>
    </div>
  )
}

/* ------------------------------------------------------------------ */
/* Classic bilingual invoice — matches the shop's existing paper form  */
/* ------------------------------------------------------------------ */

function shortDate(d: Date) {
  const dd = String(d.getDate()).padStart(2, '0')
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const yy = String(d.getFullYear()).slice(-2)
  return `${dd}/${mm}/${yy}`
}

function MetaLine({ kh, value, align = 'left' }: { kh: string; value?: string; align?: 'left' | 'right' }) {
  return (
    <p className={`text-[12px] ${align === 'right' ? 'text-right' : ''}`}>
      <span className="font-medium">{kh}</span> {value}
    </p>
  )
}

function ClassicInvoice({ invoice }: { invoice: InvoiceDetail }) {
  const MIN_ROWS = 10
  const rows: Array<{ description: string; quantity: number; unit_price_usd: number; total_usd: number } | null> = [
    ...invoice.items,
  ]
  while (rows.length < MIN_ROWS) rows.push(null)

  // "Total in $" is the pre-VAT subtotal (after discount); the grand totals
  // below are what the customer actually owes — matches the reference form,
  // where both lines read the same number whenever there's no tax/discount.
  const totalIn = invoice.subtotal - invoice.discount

  return (
    <div className="kh-invoice hidden print:block w-full text-black text-[12px]">
      {/* @page margin:0 suppresses the browser-injected header/footer (date, page
          title, URL); the page margin is applied as padding instead. print-color-adjust
          forces the black header + zebra-striped rows to actually render on paper. */}
      <style>{`
        @page { size: A4; margin: 0; }
        .kh-invoice { padding: 15mm; box-sizing: border-box; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
      `}</style>

      <h1 className="mb-5 text-center text-[22px] font-bold">Invoice</h1>

      <div className="mb-3 grid grid-cols-2 gap-x-8">
        <div className="space-y-1">
          <MetaLine kh="អ្នកលក់ៈ" value={invoice.created_by_name || 'Admin'} />
          <MetaLine kh="អតិថិជនៈ" value={invoice.customer_name} />
          <MetaLine kh="ស្លាកលេខៈ" value={invoice.plate_number} />
          <MetaLine kh="ម៉ាកយានយន្តៈ" value={invoice.vehicle_info} />
        </div>
        <div className="space-y-1">
          <MetaLine kh="កាលបរិច្ឆេទៈ" value={shortDate(issuedDate(invoice))} align="right" />
          <MetaLine kh="អត្រាៈ" value={`$1 = ${Math.round(invoice.exchange_rate).toLocaleString()} ៛`} align="right" />
          <MetaLine kh="ទូរស័ព្ទអតិថិជនៈ" value={invoice.customer_phone} align="right" />
          <MetaLine kh="គីឡូម៉ែត្រប្រើប្រាស់ៈ" value={invoice.mileage != null ? `${invoice.mileage.toLocaleString()} km` : undefined} align="right" />
        </div>
      </div>

      <table className="w-full border-collapse border border-black">
        <thead>
          <tr className="bg-black text-white">
            <th className="border border-black px-2 py-1 w-[8%]">
              ល.រ<br /><span className="text-[9px] font-normal italic">N</span>
            </th>
            <th className="border border-black px-2 py-1 text-left">
              មុខទំនិញ<br /><span className="text-[9px] font-normal italic">Item</span>
            </th>
            <th className="border border-black px-2 py-1 w-[12%]">
              បរិមាណ<br /><span className="text-[9px] font-normal italic">QTY</span>
            </th>
            <th className="border border-black px-2 py-1 w-[18%]">
              តម្លៃ<br /><span className="text-[9px] font-normal italic">Price</span>
            </th>
            <th className="border border-black px-2 py-1 w-[18%]">
              សរុប<br /><span className="text-[9px] font-normal italic">Total</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className={i % 2 === 1 ? 'bg-[#e9e9e9]' : ''}>
              <td className="border border-black px-2 py-1.5 text-center">{i + 1}</td>
              <td className="border border-black px-2 py-1.5 font-medium">{row?.description}</td>
              <td className="border border-black px-2 py-1.5 text-center">{row?.quantity}</td>
              <td className="border border-black px-2 py-1.5 text-right">
                {row ? `$ ${row.unit_price_usd.toFixed(2)}` : ''}
              </td>
              <td className="border border-black px-2 py-1.5 text-right">
                {row ? `$ ${row.total_usd.toFixed(2)}` : '$ -'}
              </td>
            </tr>
          ))}
          <tr>
            <td colSpan={4} className="border border-black px-2 py-1.5 text-right font-semibold">Total in $</td>
            <td className="border border-black px-2 py-1.5 text-right">$ {totalIn.toFixed(2)}</td>
          </tr>
          <tr>
            <td colSpan={4} className="border border-black px-2 py-1.5 text-right font-semibold">Grand Total (VAT Included) ($)</td>
            <td className="border border-black px-2 py-1.5 text-right font-semibold">$ {invoice.total_usd.toFixed(2)}</td>
          </tr>
          <tr>
            <td colSpan={4} className="border border-black px-2 py-1.5 text-right font-semibold">Grand Total (VAT Included) (៛)</td>
            <td className="border border-black px-2 py-1.5 text-right font-semibold">{khr(invoice.total_khr)}</td>
          </tr>
        </tbody>
      </table>

      {invoice.status === 'voided' && (
        <p className="mt-4 text-center text-[16px] font-bold">*** VOIDED ***</p>
      )}

      <p className="mt-10 text-center text-[14px] font-bold italic">Thank you!</p>
    </div>
  )
}

/* ------------------------------------------------------------------ */
/* A4 branded invoice                                                  */
/* ------------------------------------------------------------------ */

const NAVY = 'var(--color-brand-navy)'
const NAVY_DEEP = 'var(--color-brand-navy-deep)'
const GOLD = 'var(--color-brand-gold)'
// Force background/borders to render when printing.
const exact = { WebkitPrintColorAdjust: 'exact', printColorAdjust: 'exact' } as const

function InvoiceA4({ invoice, shop, capture }: { invoice: InvoiceDetail; shop: Shop; capture?: boolean }) {
  const voided = invoice.status === 'voided'
  return (
    <div className={`relative mx-auto w-full max-w-[190mm] text-[11px] text-black ${capture ? 'block' : 'hidden print:block'}`} style={exact}>
      <style>{`@page { size: A4; margin: 14mm; } @media print { .a4-invoice, .a4-invoice * { -webkit-print-color-adjust: exact; print-color-adjust: exact; } }`}</style>

      <div className="a4-invoice">
        {/* Header band */}
        <div
          className="flex items-center justify-between rounded-t-lg px-6 py-5 text-white"
          style={{ background: `linear-gradient(120deg, ${NAVY_DEEP} 0%, ${NAVY} 100%)`, ...exact }}
        >
          <div className="flex items-center gap-3">
            <BrandMark className="h-14 w-14" />
            <div className="leading-tight">
              <p className="text-[18px] font-bold tracking-wide">{shop.name}</p>
              <p className="text-[9px] uppercase tracking-[0.25em] text-white/50">Wheel &amp; Tyre Service</p>
            </div>
          </div>
          <div className="text-right">
            <p className="text-[24px] font-bold tracking-widest" style={{ color: GOLD, ...exact }}>INVOICE</p>
            <p className="text-[11px] text-white/80">{invoice.invoice_number}</p>
          </div>
        </div>
        <div className="h-1" style={{ background: GOLD, ...exact }} />

        {/* Meta: bill-to + details */}
        <div className="grid grid-cols-2 gap-6 px-6 py-5">
          <div>
            <p className="mb-1 text-[9px] font-semibold uppercase tracking-[0.15em]" style={{ color: NAVY }}>Bill To</p>
            <p className="text-[13px] font-semibold">{invoice.customer_name || 'Walk-in customer'}</p>
            {invoice.customer_phone && <p>{invoice.customer_phone}</p>}
            {invoice.customer_address && <p className="text-black/70">{invoice.customer_address}</p>}
            {invoice.vehicle_info && (
              <p className="mt-1 text-black/70">
                {invoice.vehicle_info}{invoice.plate_number ? ` · ${invoice.plate_number}` : ''}
              </p>
            )}
            {invoice.mileage != null && <p className="text-black/70">Odometer: {invoice.mileage.toLocaleString()} km</p>}
            {invoice.job_number && <p className="text-black/70">Job: {invoice.job_number}</p>}
          </div>
          <div className="text-right">
            <MetaRow label="Date" value={issuedDate(invoice).toLocaleDateString()} />
            <MetaRow
              label="Payment"
              value={`${invoice.payment_status}${invoice.payment_method ? ` (${invoice.payment_method})` : ''}`}
            />
            {invoice.paid_amount > 0 && <MetaRow label="Paid" value={usd(invoice.paid_amount)} />}
          </div>
        </div>

        {/* Items */}
        <div className="px-6">
          <table className="w-full border-collapse">
            <thead>
              <tr className="text-white" style={{ background: NAVY, ...exact }}>
                <th className="px-3 py-2 text-left text-[9px] font-semibold uppercase tracking-wider">Description</th>
                <th className="px-3 py-2 text-center text-[9px] font-semibold uppercase tracking-wider">Qty</th>
                <th className="px-3 py-2 text-right text-[9px] font-semibold uppercase tracking-wider">Unit</th>
                <th className="px-3 py-2 text-right text-[9px] font-semibold uppercase tracking-wider">Amount</th>
              </tr>
            </thead>
            <tbody>
              {invoice.items.map((item, i) => (
                <tr key={item.id} style={i % 2 === 1 ? { background: 'rgba(0,0,0,0.03)', ...exact } : undefined}>
                  <td className="px-3 py-2 align-top">{item.description}</td>
                  <td className="px-3 py-2 text-center align-top tabular-nums">{item.quantity}</td>
                  <td className="px-3 py-2 text-right align-top tabular-nums">{usd(item.unit_price_usd)}</td>
                  <td className="px-3 py-2 text-right align-top tabular-nums">{usd(item.total_usd)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Totals */}
        <div className="flex justify-end px-6 py-4">
          <div className="w-[70mm] space-y-1">
            <TotalRow label="Subtotal" value={usd(invoice.subtotal)} />
            {invoice.discount > 0 && <TotalRow label="Discount" value={`-${usd(invoice.discount)}`} />}
            {invoice.tax_amount > 0 && <TotalRow label={`Tax (${invoice.tax_rate}%)`} value={usd(invoice.tax_amount)} />}
            <div
              className="mt-1 flex items-center justify-between rounded px-3 py-2 text-white"
              style={{ background: NAVY, ...exact }}
            >
              <span className="text-[10px] font-semibold uppercase tracking-wider">Total USD</span>
              <span className="text-[15px] font-bold tabular-nums">{usd(invoice.total_usd)}</span>
            </div>
            <div className="flex justify-between px-3 text-[10px] text-black/70">
              <span>Total KHR</span>
              <span className="tabular-nums">{khr(invoice.total_khr)}</span>
            </div>
          </div>
        </div>

        {invoice.notes && (
          <div className="px-6 pb-2">
            <p className="text-[9px] font-semibold uppercase tracking-[0.15em]" style={{ color: NAVY }}>Notes</p>
            <p className="text-black/70">{invoice.notes}</p>
          </div>
        )}

        {/* Footer */}
        <div className="mt-2 border-t px-6 py-4 text-center" style={{ borderColor: GOLD, ...exact }}>
          <p className="text-[12px] font-semibold" style={{ color: NAVY }}>Thank you! សូមអរគុណ!</p>
          <p className="mt-1 text-[9px] text-black/60">
            {[shop.address, shop.phone, shop.email].filter(Boolean).join('  ·  ')}
          </p>
        </div>
      </div>

      {voided && (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
          <p className="rotate-[-24deg] text-[80px] font-bold uppercase tracking-widest" style={{ color: 'rgba(200,0,0,0.18)', ...exact }}>
            Voided
          </p>
        </div>
      )}
    </div>
  )
}

function MetaRow({ label, value }: { label: string; value: string }) {
  return (
    <p className="text-[11px]">
      <span className="text-black/50">{label}: </span>
      <span className="font-medium">{value}</span>
    </p>
  )
}

function TotalRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between px-3 text-[11px]">
      <span className="text-black/70">{label}</span>
      <span className="tabular-nums">{value}</span>
    </div>
  )
}
