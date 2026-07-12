import type { ReactNode, ThHTMLAttributes } from 'react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

// Shared shell for list pages: filter toolbar and pagination footer attached
// to the same surface as the table, so the page reads as one object.
export function TableCard({ toolbar, footer, children }: {
  toolbar?: ReactNode
  footer?: ReactNode
  children: ReactNode
}) {
  return (
    <div className="bg-card rounded-lg shadow-sm">
      {toolbar && (
        <div className="flex flex-wrap items-center gap-2.5 border-b p-3">{toolbar}</div>
      )}
      <div className="max-h-[65vh] overflow-y-auto overflow-x-auto">{children}</div>
      {footer}
    </div>
  )
}

export function Th({ className, children, ...props }: ThHTMLAttributes<HTMLTableCellElement>) {
  return (
    <th
      className={cn('px-4 py-2.5 text-left text-xs font-medium text-muted-foreground whitespace-nowrap', className)}
      {...props}
    >
      {children}
    </th>
  )
}

export function ActionsTh() {
  return (
    <th className="px-4 py-2.5 text-right">
      <span className="sr-only">Actions</span>
    </th>
  )
}

export function TableFooter({ total, page, totalPages, onPage, noun = 'items' }: {
  total: number
  page: number
  totalPages: number
  onPage: (page: number) => void
  noun?: string
}) {
  return (
    <div className="flex items-center justify-between border-t px-4 py-2.5 text-sm">
      <span className="text-muted-foreground tabular-nums">
        {total} {noun}
        {totalPages > 1 && <span> · page {page} of {totalPages}</span>}
      </span>
      {totalPages > 1 && (
        <div className="flex gap-1">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => onPage(page - 1)}>
            Previous
          </Button>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => onPage(page + 1)}>
            Next
          </Button>
        </div>
      )}
    </div>
  )
}
