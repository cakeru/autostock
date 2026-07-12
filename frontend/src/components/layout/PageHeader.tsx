import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'

interface PageHeaderProps {
  title: string
  backTo?: string | number
  breadcrumb?: string
  badges?: ReactNode
  actions?: ReactNode
  subtitle?: string
}

export function PageHeader({ title, backTo, breadcrumb, badges, actions, subtitle }: PageHeaderProps) {
  const navigate = useNavigate()

  const handleBack = () => {
    if (typeof backTo === 'number') {
      navigate(backTo)
    } else if (backTo) {
      navigate(backTo)
    }
  }

  return (
    <div className="pb-2">
      {breadcrumb && (
        <p className="text-xs text-muted-foreground mb-1">{breadcrumb}</p>
      )}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
        <div className="flex min-w-0 items-center gap-2 sm:gap-3">
          {backTo !== undefined && (
            <Button variant="ghost" size="sm" onClick={handleBack} className="-ml-2 flex-shrink-0">
              ← Back
            </Button>
          )}
          <h1 className="truncate text-2xl font-semibold">{title}</h1>
        </div>
        {badges && (
          <div className="flex flex-wrap items-center gap-1.5">
            {badges}
          </div>
        )}
        {subtitle && (
          <p className="text-sm text-muted-foreground sm:ml-2">{subtitle}</p>
        )}
        {actions && (
          <div className="flex flex-wrap items-center gap-2 sm:ml-auto">
            {actions}
          </div>
        )}
      </div>
    </div>
  )
}
