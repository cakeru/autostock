import { Disc3, Cog, Wrench, Droplet } from 'lucide-react'
import { cn } from '@/lib/utils'
import { imageSrc } from '@/utils/imageUrl'
import type { Product } from '@/types/product'

const typeConfig: Record<string, { Icon: typeof Disc3; tint: string }> = {
  tire: { Icon: Disc3, tint: 'bg-primary/10 text-primary' },
  part: { Icon: Cog, tint: 'bg-accent/10 text-accent' },
  labor: { Icon: Wrench, tint: 'bg-muted text-muted-foreground' },
  consumable: { Icon: Droplet, tint: 'bg-primary/10 text-primary' },
}

// Product image tile: renders the real photo when present, otherwise a tinted
// type-icon placeholder so every row keeps a consistent visual anchor.
export function ProductThumb({ product, className }: { product: Product; className?: string }) {
  const { Icon, tint } = typeConfig[product.type] || typeConfig.part
  const src = imageSrc(product.image_url)

  if (src) {
    return (
      <img
        src={src}
        alt={product.name}
        loading="lazy"
        className={cn('rounded-md object-cover bg-muted', className)}
      />
    )
  }

  return (
    <div className={cn('grid place-items-center rounded-md', tint, className)}>
      <Icon className="h-1/2 w-1/2" strokeWidth={1.75} />
    </div>
  )
}

// Type-appropriate secondary spec line for the list view.
export function productSpec(product: Product): string {
  if (product.type === 'tire') {
    return [product.tire_size, product.tire_brand].filter(Boolean).join(' · ')
  }
  return product.category || ''
}
