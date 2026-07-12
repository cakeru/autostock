export function Badge({ className, children, ...props }: React.HTMLAttributes<HTMLSpanElement>) {
  return <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${className || ''}`} {...props}>{children}</span>
}
