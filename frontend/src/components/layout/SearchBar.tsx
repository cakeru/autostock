import { useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useDebounce } from '@/hooks/useDebounce'
import api from '@/services/api'

interface SearchResult {
  type: 'customer' | 'vehicle' | 'invoice' | 'job'
  id: number
  label: string
  sub: string
  url: string
}

const typeLabels: Record<string, string> = {
  customer: 'Customer',
  vehicle: 'Vehicle',
  invoice: 'Invoice',
  job: 'Service Job',
}

export function SearchBar({ dark = false }: { dark?: boolean }) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const debounced = useDebounce(query, 300)
  const navigate = useNavigate()
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (debounced.length < 2) {
      setResults([])
      setOpen(false)
      return
    }
    setLoading(true)
    api.get('/search', { params: { q: debounced } })
      .then(r => {
        setResults(r.data.data || [])
        setOpen(true)
      })
      .catch(() => setResults([]))
      .finally(() => setLoading(false))
  }, [debounced])

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const handleSelect = (r: SearchResult) => {
    setOpen(false)
    setQuery('')
    navigate(r.url)
  }

  return (
    <div ref={ref} className="relative">
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search..."
        className={
          dark
            ? 'w-full rounded-lg border border-white/15 bg-white/10 px-3 py-2 text-sm text-white placeholder:text-white/45 focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-gold)]/40'
            : 'w-full rounded-md border border-input bg-background px-3 py-2 text-sm'
        }
        onFocus={() => results.length > 0 && setOpen(true)}
      />
      {open && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-card border rounded-md shadow-lg z-50 max-h-64 overflow-y-auto">
          {loading ? (
            <p className="text-xs text-muted-foreground p-3">Searching...</p>
          ) : results.length === 0 ? (
            <p className="text-xs text-muted-foreground p-3">No results</p>
          ) : (
            results.map((r, i) => (
              <button
                key={`${r.type}-${r.id}-${i}`}
                className="w-full text-left px-3 py-2 text-sm hover:bg-muted transition-colors flex items-center justify-between"
                onClick={() => handleSelect(r)}
              >
                <div className="min-w-0">
                  <span className="font-medium">{r.label}</span>
                  {r.sub && <span className="text-muted-foreground ml-2 text-xs">{r.sub}</span>}
                </div>
                <span className="text-[10px] text-muted-foreground uppercase flex-shrink-0 ml-2">{typeLabels[r.type]}</span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  )
}
