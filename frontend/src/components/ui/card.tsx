import { cn } from '@/lib/utils'

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  title?: string
}

export function Card({ className, title, children, ...props }: CardProps) {
  return (
    <div className={cn('bg-card rounded-lg p-5 shadow-sm', className)} {...props}>
      {title && <p className="text-sm font-semibold mb-3">{title}</p>}
      {children}
    </div>
  )
}
