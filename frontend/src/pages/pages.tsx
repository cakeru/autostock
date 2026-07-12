export function Placeholder({ title }: { title: string }) {
  return (
    <div className="space-y-4">
      <h1 className="text-lg font-semibold">{title}</h1>
      <p className="text-sm text-muted-foreground">Coming soon.</p>
    </div>
  )
}
