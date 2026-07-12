import {
  Children,
  isValidElement,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ChangeEvent,
  type KeyboardEvent,
  type ReactNode,
  type SelectHTMLAttributes,
} from 'react'
import { Check, ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Option {
  value: string
  label: string
  disabled: boolean
}

function collectOptions(children: ReactNode, out: Option[] = []): Option[] {
  Children.forEach(children, (child) => {
    if (!isValidElement(child)) return
    const props = child.props as {
      value?: string | number
      disabled?: boolean
      children?: ReactNode
    }
    if (child.type === 'option') {
      const label = flattenText(props.children)
      out.push({
        value: props.value !== undefined ? String(props.value) : label,
        label,
        disabled: !!props.disabled,
      })
    } else if (props.children) {
      collectOptions(props.children, out)
    }
  })
  return out
}

function flattenText(node: ReactNode): string {
  if (node === null || node === undefined || typeof node === 'boolean') return ''
  if (typeof node === 'string' || typeof node === 'number') return String(node)
  if (Array.isArray(node)) return node.map(flattenText).join('')
  if (isValidElement(node)) return flattenText((node.props as { children?: ReactNode }).children)
  return ''
}

const SEARCH_THRESHOLD = 10

// Styled replacement for a native <select>: same controlled value/onChange
// contract and <option> children, but renders its own listbox so the dropdown
// matches the design system. Adds a filter input on long option lists.
const Select = ({ className, children, value, onChange, disabled, id }: SelectHTMLAttributes<HTMLSelectElement>) => {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [highlighted, setHighlighted] = useState(0)
  const rootRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLDivElement>(null)
  const searchRef = useRef<HTMLInputElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)

  const options = useMemo(() => collectOptions(children), [children])
  const searchable = options.length > SEARCH_THRESHOLD
  const filtered = useMemo(
    () => (query ? options.filter((o) => o.label.toLowerCase().includes(query.toLowerCase())) : options),
    [options, query]
  )
  const selected = options.find((o) => o.value === String(value ?? ''))

  useEffect(() => {
    if (!open) return
    const onPointerDown = (e: MouseEvent | TouchEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onPointerDown)
    document.addEventListener('touchstart', onPointerDown)
    return () => {
      document.removeEventListener('mousedown', onPointerDown)
      document.removeEventListener('touchstart', onPointerDown)
    }
  }, [open])

  useEffect(() => {
    if (open) {
      setQuery('')
      const idx = filtered.findIndex((o) => o.value === selected?.value)
      setHighlighted(idx >= 0 ? idx : 0)
      searchRef.current?.focus()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  useEffect(() => {
    listRef.current
      ?.querySelector('[data-highlighted="true"]')
      ?.scrollIntoView({ block: 'nearest' })
  }, [highlighted, open])

  const commit = (opt: Option) => {
    if (opt.disabled) return
    setOpen(false)
    triggerRef.current?.focus()
    onChange?.({ target: { value: opt.value } } as unknown as ChangeEvent<HTMLSelectElement>)
  }

  const move = (delta: number) => {
    if (filtered.length === 0) return
    let i = highlighted
    do {
      i = (i + delta + filtered.length) % filtered.length
    } while (filtered[i].disabled && i !== highlighted)
    setHighlighted(i)
  }

  const onKeyDown = (e: KeyboardEvent) => {
    if (!open) {
      if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        e.preventDefault()
        setOpen(true)
      }
      return
    }
    switch (e.key) {
      case 'Escape':
        e.preventDefault()
        setOpen(false)
        triggerRef.current?.focus()
        break
      case 'ArrowDown':
        e.preventDefault()
        move(1)
        break
      case 'ArrowUp':
        e.preventDefault()
        move(-1)
        break
      case 'Enter':
        e.preventDefault()
        if (filtered[highlighted]) commit(filtered[highlighted])
        break
      case 'Tab':
        setOpen(false)
        break
    }
  }

  return (
    <div ref={rootRef} className={cn('relative', className)} onKeyDown={onKeyDown}>
      <button
        ref={triggerRef}
        type="button"
        id={id}
        disabled={disabled}
        role="combobox"
        aria-expanded={open}
        aria-haspopup="listbox"
        onClick={() => setOpen(!open)}
        className={cn(
          'flex h-9 w-full items-center justify-between gap-2 rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm transition-all duration-150',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30',
          'disabled:cursor-not-allowed disabled:opacity-50'
        )}
      >
        <span className={cn('truncate text-left', !selected?.label && 'text-muted-foreground')}>
          {selected?.label || 'Select...'}
        </span>
        <ChevronDown className={cn('h-4 w-4 flex-shrink-0 text-muted-foreground transition-transform', open && 'rotate-180')} />
      </button>

      {open && (
        <div className="absolute left-0 right-0 top-full z-50 mt-1 min-w-[8rem] overflow-hidden rounded-md border bg-card shadow-lg">
          {searchable && (
            <input
              ref={searchRef}
              value={query}
              onChange={(e) => {
                setQuery(e.target.value)
                setHighlighted(0)
              }}
              placeholder="Search..."
              className="w-full border-b bg-transparent px-3 py-2 text-sm focus:outline-none"
            />
          )}
          <div ref={listRef} role="listbox" className="max-h-60 overflow-y-auto p-1">
            {filtered.length === 0 && (
              <div className="px-2 py-2 text-sm text-muted-foreground">No matches</div>
            )}
            {filtered.map((opt, i) => (
              <div
                key={`${opt.value}-${i}`}
                role="option"
                aria-selected={opt.value === selected?.value}
                data-highlighted={i === highlighted}
                onMouseEnter={() => setHighlighted(i)}
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => commit(opt)}
                className={cn(
                  'flex cursor-pointer items-center justify-between gap-2 rounded px-2 py-1.5 text-sm',
                  i === highlighted && 'bg-primary/10 text-primary',
                  opt.disabled && 'cursor-not-allowed opacity-50'
                )}
              >
                <span className="truncate">{opt.label}</span>
                {opt.value === selected?.value && <Check className="h-3.5 w-3.5 flex-shrink-0" />}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
Select.displayName = 'Select'

export { Select }
