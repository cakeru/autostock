import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'destructive' | 'ghost' | 'outline'
  size?: 'default' | 'sm' | 'lg' | 'icon'
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'default', ...props }, ref) => {
    return (
      <button
        className={cn(
          'inline-flex items-center justify-center rounded-md text-sm font-medium transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
          variant === 'default' && 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm hover:shadow-md hover:-translate-y-0.5 active:translate-y-0 active:shadow-sm',
          variant === 'destructive' && 'bg-destructive text-destructive-foreground hover:bg-destructive/90 shadow-sm',
          variant === 'ghost' && 'hover:bg-muted hover:text-foreground',
          variant === 'outline' && 'border border-input bg-background hover:bg-muted hover:shadow-sm',
          size === 'default' && 'h-9 px-3',
          size === 'sm' && 'h-8 px-2 text-xs',
          size === 'lg' && 'h-10 px-4',
          size === 'icon' && 'h-9 w-9',
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = 'Button'

export { Button }
