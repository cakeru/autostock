import type { ServiceJobDetail } from '@/types/servicejob'

const usd = (n: number) => `$${n.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`

export function PrintQuote({ job }: { job: ServiceJobDetail }) {
  return (
    <div className="hidden print:block max-w-[80mm] mx-auto text-black text-[12px] leading-snug font-mono">
      <div className="text-center mb-3">
        <p className="text-[16px] font-bold">AutoStock Garage</p>
        <p>Phnom Penh, Cambodia</p>
        <p>+855 12 345 678</p>
      </div>

      <div className="border-t border-b border-black border-dashed py-1 mb-2">
        <div className="flex justify-between"><span>Quotation</span><span>{job.job_number}</span></div>
        <div className="flex justify-between">
          <span>Date</span>
          <span>{new Date(job.created_at).toLocaleString()}</span>
        </div>
        {job.customer_name && <div className="flex justify-between"><span>Customer</span><span>{job.customer_name}</span></div>}
        {job.plate_number && <div className="flex justify-between"><span>Vehicle</span><span>{job.plate_number}</span></div>}
        {job.vehicle_info && <div className="flex justify-between"><span>Model</span><span>{job.vehicle_info}</span></div>}
      </div>

      <p className="font-bold mb-1">Description: {job.description}</p>

      <table className="w-full mb-2">
        <thead>
          <tr className="border-b border-black border-dashed">
            <th className="text-left">Item</th>
            <th className="text-right">Price</th>
          </tr>
        </thead>
        <tbody>
          {job.items.map((item) => (
            <tr key={item.id} className="align-top">
              <td className="pr-1">
                {item.product_name}
                <br />
                <span className="text-[10px]">{item.quantity} x {usd(item.unit_price)}</span>
              </td>
              <td className="text-right whitespace-nowrap">{usd(item.total_price)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="border-t border-black border-dashed pt-1">
        <div className="flex justify-between text-[14px] font-bold"><span>TOTAL</span><span>{usd(job.total_amount)}</span></div>
      </div>

      {job.diagnosis && <p className="mt-2 text-[10px]">Diagnosis: {job.diagnosis}</p>}
      {job.work_performed && <p className="text-[10px]">Work: {job.work_performed}</p>}

      <p className="text-center mt-3">This quote is valid for 7 days</p>
    </div>
  )
}
