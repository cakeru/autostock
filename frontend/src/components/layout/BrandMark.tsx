import { cn } from '@/lib/utils'

// K&S Wheel-Tyre monogram: a gold-ringed navy disc with a tire-tread notched
// outer edge, echoing the shop's business-card logo without the full artwork.
export function BrandMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 48 48" className={cn('flex-shrink-0', className)} role="img" aria-label="K&S Wheel-Tyre">
      <circle cx="24" cy="24" r="23" fill="var(--color-brand-navy-deep)" />
      {Array.from({ length: 24 }).map((_, i) => {
        const a = (i / 24) * Math.PI * 2
        return (
          <rect
            key={i}
            x="23"
            y="1.5"
            width="2"
            height="3.5"
            rx="0.6"
            fill="var(--color-brand-gold)"
            transform={`rotate(${(a * 180) / Math.PI} 24 24)`}
          />
        )
      })}
      <circle cx="24" cy="24" r="17.5" fill="none" stroke="var(--color-brand-gold)" strokeWidth="1.5" />
      <text
        x="24"
        y="24"
        textAnchor="middle"
        dominantBaseline="central"
        fontFamily="'Plus Jakarta Sans', sans-serif"
        fontWeight="700"
        fontSize="17"
        letterSpacing="-0.5"
        fill="var(--color-brand-gold)"
      >
        KS
      </text>
    </svg>
  )
}
